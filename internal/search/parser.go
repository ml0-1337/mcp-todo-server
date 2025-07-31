package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/logging"
	"gopkg.in/yaml.v3"
)

// parseTodoFile parses a todo file content and returns a domain.Todo
func parseTodoFile(id, content string) (*domain.Todo, error) {
	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid todo file format: missing frontmatter")
	}

	// Parse YAML frontmatter
	var frontmatter struct {
		TodoID    string   `yaml:"todo_id"`
		Started   string   `yaml:"started"`
		Completed string   `yaml:"completed,omitempty"`
		Status    string   `yaml:"status"`
		Priority  string   `yaml:"priority"`
		Type      string   `yaml:"type"`
		ParentID  string   `yaml:"parent_id,omitempty"`
		Tags      []string `yaml:"tags,omitempty"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Parse started time
	started, err := time.Parse(time.RFC3339, frontmatter.Started)
	if err != nil {
		// Try legacy format
		started, err = time.Parse("2006-01-02 15:04:05", frontmatter.Started)
		if err != nil {
			started = time.Now()
		}
	}

	// Parse completed time if present
	var completed time.Time
	if frontmatter.Completed != "" {
		var err error
		completed, err = time.Parse(time.RFC3339, frontmatter.Completed)
		if err != nil || completed.IsZero() {
			completed, err = time.Parse("2006-01-02 15:04:05", frontmatter.Completed)
			if err != nil {
				// Log warning but continue - completed time is optional
				logging.Warnf("Failed to parse completed time '%s': %v", frontmatter.Completed, err)
			}
		}
	}

	// Extract task from content
	var task string
	contentLines := strings.Split(parts[2], "\n")
	for _, line := range contentLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# Task:") {
			task = strings.TrimSpace(strings.TrimPrefix(line, "# Task:"))
			break
		}
	}

	// Create Todo
	todo := &domain.Todo{
		ID:        frontmatter.TodoID,
		Task:      task,
		Started:   started,
		Completed: completed,
		Status:    frontmatter.Status,
		Priority:  frontmatter.Priority,
		Type:      frontmatter.Type,
		ParentID:  frontmatter.ParentID,
		Tags:      frontmatter.Tags,
		Sections:  make(map[string]*domain.SectionDefinition),
	}

	// Use provided ID if todo ID is empty
	if todo.ID == "" {
		todo.ID = id
	}

	return todo, nil
}
