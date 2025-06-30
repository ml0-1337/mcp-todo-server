package handlers

import (
	"context"
	"fmt"

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

	// Get the manager for this context
	manager := h.getManagerForContext(ctx)

	// Handle single todo read
	if params.ID != "" {
		// Regular single todo read
		todo, _, err := manager.ReadTodoWithContent(params.ID)
		if err != nil {
			return HandleError(err), nil
		}

		// Create response based on format
		return FormatTodoReadResponse([]*core.Todo{todo}, params.Format, true), nil
	}

	// Handle list todos
	todos, err := manager.ListTodos(
		params.Filter.Status,
		params.Filter.Priority,
		params.Filter.Days,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
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
	results, err := h.search.SearchTodos(params.Query, filterMap, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Create response
	return FormatTodoSearchResponse(results), nil
}