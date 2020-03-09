package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "\nprovide one file name!\n")
		os.Exit(1)
	}

	// create a temp file
	buf1 := []byte("12345678901234567890")
	buf2 := []byte("abcdefghijklmnopqrst")
	tmpfile, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(buf1); err != nil {
		log.Fatal(err)
	}
	cur0, _ := tmpfile.Seek(0, os.SEEK_CUR)
	fmt.Printf("After buf1, offset = %v\n", cur0)

	// seek forward
	// the following code make sure we have a blocksize hole
	blocksize := 4096
	forward := blocksize - len(buf1)%blocksize + blocksize

	cur1, _ := tmpfile.Seek(int64(forward), os.SEEK_CUR)

	fmt.Printf("After forward %v, now offset = %v\n", forward, cur1)

	// write 2nd block and create a hole
	if _, err := tmpfile.Write(buf2); err != nil {
		log.Fatal(err)
	}

	cur2, _ := tmpfile.Seek(0, os.SEEK_CUR)
	fmt.Printf("After buf2, offset = %v\n", cur2)

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created", os.Args[1])
}
