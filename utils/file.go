package utils

import (
	"fmt"
	"os"
	"strings"
)

// StringExistsInFile checks if the given searchString exists within the file
// specified by filePath. Returns true if the string is found, false otherwise.
// Returns an error if the file cannot be read.
func StringExistsInFile(filePath, searchString string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), searchString), nil
}

// SetReadOnly sets the file at filePath to read-only mode for owner, group,
// and others. Returns an error if the operation fails.
func SetReadOnly(filePath string) error {
	if err := os.Chmod(filePath, 0444); err != nil {
		return fmt.Errorf("failed to set read-only for %s: %w", filePath, err)
	}
	return nil
}

// FileExists checks whether a file or directory exists at the given path.
// Returns true if the file exists, false otherwise.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	// For any other error, assume the file does not exist
	return err == nil && info != nil
}
