package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Enable the edit of selected file by locking it",
	Run:   lock,
}

func lock(cmd *cobra.Command, args []string) {
	relPath := args[0] //relative path of the file
	// Check if file is locked
	lock, err := utils.GetLockStatus(relPath)
	if lock != (utils.Lock{}) {
		fmt.Printf("File %s is already locked by %s\n", relPath, lock.Owner.Name)
		return
	} else if err != nil && lock != (utils.Lock{}) {
		fmt.Println("Error in locking:", err)
		return
	}

	// File can be unlocked, check if file has changes on another branch
	changes, err := utils.HasDiff(relPath)
	if len(changes) > 0 {
		// file has changes on another branch -> need update with checkout
		fmt.Printf("The file was edited in another branch.\nUse the following command to retrive the last version\n\n\tgit checkout %s -- \"%s\"\n\n", changes[0], relPath)
		return
	}

	// Lock file
	status, lock, err := utils.LockFile(relPath)
	if err != nil {
		fmt.Println("Error in locking:", err)
		return
	}
	if status {
		fmt.Printf("Successfully enabled editing for \"%s\"\n", relPath)
	} else {
		fmt.Printf("File %s is already locked by %s\n", relPath, lock.Owner.Name)
		return
	}
}

func init() {
	rootCmd.AddCommand(lockCmd)
}
