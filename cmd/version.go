package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "v0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the git-pdm version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("git-pdm version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
