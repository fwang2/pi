package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fwang2/pi/fs"
	"github.com/fwang2/pi/util"
	"github.com/spf13/cobra"
)

var ws *fs.WalkStat = new(fs.WalkStat)
var wc *fs.WalkControl = new(fs.WalkControl)

func init() {

	profileCmd.Flags().BoolVar(&wc.DoHist, "hist", false, "Do histogram")
	profileCmd.Flags().BoolVar(&wc.DoSparse, "sparse", false, "Check sparse file")
	var bins = topnCmd.Flags().String("bins",
		"4k,8k,16k,32k,64k,256k,512k,1m,4m,16m,512m,1g,16g,64g,128g,256g,1t,32t", "histogram bins")
	ws.HistBins = util.BinsToNum(*bins)
	ws.HistCounter = make([]int64, len(ws.HistBins), len(ws.HistBins))

	// check exclusions

	wc.ExcludeMap = map[string]bool{
		"/Volumes/GoogleDrive": true,
		"/Volumes/Recovery":    true,
	}

	// check env PI_EXCLUDE
	checkExcludeEnv()

	log.Debug("Exclusion:", wc.ExcludeMap)

	rootCmd.AddCommand(profileCmd)
}

func checkExcludeEnv() {
	exstring := os.Getenv("PI_EXCLUDE")
	if exstring == "" {
		return
	}
	expath := strings.Split(exstring, ":")
	for _, v := range expath {
		if !wc.ExcludeMap[v] {
			wc.ExcludeMap[v] = true
		}
	}
}

func printHistogram() {
	fmt.Printf("\nHistogram\n\n")
	// minwidth, tabwidth, padding, padchar
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 0, ' ', tabwriter.AlignRight)
	for k, v := range ws.HistBins {
		var label string
		var bucket string
		if v == util.GUARD {
			nextToLast := ws.HistBins[len(ws.HistBins)-2]
			bucket = util.ShortByte(nextToLast)
			label = "> "
		} else {
			label = "<= "
			bucket = util.ShortByte(v)
		}
		fmt.Fprintf(w,
			"%4s%10s\t\t%s\t\t%.2f%%\t\t\n",
			label, bucket, util.Comma(ws.HistCounter[k]),
			float64(ws.HistCounter[k])/float64(ws.TotFileCnt)*100)
	}
	w.Flush()
	fmt.Println()
}

func printSummary() {
	const padding = 10
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, '.', tabwriter.Debug)
	fmt.Fprintf(w, "Total # of files \t %s\n", util.Comma(ws.TotFileCnt))
	fmt.Fprintf(w, "Total # of dirs \t %s\n", util.Comma(ws.TotDirCnt))
	fmt.Fprintf(w, "Total # of symlinks \t %s\n", util.Comma(ws.TotSymlinkCnt))
	fmt.Fprintf(w, "Total # of pipes \t %s\n", util.Comma(ws.TotPipeCnt))
	if wc.DoSparse {
		fmt.Fprintf(w, "Total # of sparse files \t %s\n", util.Comma(ws.TotSparseCnt))
	}
	fmt.Fprintf(w, "Avg file size \t %s\n", util.ShortByte(ws.TotFileSize/ws.TotFileCnt))
	if ws.TotDirCnt != 0 {
		fmt.Fprintf(w, "Avg # of entries per directory \t %s\n", util.Comma(ws.TotFileCnt/ws.TotDirCnt))
	}
	fmt.Fprintf(w, "Aggregated file size \t %s\n", util.ShortByte(ws.TotFileSize))
	fmt.Fprintf(w, "Skipped \t %s\n", util.Comma(ws.TotSkipped))
	fmt.Fprintf(w, "Scanning rate \t %d/s \n", ws.Rate)
	fmt.Fprintf(w, "Elapsed time \t %v\n\n", ws.Elapsed)

	w.Flush()
}

func profileEpilogue(ws *fs.WalkStat) {
	if wc.DoHist {
		printHistogram()
	}
	printSummary()
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "General file system profiling",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine path
		ws.NumOfWorkers = NumOfWorkers
		ws.RootPath = fs.ParseRootPath(args)
		wc.TopNdirs = false
		wc.TopNfiles = false
		fs.WalkPrologue(ws)
		start := time.Now()
		fs.RunProfile(wc, ws)
		fs.CalcRate(start, ws)
		profileEpilogue(ws)
	},
}
