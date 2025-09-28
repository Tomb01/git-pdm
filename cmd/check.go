package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the CAD file status on the branch",
	Run:   check,
}

var outputPath string
var checkBranch string

type PdmFileDiff struct {
	Path    string                      `json:"path"`
	Changes []utils.PdmDiffBranchStatus `json:"changes"`
}

func check(cmd *cobra.Command, args []string) {
	// select branch
	branch := ""
	if checkBranch == "" {
		currentBranch, err := utils.GetCurrentBranch()
		if err != nil {
			fmt.Println("Error in check:", err)
			return
		}
		branch = currentBranch
	}

	var files []string
	for _, s := range args {
		if s != "--" {
			files = append(files, s)
		}
	}

	diff, err := utils.GetBranchDiff(branch, "origin/main", files)
	out := []PdmFileDiff{}
	if err != nil {
		fmt.Println("Error in check:", err)
		return
	}

	tot := len(diff)
	branches, err := utils.GetRemoteBranches()
	if err != nil {
		fmt.Println("Error in retriving remote branches", err)
	}
	for i, relPath := range diff {
		fmt.Printf("Check %s (%d/%d)\n", relPath, i+1, tot)
		changes, err := utils.FileDiff(relPath, branches, false)
		if err != nil {
			fmt.Println("Error in retriving file changes ("+relPath+"):", err)
			return
		}
		if len(changes) > 0 {

			//check changes for output
			multipleChanges := false
			fileHash := ""
			for _, change := range changes {
				if multipleChanges && change.Status == 1 && fileHash != change.File {
					fmt.Println("The file " + relPath + " has changes in multiple branches. NEED FIX")
					break
				}
				if change.Status == 1 {
					multipleChanges = true
					fileHash = change.File
				}
			}

			fileStatus := PdmFileDiff{}
			fileStatus.Path = relPath
			fileStatus.Changes = changes
			out = append(out, fileStatus)
		}
	}

	var encoder *json.Encoder
	if outputPath != "" {
		outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}

		// Encode the struct to JSON and write to file
		encoder = json.NewEncoder(outputFile)
		encoder.SetIndent("", "  ") // pretty-print

		if err := encoder.Encode(out); err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}

		defer outputFile.Close()
	}
}

func init() {
	checkCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output JSON file")
	checkCmd.Flags().StringVarP(&checkBranch, "branch", "b", "", "branch to check (current is default)")
	rootCmd.AddCommand(checkCmd)
}
