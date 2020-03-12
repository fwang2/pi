package cmd

import (
	"fmt"
	"runtime"

	"github.com/fwang2/pi/fs"
	"github.com/spf13/cobra"
)

var holes int  // number of holes
var end bool   // have holes at end?
var debug bool // save extent map to file if true

func init() {
	spgenCmd.Flags().IntVar(&holes, "holes", 3, "Number of holes")
	spgenCmd.Flags().BoolVar(&end, "end", false, "Have holes at end")
	spgenCmd.Flags().BoolVar(&debug, "debug", true, "Save extent map")
	rootCmd.AddCommand(spgenCmd)
}

var spgenCmd = &cobra.Command{
	Use:   "sparse-gen",
	Short: "Generate a sparse file (linux only)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			fmt.Println("Sparse file info detection only works on Linux")
			return
		}
		ok := fs.CreateSparseFile(args[0], holes, debug)
		if ok {
			fmt.Printf("Sparse file [%s] successfully created\n", args[0])
		} else {
			fmt.Printf("Sparse file [%s] creation failed\n", args[0])
		}

	},
}
