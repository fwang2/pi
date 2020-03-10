package cmd

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/fwang2/pi/fs"
	"github.com/spf13/cobra"
)

var checkData bool
var checkHole bool

func init() {
	spcheckCmd.Flags().BoolVar(&checkData, "data", false, "Scan data")
	spcheckCmd.Flags().BoolVar(&checkHole, "hole", false, "Scan data")
	rootCmd.AddCommand(spcheckCmd)
}

var spcheckCmd = &cobra.Command{
	Use:   "sparse-check",
	Short: "Check sparse information (linux only)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			fmt.Println("Sparse file info detection only works on Linux")
			return
		}

		fs.CheckFilePath(args)
		if checkData {
			printData(args[0])
		} else {
			printHoles(args[0])
		}
	},
}

func printData(file string) {

	extents, _ := fs.ScanData(file)
	if len(extents) == 0 {
		fmt.Println("Not holes detected")
	} else {
		fmt.Printf("Sparse file with %d data segments \n", len(extents))
		const padding = 10
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, '.', tabwriter.Debug)

		for i := 0; i < len(extents); i++ {
			fmt.Fprintf(w, "Data offset=%d \t length=%d\n",
				extents[i].Ext_logical, extents[i].Ext_length)
		}
		w.Flush()
	}
	fmt.Println()

}

func printHoles(file string) {
	holes, err := fs.ScanHoles(file)
	if err != nil {
		fmt.Println("Error detected: ", err)
		os.Exit(1)
	}
	if len(holes) == 0 {
		fmt.Println("Not holes detected")
	} else {
		fmt.Printf("Sparse file with %d holes \n", len(holes))
		for k, v := range holes {
			fmt.Printf("\tHole %d at: \t %d\n", k+1, v)
		}
	}
	fmt.Println()
}
