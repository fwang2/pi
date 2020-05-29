package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fwang2/pi/pool"
	"github.com/fwang2/pi/util"
)

/**

## Design note:

Each job is defined with:
	src
	dst
	type: [prep, copy]

if it is prep job, then readdir(src), return list of files/directories under it.

if it is copy job, then prep src as file, prep dst as destination, and invoke copy tool.

so the job return is a res.dirs (list of directories), res.files (list of files)
and res.err, not nil if error happens

the main routine examines the return as the following:
	if job type is copy, then error shows if copy is successful.
	if job type is walk, then expect a list of files/directory
		for files, inject another copy job
		for directories, inject another walk job


## Execution flow

RunCopy() -> init_work_pool() // this will throw all command line args into the pool
		  -> wait for each job output:
				  -> process res.dirs: list of directory
					  which means we will generate J_PREP
				  -> process res.files: list of files
						which means we will generate J_COPY
				  // not going to deal with symlink for now

## TODO:

* report file number, file size copied
* report progress
* resume (either full checksum or spot check)
* handle sparse files
* handle permissions
* handle special properties
* performance benchmark
*/

type JobType int

const (
	J_PREP = iota
	J_COPY
	J_SYMLINK
)

type CopyControl struct {
	NumOfWorkers int
}

type CopyStat struct {
}

type CopyJob struct {
	srcPath string
	// relPath  string
	// fileName string
	dstPath string
	jtype   JobType
}

type CopyResult struct {
	skipCnt int64
	dirs    []string
	files   []string
	syms    []string
	err     error
}

// func handler(jo *CopyJob) CopyResult {
// the following interface signature is fixed
// as required by pool.Add() interface
// args[0] - the first argument
// args[1] - the second, so on and so forth

func handler(args ...interface{}) interface{} {
	jo := args[0].(CopyJob)
	var res CopyResult

	if jo.jtype == J_PREP {
		files, err := ioutil.ReadDir(jo.srcPath)
		if err != nil {
			log.Debugf("Can't ReadDir() of: %v\n", jo.srcPath)
			res.err = err
			return res
		}
		for _, file := range files {
			fullName := path.Join(jo.srcPath, file.Name())
			mode := file.Mode()

			switch {
			case mode.IsDir():
				res.dirs = append(res.dirs, fullName)
			case mode.IsRegular():
				// handle regular files
				res.files = append(res.files, fullName)
			case mode&os.ModeSymlink != 0:
				// not handle symlinks
				// res.syms = append(res.files, fullName)
				fmt.Printf("Skip symoblic link %v\n", fullName)
			}
		}
	}

	if jo.jtype == J_COPY {
		log.Debugf("src = %v, dest=%v\n", jo.srcPath, jo.dstPath)
		dstParentDir, _ := filepath.Split(jo.dstPath)
		isExist, _, _ := CheckPath(dstParentDir)

		if !isExist {
			os.MkdirAll(dstParentDir, 0744)
		}
		res.err = CopyFile(jo.srcPath, jo.dstPath)
	}

	if jo.jtype == J_SYMLINK {
		// do nothing
	}
	return res
}

// derive the source base from: /a/b/c/file
// as /a/b/c
// we assume that this base won't change
// so if arg list have multiple bases, this will not work
// need to verify if this is 'cp' bahavior
func get_srcbase(srcs []string) (srcBase string) {
	for _, src := range srcs {
		isExist, isDir, isFile := CheckPath(src)
		if !isExist {
			continue
		}
		if isFile {
			srcBase, _ = filepath.Split(src)
			return
		}
		if isDir {
			srcBase, _ = filepath.Abs(src)
			return
		}
	}
	return
}

func init_work_pool(cc *CopyControl, srcs []string, srcBase string, dstAbs string) (mypool *pool.Pool) {

	mypool = pool.New(cc.NumOfWorkers)
	mypool.Run()

	// initialize the pool job items with command line args
	for _, src := range srcs {
		finfo, err := os.Stat(src)
		if os.IsNotExist(err) {
			continue
		}
		var jo CopyJob
		jo.srcPath, _ = filepath.Abs(src)

		switch mode := finfo.Mode(); {
		case mode.IsRegular():
			jo.jtype = J_COPY
			fileName := filepath.Base(jo.srcPath)
			jo.dstPath = filepath.Join(dstAbs, fileName)
		case mode.IsDir():
			jo.jtype = J_PREP
		case mode&os.ModeSymlink != 0:
			jo.jtype = J_SYMLINK
			fmt.Printf("Skip symoblic link %v\n", src)
		}

		mypool.Add(handler, jo)
	}
	return
}

