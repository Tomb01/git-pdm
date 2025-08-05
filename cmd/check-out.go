package cmd

import (
	"fmt"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "check-out",
	Short: "Enable the edit of selected file",
	Run:   checkOut,
}

func checkOut(cmd *cobra.Command, args []string) {
	target := args[0] //absolute path of the file
	id, err := utils.GetLFSLockID(target)
	if err != nil {
		fmt.Println("Error in retriving file lock id")
		return
	}
	isLocked, locker, err := utils.IsLocked(id)
	if err != nil {
		fmt.Println("Error in retriving file lock status")
		return
	}
	if isLocked {
		//if is locked retrive error and print the locker
		fmt.Println("The file is currently checked out by another user: ", locker)
		return
	}

	//check if file is not locked by another user

}

func init() {
	//installCmd.Flags().StringVarP(&software, "software", "s", "", "Custom installation based on specific CAD software\nSOLIDWORS = Dassault System SOLIDWORKS")
	rootCmd.AddCommand(checkoutCmd)
}
