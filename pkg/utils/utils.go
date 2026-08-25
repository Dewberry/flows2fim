package utils

import (
	"log/slog"
	"os"
)

// SliceContains checks if a slice contains a specific element
func SliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func WriteListToTempFile(list []string) (string, error) {
	tmpfile, err := os.CreateTemp("", "list-*.txt")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	for _, item := range list {
		if _, err := tmpfile.WriteString(item + "\n"); err != nil {
			return "", err
		}
	}

	slog.Debug("Created temporary list file", "path", tmpfile.Name(), "line_count", len(list))
	return tmpfile.Name(), nil
}
