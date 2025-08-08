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
	locks, err := utils.GetLocks("--local")
	if err != nil {
		fmt.Println("Error in pre-push routine", err)
		return
	}

	if err != nil {
		fmt.Println("Error in pre-push routine", err)
		return
	}
	if len(locks) <= 0 {
		fmt.Println("No file to unlock")
		return
	}
	userName, err := utils.GetGitUserName()
	if err != nil {
		fmt.Println("Fail to retrive Git username", err)
		return
	}
	for _, lock := range locks {
		if lock.Owner.Name == userName {
			absPath, _ := utils.GetAbsoluteFilePath(lock.Path)
			if !utils.FileExists(absPath) && err != nil {
				fmt.Printf("%s doesn't exist anymore. Skipping lock\n", lock.Path)
			} else {
				status, _, err := utils.UnLockFile(lock.Path)
				if err != nil || !status {
					fmt.Println("Error in unlocking "+lock.Path, err)
					return
				} else if verbose {
					fmt.Printf("Successfully unlocked %s\n", lock.Path)
				}
			}
		}
	}

	fmt.Println("Pre-push routine completed")
}

func init() {
	rootCmd.AddCommand(prePushCmd)
}
