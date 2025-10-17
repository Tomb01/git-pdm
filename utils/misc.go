package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var Verbose bool // Verbose enables verbose logging
var OutJson bool // OutJson enables JSON output

// JoinWithQuotes joins a slice of strings with a separator,
// wrapping each element in double quotes.
func JoinWithQuotes(parts []string, sep string) string {
	for i, v := range parts {
		parts[i] = `"` + v + `"`
	}
	return strings.Join(parts, sep)
}

// ToJson serializes a Go value into a JSON string.
func ToJson(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// LogVerbose prints a formatted message if Verbose mode is enabled.
func LogVerbose(str string, args ...interface{}) {
	if Verbose {
		fmt.Printf(str+"\n", args...)
	}
}

// Println prints a formatted message unless JSON output mode is enabled.
func Println(str string, args ...interface{}) {
	if !OutJson {
		fmt.Printf(str+"\n", args...)
	}
}

// PrintJSON encodes a value to JSON and prints it to stdout if OutJson is true.
// It returns an error if JSON encoding fails.
func PrintJSON(data interface{}) error {
	if !OutJson {
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}

// SplitArgs splits command-line arguments at the first "--" separator.
// Returns two slices: beforeArgs (before "--") and afterArgs (after "--").
// If no separator exists, afterArgs will be empty.
func SplitArgs(args []string) (beforeArgs []string, afterArgs []string) {
	separatorIndex := -1
	for i, a := range args {
		if a == "--" {
			separatorIndex = i
			break
		}
	}

	if separatorIndex != -1 {
		beforeArgs = args[:separatorIndex]
		if separatorIndex+1 < len(args) {
			afterArgs = args[separatorIndex+1:]
		}
	} else {
		beforeArgs = args
	}

	return beforeArgs, afterArgs
}
