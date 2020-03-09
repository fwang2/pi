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
