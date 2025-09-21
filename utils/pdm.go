package utils

import (
	"fmt"
)

type PdmFileDiff struct {
	Path     string                `json:"path"`
	Branches []PdmDiffBranchStatus `json:"branches"`
}

/*
	Status

- 0 -> same file, no changes
- 1 -> previous file version, need reload
- 2 -> no match between branch, new file
*/
type PdmDiffBranchStatus struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Commit string `json:"commit"`
}

func FileDiff(relPath string, fast bool) ([]PdmDiffBranchStatus, error) {
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	branches, err := GetRemoteBranches()
	if err != nil {
		return nil, err
	}

	//root := GetGitRoot()
	current_hash, err := GetFileHash(relPath, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("Error in retriving the file current hash", err)
	}
	file := []string{relPath}
	//fmt.Println(current_hash)

	// Fetch latest remote data
	if output, err := execGitCommand("git", "fetch", "origin"); err != nil {
		return nil, fmt.Errorf("git fetch failed: %w\n%s", err, output)
	}

	changedEntries := []PdmDiffBranchStatus{}
	for _, branch := range branches {
		// Skip origin/<currentBranch>
		if branch == "origin/"+currentBranch {
			continue
		}
		LogVerbose(fmt.Sprintf("Comparing hash in %s\n", branch))
		// get the common ancestor between the two branches
		commonAncestor, err := GetCommonAncestor(branch, currentBranch)
		if err != nil {
			return nil, err
		}
		// get list of commit hash from common ancestor to current branch commit
		history, err := GetCommitHystory(commonAncestor, branch, file)
		if err != nil {
			return nil, fmt.Errorf("Error in retriving commit history", err)
		}
		history = append([]string{commonAncestor}, history...)
		common_hash := false
		for i, commit := range history {
			prev_hash, err := GetFileHash(relPath, commit)
			//LogVerbose(prev_hash)
			if err != nil {
				return changedEntries, fmt.Errorf("Error in retriving the file hash in commit "+commit, err)
			}
			if prev_hash == "" {
				// file doesn't exist in previous commit -> skip to next
				continue
			}
			if prev_hash == current_hash {
				if i != 0 {
					// previous hash equal -> the file is a previous version of this branch and must be updated
					LogVerbose(fmt.Sprintf("There is a newer version of the file in %s\n", branch))
					changedEntries = append(changedEntries, PdmDiffBranchStatus{Name: branch, Status: 1, Commit: commit})
					if fast {
						return changedEntries, nil
					}
				} else {
					// file is equal to the common ancestor -> no changes in the branch
					LogVerbose(fmt.Sprintf("No file changes in %s\n", branch))
				}
				// go to next branch
				common_hash = true
				break
			}
		}
		if !common_hash {
			LogVerbose(fmt.Sprintf("No common hash found in %s\n", branch))
		}
	}

	return changedEntries, nil
}
