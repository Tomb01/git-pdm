package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-pdm",
	Short: "git-pdm is a Git plugin for CAD project dependency management",
	Long:  "git-pdm helps manage and automate project dependencies, templates, and tasks in Git repositories.",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
