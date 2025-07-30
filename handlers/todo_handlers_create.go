package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

	// Get managers for the current context
	manager, search, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to get context-aware managers")
	}

	// Check if we need to use a template based on type
	var templateContent string
	if params.Type == "prd" || params.Template != "" {
		// Determine which template to use
		templateName := params.Template
		if templateName == "" && params.Type == "prd" {
			templateName = "prd"
		}

		// Get the template manager
		basePath := manager.GetBasePath()
		templatesDir := filepath.Join(basePath, "templates")
		templateManager := core.NewTemplateManager(templatesDir)

		// Load the template
		template, err := templateManager.LoadTemplate(templateName)
		if err != nil {
			// If template not found, log warning and continue with default sections
			fmt.Fprintf(os.Stderr, "Warning: failed to load template '%s': %v\n", templateName, err)
		} else {
			// Execute the template with variables
			vars := map[string]interface{}{
				"feature_name": params.Task,
				"author":       "Claude Code",
				"date":         time.Now().Format("2006-01-02"),
			}
			content, err := templateManager.ExecuteTemplate(template, vars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to execute template: %v\n", err)
			} else {
				templateContent = content
			}
		}
	}

	// Create todo with template content if available
	var todo *core.Todo
	if concreteManager, ok := manager.(*core.TodoManager); ok && templateContent != "" {
		todo, err = concreteManager.CreateTodoWithTemplate(params.Task, params.Priority, params.Type, templateContent)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to create todo with template")
		}
	} else {
		todo, err = manager.CreateTodo(params.Task, params.Priority, params.Type)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to create todo")
		}
	}

	// Handle parent-child relationship if parent_id is provided
	if params.ParentID != "" {
		// Update the todo with parent_id
		todo.ParentID = params.ParentID
		err = manager.SaveTodo(todo)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to update todo with parent_id")
		}

		// Create the link using the TodoLinker
		// Need to use a concrete TodoManager for linking
		if concreteManager, ok := manager.(*core.TodoManager); ok {
			linker := core.NewTodoLinker(concreteManager)
			err = linker.LinkTodos(params.ParentID, todo.ID, "parent-child")
			if err != nil {
				// Log but don't fail
				fmt.Fprintf(os.Stderr, "Warning: failed to create parent-child link: %v\n", err)
			}
		}
	}

	// Index the todo for search
	if search != nil {
		content, _ := manager.ReadTodoContent(todo.ID)
		if err := search.IndexTodo(todo, content); err != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to index todo %s: %v\n", todo.ID, err)
		}
	}

	// Create response with parent context if applicable
	filePath := filepath.Join(manager.GetBasePath(), todo.ID+".md")
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

	// Get managers for the current context
	manager, search, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to get context-aware managers")
	}

	// Create parent todo first
	parentTodo, err := manager.CreateTodo(
		params.Parent.Task,
		params.Parent.Priority,
		params.Parent.Type,
	)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to create parent todo")
	}

	// Index parent todo
	if search != nil {
		content, _ := manager.ReadTodoContent(parentTodo.ID)
		if err := search.IndexTodo(parentTodo, content); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index parent todo %s: %v\n", parentTodo.ID, err)
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

		childTodo, err := manager.CreateTodo(
			childParam.Task,
			childParam.Priority,
			childType,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create child todo '%s': %v\n", childParam.Task, err)
			continue
		}

		// Set parent ID
		childTodo.ParentID = parentTodo.ID
		if err := manager.SaveTodo(childTodo); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update child todo with parent_id: %v\n", err)
		}

		// Create parent-child link
		if concreteManager, ok := manager.(*core.TodoManager); ok {
			linker := core.NewTodoLinker(concreteManager)
			if err := linker.LinkTodos(parentTodo.ID, childTodo.ID, "parent-child"); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create parent-child link: %v\n", err)
			}
		}

		// Index child todo
		if search != nil {
			content, _ := manager.ReadTodoContent(childTodo.ID)
			if err := search.IndexTodo(childTodo, content); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to index child todo %s: %v\n", childTodo.ID, err)
			}
		}

		childTodos = append(childTodos, childTodo)
	}

	// Create response
	return FormatTodoCreateMultiResponse(parentTodo, childTodos), nil
}
