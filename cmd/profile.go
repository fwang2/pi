package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fwang2/pi/fs"
	"github.com/fwang2/pi/util"
	"github.com/spf13/cobra"
)

var ws *fs.WalkStat = new(fs.WalkStat)
var wc *fs.WalkControl = new(fs.WalkControl)

func init() {

	profileCmd.Flags().BoolVar(&wc.DoHist, "hist", true, "Do histogram")
	profileCmd.Flags().BoolVar(&wc.DoSparse, "sparse", false, "Check sparse file")
	var bins = topnCmd.Flags().String("bins", "4k,1m,1g,1t,32t", "histogram bins")
	ws.HistBins = util.BinsToNum(*bins)
	ws.HistCounter = make([]int64, len(ws.HistBins), len(ws.HistBins))
	rootCmd.AddCommand(profileCmd)
}

func printHistogram() {
	fmt.Printf("\nHistogram\n\n")
	// minwidth, tabwidth, padding, padchar
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 0, ' ', tabwriter.AlignRight)
	for k, v := range ws.HistBins {
		var label string
		if v == util.GUARD {
			nextToLast := ws.HistBins[len(ws.HistBins)-2]
			label = "> " + util.ShortByte(nextToLast)
		} else {
			label = "<= " + util.ShortByte(v)
		}
		fmt.Fprintf(w,
			"%s\t\t%s\t\t%.2f%%\t\t\n",
			label, util.Comma(ws.HistCounter[k]),
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
		fs.Run(wc, ws)
		fs.CalcRate(start, ws)
		profileEpilogue(ws)
	},
}
