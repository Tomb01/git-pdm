package utils

import (
	"fmt"
	"os"
	"strings"
)

func StringExistsInFile(filePath, searchString string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(content), searchString), nil
}

// SetReadOnly sets the file at the given path to read-only mode
func SetReadOnly(filePath string) error {
	err := os.Chmod(filePath, 0444) // Owner, group, others: read-only
	if err != nil {
		return fmt.Errorf("failed to set read-only: %w", err)
	}
	return nil
}
