package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// HandleTodoArchive handles the todo_archive tool
func (h *TodoHandlers) HandleTodoArchive(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoArchiveParams(request)
	if err != nil {
		return nil, err
	}

	// Get managers for the current context
	manager, search, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Read todo BEFORE archiving to get its metadata
	todo, readErr := manager.ReadTodo(params.ID)

	// Archive todo
	err = manager.ArchiveTodo(params.ID)
	if err != nil {
		return HandleError(err), nil
	}

	// Construct archive path
	var archivePath string
	if readErr == nil && todo != nil {
		// Use the todo's started date for archive path (matches ArchiveTodo behavior)
		dayPath := core.GetDailyPath(todo.Started)
		archivePath = filepath.Join(".claude", "archive", dayPath, params.ID+".md")
	} else {
		// Fallback path when we couldn't read the todo
		// This might happen with timestamp parsing errors
		archivePath = filepath.Join(".claude", "archive", params.ID+".md")
	}

	// Remove from search index if available
	if search != nil {
		err = search.DeleteTodo(params.ID)
		if err != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to remove from search index: %v\n", err)
		}
	}

	// Get todo type for prompts (default to empty string if not available)
	todoType := ""
	if todo != nil {
		todoType = todo.Type
	}
	
	return FormatTodoArchiveResponse(params.ID, archivePath, todoType), nil
}

// HandleTodoClean performs cleanup operations
func (h *TodoHandlers) HandleTodoClean(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Get operation type
	operation := request.GetString("operation", "archive_old")

	switch operation {
	case "archive_old":
		// Archive todos older than specified days
		days := request.GetInt("days", 90)
		count, err := manager.ArchiveOldTodos(days)
		if err != nil {
			return HandleError(err), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Archived %d todos older than %d days", count, days)), nil

	case "find_duplicates":
		// Find duplicate todos
		duplicates, err := manager.FindDuplicateTodos()
		if err != nil {
			return HandleError(err), nil
		}

		if len(duplicates) == 0 {
			return mcp.NewToolResultText("No duplicate todos found"), nil
		}

		response := fmt.Sprintf("Found %d sets of duplicates:\n", len(duplicates))
		for _, group := range duplicates {
			response += fmt.Sprintf("\n- %s\n", group[0])
			for _, dup := range group[1:] {
				response += fmt.Sprintf("  - %s\n", dup)
			}
		}
		return mcp.NewToolResultText(response), nil

	default:
		return HandleError(fmt.Errorf("unknown operation: %s", operation)), nil
	}
}