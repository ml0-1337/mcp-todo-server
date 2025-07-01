package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	
	interrors "github.com/user/mcp-todo-server/internal/errors"
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
		if interrors.IsNotFound(err) {
			return interrors.NewNotFoundError("parent todo", parentID)
		}
		return interrors.Wrap(err, "failed to read parent todo")
	}

	_, err = tl.manager.ReadTodo(childID)
	if err != nil {
		if interrors.IsNotFound(err) {
			return interrors.NewNotFoundError("child todo", childID)
		}
		return interrors.Wrap(err, "failed to read child todo")
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
	return interrors.NewValidationError("linkType", linkType, "unsupported link type")
}

// CreateTodoWithParent creates a new todo with a parent reference
func (tm *TodoManager) CreateTodoWithParent(task, priority, todoType, parentID string) (*Todo, error) {
	// Validate parent exists
	if parentID != "" {
		_, err := tm.ReadTodo(parentID)
		if err != nil {
			if interrors.IsNotFound(err) {
				return nil, interrors.NewNotFoundError("parent todo", parentID)
			}
			return nil, interrors.Wrap(err, "failed to validate parent")
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
			os.Remove(filepath.Join(tm.basePath, ".claude", "todos", todo.ID+".md"))
			return nil, interrors.Wrap(err, "failed to set parent_id")
		}
		todo.ParentID = parentID
	}

	return todo, nil
}

// GetChildren returns all todos that have the given parent_id
func (tm *TodoManager) GetChildren(parentID string) ([]*Todo, error) {
	// Read all todos in the directory
	todosDir := filepath.Join(tm.basePath, ".claude", "todos")
	files, err := ioutil.ReadDir(todosDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No todos directory exists yet
			return []*Todo{}, nil
		}
		return nil, interrors.Wrap(err, "failed to read todos directory")
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
