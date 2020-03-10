// +build linux

package fs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createNonSparseFile() string {
	// create a temp file
	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "non-sparse")
	if err != nil {
		log.Fatal(err)
	}

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

	// seek forward calulation
	// the following code make sure we have a blocksize hole
	blocksize := 4096
	forward := blocksize - len(buf1)%blocksize + blocksize

	// Ignore all errors
	tmpfile.Write(buf1)
	tmpfile.Seek(int64(forward), os.SEEK_CUR)
	tmpfile.Write(buf2)
	tmpfile.Seek(int64(forward), os.SEEK_CUR)
	tmpfile.Write(buf1)
	tmpfile.Seek(int64(forward), os.SEEK_CUR)
	tmpfile.Write(buf2)

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()
}
func TestIsSparse(t *testing.T) {

	nonSparseFile := createNonSparseFile()
	defer os.Remove(nonSparseFile)

	ok, err := IsSparseFile(nonSparseFile)

	assert.Nil(t, err)
	assert.NotEqual(t, true, ok)

	sparseFile := createSparseFile()
	defer os.Remove(sparseFile)

	ok, _ = IsSparseFile(sparseFile)
	assert.Equal(t, true, ok)
}

func TestScanHoles(t *testing.T) {

	nonSparseFile := createNonSparseFile()
	defer os.Remove(nonSparseFile)

	if holes, _ := ScanHoles(nonSparseFile); len(holes) != 0 {
		t.Error("Expected no holes")
	}

	sparseFile := createSparseFile()
	defer os.Remove(sparseFile)

	if holes, _ := ScanHoles(sparseFile); len(holes) != 3 {
		t.Errorf("Expected 3 hole, got: %d holes \n", len(holes))
	}

}

func TestScanData(t *testing.T) {

	sparseFile := createSparseFile()
	defer os.Remove(sparseFile)

	// if offsets, _, _ := ScanData(sparseFile); len(offsets) != 4 {
	// 	t.Errorf("Expected 3 data chunk, got: %d \n", len(offsets))
	// }

}

func TestCreateSparseFile(t *testing.T) {

	ok := CreateSparseFile("sparse3.data", 3, true)
	assert.Equal(t, true, ok)
	assert.FileExists(t, "sparse3.map")

}
