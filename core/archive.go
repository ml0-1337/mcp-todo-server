package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// GetDailyPath returns the daily archive path YYYY/MM/DD for a given time
func GetDailyPath(t time.Time) string {
	return filepath.Join(t.Format("2006"), t.Format("01"), t.Format("02"))
}

// ArchiveTodo moves a todo to the archive folder and sets completed timestamp
func (tm *TodoManager) ArchiveTodo(id string) error {
	// Check if todo has active children first (before locking)
	children, err := tm.GetChildren(id)
	if err != nil {
		return interrors.Wrap(err, "failed to check for children")
	}

	// Check for incomplete children
	incompleteChildren := 0
	for _, child := range children {
		if child.Status != "completed" {
			incompleteChildren++
		}
	}

	if incompleteChildren > 0 {
		return interrors.NewOperationError("archive", "todo", fmt.Sprintf("cannot archive todo %s: has %d active children", id, incompleteChildren), nil)
	}

	// Now lock for the actual archive operation
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Construct source file path using ResolveTodoPath
	sourcePath, err := ResolveTodoPath(tm.basePath, id)
	if err != nil {
		if os.IsNotExist(err) {
			return interrors.NewNotFoundError("todo", id)
		}
		return interrors.Wrap(err, "failed to resolve todo path")
	}

	// Check if todo exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return interrors.NewNotFoundError("todo", id)
	}

	// Read the todo file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return interrors.Wrap(err, "failed to read todo file")
	}

	// Parse the file to update completed timestamp
	contentStr := string(content)
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		return interrors.NewValidationError("content", contentStr, "invalid markdown format: missing frontmatter delimiters")
	}

	// Parse existing frontmatter
	var frontmatter map[string]interface{}
	err = yaml.Unmarshal([]byte(parts[1]), &frontmatter)
	if err != nil {
		return interrors.Wrap(err, "failed to parse YAML frontmatter")
	}

	// First, determine archive path and create directory
	completedTime := time.Now()

	// Parse the todo to get the started date
	todo, err := tm.parseTodoFile(string(content))
	if err != nil {
		return interrors.Wrap(err, "failed to parse todo for archive date")
	}

	// Use the todo's started date for archiving
	archivePath := GetDailyPath(todo.Started)

	// Create archive directory structure within .claude
	archiveDir := filepath.Join(tm.basePath, ".claude", "archive", archivePath)
	err = os.MkdirAll(archiveDir, 0750)
	if err != nil {
		return interrors.NewOperationError("create", "archive directory", "failed to create archive directory", err)
	}

	// Now update the frontmatter
	frontmatter["completed"] = completedTime.Format(time.RFC3339)
	frontmatter["status"] = "completed"

	// Marshal back to YAML
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return interrors.Wrap(err, "failed to marshal YAML")
	}

	// Reconstruct content with updated frontmatter
	updatedContent := "---\n" + string(yamlData) + "---\n" + parts[2]

	// Write updated content to temp file in archive directory
	tempPath := filepath.Join(archiveDir, id+".md.tmp")
	err = os.WriteFile(tempPath, []byte(updatedContent), 0600)
	if err != nil {
		return interrors.NewOperationError("write", "temp archive file", "failed to write temp file", err)
	}

	// Rename temp file to final location
	finalPath := filepath.Join(archiveDir, id+".md")
	err = os.Rename(tempPath, finalPath)
	if err != nil {
		os.Remove(tempPath) // Clean up temp file
		return interrors.NewOperationError("rename", "archive file", "failed to finalize archive", err)
	}

	// Remove original file only after successful archive
	err = os.Remove(sourcePath)
	if err != nil {
		// Archive succeeded but couldn't remove original
		// Try to clean up the archive to maintain consistency
		os.Remove(finalPath)
		return interrors.NewOperationError("remove", "original todo file", "failed to remove original file after archive", err)
	}

	return nil
}

// BulkResult represents the result of a bulk operation on a single item
type BulkResult struct {
	ID      string
	Success bool
	Error   error
}

// BulkArchiveTodos archives multiple todos and returns per-item results
func (tm *TodoManager) BulkArchiveTodos(ids []string) []BulkResult {
	results := make([]BulkResult, len(ids))

	// Process each todo independently
	for i, id := range ids {
		results[i].ID = id

		// Attempt to archive
		err := tm.ArchiveTodo(id)
		if err != nil {
			results[i].Success = false
			results[i].Error = err
		} else {
			results[i].Success = true
		}
	}

	return results
}

// isArchived checks if a todo is already archived
func isArchived(basePath, id string) bool {
	// Try to resolve the todo path
	todoPath, err := ResolveTodoPath(basePath, id)
	if err != nil {
		// If we can't find it in todos directory, assume it's archived
		return true
	}

	// Check if the resolved path is in the archive directory
	archivePath := filepath.Join(basePath, ".claude", "archive")
	return strings.HasPrefix(todoPath, archivePath)
}

// ArchiveTodoWithCascade archives a todo and optionally its children
func (tm *TodoManager) ArchiveTodoWithCascade(id string, cascade bool) error {
	if !cascade {
		// Just regular archive
		return tm.ArchiveTodo(id)
	}

	// Get all children first
	children, err := tm.GetChildren(id)
	if err != nil {
		return interrors.Wrap(err, "failed to get children")
	}

	// Archive all completed children first
	for _, child := range children {
		if child.Status == "completed" && !isArchived(tm.basePath, child.ID) {
			err := tm.ArchiveTodo(child.ID)
			if err != nil {
				return interrors.Wrapf(err, "failed to archive child %s", child.ID)
			}
		}
	}

	// Now archive the parent
	return tm.ArchiveTodo(id)
}
