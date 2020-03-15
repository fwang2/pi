package fs

import (
	"io"
	"os"
	"path"

	"github.com/fwang2/pi/util"
	gzip "github.com/klauspost/pgzip"
)

func Compress(fname string) (zfname string, err error) {
	zfname = fname[0:len(fname)-len(path.Ext(fname))] + ".gz"
	log.Debugf("zip filename = %s", zfname)
	rfile, err := os.Open(fname)
	if err != nil {
		return
	}
	defer rfile.Close()

	wfile, err := os.OpenFile(zfname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer wfile.Close()

	zw := gzip.NewWriter(wfile)
	zw.Name = fname
	zw.Comment = "pi - rules"
	// set block size, and #cpu
	zw.SetConcurrency(int(16*util.MiB), 10)

	io.Copy(zw, rfile)
	zw.Close()
	return
}

func ConcatZip(file *os.File, zfname string) error {
	zr, err := os.Open(zfname)
	if err != nil {
		return err
	}
	io.Copy(file, zr)
	zr.Close()
	os.Remove(zfname)
	return nil
}
