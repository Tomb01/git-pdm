package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// execGitCommand executes a Git command with the given arguments in the current repository.
// If the first argument is "git", it will be stripped automatically.
// Returns the command output or an error.
func execGitCommand(args ...string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("no arguments provided to execGitCommand")
	}

	if args[0] == "git" {
		args = args[1:]
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = GetGitRoot()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, errors.New(stderr.String())
		}
		return stdout.Bytes(), nil
	}

	return stdout.Bytes(), nil
}

// Fetch executes 'git fetch origin' to update remote tracking branches.
func Fetch() error {
	_, err := execGitCommand("git", "fetch", "origin")
	return err
}

// GetGitRoot returns the absolute path to the top-level Git repository.
// Returns an empty string if the repository cannot be determined.
func GetGitRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetHooksPath returns the absolute path to the Git hooks directory.
func GetHooksPath() string {
	repoRoot := GetGitRoot()
	if repoRoot == "" {
		return ""
	}
	hooksRelBytes, err := execGitCommand("rev-parse", "--git-path", "hooks")
	if err != nil {
		return ""
	}
	hooksRel := strings.TrimSpace(string(hooksRelBytes))
	return filepath.Join(repoRoot, hooksRel)
}

// GitRelativeFilepath returns the path of a file relative to the Git root directory.
func GitRelativeFilepath(absPath string) (string, error) {
	gitRoot := GetGitRoot()
	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	return relPath, nil
}

// HasBranchDiff checks if a file differs from its version in the remote branch.
func HasBranchDiff(relPath, remoteBranch string) (bool, error) {
	output, err := execGitCommand("ls-tree", "-r", remoteBranch, "--", relPath)
	if err != nil {
		return false, fmt.Errorf("failed to check if file exists in remote branch: %w", err)
	}

	if len(output) == 0 {
		return false, nil
	}

	output, err = execGitCommand("git", "diff", remoteBranch, "--", relPath)
	if err != nil {
		return false, fmt.Errorf("git diff failed: %w\n%s", err, string(output))
	}

	return len(output) > 0, nil
}

// GetCurrentBranch returns the name of the current Git branch.
func GetCurrentBranch() (string, error) {
	out, err := execGitCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentCommit returns the commit hash for the specified branch.
func GetCurrentCommit(branch string) (string, error) {
	out, err := execGitCommand("git", "rev-parse", branch)
	if err != nil {
		return "", fmt.Errorf("failed to get current HEAD commit: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRemoteBranches returns a list of remote branches (e.g., origin/main).
func GetRemoteBranches() ([]string, error) {
	out, err := execGitCommand("git", "--no-pager", "branch", "-r")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}
	lines := strings.Split(string(out), "\n")
	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" && !strings.Contains(branch, "->") {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// GetGitUserName returns the configured Git user.name.
func GetGitUserName() (string, error) {
	out, err := execGitCommand("git", "config", "user.name")
	if err != nil {
		return "", fmt.Errorf("failed to get git user.name: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GetAbsoluteFilePath returns the absolute path of a file relative to the Git root.
func GetAbsoluteFilePath(gitRelPath string) (string, error) {
	root, err := execGitCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}
	return filepath.Join(strings.TrimSpace(string(root)), gitRelPath), nil
}

// DiffEntry represents a file difference between two commits.
type DiffEntry struct {
	Status   string // e.g., M = modified, A = added, D = deleted
	Filename string
}

// GetDiff returns the diff entries between two commits for the specified files.
func GetDiff(commit1, commit2 string, files []string) ([]DiffEntry, error) {
	args := append([]string{"diff", commit1, commit2, "--name-status", "--"}, files...)
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff command: %w", err)
	}

	var entries []DiffEntry
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) >= 2 {
			entries = append(entries, DiffEntry{
				Status:   parts[0],
				Filename: parts[1],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading git diff output: %w", err)
	}
	return entries, nil
}

// GetCommonAncestor returns the merge-base commit hash between two branches.
func GetCommonAncestor(baseBranch, sourceBranch string) (string, error) {
	out, err := execGitCommand("merge-base", baseBranch, sourceBranch)
	if err != nil {
		return "", fmt.Errorf("error in merge-base: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CheckoutFile checks out a specific file from the given branch.
func CheckoutFile(branch, filePath string) error {
	out, err := execGitCommand("checkout", branch, "--", filePath)
	if err != nil {
		return fmt.Errorf("error checking out file '%s' from branch '%s': %w\n%s", filePath, branch, err, string(out))
	}
	return nil
}

// GetCommitHistory returns a list of commit hashes between start and end for the specified files.
func GetCommitHistory(start, end string, files []string) ([]string, error) {
	args := append([]string{"--no-pager", "log", end, start, "--pretty=format:%H", "--"}, files...)
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving commit history: %w", err)
	}
	if len(out) == 0 {
		return []string{}, nil
	}
	return strings.Split(string(out), "\n"), nil
}

// GetFileHash returns the blob hash of a file at a specific commit.
func GetFileHash(file, commit string) (string, error) {
	out, err := execGitCommand("ls-tree", commit, "--", file)
	if err != nil {
		return "", fmt.Errorf("error retrieving file hash: %w", err)
	}
	str := string(out)
	if str == "" {
		return "", nil
	}
	fields := strings.Fields(str)
	if len(fields) < 3 {
		return "", fmt.Errorf("unexpected hash format")
	}
	return fields[2], nil
}

// GetBranchDiff returns the list of files that differ between two branches, optionally filtered.
func GetBranchDiff(source, dest string, filter []string) ([]string, error) {
	args := []string{"--no-pager", "diff", "--name-only", dest + ".." + source, "--"}
	args = append(args, filter...)
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving branch diff: %w", err)
	}
	str := string(out)
	if str == "" {
		return []string{}, nil
	}
	lines := strings.Split(str, "\n")
	var filtered []string
	for _, s := range lines {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func GetLogHistory(branch string, file string) ([]string, error) {
	out, err := execGitCommand("--no-pager", "log", "--format=%H", branch, "--", file)
	if err != nil {
		return nil, fmt.Errorf("error retrieving branch log: %w", err)
	}
	str := string(out)
	if str == "" {
		return []string{}, nil
	}
	lines := strings.Split(str, "\n")
	var filtered []string
	for _, s := range lines {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func GetFiles(filter []string) ([]string, error) {
	args := []string{"--no-pager", "ls-files"}
	args = append(args, filter...)
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving files: %w", err)
	}
	str := string(out)
	if str == "" {
		return []string{}, nil
	}
	lines := strings.Split(str, "\n")
	var filtered []string
	for _, s := range lines {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}
