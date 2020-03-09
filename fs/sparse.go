package fs

import (
	"fmt"
	"os"
)

const (
	SEEK_DATA = 3 // seek to next data
	SEEK_HOLE = 4 // seek to next hole
)

// IsSparse checks if a file is sparse
// TODO: need better logging
func IsSparse(fd *os.File) bool {

	holeOffset, err := fd.Seek(0, SEEK_HOLE)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}

	endOffset, _ := fd.Seek(0, os.SEEK_END)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}

	return !(endOffset == holeOffset)

}

// IsSparse checks if a file is sparse
func IsSparseFile(file string) bool {
	fd, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	defer fd.Close()
	return IsSparse(fd)
}
