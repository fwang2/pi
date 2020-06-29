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
var copyMode int // control which function to run

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
		sourceExist, _, sourceIsFile := fs.CheckPath(args[0])
		destExist, destIsDir, destIsFile := fs.CheckPath(args[len(args)-1])
		if sourceExist && sourceIsFile && len(args) == 2 {
			sources = append(sources, args[0])
			dest = args[1]
			copyMode = fs.COPY_F2F
			if destExist && destIsDir {
				// destination exists as a directory
				dest, err = filepath.Abs(dest)
				sourceBase := filepath.Base(args[0])
				dest = filepath.Join(dest, sourceBase)
			}
			return
		}

		// file to directory copy
		copyMode = fs.COPY_F2D
		lastArg := args[len(args)-1]
		if destExist {
			if destIsFile {
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
		if copyMode == fs.COPY_F2F {
			fs.CopyFile(sources[0], dest)
		} else {
			fs.RunCopy(cc, sources, dest)
		}
	},
}
