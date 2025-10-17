package cmd

import (
	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-pdm",
	Short: "git-pdm is a Git plugin for CAD project dependency management",
	Long:  "git-pdm is a Git plugin for CAD project dependency management",
}

func init() {
	// Global persistent flag, available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&utils.OutJson, "json", "", false, "Writes lock info as JSON to STDOUT if the command exits successfully. Intended for interoperation with external tools")
	rootCmd.PersistentFlags().BoolVarP(&utils.Verbose, "verbose", "v", false, "Enable verbose output for debug purposes")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
