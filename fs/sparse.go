package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fwang2/pi/util"
	"golang.org/x/sys/unix"
)

const (
	SEEK_DATA = 3 // seek to next data
	SEEK_HOLE = 4 // seek to next hole
)

// Use fallocate to punch a hole at offset.
// https://pkg.go.dev/golang.org/x/sys/unix?tab=doc
func PunchHole(file *os.File, offset int64, length int64) bool {
	// fd int, mode uint32, off int64, len int64
	err := unix.Fallocate(int(file.Fd()),
		unix.FALLOC_FL_PUNCH_HOLE|unix.FALLOC_FL_KEEP_SIZE, offset, length)
	if err != nil {
		return false
	} else {
		return true
	}
}

// Create a hole at the end of a file,
// Avoid preallication ...
func CreateHole(file *os.File, size int64) bool {
	endFile, _ := file.Seek(size, os.SEEK_CUR)
	return PunchHole(file, endFile-size, size)
}

// IsSparse checks if a file is sparse
// TODO: need better logging
func IsSparse(fd *os.File) (bool, error) {

	holeOffset, err := fd.Seek(0, SEEK_HOLE)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return false, err
	}

	endOffset, _ := fd.Seek(0, os.SEEK_END)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return false, err
	}

	return !(endOffset == holeOffset), nil

}

// IsSparse checks if a file is sparse
func IsSparseFile(file string) (bool, error) {
	fd, err := os.Open(file)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v", err)
		return false, err
	}
	defer fd.Close()
	return IsSparse(fd)
}

func ScanHoles(file string) ([]int64, error) {
	var holes []int64
	fd, err := os.Open(file)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v", err)
		return nil, err
	}
	defer fd.Close()

	endOffset, _ := fd.Seek(0, os.SEEK_END)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil, err
	}

	holeOffset, err := fd.Seek(0, SEEK_HOLE)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil, err
	}

	if holeOffset != endOffset {
		var dataOffset int64
		for holeOffset < endOffset {
			holes = append(holes, holeOffset)
			dataOffset, err = fd.Seek(holeOffset, SEEK_DATA)
			if dataOffset < endOffset {
				holeOffset, err = fd.Seek(dataOffset, SEEK_HOLE)
			} else {
				break
			}
		}
	}
	return holes, nil
}

// ScanData returns offset and length
// Maybe a struct is better
func ScanData(file string) ([]ExtentInfo, error) {
	var extents []ExtentInfo

	fd, err := os.Open(file)
	if err != nil {
		log.Errorf("open: %v", err)
		return extents, err
	}
	defer fd.Close()

	endOffset, err := fd.Seek(0, os.SEEK_END)
	if err != nil {
		log.Errorf("open: %v", err)
		return extents, err
	}

	dataOffset, err := fd.Seek(0, SEEK_DATA)
	if err != nil {
		log.Printf("open: %v", err)
		return extents, err
	}

	if dataOffset != endOffset {
		var holeOffset int64
		for dataOffset < endOffset {
			holeOffset, err = fd.Seek(dataOffset, SEEK_HOLE)
			extents = append(extents, ExtentInfo{
				Ext_logical: dataOffset,
				Ext_length:  holeOffset - dataOffset})
			if holeOffset < endOffset {
				dataOffset, err = fd.Seek(holeOffset, SEEK_DATA)
			}
			if holeOffset >= endOffset || dataOffset >= endOffset {
				break
			}
		}
	}
	if err != nil {
		log.Errorf("End of ... : %v", err)
	}
	return extents, err
}

func randNumber(min int64, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n((max - min + 1) + min)
}

func randBytes(size int64) (blk []byte, err error) {
	blk = make([]byte, size)
	_, err = rand.Read(blk) // for crypto rand, import cryto/rand
	return
}

func CreateSparseFile(filename string, num_holes int, debug bool) bool {
	fileptr, err := os.Create(filename)
	if err != nil {
		log.Error(err)
		return false
	}
	var offsets, lengths []int64
	// ignore errors
	for i := 0; i < num_holes; i++ {
		forward := randNumber(10*util.MiB, 10*util.GiB)
		buf, _ := randBytes(randNumber(1*util.MiB, 10*util.MiB))
		offset, err := fileptr.Seek(forward, os.SEEK_CUR)
		if err != nil {
			log.Error(err)
			return false
		}
		offsets = append(offsets, offset)
		lengths = append(lengths, int64(len(buf)))
		fileptr.Write(buf)
	}

	if err := fileptr.Close(); err != nil {
		log.Error(err)
		return false
	}

	if debug {
		var str strings.Builder
		for i := 0; i < len(offsets); i++ {
			str.WriteString(fmt.Sprintf("offset = %d \t len=%d\n", offsets[i], lengths[i]))
		}
		extension := filepath.Ext(filename)
		newfname := filename[0:len(filename)-len(extension)] + ".map"
		log.Debugf("Map file: %s", newfname)

		ioutil.WriteFile(newfname, []byte(str.String()), 0644)

	}

	return true
}

func ExtentCopy(srcfile string, destfile string) (int64, bool) {
	//
	extents, err := ScanData(srcfile)
	if err != nil {
		log.Fatalf("Scan data: %v\n", err)
	}
	srcfd, _ := os.Open(srcfile)
	defer srcfd.Close()
	var tot int64
	destfd, _ := os.OpenFile(destfile, os.O_CREATE|os.O_WRONLY, 0666)
	for i := 0; i < len(extents); i++ {
		ext_start := extents[i].Ext_logical
		ext_length := extents[i].Ext_length
		srcfd.Seek(ext_start, os.SEEK_SET)
		destfd.Seek(ext_start, os.SEEK_SET)
		written, _ := io.CopyN(destfd, srcfd, ext_length)
		tot += written
	}

	// check if hole is needed at the end
	end_offset, _ := srcfd.Seek(0, os.SEEK_END)
	last_extent := extents[len(extents)-1]
	dest_offset := last_extent.Ext_logical + last_extent.Ext_length
	if dest_offset < end_offset {
		// we need to punch hole at the end
		hole_ext := end_offset - dest_offset
		PunchHole(destfd, dest_offset, hole_ext)
	}

	return tot, true
}
