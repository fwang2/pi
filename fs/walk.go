package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/fwang2/fnmatch"
	"github.com/fwang2/pi/pool"
	"github.com/fwang2/pi/util"
)

// ScanResult ... results from single dir scan
type ScanResult struct {
	dirPath     string
	fileCnt     int64
	symlinkCnt  int64
	pipeCnt     int64
	dirCnt      int64
	sparseCnt   int64
	fileSizeAgg int64
	fileSizeMax int64
	skipCnt     int64
	dirs        []string // new dirs, new jobs
}

// WalkStat ...
type WalkStat struct {
	RootPath      string
	NumOfWorkers  int
	TotSkipped    int64
	TotFileCnt    int64
	TotFileSize   int64
	TotDirCnt     int64
	TotSymlinkCnt int64
	TotPipeCnt    int64
	TotSparseCnt  int64
	Rate          int64
	TopNFileQ     *util.SortedQueue
	TopNDirQ      *util.SortedQueue
	Elapsed       time.Duration

	// Histogram
	HistBins    []int64
	HistCounter []int64
}

// WalkControl ...
type WalkControl struct {
	Verbose    bool
	TopNfiles  bool
	TopNdirs   bool
	DoHist     bool
	DoSparse   bool
	ExcludeMap map[string]bool
	Findc      *FindControl
	DoProgress bool
}

func check_fsize(findc *FindControl, fsize int64) bool {
	switch findc.SizeOp {
	case GREAT_THAN:
		if fsize > findc.Size {
			return true
		}
	case LESS_THAN:
		if fsize < findc.Size {
			return true
		}
	case EQUAL:
		if fsize == findc.Size {
			return true
		}
	}
	return false
}

func check_fname(findc *FindControl, fname string) bool {
	pattern := findc.Name
	// convert a glob pattern (wildcard form) to regex needs extra work
	// python has this nice fnmatch built in. For golang, this one seem works
	// found, err := regexp.MatchString(pattern, fname)
	found := fnmatch.Match(pattern, fname, 0)
	return found

}

// depreciated
func compare_time(findc *FindControl, t time.Time) bool {
	unix_time := t.UnixNano()
	start_time := findc.StartTime.UnixNano()
	end_time := findc.EndTime.UnixNano()
	if unix_time > start_time && unix_time < end_time {
		return true
	} else {
		return false
	}
}

func check_time(findc *FindControl, fi os.FileInfo) bool {
	// The following check won't do combination of
	// of acm time, yet.
	// https://golang.org/pkg/syscall/#Stat_t
	stat := fi.Sys().(*syscall.Stat_t)
	atime, mtime, ctime := util.StatsTime(stat)

	switch {
	case Has(findc.Flags, FB_ATIME):
		return atime.Before(findc.EndTime) &&
			atime.After(findc.StartTime)
	case Has(findc.Flags, FB_CTIME):
		return ctime.Before(findc.EndTime) &&
			ctime.After(findc.StartTime)
	case Has(findc.Flags, FB_MTIME):
		return mtime.Before(findc.EndTime) &&
			mtime.After(findc.StartTime)

	}
	return false
}

func check_ftype(findc *FindControl, mode os.FileMode) bool {
	switch {
	case mode.IsDir():
		return Has(findc.Flags, FB_TYPE_D)
	case mode.IsRegular():
		return Has(findc.Flags, FB_TYPE_F)
	case mode&os.ModeSymlink != 0:
		return Has(findc.Flags, FB_TYPE_L)
	}
	return false
}

// find_ioi ... locate the item of interests (ioi). ioi is checked upon
// name, size, type, time. We don't have support for OR combination. So
// each condition is combined with AND.
func find_ioi(findc *FindControl, dir string, file os.FileInfo) (yes bool) {
	var ioi_flag Bits

	if Has(findc.Flags, FB_NAME) {
		if check_fname(findc, file.Name()) {
			ioi_flag = Set(ioi_flag, IOI_NAME)
		} else {
			yes = false
			return
		}
	} else {
		ioi_flag = Set(ioi_flag, IOI_NAME)
	}

	if Has(findc.Flags, FB_SIZE) {
		// size info is given
		// we check only aginst files
		// in the case of directory (-type d)
		// the size (filesize aggregate) must be checked at the end
		// of directory scan

		if check_fsize(findc, file.Size()) {
			ioi_flag = Set(ioi_flag, IOI_SIZE)
		} else {
			yes = false
			return
		}
	} else {
		ioi_flag = Set(ioi_flag, IOI_SIZE)
	}

	if Has(findc.Flags, FB_ATIME|FB_CTIME|FB_MTIME) {
		if check_time(findc, file) {
			ioi_flag = Set(ioi_flag, IOI_TIME)
		} else {
			yes = false
			return
		}
	} else {
		ioi_flag = Set(ioi_flag, IOI_TIME)
	}

	if Has(findc.Flags, FB_TYPE_D|FB_TYPE_F|FB_TYPE_L) {
		if check_ftype(findc, file.Mode()) {
			ioi_flag = Set(ioi_flag, IOI_TYPE)
		} else {
			yes = false
			return
		}
	} else {
		ioi_flag = Set(ioi_flag, IOI_TYPE)
	}

	return Has(ioi_flag, IOI_NAME) && Has(ioi_flag, IOI_TYPE) &&
		Has(ioi_flag, IOI_TIME) && Has(ioi_flag, IOI_SIZE)

}

