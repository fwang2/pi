package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fwang2/pi/fs"
	"github.com/spf13/cobra"
)

var cc *fs.CopyControl = new(fs.CopyControl)
var sources []string
var dest string

func init() {
	rootCmd.AddCommand(cpCmd)
}

var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "parallel copy",
	// use a custom validator
	// so we can emit more reasonable errors
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) < 2 {
			return errors.New("need at least 2 args, one source, one destination")
		}

		// case 1:
		// if the last argument is a file
		// if it is a file, then it is expected to be file-to-file copy

		// case 2:
		// if the last argument is a directory
		// then the source is all the files and/or directories
		// present in the arg list, and we copy them over

		// case 3:
		// directory to directory
		//
		log.Debugf("command line args: %v\n", args)
		lastArg := args[len(args)-1]
		isExist, _, isFile := fs.CheckPath(lastArg)
		if isExist {
			if isFile {
				log.Fatalf("Taget exists as file: %s\n", lastArg)
			} else {
				log.Warningf("Target exists as directory: %s\n", lastArg)
				dest, err = filepath.Abs(lastArg)
			}

		} else {
			// target doesn't exist, create
			err = os.MkdirAll((lastArg), 0744)
			if err != nil {
				log.Fatalf("Can't create target dir: %s\n", lastArg)
			}
			dest, err = filepath.Abs(lastArg)
		}

		for i := len(args) - 2; i >= 0; i-- {
			src, _ := filepath.Abs(args[i])
			sources = append(sources, src)
		}
		return
	},
	Run: func(cmd *cobra.Command, args []string) {
		cc.NumOfWorkers = NumOfWorkers
		// start := time.Now()
		log.Debugf("sources = %v, dest = %s \n", sources, dest)
		fs.RunCopy(cc, sources, dest)
	},
}
