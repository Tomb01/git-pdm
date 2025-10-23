// The 'check' command verifies the CAD file status on a specified or current branch,
// showing differences across branches.
package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var fastMode bool

// checkCmd represents the "pdm check" command. It inspects CAD files for
// differences between the specified branch (or current one) and "origin/main".
var checkCmd = &cobra.Command{
	Use:           "check",
	Short:         "Check the CAD file status on the branch",
	RunE:          check,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// PdmFileDiff holds the path of a CAD file and its corresponding
// differences across Git branches.
type PdmFileDiff struct {
	Path    string                      `json:"path"`    // Relative path to the CAD file
	Changes []utils.PdmDiffBranchStatus `json:"changes"` // Branch-specific change information
}

// check executes the "pdm check" command logic.
// Returns a non-nil error on failure, which Cobra uses to set the exit status.
func check(cmd *cobra.Command, args []string) error {
	filters := args

	//files, err := utils.GetBranchDiff("HEAD", branch, files)
	files, err := utils.GetFiles(filters)
	if err != nil {
		return fmt.Errorf("error getting branch diff: %w", err)
	}

	out := []PdmFileDiff{}
	tot := len(files)
	outstr := ""

	branches, err := utils.GetRemoteBranches()
	if err != nil {
		return fmt.Errorf("error retrieving remote branches: %w", err)
	}

	//utils.Println("")
	if fastMode {
		utils.Println("!!! Fast mode ON !!!")
	}

	for i, relPath := range files {
		// Update the counter line using carriage return and padding
		utils.PrintCounter(i+1, tot, "Checking "+relPath)

		// Process file diff
		changes, err := utils.FileDiff(relPath, branches, fastMode)
		if err != nil {
			utils.Println("") // move to a new line before returning
			return fmt.Errorf("error retrieving file changes (%s): %w", relPath, err)
		}

		// Print messages below the counter
		if len(changes) > 0 {
			multipleChanges := false
			fileHash := ""
			for _, change := range changes {
				if multipleChanges && change.Status == 1 && fileHash != change.File {
					outstr = outstr + fmt.Sprintf("The file %s has changes in multiple branches. Manual fixing required\n", relPath)
					break
				}
				if change.Status == 1 {
					multipleChanges = true
					fileHash = change.File
				}
			}

			fileStatus := PdmFileDiff{
				Path:    relPath,
				Changes: changes,
			}
			out = append(out, fileStatus)
		} else {
			utils.LogVerbose("\nNo changes")
		}
	}
	utils.Print("\r[%d/%d] Check completed!%-100s", tot, tot, "")

	if outstr != "" {
		utils.Print("\n\n" + outstr)
	}

	if err := utils.PrintJSON(out); err != nil {
		return err
	}

	return nil
}

// init initializes the "check" command and registers it with the root command.
func init() {
	checkCmd.Flags().BoolVar(&fastMode, "fast", false, "The check stops at the first multiple branch edit")
	rootCmd.AddCommand(checkCmd)
}
