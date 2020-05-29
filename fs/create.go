package fs

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/fwang2/pi/util"
)

func CreateNonSparseFile(fsize int64) string {
	// create a temp file
	num_chunks := fsize / util.MiB
	remainder := fsize % util.MiB

	tmpfile, err := ioutil.TempFile("", "non-sparse")
	//tmpfile, err := os.OpenFile("src.file", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}

	for i := int64(0); i < num_chunks; i++ {
		randBuf := make([]byte, 1*util.MiB)
		rand.Read(randBuf)
		if _, err := tmpfile.Write(randBuf); err != nil {
			log.Fatal(err)
		}
	}

	if remainder != 0 {
		randBuf := make([]byte, remainder)
		rand.Read(randBuf)
		if _, err := tmpfile.Write(randBuf); err != nil {
			log.Fatal(err)
		}

	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()

}

func Md5Checksum(fp string) (string, error) {
	var signature string

	//Open the passed argument and check for any error
	file, err := os.Open(fp)
	if err != nil {
		return signature, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return signature, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	signature = hex.EncodeToString(hashInBytes)

	return signature, nil
}
