package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetGitRoot() string {
	// Get top-level repo dir (absolute)
	repoRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	repoRootBytes, err := repoRootCmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(repoRootBytes))

}
func GetHooksPath() string {
	repoRoot := GetGitRoot()
	if repoRoot == "" {
		return ""
	}
	// Get the relative hooks path
	hooksPathCmd := exec.Command("git", "rev-parse", "--git-path", "hooks")
	hooksRelBytes, err := hooksPathCmd.Output()
	if err != nil {
		return ""
	}
	hooksRel := strings.TrimSpace(string(hooksRelBytes))

	// Join to get absolute path
	hooksAbs := filepath.Join(repoRoot, hooksRel)
	return hooksAbs
}

// GitRelFilepath gets the path relative to the git root
func GitRelFilepath(absPath string) (string, error) {
	// Get git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitRootBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get git root: %w", err)
	}
	gitRoot := string(gitRootBytes)
	gitRoot = filepath.Clean(string(gitRoot))

	// Compute relative path
	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("Failed to get relative path: %w", err)
	}

	return relPath, nil
}

// HasBranchDiff checks if a file differs from its version in the remote branch.
func HasBranchDiff(relPath, remoteBranch string) (bool, error) {

	// Step 1: Check if the file exists in the remote branch
	checkCmd := exec.Command("git", "ls-tree", "-r", remoteBranch, "--", relPath)
	checkCmd.Dir = GetGitRoot()
	output, err := checkCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check if file exists in remote branch: %w", err)
	}

	if len(output) == 0 {
		// File does NOT exist in remote branch → it was created in current branch
		return false, nil
	}

	// Step 2: Compare file in current HEAD vs remote
	diffCmd := exec.Command("git", "diff", fmt.Sprintf(remoteBranch), "--", relPath)
	diffCmd.Dir = GetGitRoot()
	var out bytes.Buffer
	diffCmd.Stdout = &out
	diffCmd.Stderr = &out

	if err := diffCmd.Run(); err != nil {
		return false, fmt.Errorf("git diff failed: %w\n%s", err, out.String())
	}

	// If output is not empty, file has changed
	changed := out.Len() > 0
	return changed, nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func HasDiff(filePath string) ([]string, error) {
	branches, err := GetRemoteBranches()
	if err != nil {
		return nil, err
	}

	// Get the current branch name (e.g., "main")
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	// Fetch latest remote data
	fetchCmd := exec.Command("git", "fetch", "origin")
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git fetch failed: %w\n%s", err, output)
	}

	var changedBranches []string
	for _, branch := range branches {
		// Skip origin/<currentBranch>
		if branch == "origin/"+currentBranch {
			continue
		}

		changed, err := HasBranchDiff(filePath, branch)
		if err != nil {
			fmt.Printf("Warning: skipping branch %s due to error: %v\n", branch, err)
			continue
		}
		if changed {
			changedBranches = append(changedBranches, branch)
		}
	}

	return changedBranches, nil
}

// GetRemoteBranches returns all remote branches (e.g., origin/main, origin/dev)
func GetRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}
	lines := strings.Split(string(out), "\n")
	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" && !strings.Contains(branch, "->") { // ignore symbolic refs
			branches = append(branches, branch)
		}
	}
	return branches, nil
}
