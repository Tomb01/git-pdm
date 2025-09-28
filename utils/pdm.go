package utils

import (
	"fmt"
)

/*
	Status

- 0 -> older version or no changes, no action required
- 1 -> previous file version, need reload
- 2 -> no match between branch, new file
*/
type PdmDiffBranchStatus struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Commit string `json:"commit"`
	File   string `json:"file"`
}

func FileDiff(relPath string, branches []string, fast bool) ([]PdmDiffBranchStatus, error) {
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	file := []string{relPath}

	changedEntries := []PdmDiffBranchStatus{}
	var previousCommonAncestor string
	for _, branch := range branches {
		// Skip origin/<currentBranch>
		if branch == "origin/"+currentBranch || branch == "origin/main" {
			continue
		}
		LogVerbose(fmt.Sprintf("Comparing hash in %s\n", branch))
		// get the common ancestor between the two branches
		commonAncestor, err := GetCommonAncestor(currentBranch, branch)
		if err != nil {
			return nil, err
		}

		LogVerbose(fmt.Sprintf("Common ancestor of %s and HEAD is %s\n", branch, commonAncestor))
		if commonAncestor == previousCommonAncestor {
			// same ancestor than the previous branch -> same results
			if !fast {
				changedEntries = append(changedEntries, PdmDiffBranchStatus{Name: branch, Status: changedEntries[len(changedEntries)-1].Status, Commit: commonAncestor, File: changedEntries[len(changedEntries)-1].File})
			}
			continue
		}
		// get current branch hash
		current_hash, err := GetFileHash(relPath, branch)
		if err != nil {
			return nil, fmt.Errorf("Error in retriving the file current hash", err)
		}
		if current_hash == "" {
			// the file does not exist in branch -> deal with deleted file
			LogVerbose(fmt.Sprintf("The file does not exists in %s \n", branch))
			if !fast {
				changedEntries = append(changedEntries, PdmDiffBranchStatus{Name: branch, Status: 2, Commit: commonAncestor, File: ""})
			}
			continue
		}
		// get list of commit hash from common ancestor to current branch commit
		history, err := GetCommitHystory(commonAncestor, currentBranch, file)
		if err != nil {
			return nil, fmt.Errorf("Error in retriving commit history", err)
		}
		//history = append([]string{commonAncestor}, history...)
		//fmt.Println(commonAncestor)
		common_hash := false
		for _, commit := range history {
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
				LogVerbose(fmt.Sprintf("The file in %s is a previous version of the one in %s\n", branch, currentBranch))
				if !fast {
					changedEntries = append(changedEntries, PdmDiffBranchStatus{Name: branch, Status: 0, Commit: commit, File: current_hash})
				}
				common_hash = true
				break
			}
		}
		if !common_hash {
			LogVerbose(fmt.Sprintf("There is a newer version of the file in %s\n", branch))
			currentCommit, err := GetCurrentCommit(branch)
			if err != nil {
				return changedEntries, fmt.Errorf("Error in retriving the current commit", err)
			}
			fileHash, _ := GetFileHash(relPath, currentCommit)
			changedEntries = append(changedEntries, PdmDiffBranchStatus{Name: branch, Status: 1, Commit: currentCommit, File: fileHash})
			if fast {
				return changedEntries, nil
			}
		}
	}

	return changedEntries, nil
}
