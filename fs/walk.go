package fs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

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
	fileSizeAgg int64
	fileSizeMax int64
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

		switch mode := file.Mode(); {
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
				ws.TopNFileQ.Put(util.Item{Name: path.Join(res.dirPath, file.Name()),
					Val: fSize})
			}

			// handle histogram
			if wc.DoHist {
				util.InsertLeft(ws.HistBins, ws.HistCounter, fSize)
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
