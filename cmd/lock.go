// The 'lock' command enables editing of a specified CAD file by locking it
// in the PDM system, preventing concurrent modifications by other users.
package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

// lockCmd represents the "pdm lock" command. It locks a specified CAD file,
// granting the current user edit access while ensuring no conflicts exist
// with other branches or locks.
var lockCmd = &cobra.Command{
	Use:           "lock <file>",
	Short:         "Enable the edit of a selected file by locking it",
	Args:          cobra.ExactArgs(1),
	RunE:          lock,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// lock executes the logic for the "pdm lock" command.
//
// It verifies that the target file is not already locked, checks for
// unmerged changes on remote branches, and locks the file if safe.
// Returns a non-nil error if any operation fails.
func lock(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	relPath, _ := utils.GitRelativeFilepath(filePath)
	if relPath == "" {
		relPath = filePath
	}

	// Check if the file is already locked
	lockStatus, err := utils.GetLockStatus(relPath)
	if err != nil {
		return fmt.Errorf("error while checking lock status: %w", err)
	}
	if lockStatus != (utils.Lock{}) {
		return fmt.Errorf("file %s is already locked by %s", relPath, lockStatus.Owner.Name)
	}

	// Verify file has no conflicting changes on remote branches
	branches, err := utils.GetRemoteBranches()
	if err != nil {
		return fmt.Errorf("error retrieving remote branches: %w", err)
	}

	changes, err := utils.FileDiff(relPath, branches, true)
	if err != nil {
		return fmt.Errorf("error while checking file differences: %w", err)
	}
	if len(changes) > 0 {
		changedBranch := changes[0].Name
		return fmt.Errorf(
			"the file was edited in another branch.\nUse the following command to retrieve the latest version:\n\n\tgit checkout %s -- \"%s\"",
			changedBranch,
			relPath,
		)
	}

	// Lock the file for editing
	status, lockStatus, err := utils.LockFile(relPath)
	if err != nil {
		return fmt.Errorf("error while locking file: %w", err)
	}
	if status {
		utils.Println("Successfully enabled editing for \"%s\"", relPath)
	} else {
		return fmt.Errorf("file %s is already locked by %s", relPath, lockStatus.Owner.Name)
	}

	// Output JSON if requested
	if err := utils.PrintJSON(lockStatus); err != nil {
		return err
	}

	return nil
}

// init registers the "lock" command with the root command.
func init() {
	rootCmd.AddCommand(lockCmd)
}
