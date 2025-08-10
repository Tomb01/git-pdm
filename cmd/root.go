package cmd

import (
	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-pdm",
	Short: "git-pdm is a Git plugin for CAD project dependency management",
	Long:  "git-pdm helps manage and automate project dependencies, templates, and tasks in Git repositories.",
}

func init() {
	// Global persistent flag, available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&utils.Verbose, "verbose", "v", false, "enable verbose output")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
