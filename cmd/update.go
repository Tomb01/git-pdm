package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var filePatterns []string
var outJson bool
var onlyConflict bool

var diffCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the changed file from another branch to the current branch",
	Args:  cobra.MinimumNArgs(1),
	Run:   update,
}

func update(cmd *cobra.Command, args []string) {
	// select branch
	//branches, files := utils.SplitArgs(args)
	branch := args[0]
	files := args[1:]

	//fmt.Println("Arguments before --:", files[0])

	diff, err := utils.GetBranchDiff(branch, "HEAD", files)
	//out := []PdmFileDiff{}
	if err != nil {
		fmt.Println("Error in update:", err)
		return
	}

	tot := len(diff)
	for i, relPath := range diff {
		fmt.Printf("Check for update %s (%d/%d)\n", relPath, i+1, tot)
		changes, err := utils.FileDiff(relPath, []string{branch}, true)
		if err != nil {
			fmt.Println("Error in retriving file changes ("+relPath+"):", err)
			return
		}
		if len(changes) > 0 {
			err := utils.CheckoutFile(branch, relPath)
			if err != nil {
				fmt.Println("Checkout failed:", err)
			}
			fmt.Printf("%s has changes -> updated \n", relPath)
		}
	}

}

func init() {
	diffCmd.Flags().StringSliceVarP(&filePatterns, "filepath", "f", []string{}, "Check the difference of a specific file or file type (use * as a wildcard). If empty checks all the locked file types specified in .gitattributes.")
	diffCmd.Flags().BoolVar(&outJson, "json", false, "If used, the output will be formatted in JSON")
	diffCmd.Flags().BoolVarP(&onlyConflict, "only-conflict", "c", false, "Only edit conflict will be printed")
	rootCmd.AddCommand(diffCmd)
}
