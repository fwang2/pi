package fs

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func createNonSparseFile() string {
	// create a temp file
	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	// defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}

func createSparseFile() string {
	// create a temp file
	buf1 := []byte("12345678901234567890")
	buf2 := []byte("abcdefghijklmnopqrst")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(buf1); err != nil {
		log.Fatal(err)
	}
	// seek forward
	// the following code make sure we have a blocksize hole
	blocksize := 4096
	forward := blocksize - len(buf1)%blocksize + blocksize
	tmpfile.Seek(int64(forward), os.SEEK_CUR)

	if _, err := tmpfile.Write(buf2); err != nil {
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
	return tmpfile.Name()
}
func TestIsSparse(t *testing.T) {
	nonSparseFile := createNonSparseFile()
	defer os.Remove(nonSparseFile)
	sparseFile := createSparseFile()
	defer os.Remove(sparseFile)

	if ok, _ := IsSparseFile(nonSparseFile); ok {
		t.Error("Expected non-sparse file")
	}

	if ok, _ := IsSparseFile(sparseFile); !ok {
		t.Error("Expected sparse file")
	}
}
