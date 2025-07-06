package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// HandleTodoRead handles the todo_read tool
func (h *TodoHandlers) HandleTodoRead(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoReadParams(request)
	if err != nil {
		return nil, err
	}

	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Handle single todo read
	if params.ID != "" {
		// Regular single todo read
		todo, content, err := manager.ReadTodoWithContent(params.ID)
		if err != nil {
			return HandleError(err), nil
		}

		// Create response based on format
		if params.Format == "full" {
			return formatSingleTodoWithContent(todo, content, params.Format), nil
		}
		return FormatTodoReadResponse([]*core.Todo{todo}, params.Format, true), nil
	}

	// Handle list todos
	fmt.Fprintf(os.Stderr, "HandleTodoRead: Listing todos with status=%s, priority=%s, days=%d\n", 
		params.Filter.Status, params.Filter.Priority, params.Filter.Days)
	todos, err := manager.ListTodos(
		params.Filter.Status,
		params.Filter.Priority,
		params.Filter.Days,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}
	fmt.Fprintf(os.Stderr, "HandleTodoRead: Found %d todos\n", len(todos))

	// For full format with multiple todos, we need to get content for each
	if params.Format == "full" && len(todos) > 0 {
		contents := make(map[string]string)
		for _, todo := range todos {
			content, err := manager.ReadTodoContent(todo.ID)
			if err != nil {
				// Skip todos we can't read content for
				continue
			}
			contents[todo.ID] = content
		}
		return formatTodosFullWithContent(todos, contents), nil
	}

	// Create response
	return FormatTodoReadResponse(todos, params.Format, false), nil
}

// HandleTodoSearch handles the todo_search tool
func (h *TodoHandlers) HandleTodoSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoSearchParams(request)
	if err != nil {
		return nil, err
	}

	// Get managers for the current context
	_, search, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Convert filters to map
	filterMap := make(map[string]string)
	if params.Filters.Status != "" {
		filterMap["status"] = params.Filters.Status
	}
	if params.Filters.DateFrom != "" {
		filterMap["date_from"] = params.Filters.DateFrom
	}
	if params.Filters.DateTo != "" {
		filterMap["date_to"] = params.Filters.DateTo
	}

	// Perform search
	results, err := search.SearchTodos(params.Query, filterMap, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Create response
	return FormatTodoSearchResponse(results), nil
}