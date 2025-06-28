package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetEnv returns environment variable or fallback
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FindProjectRoot finds the project root by looking for marker files/directories
func FindProjectRoot(startPath string) (string, error) {
	markers := []string{
		".claude",      // Claude project directory
		".git",         // Git repository
		"go.mod",       // Go module
		"package.json", // Node.js project
		".mcp.json",    // MCP configuration
	}
	
	current := startPath
	for {
		// Check each marker
		for _, marker := range markers {
			markerPath := filepath.Join(current, marker)
			if FileExists(markerPath) {
				return current, nil
			}
		}
		
		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			return "", fmt.Errorf("no project root found (checked for: %v)", markers)
		}
		current = parent
	}
}