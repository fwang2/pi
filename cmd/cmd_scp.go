// +build linux

package cmd

import (
	"runtime"

	"github.com/fwang2/pi/fs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cpCmd)
}

var cpCmd = &cobra.Command{
	Use:   "scp",
	Short: "copy file to file (sparse aware",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS == "linux" {
			linux_copy(args)
		} else {
			non_linux_copy(args)
		}
	},
}

func non_linux_copy(args []string) {

}

func linux_copy(args []string) {
	srcfile := args[0]
	destfile := fs.DestPath(args[0], args[1])
	tot, ok := fs.ExtentCopy(srcfile, destfile)
	if !ok {
		log.Fatalf("Error occured, %d bytes written\n", tot)
	}
}
