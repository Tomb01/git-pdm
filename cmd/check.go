// The 'check' command verifies the CAD file status on a specified or current branch,
// showing differences across branches.
package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

// checkCmd represents the "pdm check" command. It inspects CAD files for
// differences between the specified branch (or current one) and "origin/main".
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the CAD file status on the branch",
	RunE:  check,
}

// checkBranch stores the name of the branch to check. Defaults to the current Git branch.
var checkBranch string

// PdmFileDiff holds the path of a CAD file and its corresponding
// differences across Git branches.
type PdmFileDiff struct {
	Path    string                      `json:"path"`    // Relative path to the CAD file
	Changes []utils.PdmDiffBranchStatus `json:"changes"` // Branch-specific change information
}

// check executes the "pdm check" command logic.
// Returns a non-nil error on failure, which Cobra uses to set the exit status.
func check(cmd *cobra.Command, args []string) error {
	branch := checkBranch
	if branch == "" {
		currentBranch, err := utils.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("error determining current branch: %w", err)
		}
		branch = currentBranch
	}

	// Filter file arguments (ignore "--")
	var files []string
	for _, s := range args {
		if s != "--" {
			files = append(files, s)
		}
	}

	diff, err := utils.GetBranchDiff(branch, "origin/main", files)
	if err != nil {
		return fmt.Errorf("error getting branch diff: %w", err)
	}

	out := []PdmFileDiff{}
	tot := len(diff)

	branches, err := utils.GetRemoteBranches()
	if err != nil {
		return fmt.Errorf("error retrieving remote branches: %w", err)
	}

	for i, relPath := range diff {
		counter := fmt.Sprintf("[%d/%d]", i+1, tot)
		fmt.Printf("\r%s Checking %s ...", counter, relPath) // \r moves to start of line

		changes, err := utils.FileDiff(relPath, branches, false)
		if err != nil {
			fmt.Printf("\n") // ensure the line is not overwritten
			return fmt.Errorf("error retrieving file changes (%s): %w", relPath, err)
		}

		if len(changes) > 0 {
			multipleChanges := false
			fileHash := ""
			for _, change := range changes {
				if multipleChanges && change.Status == 1 && fileHash != change.File {
					fmt.Printf("\n") // newline before printing message
					utils.Println("The file %s has changes in multiple branches. Manual fixing required", relPath)
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
		}
	}
	utils.Println("\n") // move to a new line after finishing the loop

	if err := utils.PrintJSON(out); err != nil {
		return err
	}

	return nil
}

// init initializes the "check" command and registers it with the root command.
func init() {
	checkCmd.Flags().StringVarP(&checkBranch, "branch", "b", "", "branch to check (current is default)")
	rootCmd.AddCommand(checkCmd)
}
