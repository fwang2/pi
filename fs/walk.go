package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
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
	Verbose   bool
	TopNfiles bool
	TopNdirs  bool
	DoHist    bool
	DoSparse  bool
	Findc     *FindControl
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

func find_ioi(findc *FindControl, dir string, file os.FileInfo) bool {

	if Has(findc.Flags, FB_NAME) {
		if check_fname(findc, file.Name()) {
			return true
		} else {
			return false
		}
	}

	if Has(findc.Flags, FB_SIZE) {
		// size info is given
		// we check only aginst files
		// in the case of directory (-type d)
		// the size (filesize aggregate) must be checked at the end
		// of directory scan
		if Has(findc.Flags, FB_TYPE_A) && check_fsize(findc, file.Size()) {
			return true
		} else {
			return false
		}
	}

	// not searching name, not searching size
	// only type remains

	return check_ftype(findc, file.Mode())

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
		}

		mode := file.Mode()

		switch {
		case mode.IsDir():
			res.dirCnt++
			newDir := path.Join(res.dirPath, file.Name())
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
		ws.TopNDirQ.Put(util.Item{Name: res.dirPath, Val: res.dirCnt})
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

// Run ...
func Run(wc *WalkControl, ws *WalkStat) {
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
			WalkProgressReport(ws)
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
