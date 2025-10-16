package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "v1.0.1"

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
