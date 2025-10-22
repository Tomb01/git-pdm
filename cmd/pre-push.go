package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

// prePushCmd represents the "pdm pre-push" command, which runs before a git push
// to automatically unlock files that were previously locked.
var prePushCmd = &cobra.Command{
	Use:           "pre-push",
	Short:         "Pre-push command hooks",
	RunE:          prePush,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// prePush executes the pre-push routine.
// It unlocks all files that are currently locked and exist in the repository.
// Returns a non-nil error if any unlock operation fails.
func prePush(cmd *cobra.Command, args []string) error {
	locks, err := utils.GetLocks(true)
	if err != nil {
		return fmt.Errorf("error retrieving locks in pre-push routine: %w", err)
	}

	if len(locks) == 0 {
		utils.Println("No files to unlock")
		return nil
	}

	tot := len(locks)
	locked := []utils.Lock{}
	for i, lock := range locks {
		utils.PrintCounter(i+1, tot, "Unlocking "+lock.Path)

		absPath, _ := utils.GetAbsoluteFilePath(lock.Path)
		if !utils.FileExists(absPath) {
			utils.LogVerbose("\n%s File doesn't exist. Skipping.\n")
			continue
		}

		status, newlock, err := utils.UnLockFile(lock.Path)
		locked = append(locked, newlock)
		if err != nil || !status {
			fmt.Printf("\n")
			return fmt.Errorf("error unlocking file %s: %w", lock.Path, err)
		}
	}

	utils.Print("\r[%d/%d] Unlock completed!%-100s", tot, tot, "")
	if err := utils.PrintJSON(locked); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(prePushCmd)
}
