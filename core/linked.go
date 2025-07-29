package core

import (
	"os"
	"path/filepath"
	
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
	// Use ListTodos to get all todos and filter by parent_id
	allTodos, err := tm.ListTodos("", "", 0)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to list todos")
	}

	var children []*Todo
	for _, todo := range allTodos {
		if todo.ParentID == parentID {
			children = append(children, todo)
		}
	}

	return children, nil
}
