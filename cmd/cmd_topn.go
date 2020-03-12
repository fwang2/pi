package cmd

import (
	"fmt"
	"time"

	"github.com/fwang2/pi/fs"
	"github.com/fwang2/pi/util"
	"github.com/spf13/cobra"
)

var topNfiles int
var topNdirs int

func init() {
	topnCmd.Flags().IntVarP(&topNdirs, "dirs", "d", 5, "top N directories")
	topnCmd.Flags().IntVarP(&topNfiles, "files", "f", 5, "top N files")
	rootCmd.AddCommand(topnCmd)
}

func topnEpilogue(ws *fs.WalkStat) {
	printTopNdir(ws.TopNDirQ.Items())
	printTopNfile(ws.TopNFileQ.Items())
}

func printTopNdir(items util.ItemList) {
	fmt.Printf("\n\nTop count on directory entries\n\n")
	for i := len(items) - 1; i >= 0; i-- {
		fmt.Printf("\t%s (%s) \n", items[i].Name, util.ShortNum(items[i].Val))
	}
	fmt.Printf("\n")
}

func printTopNfile(items util.ItemList) {
	fmt.Printf("\n\nTop count on large files\n\n")
	for i := len(items) - 1; i >= 0; i-- {
		fmt.Printf("\t%s (%s) \n", items[i].Name, util.ShortByte(items[i].Val))
	}
	fmt.Printf("\n")
}

var topnCmd = &cobra.Command{
	Use:   "topn",
	Short: "Find top N items of interest",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine path
		var ws *fs.WalkStat = new(fs.WalkStat)
		ws.NumOfWorkers = NumOfWorkers
		ws.RootPath = fs.ParseRootPath(args)
		ws.TopNDirQ = util.NewSortedQueue(topNdirs)
		ws.TopNFileQ = util.NewSortedQueue(topNfiles)
		var wc *fs.WalkControl = new(fs.WalkControl)
		wc.TopNdirs = true
		wc.TopNfiles = true
		fs.WalkPrologue(ws)
		start := time.Now()
		fs.Run(wc, ws)
		fs.CalcRate(start, ws)
		topnEpilogue(ws)
	},
}
