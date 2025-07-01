package handlers

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// HandleTodoCreate handles the todo_create tool
func (h *TodoHandlers) HandleTodoCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoCreateParams(request)
	if err != nil {
		// Return validation errors as tool results, not Go errors
		return HandleError(err), nil
	}

	// Create todo
	todo, err := h.manager.CreateTodo(params.Task, params.Priority, params.Type)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to create todo")
	}

	// Handle parent-child relationship if parent_id is provided
	if params.ParentID != "" {
		// Update the todo with parent_id
		todo.ParentID = params.ParentID
		err = h.manager.SaveTodo(todo)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to update todo with parent_id")
		}

		// Create the link using the TodoLinker
		if h.baseManager != nil {
			linker := core.NewTodoLinker(h.baseManager)
			err = linker.LinkTodos(params.ParentID, todo.ID, "parent-child")
			if err != nil {
				// Log but don't fail
				log.Printf("Warning: failed to create parent-child link: %v", err)
			}
		}
	}

	// Index the todo for search
	if h.search != nil {
		content, _ := h.manager.ReadTodoContent(todo.ID)
		if err := h.search.IndexTodo(todo, content); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to index todo %s: %v\n", todo.ID, err)
		}
	}

	// Create response with parent context if applicable
	filePath := filepath.Join(h.manager.GetBasePath(), todo.ID+".md")
	// Create the tool result with formatted response
	return FormatTodoCreateResponse(todo, filePath), nil
}

// HandleTodoCreateMulti handles the todo_create_multi tool
func (h *TodoHandlers) HandleTodoCreateMulti(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoCreateMultiParams(request)
	if err != nil {
		return nil, err
	}

	// Create parent todo first
	parentTodo, err := h.manager.CreateTodo(
		params.Parent.Task,
		params.Parent.Priority,
		params.Parent.Type,
	)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to create parent todo")
	}

	// Index parent todo
	if h.search != nil {
		content, _ := h.manager.ReadTodoContent(parentTodo.ID)
		if err := h.search.IndexTodo(parentTodo, content); err != nil {
			log.Printf("Warning: failed to index parent todo %s: %v", parentTodo.ID, err)
		}
	}

	// Create child todos
	var childTodos []*core.Todo
	for _, childParam := range params.Children {
		// Set default type if not specified
		childType := childParam.Type
		if childType == "" {
			childType = "phase"
		}

		childTodo, err := h.manager.CreateTodo(
			childParam.Task,
			childParam.Priority,
			childType,
		)
		if err != nil {
			log.Printf("Failed to create child todo '%s': %v", childParam.Task, err)
			continue
		}

		// Set parent ID
		childTodo.ParentID = parentTodo.ID
		if err := h.manager.SaveTodo(childTodo); err != nil {
			log.Printf("Failed to update child todo with parent_id: %v", err)
		}

		// Create parent-child link
		if h.baseManager != nil {
			linker := core.NewTodoLinker(h.baseManager)
			if err := linker.LinkTodos(parentTodo.ID, childTodo.ID, "parent-child"); err != nil {
				log.Printf("Warning: failed to create parent-child link: %v", err)
			}
		}

		// Index child todo
		if h.search != nil {
			content, _ := h.manager.ReadTodoContent(childTodo.ID)
			if err := h.search.IndexTodo(childTodo, content); err != nil {
				log.Printf("Warning: failed to index child todo %s: %v", childTodo.ID, err)
			}
		}

		childTodos = append(childTodos, childTodo)
	}

	// Create response
	return FormatTodoCreateMultiResponse(parentTodo, childTodos), nil
}