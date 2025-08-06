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

var verbose bool

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
	count := len(locks)
	for i, lock := range locks {
		status, _, err := utils.UnLockFile(lock.Path)
		if err != nil || !status {
			fmt.Println("Error in unlocking "+lock.Path, err)
			return
		} else if verbose {
			fmt.Printf("Successfully unlocked %s (%d/%d)\n", lock.Path, i+1, count)
		}
	}

	fmt.Println("Pre-push routine completed")
}

func init() {
	prePushCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print complete output")
	rootCmd.AddCommand(prePushCmd)
}
