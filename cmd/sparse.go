package cmd

import (
	"fmt"
	"os"

	"github.com/fwang2/pi/fs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sparseCmd)
}

var sparseCmd = &cobra.Command{
	Use:   "sparse",
	Short: "Check sparse information",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fs.CheckFilePath(args)
		holes, err := fs.ScanHoles(args[0])
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
	},
}
