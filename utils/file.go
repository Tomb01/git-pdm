package utils

import (
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
