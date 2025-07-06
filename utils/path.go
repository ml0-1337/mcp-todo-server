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

// ResolveTodoPath finds the project-level todo directory
func ResolveTodoPath() (string, error) {
	// 1. Check for override (rare, but supported)
	if customPath := GetEnv("CLAUDE_TODO_PATH", ""); customPath != "" {
		// Use fmt.Fprintf to stderr to avoid stdout pollution
		fmt.Fprintf(os.Stderr, "Using custom todo path: %s\n", customPath)
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
		fmt.Fprintf(os.Stderr, "No project root found, using current directory: %s\n", cwd)
		projectRoot = cwd
	} else {
		fmt.Fprintf(os.Stderr, "Found project root: %s\n", projectRoot)
	}

	// 4. Build todo path (ALWAYS project-level)
	todoPath := filepath.Join(projectRoot, ".claude", "todos")

	// 5. Ensure directory exists
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create todo directory: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Using todo directory: %s\n", todoPath)
	return todoPath, nil
}

// ResolveTodoPathFromWorkingDir resolves todo path from a specific working directory
func ResolveTodoPathFromWorkingDir(workingDir string) (string, error) {
	// Build todo path from provided working directory
	todoPath := filepath.Join(workingDir, ".claude", "todos")
	
	// Ensure directory exists
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create todo directory: %w", err)
	}
	
	fmt.Fprintf(os.Stderr, "Using todo directory from working dir: %s\n", todoPath)
	return todoPath, nil
}

// ResolveTemplatePath finds the template directory based on configured mode
func ResolveTemplatePath() (string, error) {
	// 1. Check for explicit override
	if customPath := GetEnv("CLAUDE_TEMPLATE_PATH", ""); customPath != "" {
		fmt.Fprintf(os.Stderr, "Using custom template path: %s\n", customPath)
		return customPath, nil
	}

	// 2. Get mode and debug flag
	mode := GetEnv("CLAUDE_TEMPLATE_MODE", "auto")
	debug := GetEnv("CLAUDE_DEBUG", "false") == "true"

	// 3. Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not get working directory: %v\n", err)
		cwd = "."
	}

	// 4. Find project root for project-relative paths
	projectRoot, _ := FindProjectRoot(cwd)
	if projectRoot == "" {
		projectRoot = cwd
	}

	// 5. Build candidate paths based on mode
	var candidates []string
	homeDir, _ := os.UserHomeDir()

	switch mode {
	case "auto":
		candidates = []string{
			filepath.Join(projectRoot, ".claude", "templates"),
			filepath.Join(projectRoot, "templates"),
			filepath.Join(homeDir, ".claude", "templates"),
		}
	case "project":
		candidates = []string{
			filepath.Join(projectRoot, ".claude", "templates"),
		}
	case "user":
		candidates = []string{
			filepath.Join(homeDir, ".claude", "templates"),
		}
	case "hybrid":
		// Special handling needed - return both paths
		projectPath := filepath.Join(projectRoot, ".claude", "templates")
		userPath := filepath.Join(homeDir, ".claude", "templates")

		// For hybrid mode, we'd need to modify the template manager
		// to handle multiple directories
		fmt.Fprintf(os.Stderr, "Hybrid mode: checking project (%s) and user (%s)\n", projectPath, userPath)

		// For now, use project if exists, otherwise user
		if IsDirectory(projectPath) {
			return projectPath, nil
		}
		return userPath, nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown CLAUDE_TEMPLATE_MODE: %s, using 'auto'\n", mode)
		mode = "auto"
		return ResolveTemplatePath() // Recursive call with auto mode
	}

	// 6. Find first existing directory
	for _, path := range candidates {
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Checking template path: %s\n", path)
		}
		if IsDirectory(path) {
			fmt.Fprintf(os.Stderr, "Using template directory: %s\n", path)
			return path, nil
		}
	}

	// 7. No directory found - show helpful error
	fmt.Fprintf(os.Stderr, "Warning: No template directory found. Searched:\n")
	for _, path := range candidates {
		fmt.Fprintf(os.Stderr, "  - %s\n", path)
	}

	// Return first candidate as default (will be created if needed)
	return candidates[0], nil
}
