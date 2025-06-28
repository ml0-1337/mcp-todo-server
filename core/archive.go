package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"gopkg.in/yaml.v3"
	"strings"
)

// GetQuarter returns the quarter in YYYY-QQ format for a given time
func GetQuarter(t time.Time) string {
	quarter := (int(t.Month()) + 2) / 3
	return fmt.Sprintf("%d-Q%d", t.Year(), quarter)
}

// ArchiveTodo moves a todo to the archive folder and sets completed timestamp
func (tm *TodoManager) ArchiveTodo(id string, quarterOverride string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Construct source file path
	sourcePath := filepath.Join(tm.basePath, id+".md")

	// Check if todo exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("todo not found: %s", id)
	}

	// Read the todo file
	content, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}

	// Parse the file to update completed timestamp
	contentStr := string(content)
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		return fmt.Errorf("invalid markdown format: missing frontmatter delimiters")
	}

	// Parse existing frontmatter
	var frontmatter map[string]interface{}
	err = yaml.Unmarshal([]byte(parts[1]), &frontmatter)
	if err != nil {
		return fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Update completed timestamp
	completedTime := time.Now()
	frontmatter["completed"] = completedTime.Format("2006-01-02 15:04:05")
	frontmatter["status"] = "completed"

	// Marshal back to YAML
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Reconstruct content with updated frontmatter
	updatedContent := "---\n" + string(yamlData) + "---\n" + parts[2]

	// Determine archive quarter
	quarter := quarterOverride
	if quarter == "" {
		quarter = GetQuarter(completedTime)
	}

	// Create archive directory structure
	archiveDir := filepath.Join(filepath.Dir(tm.basePath), "archive", quarter)
	err = os.MkdirAll(archiveDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Write updated content to temp file in archive directory
	tempPath := filepath.Join(archiveDir, id+".md.tmp")
	err = ioutil.WriteFile(tempPath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rename temp file to final location
	finalPath := filepath.Join(archiveDir, id+".md")
	err = os.Rename(tempPath, finalPath)
	if err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Remove original file only after successful archive
	err = os.Remove(sourcePath)
	if err != nil {
		// Archive succeeded but couldn't remove original
		// This is not ideal but data is safe
		return fmt.Errorf("archived successfully but failed to remove original: %w", err)
	}

	return nil
}