package utils

import (
	"fmt"
)

/*
Status codes for PdmDiffBranchStatus:

0 -> older version or no changes, no action required
1 -> previous file version, needs reload
2 -> no match between branches, new file
*/

// PdmDiffBranchStatus holds the status of a file in a specific branch relative to the current branch.
type PdmDiffBranchStatus struct {
	Name   string `json:"name"`   // Branch name
	Status int    `json:"status"` // Status code (0, 1, 2)
	Commit string `json:"commit"` // Commit hash for the relevant version
	File   string `json:"file"`   // File hash
}

// FileDiff compares a file across multiple branches to determine differences
// relative to the current branch.
//
// Parameters:
//   - relPath: The file path relative to the git root
//   - branches: List of branches to compare against
//   - fast: If true, returns on first detected difference; otherwise, continues checking all branches
//
// Returns a slice of PdmDiffBranchStatus for each branch that differs from the current branch.
func FileDiff(relPath string, branches []string, fast bool) ([]PdmDiffBranchStatus, error) {
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	file := []string{relPath}
	changedEntries := []PdmDiffBranchStatus{}
	var previousCommonAncestor string

	for _, branch := range branches {
		// Skip origin/currentBranch and origin/main
		if branch == "origin/"+currentBranch || branch == "origin/main" {
			continue
		}

		LogVerbose(fmt.Sprintf("Comparing hash in %s", branch))

		// Get the common ancestor between current branch and target branch
		commonAncestor, err := GetCommonAncestor(currentBranch, branch)
		if err != nil {
			return nil, err
		}

		LogVerbose(fmt.Sprintf("Common ancestor of %s and HEAD is %s", branch, commonAncestor))

		// Skip if the common ancestor matches the previous one (reuse result)
		if commonAncestor == previousCommonAncestor {
			if !fast && len(changedEntries) > 0 {
				last := changedEntries[len(changedEntries)-1]
				changedEntries = append(changedEntries, PdmDiffBranchStatus{
					Name:   branch,
					Status: last.Status,
					Commit: commonAncestor,
					File:   last.File,
				})
			}
			continue
		}

		// Get file hash in target branch
		currentHash, err := GetFileHash(relPath, branch)
		if err != nil {
			return nil, fmt.Errorf("error retrieving the file hash for branch %s: %w", branch, err)
		}
		if currentHash == "" {
			// File does not exist in branch
			LogVerbose(fmt.Sprintf("The file does not exist in %s", branch))
			if !fast {
				changedEntries = append(changedEntries, PdmDiffBranchStatus{
					Name:   branch,
					Status: 2,
					Commit: commonAncestor,
					File:   "",
				})
			}
			continue
		}

		// Get commit history from common ancestor to current branch
		history, err := GetCommitHistory(commonAncestor, currentBranch, file)
		if err != nil {
			return nil, fmt.Errorf("error retrieving commit history for %s: %w", relPath, err)
		}

		commonHash := false
		for _, commit := range history {
			prevHash, err := GetFileHash(relPath, commit)
			if err != nil {
				return changedEntries, fmt.Errorf("error retrieving file hash in commit %s: %w", commit, err)
			}
			if prevHash == "" {
				continue
			}
			if prevHash == currentHash {
				LogVerbose(fmt.Sprintf("The file in %s is a previous version of the one in %s", branch, currentBranch))
				if !fast {
					changedEntries = append(changedEntries, PdmDiffBranchStatus{
						Name:   branch,
						Status: 0,
						Commit: commit,
						File:   currentHash,
					})
				}
				commonHash = true
				break
			}
		}

		if !commonHash {
			LogVerbose(fmt.Sprintf("There is a newer version of the file in %s", branch))
			currentCommit, err := GetCurrentCommit(branch)
			if err != nil {
				return changedEntries, fmt.Errorf("error retrieving current commit for branch %s: %w", branch, err)
			}
			fileHash, _ := GetFileHash(relPath, currentCommit)
			changedEntries = append(changedEntries, PdmDiffBranchStatus{
				Name:   branch,
				Status: 1,
				Commit: currentCommit,
				File:   fileHash,
			})
			if fast {
				return changedEntries, nil
			}
		}

		previousCommonAncestor = commonAncestor
	}

	return changedEntries, nil
}
