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
	Use:   "diff",
	Short: "Check the difference between current branch and other repository branch. This command must be used to check the ability to lock a file for edit in the current branch",
	Run:   diff,
}

func diff(cmd *cobra.Command, args []string) {
	var search []string
	if len(filePatterns) == 0 {
		lockable, err := utils.GetLockableFiles()
		if err != nil {
			fmt.Println("Error in reading .gitattributes", err)
			return
		}
		search = lockable
	} else {

		search = filePatterns
	}

	diffEntries, err := utils.Diff(search, false)
	if err != nil {
		fmt.Println("Error in diff research", err)
		return
	}
	utils.LogVerbose("\n")

	if onlyConflict {
		for key, entry := range diffEntries {
			for i := len(entry.Branches) - 1; i >= 0; i-- {
				if entry.Branches[i].Status != "M" {
					// swap with last and shrink
					entry.Branches[i] = entry.Branches[len(entry.Branches)-1]
					entry.Branches = entry.Branches[:len(entry.Branches)-1]
				}
			}
			diffEntries[key] = entry
			if len(entry.Branches) <= 1 {
				delete(diffEntries, key)
			}
		}
	}

	if outJson {
		out, err := utils.ToJson(diffEntries)
		if err != nil {
			fmt.Println("Error in JSON conversion", err)
		} else {
			fmt.Println(out)
			return
		}
	} else {
		for k, diff := range diffEntries {
			fmt.Printf("%s has changes in current and %d other branches\n", k, len(diff.Branches))
		}
		fmt.Print("\nTo show more information use the --json option")
		return
	}
}

func init() {
	diffCmd.Flags().StringSliceVarP(&filePatterns, "filepath", "f", []string{}, "Check the difference of a specific file or file type (use * as a wildcard). If empty checks all the locked file types specified in .gitattributes.")
	diffCmd.Flags().BoolVar(&outJson, "json", false, "If used, the output will be formatted in JSON")
	diffCmd.Flags().BoolVarP(&onlyConflict, "only-conflict", "c", false, "Only edit conflict will be printed")
	rootCmd.AddCommand(diffCmd)
}
