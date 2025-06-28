package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// TodoLinker handles linking related todos
type TodoLinker struct {
	manager *TodoManager
}

// NewTodoLinker creates a new todo linker
func NewTodoLinker(manager *TodoManager) *TodoLinker {
	return &TodoLinker{
		manager: manager,
	}
}

// LinkTodos creates a link between two todos
func (tl *TodoLinker) LinkTodos(parentID, childID, linkType string) error {
	// Validate both todos exist
	_, err := tl.manager.ReadTodo(parentID)
	if err != nil {
		return fmt.Errorf("parent todo not found: %w", err)
	}
	
	_, err = tl.manager.ReadTodo(childID)
	if err != nil {
		return fmt.Errorf("child todo not found: %w", err)
	}
	
	// For parent-child link, update the child's parent_id
	if linkType == "parent-child" {
		metadata := map[string]string{
			"parent_id": parentID,
		}
		return tl.manager.UpdateTodo(childID, "", "", "", metadata)
	}
	
	// For other link types, could store in a separate links file or in metadata
	// For now, only support parent-child
	return fmt.Errorf("unsupported link type: %s", linkType)
}

// CreateTodoWithParent creates a new todo with a parent reference
func (tm *TodoManager) CreateTodoWithParent(task, priority, todoType, parentID string) (*Todo, error) {
	// Validate parent exists
	if parentID != "" {
		_, err := tm.ReadTodo(parentID)
		if err != nil {
			if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("parent todo not found: %s", parentID)
			}
			return nil, fmt.Errorf("failed to validate parent: %w", err)
		}
	}
	
	// Create the todo
	todo, err := tm.CreateTodo(task, priority, todoType)
	if err != nil {
		return nil, err
	}
	
	// Update with parent_id
	if parentID != "" {
		metadata := map[string]string{
			"parent_id": parentID,
		}
		err = tm.UpdateTodo(todo.ID, "", "", "", metadata)
		if err != nil {
			// Clean up the created todo on failure
			os.Remove(filepath.Join(tm.basePath, todo.ID + ".md"))
			return nil, fmt.Errorf("failed to set parent_id: %w", err)
		}
		todo.ParentID = parentID
	}
	
	return todo, nil
}

// GetChildren returns all todos that have the given parent_id
func (tm *TodoManager) GetChildren(parentID string) ([]*Todo, error) {
	// Read all todos in the directory
	files, err := ioutil.ReadDir(tm.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read todos directory: %w", err)
	}
	
	var children []*Todo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			// Extract ID from filename
			todoID := strings.TrimSuffix(file.Name(), ".md")
			
			// Read the todo
			todo, err := tm.ReadTodo(todoID)
			if err != nil {
				// Skip files that can't be read as todos
				continue
			}
			
			// Check if it's a child of the requested parent
			if todo.ParentID == parentID {
				children = append(children, todo)
			}
		}
	}
	
	return children, nil
}