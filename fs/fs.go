package fs

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/fwang2/pi/util"
)

// ParseRootPath ... only allow one root
// probably should check if (1) directory (2) accessible
func ParseRootPath(args []string) string {

	if len(args) == 0 {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return path
	}
	return args[0]

}

// FileSize ...
func FileSize(dirPath string, fi os.FileInfo) int64 {
	var reportSize int64

	// Most of the time, fi.Size() gives out incorrect information
	// du has --aparent_size option to use it without further checking
	// we will double check it by comparing to diskSize

	fileSize := fi.Size()

	sysstat := fi.Sys().(*syscall.Stat_t)

	if sysstat != nil {
		// ssystat.Blksize or st_blksize is "optimal" block size for I/O
		// it should *not* be used here as "st_blocks" explicitly stats
		// that this is number of 512B blocks allocated.

		diskSize := 512 * sysstat.Blocks
		if diskSize < fileSize {
			reportSize = diskSize
		} else {
			reportSize = fileSize
		}
	}

	return reportSize
}

// InfoT ... defines FS information struct
type InfoT struct {
	fstype             string
	totFileSystemSize  int64
	freeFileSystemSize int64
	totInodes          int64
	freeInodes         int64
}

// StatInfo ... return fs information
// man statfs
// golang's naming convention is horrible
func StatInfo(path string) InfoT {
	fsinfo := InfoT{}

	s := syscall.Statfs_t{}
	if err := syscall.Statfs(path, &s); err != nil {
		log.Fatalln(err)
	}

	fsinfo.freeFileSystemSize = int64(s.Bfree) * int64(s.Bsize) // Bfree = free blocks
	fsinfo.totFileSystemSize = int64(s.Blocks) * int64(s.Bsize) // Blocks = total blocks
	fsinfo.totInodes = int64(s.Files)                           // Files = total inodes
	fsinfo.freeInodes = int64(s.Ffree)                          // Ffree = free inodes

	return fsinfo
}

// InfoStr ... return as string
func InfoStr(path string) string {

	fsinfo := StatInfo(path)

	f1 := util.ShortByte(fsinfo.totFileSystemSize)
	f2 := float64(fsinfo.freeFileSystemSize) / float64(fsinfo.totFileSystemSize) * 100
	f4 := float64(fsinfo.totInodes-fsinfo.freeInodes) / float64(fsinfo.totInodes) * 100
	f3 := util.ShortNum(fsinfo.totInodes)
	return fmt.Sprintf("%s      Used FS: %.2f%%      Inodes: %s      Used Inodes: %.2f%%",
		f1, f2, f3, f4)
}
