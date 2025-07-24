package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PathCache provides a thread-safe cache for todo file paths
type PathCache struct {
	mu      sync.RWMutex
	cache   map[string]string
	order   []string // Track insertion order for LRU
	maxSize int
}

// NewPathCache creates a new path cache with the specified maximum size
func NewPathCache(maxSize int) *PathCache {
	return &PathCache{
		cache:   make(map[string]string),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a path from the cache
func (pc *PathCache) Get(todoID string) (string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	path, found := pc.cache[todoID]
	return path, found
}

// Set stores a path in the cache with LRU eviction
func (pc *PathCache) Set(todoID, path string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// If already exists, just update
	if _, exists := pc.cache[todoID]; exists {
		pc.cache[todoID] = path
		return
	}

	// If cache is full, remove oldest entry (LRU)
	if len(pc.cache) >= pc.maxSize {
		if len(pc.order) > 0 {
			oldest := pc.order[0]
			delete(pc.cache, oldest)
			pc.order = pc.order[1:]
		}
	}

	// Add new entry
	pc.cache[todoID] = path
	pc.order = append(pc.order, todoID)
}

// Delete removes a path from the cache
func (pc *PathCache) Delete(todoID string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	delete(pc.cache, todoID)
	
	// Remove from order slice
	for i, id := range pc.order {
		if id == todoID {
			pc.order = append(pc.order[:i], pc.order[i+1:]...)
			break
		}
	}
}

// Clear empties the cache
func (pc *PathCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	pc.cache = make(map[string]string)
	pc.order = make([]string, 0, pc.maxSize)
}

// Global path cache instance
var globalPathCache = NewPathCache(1000)

// GetDateBasedTodoPath returns the full path for a todo based on its started date
func GetDateBasedTodoPath(basePath string, todoID string, started time.Time) string {
	datePath := GetDailyPath(started)
	return filepath.Join(basePath, ".claude", "todos", datePath, todoID+".md")
}

// ResolveTodoPath finds where a todo is stored, checking both flat and date-based structures
func ResolveTodoPath(basePath string, todoID string) (string, error) {
	// Strategy 1: Check cache first
	if cachedPath, found := globalPathCache.Get(todoID); found {
		// Verify the file still exists
		if _, err := os.Stat(cachedPath); err == nil {
			return cachedPath, nil
		}
		// Cache entry is stale, remove it
		globalPathCache.Delete(todoID)
	}

	// Strategy 2: Check flat structure first (backwards compatibility)
	flatPath := filepath.Join(basePath, ".claude", "todos", todoID+".md")
	if _, err := os.Stat(flatPath); err == nil {
		globalPathCache.Set(todoID, flatPath)
		return flatPath, nil
	}

	// Strategy 3: Recursive search in date-based structure
	todosRoot := filepath.Join(basePath, ".claude", "todos")
	var foundPath string

	err := filepath.Walk(todosRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue search
		}
		
		// Skip if it's a directory
		if info.IsDir() {
			return nil
		}
		
		// Check if this is the file we're looking for
		if info.Name() == todoID+".md" {
			foundPath = path
			return filepath.SkipDir // Stop searching
		}
		
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error searching for todo: %w", err)
	}

	if foundPath == "" {
		return "", os.ErrNotExist
	}

	// Cache the result
	globalPathCache.Set(todoID, foundPath)
	return foundPath, nil
}

// ScanDateRange returns all todo files within a date range
func ScanDateRange(basePath string, startDate, endDate time.Time) ([]string, error) {
	var paths []string
	todosRoot := filepath.Join(basePath, ".claude", "todos")

	// Iterate through each day in the range
	current := startDate
	for !current.After(endDate) {
		datePath := GetDailyPath(current)
		dirPath := filepath.Join(todosRoot, datePath)

		// Check if directory exists
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			// Read all files in this directory
			files, err := os.ReadDir(dirPath)
			if err != nil {
				// Log but continue with other directories
				fmt.Fprintf(os.Stderr, "Warning: failed to read directory %s: %v\n", dirPath, err)
				current = current.AddDate(0, 0, 1)
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
					paths = append(paths, filepath.Join(dirPath, file.Name()))
				}
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	return paths, nil
}

// MigrateToDateStructure moves a todo from flat to date-based structure
func MigrateToDateStructure(basePath string, todoID string, started time.Time) error {
	// Source path (flat structure)
	sourcePath := filepath.Join(basePath, ".claude", "todos", todoID+".md")
	
	// Check if source exists
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("todo %s not found in flat structure", todoID)
		}
		return fmt.Errorf("error checking source file: %w", err)
	}

	// Destination path (date-based structure)
	destPath := GetDateBasedTodoPath(basePath, todoID, started)
	destDir := filepath.Dir(destPath)

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the file
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to move todo file: %w", err)
	}

	// Update cache
	globalPathCache.Set(todoID, destPath)

	return nil
}

// EnsureDateDirectory creates the directory structure for a given date
func EnsureDateDirectory(basePath string, date time.Time) error {
	datePath := GetDailyPath(date)
	dir := filepath.Join(basePath, ".claude", "todos", datePath)
	return os.MkdirAll(dir, 0755)
}