func RunCopy(cc *CopyControl, srcs []string, dest string) {
	srcBase := get_srcbase(srcs)
	dstAbs, _ := filepath.Abs(dest)
	mypool := init_work_pool(cc, srcs, srcBase, dstAbs)

	for {
		job := mypool.WaitForJob()
		if job == nil {
			break
		}
		result := job.Result.(CopyResult)
		for _, dir := range result.dirs {
			var jo CopyJob
			jo.jtype = J_PREP
			jo.srcPath = dir
			mypool.Add(handler, jo)
		}

		for _, file := range result.files {
			var jo CopyJob
			jo.jtype = J_COPY
			jo.srcPath = file
			// file in this case is the full path
			// we split it in the directory portion and base port
			// noted that directory portion is not the same as srcBase
			// for example: srcBase = /path/to/start/dir
			// srcDir could be = /path/to/start/dir/d1/d2
			// In this case, srcDir is two levels deep
			// we need to extract the extra depth d2/d2 using filepath.Rel()
			// then compose back to the target directory
			srcDir, fileName := filepath.Split(file)
			relPath, _ := filepath.Rel(srcBase, srcDir)
			jo.dstPath = filepath.Join(dstAbs, relPath, fileName)
			mypool.Add(handler, jo)
		}
	} // end for
	mypool.Stop()
}

/*

This note is to implement single file copy more efficiently.

source, err := os.Open("source.txt")
destination, err := os.OpenFile("dest.txt", os.O_RDWR|os.O_CREATE, 0666)
_, err = io.Copy(destination, source)  // Copy (dst Writer, src Reader)

Above code is fine EXCEPT:

* the single file is extremely large
* the single file is sparse

The first case can be addressed by:

* breaking the file into chunks
* create a thread pool, and workers copy chunks in parallel
* master will close the file when all is done.


*/

func copyn(ch chan<- error, srcfile string, dstfile string, start int64, nbytes int64) {

	srcfh, err := os.Open(srcfile)
	if err != nil {
		log.Print("Can't open file for reading: ", srcfile)
		ch <- err
		return
	}
	defer srcfh.Close()

	dstfh, err := os.OpenFile(dstfile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Print("Can't write to file: ", dstfile)
		ch <- err
		return
	}
	defer dstfh.Close()

	_, err = srcfh.Seek(start, os.SEEK_SET)
	if err != nil {
		ch <- err
		return
	}
	_, err = dstfh.Seek(start, os.SEEK_SET)
	if err != nil {
		ch <- err
		return
	}

	written, err := io.CopyN(dstfh, srcfh, nbytes)
	if written != nbytes {
		log.Print("Error of copy")
	}
	// write back error if any
	ch <- err
}

func dispatch(srcfile string, dstfile string, fsize int64, nworkers int) (err error) {

	ch := make(chan error)
	nbytes := fsize / int64(nworkers)
	remainder := fsize % int64(nworkers)

	for i := 0; i < nworkers; i++ {
		offset := int64(i) * nbytes
		go copyn(ch, srcfile, dstfile, offset, nbytes)
	}

	if remainder != 0 {
		// must update update offset first before update nworkers
		offset := int64(nworkers) * nbytes
		nworkers++
		go copyn(ch, srcfile, dstfile, offset, remainder)
	}

	// TODO: need better estimate
	timeout := time.After(60 * 60 * time.Second)
	for i := 0; i < nworkers; i++ {
		select {
		case err = <-ch:
			if err != nil {
				// one of the copy task was bad, need to cancel all?
				return
			}
		case <-timeout:
			fmt.Println("copy error")
			return
		}
	}
	return
}

// CopyFile ... copy file from srcfile to destination
func CopyFile(srcfile string, dstfile string) (err error) {

	srcfh, err := os.Open(srcfile)
	if err != nil {
		log.Print("NO => Cant open file for reading: ", srcfile)
		return err
	}
	defer srcfh.Close()

	dstfh, err := os.OpenFile(dstfile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	dstfh.Close()

	// stat file and chunk it
	// entry, _ := os.Stat(srcfile)
	fi, _ := srcfh.Stat()
	fsize := fi.Size()
	var nworkers int
	switch {
	case fsize < 64*util.MiB:
		nworkers = 4
	case fsize <= 1*util.GiB:
		nworkers = 8
	case fsize <= 8*util.GiB:
		nworkers = 16
	case fsize <= 16*util.GiB:
		nworkers = 32
	case fsize <= 32*util.GiB:
		nworkers = 64
	case fsize <= 512*util.GiB:
		nworkers = 128
	default:
		nworkers = 256
	}

	err = dispatch(srcfile, dstfile, fsize, nworkers)

	return
}
