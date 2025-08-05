package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var software string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install git-pdm on the current git repository",
	Run:   install,
}

func install(cmd *cobra.Command, args []string) {
	// get hooks path
	hooksCommand := "git-pdm pre-push"
	hooksPath := utils.GetHooksPath()
	if hooksPath == "" {
		fmt.Println("Error in edit pre-push hooks file: Hooks path not found")
		return
	}

	// STEP 1: Install git lfs
	writeHook := true
	tmpCmd := exec.Command("git-lfs", "install")
	output, err := tmpCmd.CombinedOutput()
	if err != nil {
		if strings.Split(string(output), "\n")[0] == "Hook already exists: pre-push" {
			// Check if git-pdm hooks already exist
			isInstalled, err := utils.StringExistsInFile(hooksPath+"\\pre-push", hooksCommand)
			if err == nil && isInstalled {
				writeHook = false
			} else {
				fmt.Println("There is a custom-made pre-push hook file in this repository: please follow instruction for manual installation of git-pdm")
				return
			}
		} else {
			fmt.Println("Error:", err)
			return
		}
	} else if string(output) != "Updated git hooks.\nGit LFS initialized.\n" {
		fmt.Println("Git LFS installation failed:", string(output))
	}

	// STEP 2: edit hooks and add git-pdm pre-push command

	//Write line on pre-push file
	if writeHook {
		file, err := os.OpenFile(hooksPath+"/pre-push", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error in edit pre-push hooks file:", err)
			return
		}
		defer file.Close()

		// Write the new line
		if _, err := file.WriteString(hooksCommand); err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}

	//STEP 3: write .gitattributes based on our CAD system (optional)
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
		// append to gitigore
		if gitignore != "" {
			file, _ := os.OpenFile(repoRoot+"/.gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if _, err := file.WriteString(gitignore); err != nil {
				fmt.Println("Error writing to git ignore:", err)
				return
			}
			defer file.Close()
		}
		if gitattributes != "" {
			file, _ := os.OpenFile(repoRoot+"/.gitattributes", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if _, err := file.WriteString(gitattributes); err != nil {
				fmt.Println("Error writing to git attributes:", err)
				return
			}
		}
	}

	// Finish
	fmt.Println("Successfully installed git-pdm on this repository")
}

func init() {
	installCmd.Flags().StringVarP(&software, "software", "s", "", "Custom installation based on specific CAD software\nSOLIDWORS = Dassault System SOLIDWORKS")
	rootCmd.AddCommand(installCmd)
}
