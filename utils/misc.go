package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

var Verbose bool

func JoinWithQuotes(parts []string, sep string) string {
	for i, v := range parts {
		parts[i] = `"` + v + `"`
	}
	return strings.Join(parts, sep)
}

func ToJson(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func LogVerbose(str string) {
	if Verbose {
		fmt.Print(str)
	}
}
