package cmd

import (
	"fmt"
	"path"

	"github.com/Tomb01/git-pdm/utils"
	"github.com/spf13/cobra"
)

var filePatterns []string
var outJson bool
var onlyConflict bool

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Check the difference between current branch and other repository branch. This command must be used to check the ability to lock a file for edit in the current branch",
	Run:   diff,
}

type PdmFileDiff struct {
	Path     string                `json:"path"`
	Branches []PdmDiffBranchStatus `json:"branches"`
}

type PdmDiffBranchStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func diff(cmd *cobra.Command, args []string) {
	var search []string
	if len(filePatterns) == 0 {
		lockable, err := utils.GetLockableFiles()
		if err != nil {
			fmt.Println("Error in reading .gitattributes", err)
			return
		}
		search = lockable
	} else {

		search = filePatterns
	}

	pdmDiffEntries := make(map[string]PdmFileDiff)
	branches, err := utils.GetRemoteBranches()
	currentBranch, err2 := utils.GetCurrentBranch()
	root := utils.GetGitRoot()
	if err != nil || err2 != nil {
		fmt.Println("Error in retiving remote branches", err)
		return
	}

	if onlyConflict {
		currentDiff, err := utils.GetDiff("main..."+currentBranch, search)
		Log("Comparing current branch with main...")
		if err != nil {
			fmt.Println("Error in comparing "+currentBranch, err)
			return
		}
		for _, diff := range currentDiff {
			pdmDiffEntries[diff.Filename] = PdmFileDiff{Path: path.Join(root, diff.Filename)}
		}
		Log("Complete\n")
	}
	var compare string
	for _, branch := range branches {
		if branch == "origin/"+currentBranch {
			continue
		}
		if !onlyConflict {
			compare = branch + "..." + currentBranch
		} else {
			compare = branch + "...origin/main"
		}
		Log(fmt.Sprintf("Comparing %s . . . ", compare))
		diffEntries, err := utils.GetDiff(compare, search)
		if err != nil {
			fmt.Println("Error in comparing "+branch, err)
			return
		}
		for _, diff := range diffEntries {
			pdmDiff := pdmDiffEntries[diff.Filename]
			if pdmDiff.Path == "" && !onlyConflict {
				// create new diff
				pdmDiff = PdmFileDiff{
					Path: path.Join(root, diff.Filename),
					Branches: []PdmDiffBranchStatus{
						{Name: branch, Status: diff.Status},
					}}
			} else {
				// append branch
				if !onlyConflict || diff.Status != "A" {
					pdmDiff.Branches = append(pdmDiff.Branches, PdmDiffBranchStatus{Name: branch, Status: diff.Status})
				}
			}
			pdmDiffEntries[diff.Filename] = pdmDiff
		}
		Log("Complete\n")
	}

	//retrive output
	// if --conflict is activated print only the files that have two or more entries in branches property
	if onlyConflict {
		for k, diff := range pdmDiffEntries {
			if len(diff.Branches) <= 1 {
				delete(pdmDiffEntries, k)
			}
		}
	}
	Log("\n")

	if outJson {
		out, err := utils.ToJson(pdmDiffEntries)
		if err != nil {
			fmt.Println("Error in JSON conversion", err)
		} else {
			fmt.Println(out)
			return
		}
	} else {
		for k, diff := range pdmDiffEntries {
			fmt.Printf("%s has changes in current and %d other branches\n", k, len(diff.Branches))
		}
		fmt.Print("\nTo show more information use the --json option")
		return
	}
}

func init() {
	diffCmd.Flags().StringSliceVarP(&filePatterns, "filepath", "f", []string{}, "Check the difference of a specific file or file type (use * as a wildcard). If empty checks all the locked file types specified in .gitattributes.")
	diffCmd.Flags().BoolVar(&outJson, "json", false, "If used, the output will be formatted in JSON")
	diffCmd.Flags().BoolVarP(&onlyConflict, "conflict", "c", false, "Prints only the files that were modified in current and in at least another different branch, those files can lead to conflict during merging and unlocking")
	rootCmd.AddCommand(diffCmd)
}
