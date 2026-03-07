package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Owner represents the owner of a Git LFS lock.
type Owner struct {
	Name string `json:"name"`
}

// Lock represents a single Git LFS lock.
type Lock struct {
	ID       string    `json:"id"`
	Path     string    `json:"path"`
	Owner    Owner     `json:"owner"`
	LockedAt time.Time `json:"locked_at"`
}

// UnLock represents the result of an unlock operation.
type UnLock struct {
	Path string `json:"path"`
	Unlocked bool `json:"unlocked"`
	Reason string `json:"reason"`
}

// LockVerify represents the full JSON output of `git lfs locks --json`,
// separating locks owned by the current user (Ours) and by others (Theirs).
type LockVerify struct {
	Ours   []Lock `json:"ours"`
	Theirs []Lock `json:"theirs"`
}

// LockFile locks a file using Git LFS.
// Returns true if the file was successfully locked, false if it is already locked,
// the Lock object, and an error if the operation failed.
func LockFile(file string) (bool, Lock, error) {
	lockOutputBytes, err := execGitCommand("git", "lfs", "lock", file, "--json")
	if err != nil {
		return false, Lock{}, err
	}

	lockOutput := string(lockOutputBytes)
	var lock Lock

	if strings.Contains(lockOutput, "Lock exists") {
		lockData, _ := GetLockStatus(file)
		return false, lockData, nil
	} else if strings.Contains(lockOutput, "locked_at") && !strings.Contains(lockOutput, "owner") {
		if err := UnmarshalFirst(lockOutputBytes, &lock); err != nil {
			return false, Lock{}, fmt.Errorf("error reading LFS output: %w", err)
		}
		return true, lock, nil
	} else {
		return false, Lock{}, fmt.Errorf("general error in locking operation")
	}
}

// UnLockFile unlocks a file using Git LFS.
// Returns true if successfully unlocked, the previous Lock (if any), and an error.
func UnLockFile(relPath string) (bool, Lock, error) {
	lockStatus, err := GetLockStatus(relPath)
	if err != nil {
		return false, lockStatus, err
	}

	absPath, err := GetAbsoluteFilePath(relPath)
	if err != nil {
		return false, Lock{}, fmt.Errorf("unable to retrieve complete file path: %w", err)
	}

	// Set file read-only before unlocking
	if err := SetReadOnly(absPath); err != nil {
		return false, Lock{}, fmt.Errorf("error setting file read-only before unlocking: %w", err)
	}

	unlockOutputBytes, err := execGitCommand("git", "lfs", "unlock", relPath, "--json")
	if err != nil {
		return false, Lock{}, err
	}

	unlockOutput := string(unlockOutputBytes)
	var unlock UnLock

	switch {
	case strings.Contains(unlockOutput, "Lock exists"):
		lockData, _ := GetLockStatus(relPath)
		return false, lockData, nil
	case strings.Contains(unlockOutput, "unlocked"):
		if err := UnmarshalFirst(unlockOutputBytes, &unlock); err != nil {
			return false, Lock{}, fmt.Errorf("error reading LFS object output: %w", err)
		}
		return unlock.Unlocked, Lock{}, nil
	case strings.Contains(unlockOutput, "no matching locks found"):
		return true, Lock{}, nil
	default:
		return false, Lock{}, fmt.Errorf("general error in unlocking operation")
	}
}

// GetLockStatus returns the Git LFS Lock for a given relative path.
// Returns an error if no lock exists.
func GetLockStatus(relPath string) (Lock, error) {
	locks, err := GetLocks(false)
	if err != nil {
		return Lock{}, fmt.Errorf("failed to run git lfs locks: %w", err)
	}

	for _, lock := range locks {
		if lock.Path == relPath {
			return lock, nil
		}
	}

	return Lock{}, nil
}

// GetLocks returns all Git LFS locks.
// If onlyUser is true, only returns locks owned by the current user.
func GetLocks(onlyUser bool) ([]Lock, error) {
	output, err := execGitCommand("git", "lfs", "locks", "--json", "--verify")
	if err != nil {
		return nil, fmt.Errorf("failed to run git lfs locks: %w", err)
	}

	var locks LockVerify
	if err := json.Unmarshal(output, &locks); err != nil {
		return nil, fmt.Errorf("error unmarshaling locks JSON: %w", err)
	}

	if onlyUser {
		return locks.Ours, nil
	}
	return append(locks.Ours, locks.Theirs...), nil
}

// GetLockableFiles reads the .gitattributes file and returns all file patterns
// marked as lockable, e.g., "*.sldprt".
func GetLockableFiles() ([]string, error) {
	path := filepath.Join(GetGitRoot(), ".gitattributes")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lockableExtensions := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pattern, attributes := fields[0], fields[1:]
		for _, attr := range attributes {
			if attr == "lockable" && strings.HasPrefix(pattern, "*.") {
				lockableExtensions[strings.ToUpper(pattern)] = struct{}{}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(lockableExtensions))
	for ext := range lockableExtensions {
		result = append(result, ext)
	}

	return result, nil
}
