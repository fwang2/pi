package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fwang2/pi/util"
	"github.com/spf13/cobra"
)

var Verbose bool
var NumOfWorkers int
var log = util.NewLogger()

var rootCmd = &cobra.Command{
	Use:   "pi",
	Short: "pi is a suite of file system tools",
}

func init() {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().IntVar(&NumOfWorkers, "np", cpus, "Number of worker threads")
}

// Execute ...
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
