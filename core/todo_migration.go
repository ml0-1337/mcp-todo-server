package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// MigrationStats tracks migration progress
type MigrationStats struct {
	Total    int
	Migrated int
	Failed   int
	Skipped  int
	Errors   []error
}

// MigrateToDateStructure migrates all todos from flat to date-based structure
func (tm *TodoManager) MigrateToDateStructure() (*MigrationStats, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	stats := &MigrationStats{}
	oldTodosDir := filepath.Join(tm.basePath, ".claude", "todos")
	tempDir := filepath.Join(tm.basePath, ".claude", "todos_migrating")

	// Check if migration is needed
	needsMigration, err := tm.needsMigration(oldTodosDir)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to check migration status")
	}

	if !needsMigration {
		fmt.Fprintf(os.Stderr, "Migration not needed - no flat structure todos found\n")
		return stats, nil
	}

	// Step 1: Create backup/temp directory
	fmt.Fprintf(os.Stderr, "Starting migration: creating backup...\n")
	if err := os.Rename(oldTodosDir, tempDir); err != nil {
		return nil, interrors.Wrap(err, "failed to create migration backup")
	}

	// Step 2: Create new todos directory
	if err := os.MkdirAll(oldTodosDir, 0755); err != nil {
		// Attempt rollback
		if renameErr := os.Rename(tempDir, oldTodosDir); renameErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to rollback migration: %v\n", renameErr)
		}
		return nil, interrors.Wrap(err, "failed to create new todos directory")
	}

	// Step 3: Read all files from temp directory
	files, err := os.ReadDir(tempDir)
	if err != nil {
		// Attempt rollback
		os.RemoveAll(oldTodosDir)
		if renameErr := os.Rename(tempDir, oldTodosDir); renameErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to rollback migration: %v\n", renameErr)
		}
		return nil, interrors.Wrap(err, "failed to read migration directory")
	}

	stats.Total = len(files)

	// Step 4: Process each file
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	progressChan := make(chan bool, len(files))

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			stats.Skipped++
			continue
		}

		wg.Add(1)
		go func(f os.DirEntry) {
			defer wg.Done()

			err := tm.migrateFile(tempDir, f.Name())
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", f.Name(), err)
				progressChan <- false
			} else {
				progressChan <- true
			}
		}(file)
	}

	// Wait for all migrations to complete
	go func() {
		wg.Wait()
		close(errChan)
		close(progressChan)
	}()

	// Collect results
	for success := range progressChan {
		if success {
			stats.Migrated++
		} else {
			stats.Failed++
		}
	}

	// Collect errors
	for err := range errChan {
		stats.Errors = append(stats.Errors, err)
	}

	// Step 5: Clean up or rollback
	if stats.Failed > 0 {
		fmt.Fprintf(os.Stderr, "Migration had failures, performing rollback...\n")
		// Rollback: restore original directory
		os.RemoveAll(oldTodosDir)
		if renameErr := os.Rename(tempDir, oldTodosDir); renameErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to rollback migration after errors: %v\n", renameErr)
		}
		return stats, interrors.NewOperationError("migrate", "todos",
			fmt.Sprintf("migration failed: %d errors", stats.Failed), nil)
	}

	// Success: remove temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "Migration complete: %d migrated, %d failed, %d skipped\n",
		stats.Migrated, stats.Failed, stats.Skipped)

	// Clear path cache after migration
	globalPathCache.Clear()

	return stats, nil
}

// needsMigration checks if there are any flat structure todos
func (tm *TodoManager) needsMigration(todosDir string) (bool, error) {
	files, err := os.ReadDir(todosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Check if any .md files exist in the root todos directory
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			return true, nil
		}
	}

	return false, nil
}

// migrateFile migrates a single todo file to date-based structure
func (tm *TodoManager) migrateFile(sourceDir, filename string) error {
	sourcePath := filepath.Join(sourceDir, filename)

	// Read the file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse to get the started date
	todo, err := tm.parseTodoFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse todo: %w", err)
	}

	// Handle missing started date
	if todo.Started.IsZero() {
		// Use file modification time as fallback
		info, err := os.Stat(sourcePath)
		if err != nil {
			todo.Started = time.Now()
		} else {
			todo.Started = info.ModTime()
		}

		// Update the content with the new started date
		content, err = tm.updateStartedDateInContent(string(content), todo.Started)
		if err != nil {
			return fmt.Errorf("failed to update started date: %w", err)
		}
	}

	// Create destination path
	destPath := GetDateBasedTodoPath(tm.basePath, todo.ID, todo.Started)
	destDir := filepath.Dir(destPath)

	// Create directory structure
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file to new location
	if err := os.WriteFile(destPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update path cache
	globalPathCache.Set(todo.ID, destPath)

	return nil
}

// updateStartedDateInContent updates the started date in the frontmatter
func (tm *TodoManager) updateStartedDateInContent(content string, started time.Time) ([]byte, error) {
	// Parse the frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid todo format")
	}

	// Parse existing frontmatter
	todo, err := tm.parseTodoFile(content)
	if err != nil {
		return nil, err
	}

	// Update started date
	todo.Started = started

	// Rebuild the content
	return tm.buildTodoContent(todo, parts[2])
}

// buildTodoContent rebuilds todo content with updated frontmatter
func (tm *TodoManager) buildTodoContent(todo *Todo, bodyContent string) ([]byte, error) {
	// Marshal the frontmatter
	yamlData, err := yaml.Marshal(todo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal todo: %w", err)
	}

	// Rebuild content
	var contentBuilder strings.Builder
	contentBuilder.WriteString("---\n")
	contentBuilder.Write(yamlData)
	contentBuilder.WriteString("---")
	contentBuilder.WriteString(bodyContent)

	return []byte(contentBuilder.String()), nil
}

// RollbackMigration attempts to restore the original flat structure
func (tm *TodoManager) RollbackMigration() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	todosDir := filepath.Join(tm.basePath, ".claude", "todos")
	backupDir := filepath.Join(tm.basePath, ".claude", "todos_migrating")

	// Check if backup exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return interrors.NewNotFoundError("backup", "todos_migrating")
	}

	// Remove current todos directory
	if err := os.RemoveAll(todosDir); err != nil {
		return interrors.Wrap(err, "failed to remove current todos directory")
	}

	// Restore backup
	if err := os.Rename(backupDir, todosDir); err != nil {
		return interrors.Wrap(err, "failed to restore backup")
	}

	// Clear path cache
	globalPathCache.Clear()

	fmt.Fprintf(os.Stderr, "Migration rolled back successfully\n")
	return nil
}
