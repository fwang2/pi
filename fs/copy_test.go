package fs

import (
	"os"
	"strconv"
	"testing"

	"github.com/fwang2/pi/util"
	"github.com/stretchr/testify/assert"
)

func computeSig(fsize int64) (string, string) {
	srcFile := CreateNonSparseFile(fsize)
	defer os.Remove(srcFile)

	dstFile := "dst." + strconv.Itoa(int(fsize))
	defer os.Remove(dstFile)

	CopyFile(srcFile, dstFile)

	srcChecksum, _ := Md5Checksum(srcFile)
	dstChecksum, _ := Md5Checksum(dstFile)

	return srcChecksum, dstChecksum
}

func TestCopy(t *testing.T) {

	fsizes := []int64{1024, 8192, 1*util.MiB + 10}

	for _, sz := range fsizes {
		srcChecksum, dstChecksum := computeSig(sz)
		assert.Equal(t, srcChecksum, dstChecksum, "src and dst checksum should be the same")

	}

}
