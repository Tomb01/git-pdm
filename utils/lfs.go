package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Lock represents a single lock in the JSON output of `git lfs locks --json`
type Owner struct {
	Name string `json:"name"`
}

type Lock struct {
	ID       string    `json:"id"`
	Path     string    `json:"path"`
	Owner    Owner     `json:"owner"`
	LockedAt time.Time `json:"locked_at"`
}

type UnLock struct {
	Unlocked bool `json:"unlocked"`
}

// Root struct for the full JSON
type LockVerify struct {
	Ours   []Lock `json:"ours"`
	Theirs []Lock `json:"theirs"`
}

func LockFile(file string) (bool, Lock, error) {
	lockCmd := exec.Command("git", "lfs", "lock", file, "--json")
	lockCmd.Dir = GetGitRoot()
	//lockOutputBytes, err := lockCmd.Output()
	lockOutputBytes, _ := lockCmd.CombinedOutput()
	lockOutput := string(lockOutputBytes)
	var lock Lock
	//fmt.Println(lockOutput)
	if strings.Contains(lockOutput, "Lock exists") {
		lockData, _ := GetLockStatus(file)
		return false, lockData, nil
	} else if strings.Contains(lockOutput, "locked_at") && !strings.Contains(lockOutput, "owner") {
		if err := json.Unmarshal([]byte(lockOutputBytes), &lock); err != nil {
			return false, Lock{}, fmt.Errorf("Error reading lfs output: %w", err)
		} else {
			return true, lock, nil
		}
	} else {
		return false, Lock{}, fmt.Errorf(lockOutput)
	}
}

func UnLockFile(file string) (bool, Lock, error) {
	unlockCmd := exec.Command("git", "lfs", "unlock", file, "--json")
	unlockCmd.Dir = GetGitRoot()
	absPath, _ := GetAbsoluteFilePath(file)
	//unlockOutputBytes, err := lockCmd.Output()
	unlockOutputBytes, _ := unlockCmd.CombinedOutput()
	unlockOutput := string(unlockOutputBytes)
	var unlock UnLock
	//fmt.Println(lockOutput)
	if strings.Contains(unlockOutput, "Lock exists") {
		// Locked by another user
		lockData, _ := GetLockStatus(file)
		return false, lockData, nil
	} else if strings.Contains(unlockOutput, "unlocked") {
		// Unlocked procedure completed
		if err := json.Unmarshal([]byte(unlockOutputBytes), &unlock); err != nil {
			return false, Lock{}, fmt.Errorf("Error reading lfs output: %w", err)
		} else {
			if unlock.Unlocked {
				SetReadOnly(absPath)
			}
			return unlock.Unlocked, Lock{}, nil
		}
	} else if strings.Contains(unlockOutput, "no matching locks found") {
		// No lock exist
		SetReadOnly(absPath)
		return true, Lock{}, nil
	} else {
		return false, Lock{}, fmt.Errorf(unlockOutput)
	}
}

// GetLockStatus returns the Git LFS lock ID for a given absolute file path
func GetLockStatus(relPath string) (Lock, error) {

	locks, err := GetOursLocks()
	if err != nil {
		return Lock{}, fmt.Errorf("failed to run git lfs locks: %w", err)
	}
	for _, lock := range locks {
		if lock.Path == relPath {
			return lock, nil
		}
	}

	return Lock{}, fmt.Errorf("no lock found for path: %s", relPath)
}

func GetOursLocks() ([]Lock, error) {
	// Run `git lfs locks --json`
	output, err := execGitCommand("git", "lfs", "locks", "--verify", "--json")
	if err != nil {
		return nil, fmt.Errorf("failed to run git lfs locks: %w", err)
	}

	var locks LockVerify
	if err := json.Unmarshal([]byte(output), &locks); err != nil {
		return nil, fmt.Errorf("Error unmarshaling: %w", err)
	}

	return locks.Ours, nil
}

func GetLockableFiles() ([]string, error) {
	path := GetGitRoot() + "/.gitattributes"
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lockableExtensions := make(map[string]struct{}) // Use map to avoid duplicates

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pattern := fields[0]
		attributes := fields[1:]

		// Check if "lockable" is present in attributes
		for _, attr := range attributes {
			if attr == "lockable" {
				// Extract the extension from the pattern
				if strings.HasPrefix(pattern, "*.") {
					//ext := pattern[2:] // remove "*."
					lockableExtensions[strings.ToUpper(pattern)] = struct{}{}
				}
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert map keys to a slice
	result := make([]string, 0, len(lockableExtensions))
	for ext := range lockableExtensions {
		result = append(result, ext)
	}

	return result, nil
}
