package utils

import (
	"fmt"
	"path"
)

type PdmFileDiff struct {
	Path     string                `json:"path"`
	Branches []PdmDiffBranchStatus `json:"branches"`
}

type PdmDiffBranchStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func Diff(relPath []string, fast bool) (map[string]PdmFileDiff, error) {
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	branches, err := GetRemoteBranches()
	if err != nil {
		return nil, err
	}

	root := GetGitRoot()

	// Fetch latest remote data
	if output, err := execGitCommand("git", "fetch", "origin"); err != nil {
		return nil, fmt.Errorf("git fetch failed: %w\n%s", err, output)
	}

	changedEntries := make(map[string]PdmFileDiff)
	for _, branch := range branches {
		// Skip origin/<currentBranch>
		if branch == "origin/"+currentBranch {
			continue
		}
		LogVerbose(fmt.Sprintf("Comparing changes in %s\n", branch))
		// get the commit in which the branch was separated from main
		commonAncestor, err := GetCommonAncestor(branch, "main")
		if err != nil {
			return nil, err
		}
		// get changes from common ancestor to main and the branch HEAD -> if true the changes was edited in this branch
		changes, err := GetDiff(commonAncestor, branch, relPath)
		if err != nil {
			return nil, err
		}
		if len(changes) > 0 {
			// check if the file different from the current
			if fast {
				LogVerbose(fmt.Sprintf("The file was edited in %s\nCheck if it differs from the current head . . . ", branch))
			}

			changes, err := GetDiff("HEAD", branch, relPath)
			if err != nil {
				return nil, err
			}
			if len(changes) > 0 {
				// the file was edited in the target branch
				if fast {
					LogVerbose(fmt.Sprintf("The file was edited both in %s and current branch\n", branch))
				}
				if fast {
					LogVerbose("Fast mode ON. Only the first changes will be returned\n")
					changedEntries[relPath[0]] = PdmFileDiff{Path: relPath[0], Branches: []PdmDiffBranchStatus{{Name: branch, Status: "M"}}}
					return changedEntries, nil
				} else {
					for _, diff := range changes {
						pdmDiff := changedEntries[diff.Filename]
						if pdmDiff.Path == "" {
							// create new diff
							pdmDiff = PdmFileDiff{
								Path: path.Join(root, diff.Filename),
								Branches: []PdmDiffBranchStatus{
									{Name: branch, Status: diff.Status},
								}}
						} else {
							// append branch
							pdmDiff.Branches = append(pdmDiff.Branches, PdmDiffBranchStatus{Name: branch, Status: diff.Status})
						}
						changedEntries[diff.Filename] = pdmDiff
					}
					//LogVerbose("Complete\n")
				}
			} else {
				LogVerbose("No concurrent edit\n")
			}
		}
	}

	return changedEntries, nil
}