func check_dir_size(findc *FindControl, res *ScanResult) bool {
	if Has(findc.Flags, FB_TYPE_D) && Has(findc.Flags, FB_SIZE) {
		return check_fsize(findc, res.fileSizeAgg)
	} else {
		return false
	}
}

// Walk ...
// args[0] passed as *WalkControl
// args[1] passed as *WalkStat
// args[2] passed as dir to be walked
func Walk(args ...interface{}) interface{} {
	var res ScanResult
	var wc = args[0].(*WalkControl)
	var ws = args[1].(*WalkStat)
	res.dirPath = args[2].(string)

	files, err := ioutil.ReadDir(res.dirPath)
	if err != nil {
		if wc.Verbose {
			log.Println(err)
		}
		return nil
	}

	for _, file := range files {

		fname := path.Join(res.dirPath, file.Name())

		if wc.Findc != nil && find_ioi(wc.Findc, res.dirPath, file) {
			fmt.Println(fname)
			if wc.Findc.DeleteFlag {
				err := os.Remove(fname)
				if err != nil {
					log.Warningf("Can't remove %s, %s\n", fname, err)
				}
			}
		}

		mode := file.Mode()

		switch {
		case mode.IsDir():
			res.dirCnt++
			newDir := path.Join(res.dirPath, file.Name())

			if wc.ExcludeMap[newDir] {
				log.Debug("Excluding ... ", newDir)
				break
			}

			res.dirs = append(res.dirs, newDir) // save new dirs encountered
		case mode.IsRegular():
			res.fileCnt++
			fSize := FileSize(res.dirPath, file)

			if fSize > res.fileSizeMax {
				res.fileSizeMax = fSize
			}
			res.fileSizeAgg += fSize

			// handle top N files
			if wc.TopNfiles {
				ws.TopNFileQ.Put(util.Item{Name: fname, Val: fSize})
			}

			// handle histogram
			if wc.DoHist {
				util.InsertLeft(ws.HistBins, ws.HistCounter, fSize)
			}

			// handle sparse file
			if wc.DoSparse && runtime.GOOS == "linux" {
				yes, err := IsSparseFile(path.Join(res.dirPath, file.Name()))
				if err != nil {
					res.skipCnt++
				}
				if yes {
					res.sparseCnt++
				}
			}

		case mode&os.ModeSymlink != 0:
			res.symlinkCnt++
		case mode&os.ModeNamedPipe != 0:
			res.pipeCnt++
		}
	}
	// handle top N dir
	if wc.TopNdirs {
		ws.TopNDirQ.Put(util.Item{Name: res.dirPath, Val: int64(len(files))})
	}

	// handle directory level find
	if wc.Findc != nil && check_dir_size(wc.Findc, &res) {
		fmt.Printf("%s (%d)\n", res.dirPath, res.fileSizeAgg)
	}
	return res
}

// WalkPrologue ...
func WalkPrologue(ws *WalkStat) {
	fmt.Printf("\nRunning: [%d] threads\n", ws.NumOfWorkers)
	fmt.Printf("\nFS: %s \n\n", InfoStr(ws.RootPath))
}

// WalkProgressReport ...
func WalkProgressReport(ws *WalkStat) {
	fmt.Printf("Scanned: %s, skipped: %s \r",
		util.Comma(ws.TotDirCnt+ws.TotFileCnt), util.Comma(ws.TotSkipped))
}

// Run ... this is entry function for both profile, find, and topn operation
// In this function, we set up a pool with a fixed number of workers
// We put the initial work item(s) or job(s) in the pool by:
// mypool.Add(...)
// It takes a functions and its argument list
// mypool will run this function, in this case the `Walk()` function
//
// Walk() function in this case will scan the directory, tally the stats
// and return another list of sub-directories for further scan
//
// Run() will wait for job and its return - the returned sub-directories are each
// put into the pool as a different job to walk.
//
// Esssentially, the Walk() expects a directory to talk, that is the job.
// it recursively traverse into subdirectories, until no more no job to put
// in to the pool. When all jobs are done, pool will return nil
// the for-loop will break out.

func RunProfile(wc *WalkControl, ws *WalkStat) {
	mypool := pool.New(ws.NumOfWorkers)
	mypool.Run()
	mypool.Add(Walk, wc, ws, ws.RootPath)

	var tick <-chan time.Time
	tick = time.Tick(500 * time.Millisecond)

	for {
		job := mypool.WaitForJob()
		if job == nil {
			break
		}
		if job.Result == nil {
			ws.TotSkipped++
		} else {
			result := job.Result.(ScanResult)
			ws.TotFileCnt += result.fileCnt
			ws.TotDirCnt += result.dirCnt
			ws.TotFileSize += result.fileSizeAgg
			ws.TotPipeCnt += result.pipeCnt
			ws.TotSymlinkCnt += result.symlinkCnt
			ws.TotSkipped += result.skipCnt
			if wc.DoSparse {
				ws.TotSparseCnt += result.sparseCnt
			}
			for _, d := range result.dirs {
				mypool.Add(Walk, wc, ws, d)
			}
		}

		// report progress
		select {
		case <-tick:
			if wc.DoProgress {
				WalkProgressReport(ws)
			}
		default:
			break
		}
	}
	mypool.Stop()
}

// CalcRate ...
func CalcRate(start time.Time, ws *WalkStat) {
	ws.Elapsed = time.Since(start)
	rpms := float64(ws.TotDirCnt+ws.TotFileCnt) / float64(ws.Elapsed/time.Nanosecond)
	ws.Rate = int64(rpms * 1e9)
}
