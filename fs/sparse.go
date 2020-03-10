// +build linux

package fs

import (
	//"fmt"
	"os"
)

const (
	SEEK_DATA = 3 // seek to next data
	SEEK_HOLE = 4 // seek to next hole
)

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
func ScanData(file string) (offsets []int64, lengths []int64, err error) {

	fd, err := os.Open(file)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v", err)
		//return
	}
	defer fd.Close()

	endOffset, _ := fd.Seek(0, os.SEEK_END)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		//return
	}

	dataOffset, err := fd.Seek(0, SEEK_DATA)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		//return
	}

	if dataOffset != endOffset {
		var holeOffset int64
		for dataOffset < endOffset {
			offsets = append(offsets, dataOffset)
			holeOffset, err = fd.Seek(dataOffset, SEEK_HOLE)
			lengths = append(lengths, holeOffset-dataOffset)
			dataOffset, err = fd.Seek(holeOffset, SEEK_DATA)

			if holeOffset >= endOffset || dataOffset >= endOffset {
				break
			}
		}
	}
	return
}
