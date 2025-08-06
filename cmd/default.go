package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to git-pdm!")
		fmt.Println("Use 'git pdm --help' or 'git pdm [command]' to get started.")
	}
}
