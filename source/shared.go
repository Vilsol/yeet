package source

import (
	"path/filepath"
	"strings"
)

func cleanPath(path string, dirPath string) string {
	trimmed := strings.Trim(strings.ReplaceAll(filepath.Clean(dirPath), "\\", "/"), "/")
	toRemove := len(strings.Split(trimmed, "/"))

	if trimmed == "." || trimmed == "" {
		toRemove = 0
	}

	cleanedPath := strings.ReplaceAll(filepath.Clean(path), "\\", "/")

	// Remove the initial path
	cleanedPath = strings.Join(strings.Split(cleanedPath, "/")[toRemove:], "/")

	if !strings.HasPrefix(cleanedPath, "/") {
		cleanedPath = "/" + cleanedPath
	}

	return cleanedPath
}
