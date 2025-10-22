package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var onlyConflict bool
var dryRun bool

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

	if dryRun {
		utils.Println("!!! Dry mode activated - the file will not be checked out !!!")
	}

	tot := len(diff)
	outstr := ""
	for i, relPath := range diff {

		utils.PrintCounter(i+1, tot, "Checking "+relPath)

		changes, err := utils.FileDiff(relPath, []string{branch}, true)
		if err != nil {
			fmt.Printf("\n")
			return fmt.Errorf("error retrieving file changes (%s): %w", relPath, err)
		}

		if len(changes) > 0 {
			if !dryRun {
				if err := utils.CheckoutFile(branch, relPath); err != nil {
					fmt.Printf("\n")
					return fmt.Errorf("checkout failed for %s: %w", relPath, err)
				}
				utils.LogVerbose("\n%s has changes -> updated\n", relPath)
			}

			outstr += relPath + "\n"
		} else if onlyConflict {
			utils.LogVerbose("\n%s has no changes\n", relPath)
		}
	}
	utils.Print("\r[%d/%d] Update completed!%-100s", tot, tot, "")

	if outstr != "" {
		utils.Println("Updated files:\n%s\n", outstr)
	} else {
		utils.Println("\nNo file to update")
	}

	if err := utils.PrintJSON(diff); err != nil {
		return err
	}
	return nil
}

// init registers the "update" command and its flags with the root command.
//
// Flags:
//
//	--json        : output results in JSON format.
//	--dry		  : only run the check without update
func init() {
	diffCmd.Flags().BoolVarP(&onlyConflict, "only-conflict", "c", false, "Only edit conflicts will be printed")
	diffCmd.Flags().BoolVar(&dryRun, "dry", false, "only run the check without update")
	rootCmd.AddCommand(diffCmd)
}
