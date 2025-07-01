package core

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// writeTodo writes a todo to disk
func (tm *TodoManager) writeTodo(todo *Todo) error {
	// Ensure directory exists
	dir := filepath.Join(tm.basePath, ".claude", "todos")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the todo file with frontmatter
	filename := filepath.Join(dir, fmt.Sprintf("%s.md", todo.ID))
	
	// Marshal the frontmatter
	yamlData, err := yaml.Marshal(todo)
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %w", err)
	}

	// Build content with sections
	var contentBuilder strings.Builder
	contentBuilder.WriteString("---\n")
	contentBuilder.Write(yamlData)
	contentBuilder.WriteString("---\n\n")
	contentBuilder.WriteString(fmt.Sprintf("# Task: %s\n\n", todo.Task))

	// Get ordered sections
	orderedSections := GetOrderedSections(todo.Sections)

	// Write each section according to order
	for _, section := range orderedSections {
		contentBuilder.WriteString(fmt.Sprintf("## %s\n\n", section.Definition.Title))
		
		// Note: Section content is not stored in SectionDefinition
		// It's stored separately in the markdown file
	}

	// Write to file
	if err := ioutil.WriteFile(filename, []byte(contentBuilder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	return nil
}

// ReadTodo reads a todo by ID
func (tm *TodoManager) ReadTodo(id string) (*Todo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	filename := filepath.Join(tm.basePath, ".claude", "todos", fmt.Sprintf("%s.md", id))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("todo not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read todo: %w", err)
	}

	return tm.parseTodoFile(string(content))
}

// parseTodoFile parses a todo file content
// ParseTodoFileContent parses a todo from markdown content
func (tm *TodoManager) ParseTodoFileContent(id, content string) (*Todo, error) {
	todo, err := tm.parseTodoFile(content)
	if err != nil {
		return nil, err
	}
	if todo.ID == "" {
		todo.ID = id
	}
	return todo, nil
}

func (tm *TodoManager) parseTodoFile(content string) (*Todo, error) {
	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid todo file format: missing frontmatter")
	}

	// Parse frontmatter
	var todo Todo
	if err := yaml.Unmarshal([]byte(parts[1]), &todo); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract task from heading
	todo.Task = extractTask(parts[2])

	// Fix section definitions with improved validation
	if todo.Sections == nil {
		todo.Sections = make(map[string]*SectionDefinition)
	}

	// Validate and fix section metadata
	for key, section := range todo.Sections {
		if section == nil {
			todo.Sections[key] = &SectionDefinition{
				Title: key,
				Order: 100,
			}
			continue
		}
		
		// Ensure title is set
		if section.Title == "" {
			section.Title = key
		}
		
		// Initialize metadata if nil
		if section.Metadata == nil {
			section.Metadata = make(map[string]interface{})
		}
		
		// Handle started field - allow both time.Time and string
		if started, ok := section.Metadata["started"]; ok {
			switch v := started.(type) {
			case string:
				// Try to parse as timestamp
				if t, err := parseTimestamp(v); err == nil {
					section.Metadata["started"] = t
				}
			case time.Time:
				// Already a time.Time, keep as is
			default:
				// Unknown type, remove it
				delete(section.Metadata, "started")
			}
		}
	}

	return &todo, nil
}

// SaveTodo saves a todo (alias for writeTodo with locking)
func (tm *TodoManager) SaveTodo(todo *Todo) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return tm.writeTodo(todo)
}

// ReadTodoContent reads the raw content of a todo file
func (tm *TodoManager) ReadTodoContent(id string) (string, error) {
	filename := filepath.Join(tm.basePath, ".claude", "todos", fmt.Sprintf("%s.md", id))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read todo content: %w", err)
	}
	return string(content), nil
}

// ReadTodoWithContent reads both the parsed todo and raw content
func (tm *TodoManager) ReadTodoWithContent(id string) (*Todo, string, error) {
	// First parse the todo
	todo, err := tm.ReadTodo(id)
	if err != nil {
		return nil, "", err
	}

	// Then read the raw content
	content, err := tm.ReadTodoContent(id)
	if err != nil {
		return nil, "", err
	}

	return todo, content, nil
}