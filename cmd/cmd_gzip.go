package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fwang2/pi/fs"
	"github.com/fwang2/pi/util"
	gzip "github.com/klauspost/pgzip"
	"github.com/spf13/cobra"
)

var zipname string

func init() {
	gzipCmd.Flags().StringVarP(&zipname, "output", "o", "", "output file")
	rootCmd.AddCommand(gzipCmd)
}

var gzipCmd = &cobra.Command{
	Use:     "tarzip file|dir",
	Aliases: []string{"zip", "tgz"},
	Short:   "tar and parallel zip",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		info, err := os.Lstat(args[0])

		if err != nil {
			fmt.Printf("Can't stat: %s\n", args[0])
			os.Exit(1)
		}

		root := args[0]

		mode := info.Mode()

		if mode&os.ModeSymlink != 0 {
			fmt.Printf("[%s] is a symlink, not following.\n", root)
			os.Exit(1)
		}

		if !(mode.IsDir() || mode.IsRegular()) {
			fmt.Printf("target is neither a file or directory: %s\n", root)
			os.Exit(1)
		}

		if mode.IsRegular() {
			_, err := fs.Compress(root)
			if err != nil {
				fmt.Printf("Failed to compress: %s\n", root)
				os.Exit(1)
			}
			return
		}

		// handle directory
		zipDir(root)

	},
}

func zipDir(src string) error {

	if zipname == "" {
		fmt.Printf("Must provide output gz file name with -o \n")
		os.Exit(1)
	}

	zf, err := os.OpenFile(zipname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	defer zf.Close()

	zw := gzip.NewWriter(zf)
	zw.SetConcurrency(int(16*util.MiB), 10)
	defer zw.Close()

	tw := tar.NewWriter(zw)
	defer tw.Close()

	if err != nil {
		fmt.Printf("Failed to create target file: %s\n", zipname)
		os.Exit(1)
	}

	err = filepath.Walk(src,
		func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				log.Warnf("prevent panic by handling failure accessing a path %q: %v\n", file, err)
				return err
			}
			if !fi.Mode().IsRegular() {
				return nil
			}

			// create a new dir/file header
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
			// update the name to correctly reflect the desired destination when untaring
			// target := strings.Replace(file, src, "", -1)
			// header.Name = strings.TrimPrefix(target, string(filepath.Separator))
			if strings.HasPrefix(file, "/") {
				header.Name = file[1:]
			} else {
				header.Name = file
			}
			fmt.Println("a ", header.Name)
			// write the header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// open files for taring
			f, err := os.Open(file)
			if err != nil {
				return err
			}

			// copy file data into tar writer
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
			return nil
		})
	return nil
}
