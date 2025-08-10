package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var prePushCmd = &cobra.Command{
	Use:   "pre-push",
	Short: "Pre-push command hooks",
	Run:   prePush,
}

func prePush(cmd *cobra.Command, args []string) {
	// Pre push command
	// Unlock all files
	locks, err := utils.GetOursLocks()
	if err != nil {
		fmt.Println("Error in pre-push routine", err)
		return
	}

	if err != nil {
		fmt.Println("Error in pre-push routine", err)
		return
	}
	if len(locks) == 0 {
		fmt.Println("No file to unlock")
		return
	}
	for _, lock := range locks {
		utils.LogVerbose(fmt.Sprintf("Try unlocking %s . . . ", lock.Path))
		absPath, _ := utils.GetAbsoluteFilePath(lock.Path)
		if !utils.FileExists(absPath) {
			utils.LogVerbose("the file doesn't exist anymore. Skipping lock\n")
		} else {
			status, _, err := utils.UnLockFile(lock.Path)
			if err != nil || !status {
				fmt.Println("Error in unlocking "+lock.Path, err)
				return
			} else if utils.Verbose {
				utils.LogVerbose("Complete\n")
			}
		}
	}

	fmt.Println("Pre-push routine completed")
}

func init() {
	rootCmd.AddCommand(prePushCmd)
}
