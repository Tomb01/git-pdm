package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetGitPath() string {
	// Get top-level repo dir (absolute)
	repoRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	repoRootBytes, err := repoRootCmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(repoRootBytes))

}
func GetHooksPath() string {
	repoRoot := GetGitPath()
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
