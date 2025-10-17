package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var filePatterns []string
var onlyConflict bool

// diffCmd represents the "pdm update" command.
// It updates changed files from another branch to the current branch,
// ensuring the latest version is checked out for files that have differences.
var diffCmd = &cobra.Command{
	Use:   "update <branch> [files...]",
	Short: "Update the changed file from another branch to the current branch",
	Args:  cobra.MinimumNArgs(1),
	RunE:  update,
}

// update executes the "pdm update" command logic.
//
// Steps performed:
//  1. Determine the target branch and files to update.
//  2. Retrieve differences between HEAD and the target branch.
//  3. Iterate over changed files and optionally check them out.
//  4. Show a dynamic progress counter during the operation.
//  5. Optionally output results in JSON format if utils.OutJson is true.
//
// Returns a non-nil error if any operation fails.
func update(cmd *cobra.Command, args []string) error {
	branch := args[0]
	files := args[1:]

	diff, err := utils.GetBranchDiff("HEAD", branch, files)
	if err != nil {
		return fmt.Errorf("error getting branch diff: %w", err)
	}

	tot := len(diff)
	for i, relPath := range diff {
		counter := fmt.Sprintf("[%d/%d]", i+1, tot)
		utils.Println(fmt.Sprintf("\r%s Checking for update %s ...", counter, relPath))

		changes, err := utils.FileDiff(relPath, []string{branch}, true)
		if err != nil {
			fmt.Printf("\n")
			return fmt.Errorf("error retrieving file changes (%s): %w", relPath, err)
		}

		if len(changes) > 0 {
			if err := utils.CheckoutFile(branch, relPath); err != nil {
				fmt.Printf("\n")
				return fmt.Errorf("checkout failed for %s: %w", relPath, err)
			}
			utils.LogVerbose("\r%s %s has changes -> updated\n", counter, relPath)
		} else if onlyConflict {
			utils.LogVerbose("\r%s %s has no changes\n", counter, relPath)
		}
	}
	utils.Println("\nUpdate routine completed\n")

	if err := utils.PrintJSON(diff); err != nil {
		return err
	}
	return nil
}

// init registers the "update" command and its flags with the root command.
//
// Flags:
//
//	--filepath, -f: specify file(s) or patterns to update (supports wildcards).
//	--json        : output results in JSON format.
//	--only-conflict, -c: only print files with conflicts.
func init() {
	diffCmd.Flags().StringSliceVarP(&filePatterns, "filepath", "f", []string{}, "Check the difference of a specific file or file type (use * as a wildcard). If empty, checks all locked file types specified in .gitattributes.")
	diffCmd.Flags().BoolVar(&utils.OutJson, "json", false, "If used, the output will be formatted in JSON")
	diffCmd.Flags().BoolVarP(&onlyConflict, "only-conflict", "c", false, "Only edit conflicts will be printed")
	rootCmd.AddCommand(diffCmd)
}
