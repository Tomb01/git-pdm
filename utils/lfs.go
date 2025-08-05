package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Lock represents a single lock in the JSON output of `git lfs locks --json`
type Lock struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// LFSLocksOutput represents the structure of the JSON output
type LFSLocksOutput struct {
	Locks []Lock `json:"locks"`
}

func IsLocked(id string) (bool, string, error) {
	repoRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
}

// GetLFSLockID returns the Git LFS lock ID for a given absolute file path
func GetLFSLockID(absPath string) (string, error) {
	if !filepath.IsAbs(absPath) {
		return "", fmt.Errorf("path must be absolute: %s", absPath)
	}

	// Convert to relative path from git root
	relPath, err := GitRelFilepath(absPath)
	if err != nil {
		return "", err
	}

	// Run `git lfs locks --json`
	cmd := exec.Command("git", "lfs", "locks", "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git lfs locks: %w", err)
	}

	var locksOutput LFSLocksOutput
	if err := json.Unmarshal(output, &locksOutput); err != nil {
		return "", fmt.Errorf("failed to parse lfs locks output: %w", err)
	}

	for _, lock := range locksOutput.Locks {
		if filepath.Clean(lock.Path) == relPath {
			return lock.ID, nil
		}
	}

	return "", fmt.Errorf("no lock found for path: %s", absPath)
}
