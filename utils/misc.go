package utils

import (
	"bytes"
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
	Print(str+"\n", args...)
}

func Print(str string, args ...interface{}) {
	if !OutJson {
		fmt.Printf(str, args...)
	}
}

func PrintCounter(i int, tot int, str string) {
	counter := fmt.Sprintf("[%d/%d] %s", i, tot, str)
	if Verbose {
		Println(counter)
	} else {
		Print("\r%-100s", counter)
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

// UnmarshalFirst fills 'v' with either the single object or the first element of an array.
// 'v' must be a pointer to the target variable.
func UnmarshalFirst[T any](data []byte, v *T) error {
	raw := bytes.TrimSpace(data)
	if len(raw) == 0 {
		return fmt.Errorf("empty input data")
	}

	// Case: JSON Array
	if raw[0] == '[' {
		var slice []T
		if err := json.Unmarshal(raw, &slice); err != nil {
			return err
		}
		if len(slice) == 0 {
			return fmt.Errorf("json array is empty")
		}
		// Assign the first element to the pointer
		*v = slice[0]
		return nil
	}

	// Case: Single JSON Object
	return json.Unmarshal(raw, v)
}