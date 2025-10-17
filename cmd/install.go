// The 'install' command configures git-pdm on the current repository by
// setting up necessary Git hooks, LFS, and optional CAD-specific attributes.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

// software defines the CAD system for which git-pdm should customize
// installation (e.g., SOLIDWORKS).
var software string

// installCmd represents the "pdm install" command. It installs git-pdm in the
// current Git repository by configuring LFS, updating hooks, and writing
// appropriate .gitignore and .gitattributes files.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install git-pdm on the current git repository",
	RunE:  install,
}

// install sets up git-pdm on the active Git repository.
//
// Operations performed:
//  1. Ensures Git LFS is installed and initializes it.
//  2. Adds a "git-pdm pre-push" command to the Git hooks if not already present.
//  3. Optionally appends CAD-specific ignore and attribute rules based on the
//     selected software (e.g., SOLIDWORKS).
func install(cmd *cobra.Command, args []string) error {
	hooksCommand := "git-pdm pre-push"
	hooksPath := utils.GetHooksPath()
	if hooksPath == "" {
		return fmt.Errorf("unable to locate Git hooks directory")
	}

	// Install Git LFS and verify hook presence
	writeHook := true
	tmpCmd := exec.Command("git-lfs", "install")
	output, err := tmpCmd.CombinedOutput()
	if err != nil {
		if strings.Split(string(output), "\n")[0] == "Hook already exists: pre-push" {
			isInstalled, err := utils.StringExistsInFile(hooksPath+"/pre-push", hooksCommand)
			if err == nil && isInstalled {
				writeHook = false
			} else {
				return fmt.Errorf("a custom pre-push hook exists. Please install git-pdm manually")
			}
		} else {
			return fmt.Errorf("git-lfs installation error: %w", err)
		}
	} else if string(output) != "Updated git hooks.\nGit LFS initialized.\n" {
		return fmt.Errorf("git-lfs installation returned unexpected output: %s", string(output))
	}

	// Append git-pdm pre-push hook if needed
	if writeHook {
		file, err := os.OpenFile(hooksPath+"/pre-push", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error editing pre-push hook: %w", err)
		}
		defer file.Close()

		if _, err := file.WriteString(hooksCommand); err != nil {
			return fmt.Errorf("error writing to pre-push hook: %w", err)
		}
	}

	// Configure CAD-specific attributes and ignores
	var gitignore, gitattributes string
	switch software {
	case "SOLIDWORKS":
		gitattributes = "\n*.sldprt -text lockable\n*.sldasm -text lockable\n*.slddrw -text lockable\n*.SLDPRT -text lockable\n*.SLDASM -text lockable\n*.SLDDRW -text lockable"
		gitignore = "\n**/~$*.sldprt\n**/~$*.sldasm\n**/~$*.slddrw\n**/~$*.SLDDRW\n**/~$*.SLDPRT\n**/~$*.SLDASM"
	default:
		gitattributes = ""
		gitignore = ""
	}

	repoRoot := utils.GetGitRoot()
	if repoRoot != "" {
		if gitignore != "" {
			file, err := os.OpenFile(repoRoot+"/.gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("error opening .gitignore: %w", err)
			}
			defer file.Close()
			if _, err := file.WriteString(gitignore); err != nil {
				return fmt.Errorf("error writing to .gitignore: %w", err)
			}
		}
		if gitattributes != "" {
			file, err := os.OpenFile(repoRoot+"/.gitattributes", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("error opening .gitattributes: %w", err)
			}
			defer file.Close()
			if _, err := file.WriteString(gitattributes); err != nil {
				return fmt.Errorf("error writing to .gitattributes: %w", err)
			}
		}
	}

	utils.Println("Successfully installed git-pdm on this repository")
	return nil
}

// init registers the "install" command and its flags with the root command.
// The --software (-s) flag enables CAD-specific setup for supported tools.
func init() {
	installCmd.Flags().StringVarP(
		&software,
		"software",
		"s",
		"",
		"Custom installation for a specific CAD software (e.g., SOLIDWORKS)",
	)
	rootCmd.AddCommand(installCmd)
}
