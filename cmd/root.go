package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "git-pdm",
	Short: "git-pdm is a Git plugin for CAD project dependency management",
	Long:  "git-pdm helps manage and automate project dependencies, templates, and tasks in Git repositories.",
}

func Log(str string) {
	if verbose {
		fmt.Print(str)
	}
}

func init() {
	// Global persistent flag, available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
