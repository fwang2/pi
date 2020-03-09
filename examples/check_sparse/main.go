package main

import (
	"fmt"
	"os"

	"github.com/fwang2/pi/fs"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "\nprovide one file name!\n")
		os.Exit(1)
	}
	ok := fs.IsSparseFile(os.Args[1])
	fmt.Println(ok)
}
