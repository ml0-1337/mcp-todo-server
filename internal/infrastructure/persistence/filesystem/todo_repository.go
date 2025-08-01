// Package filesystem provides file system based persistence for todo items.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/domain/repository"
	"gopkg.in/yaml.v3"
)

// todoYAML represents the YAML structure for todo persistence
type todoYAML struct {
	ID        string                               `yaml:"todo_id"`
	Started   time.Time                            `yaml:"started"`
	Completed time.Time                            `yaml:"completed,omitempty"`
	Status    string                               `yaml:"status"`
	Priority  string                               `yaml:"priority"`
	Type      string                               `yaml:"type"`
	ParentID  string                               `yaml:"parent_id,omitempty"`
	Tags      []string                             `yaml:"tags,omitempty"`
	Sections  map[string]*domain.SectionDefinition `yaml:"sections,omitempty"`
}

// TodoRepository implements the repository interface using filesystem
type TodoRepository struct {
	basePath string
	mu       sync.RWMutex
}

// NewTodoRepository creates a new filesystem-based todo repository
func NewTodoRepository(basePath string) *TodoRepository {
	return &TodoRepository{
		basePath: basePath,
	}
}

// Save creates or updates a todo
func (r *TodoRepository) Save(_ context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := todo.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert domain todo to YAML structure
	yamlTodo := todoYAML{
		ID:        todo.ID,
		Started:   todo.Started,
		Completed: todo.Completed,
		Status:    todo.Status,
		Priority:  todo.Priority,
		Type:      todo.Type,
		ParentID:  todo.ParentID,
		Tags:      todo.Tags,
		Sections:  todo.Sections,
	}

	// Create YAML frontmatter
	frontmatter, err := yaml.Marshal(yamlTodo)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Construct the full content
	content := fmt.Sprintf("---\n%s---\n\n# %s\n", string(frontmatter), todo.Task)

	// Add sections as markdown content
	if len(todo.Sections) > 0 {
		// Sort sections by order
		var sectionKeys []string
		for key := range todo.Sections {
			sectionKeys = append(sectionKeys, key)
		}
		sort.Slice(sectionKeys, func(i, j int) bool {
			return todo.Sections[sectionKeys[i]].Order < todo.Sections[sectionKeys[j]].Order
		})

		// Add each section
		for _, key := range sectionKeys {
			section := todo.Sections[key]
			content += fmt.Sprintf("\n## %s\n\n%s\n", section.Title, section.Content)
		}
	}

	// Write to file
	filePath := r.todoPath(todo.ID)
	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FindByID retrieves a todo by its ID
func (r *TodoRepository) FindByID(_ context.Context, id string) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filePath := r.todoPath(id)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return r.parseTodo(content)
}

// FindByIDWithContent retrieves a todo and its full content
func (r *TodoRepository) FindByIDWithContent(_ context.Context, id string) (*domain.Todo, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filePath := r.todoPath(id)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", domain.ErrTodoNotFound
		}
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	todo, err := r.parseTodo(content)
	if err != nil {
		return nil, "", err
	}

	return todo, string(content), nil
}

// List retrieves todos based on filters
func (r *TodoRepository) List(ctx context.Context, filters repository.ListFilters) ([]*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var todos []*domain.Todo

	// Check if directory exists
	if _, err := os.Stat(r.basePath); os.IsNotExist(err) {
		// Return empty list for non-existent directory
		return todos, nil
	}

	// Walk through all todo files
	err := filepath.Walk(r.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.md files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip files in archive directory
		if strings.Contains(path, "/archive/") {
			return nil
		}

		// Read and parse todo
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		todo, err := r.parseTodo(content)
		if err != nil {
			// Skip invalid todos
			return nil
		}

		// Apply filters
		if filters.Status != "" && todo.Status != filters.Status {
			return nil
		}

		if filters.Priority != "" && todo.Priority != filters.Priority {
			return nil
		}

		if filters.Days > 0 {
			cutoff := time.Now().AddDate(0, 0, -filters.Days)
			if todo.Started.Before(cutoff) {
				return nil
			}
		}

		if filters.ParentID != "" && todo.ParentID != filters.ParentID {
			return nil
		}

		todos = append(todos, todo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return todos, nil
}

// Delete removes a todo
func (r *TodoRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	filePath := r.todoPath(id)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return domain.ErrTodoNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Archive moves a todo to archive
func (r *TodoRepository) Archive(_ context.Context, id string, archivePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	srcPath := r.todoPath(id)
	dstPath := filepath.Join(r.basePath, "archive", archivePath, fmt.Sprintf("%s.md", id))

	// Create archive directory
	if err := os.MkdirAll(filepath.Dir(dstPath), 0750); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Move file
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move file to archive: %w", err)
	}

	return nil
}

// UpdateContent updates a specific section of a todo
func (r *TodoRepository) UpdateContent(_ context.Context, id string, section string, content string) error {
	// This is a simplified implementation
	// In practice, this would parse the file, update the section, and rewrite
	return fmt.Errorf("not implemented")
}

// GetContent retrieves the full content of a todo
func (r *TodoRepository) GetContent(_ context.Context, id string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filePath := r.todoPath(id)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", domain.ErrTodoNotFound
		}
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// todoPath returns the file path for a todo ID
func (r *TodoRepository) todoPath(id string) string {
	return filepath.Join(r.basePath, fmt.Sprintf("%s.md", id))
}

// parseTodo parses todo content into domain model
func (r *TodoRepository) parseTodo(content []byte) (*domain.Todo, error) {
	// Extract frontmatter
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, fmt.Errorf("no frontmatter found")
	}

	endIndex := strings.Index(contentStr[4:], "\n---\n")
	if endIndex == -1 {
		return nil, fmt.Errorf("invalid frontmatter")
	}

	frontmatter := contentStr[4 : endIndex+4]
	remaining := contentStr[endIndex+9:]

	// Parse YAML
	var yamlTodo todoYAML
	if err := yaml.Unmarshal([]byte(frontmatter), &yamlTodo); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract task from heading
	task := ""
	lines := strings.Split(remaining, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			task = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Convert to domain model
	todo := &domain.Todo{
		ID:        yamlTodo.ID,
		Task:      task,
		Started:   yamlTodo.Started,
		Completed: yamlTodo.Completed,
		Status:    yamlTodo.Status,
		Priority:  yamlTodo.Priority,
		Type:      yamlTodo.Type,
		ParentID:  yamlTodo.ParentID,
		Tags:      yamlTodo.Tags,
		Sections:  yamlTodo.Sections,
	}

	return todo, nil
}
