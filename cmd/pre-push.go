package cmd

import (
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
}

func init() {
	rootCmd.AddCommand(prePushCmd)
}
