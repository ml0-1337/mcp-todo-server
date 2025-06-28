package utils

import (
	"fmt"
	"log"
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

// ResolveTodoPath finds the project-level todo directory
func ResolveTodoPath() (string, error) {
	// 1. Check for override (rare, but supported)
	if customPath := GetEnv("CLAUDE_TODO_PATH", ""); customPath != "" {
		log.Printf("Using custom todo path: %s", customPath)
		return customPath, nil
	}
	
	// 2. Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// 3. Find project root
	projectRoot, err := FindProjectRoot(cwd)
	if err != nil {
		// No project markers found, use current directory
		log.Printf("No project root found, using current directory: %s", cwd)
		projectRoot = cwd
	} else {
		log.Printf("Found project root: %s", projectRoot)
	}
	
	// 4. Build todo path (ALWAYS project-level)
	todoPath := filepath.Join(projectRoot, ".claude", "todos")
	
	// 5. Ensure directory exists
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create todo directory: %w", err)
	}
	
	log.Printf("Using todo directory: %s", todoPath)
	return todoPath, nil
}