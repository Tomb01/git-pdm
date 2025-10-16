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

func execGitCommand(args ...string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("no arguments provided to execGitCommand")
	}

	// Strip the leading "git" if included
	if args[0] == "git" {
		args = args[1:]
	}

	//LogVerbose(strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Dir = GetGitRoot()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	//fmt.Println(strings.Join(args, " "))
	//fmt.Println("Executing:", "git", strings.Join(args, " "))
	// If there is an error, but stderr is empty and stdout is empty (like `git diff`), return no error
	if err != nil {
		if stderr.Len() > 0 {
			// Real error (e.g., invalid command)
			return nil, errors.New(stderr.String())
		}
		// Command failed but no actual stderr — treat as non-critical
		return stdout.Bytes(), nil
	}
	//LogVerbose(strings.Join(args, " "))
	return stdout.Bytes(), nil
}

func Fetch() error {
	_, err := execGitCommand("git", "fetch", "origin")
	return err
}

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
	hooksRelBytes, err := execGitCommand("rev-parse", "--git-path", "hooks")
	if err != nil {
		return ""
	}
	hooksRel := strings.TrimSpace(string(hooksRelBytes))

	// Join to get absolute path
	hooksAbs := filepath.Join(repoRoot, hooksRel)
	return hooksAbs
}

// GitRelativeFilepath gets the path relative to the git root
func GitRelativeFilepath(absPath string) (string, error) {
	// Get git root
	gitRoot := GetGitRoot()
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
	output, err := execGitCommand("ls-tree", "-r", remoteBranch, "--", relPath)
	if err != nil {
		return false, fmt.Errorf("failed to check if file exists in remote branch: %w", err)
	}

	if len(output) == 0 {
		// File does NOT exist in remote branch → it was created in current branch
		return false, nil
	}

	// Step 2: Compare file in current HEAD vs remote
	output, err = execGitCommand("git", "diff", remoteBranch, "--", relPath)
	if err != nil {
		return false, fmt.Errorf("git diff failed: %w\n%s", err, string(output))
	}

	// If output is not empty, file has changed
	changed := len(output) > 0
	return changed, nil
}

func GetCurrentBranch() (string, error) {
	out, err := execGitCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func GetCurrentCommit(branch string) (string, error) {
	out, err := execGitCommand("git", "rev-parse", branch)
	if err != nil {
		return "", fmt.Errorf("failed to get current HEAD commit: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRemoteBranches returns all remote branches (e.g., origin/main, origin/dev)
func GetRemoteBranches() ([]string, error) {
	out, err := execGitCommand("git", "--no-pager", "branch", "-r")
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

// GetGitUserName returns the configured Git user.name
func GetGitUserName() (string, error) {
	out, err := execGitCommand("git", "config", "user.name")
	if err != nil {
		return "", fmt.Errorf("failed to get git user.name: %w\n%s", err, string(out))
	}
	return string(out), nil
}

func GetAbsoluteFilePath(gitRelPath string) (string, error) {
	// Get the top-level directory of the git repo
	output, err := execGitCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}

	root := strings.TrimSpace(string(output))

	// Join the root and relative path
	absPath := filepath.Join(root, gitRelPath)
	return absPath, nil
}

type DiffEntry struct {
	Status   string // e.g., M = modified, A = added, D = deleted
	Filename string
}

func GetDiff(commit1 string, commit2 string, file []string) ([]DiffEntry, error) {
	args := append([]string{"diff", commit1, commit2, "--name-status", "--"}, file...)
	fileBytes, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to run git diff comand: %w", err)
	}
	var entries []DiffEntry
	scanner := bufio.NewScanner(bytes.NewReader(fileBytes))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		if len(parts) >= 2 {
			entry := DiffEntry{
				Status:   parts[0],
				Filename: parts[1],
			}
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error in reading git diff output: %w", err)
	}

	return entries, nil
}

func GetCommonAncestor(baseBranch string, sourceBranch string) (string, error) {
	out, err := execGitCommand("merge-base", baseBranch, sourceBranch)
	if err != nil {
		return "", fmt.Errorf("Error in merge-base: %w", err)
	}
	return strings.ReplaceAll(string(out), "\n", ""), nil
}

// GitCheckoutFile checks out a specific file from a given branch
func CheckoutFile(branch string, filePath string) error {
	out, err := execGitCommand("checkout", branch, "--", filePath)
	if err != nil {
		return fmt.Errorf("error checking out file '%s' from branch '%s': %w\n%s", filePath, branch, err, string(out))
	}
	return nil
}

func GetCommitHystory(start string, end string, file []string) ([]string, error) {
	args := append([]string{"--no-pager", "log", end, start, "--pretty=format:%H", "--"}, file...)
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("Error retriving commit hystory: %w", err)
	}
	str := string(out)
	if str == "" {
		return []string{}, nil
	}
	return strings.Split(str, "\n"), nil
}

func GetFileHash(file string, commit string) (string, error) {
	out, err := execGitCommand("ls-tree", commit, "--", file)
	if err != nil {
		return "", fmt.Errorf("Error retriving commit hystory: %w", err)
	}
	str := string(out)
	if str == "" {
		return "", nil
	}
	fields := strings.Fields(str) // <-- correct usage
	if len(fields) < 3 {
		return "", fmt.Errorf("Unexpected hash format: %w", err)
	}
	return fields[2], nil
}

func GetBranchDiff(source string, dest string, filter []string) ([]string, error) {
	//git diff --name-only main..your-branch | grep -Ei '\.(js|ts|jsx)$'
	args := []string{"--no-pager", "diff", "--name-only", dest + ".." + source, "--"}
	args = append(args, filter...)
	LogVerbose(strings.Join(args, " ") + "\n")
	out, err := execGitCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("Error retriving diff: %w", err)
	}
	str := string(out)
	//LogVerbose(str)
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
