package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "\n [filename] [num_of_holes]\n")
		os.Exit(1)
	}
	num_holes, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// create a temp file
	buf1 := []byte("12345678901234567890")
	buf2 := []byte("abcdefghijklmnopqrst")
	tmpfile, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// seek forward
	// the following code make sure we have a blocksize hole
	blocksize := 4096
	forward := blocksize - len(buf1)%blocksize + blocksize

	// ignore errors
	for i := 0; i < num_holes; i++ {
		tmpfile.Write(buf1)
		tmpfile.Seek(int64(forward), os.SEEK_CUR)
		tmpfile.Write(buf2)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created", os.Args[1])
}
