package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleTodoTemplate creates a todo from template
func (h *TodoHandlers) HandleTodoTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get managers for the current context
	manager, search, _, templates, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Extract template parameter
	template := request.GetString("template", "")
	if template == "" {
		// List available templates
		templateList, err := templates.ListTemplates()
		if err != nil {
			return HandleError(err), nil
		}

		response := "Available templates:\n"
		for _, t := range templateList {
			response += fmt.Sprintf("- %s\n", t)
		}
		return mcp.NewToolResultText(response), nil
	}

	// Create from template
	task, _ := request.RequireString("task")
	priority := request.GetString("priority", "high")
	todoType := request.GetString("type", "feature")

	todo, err := templates.CreateFromTemplate(template, task, priority, todoType)
	if err != nil {
		return HandleError(err), nil
	}

	// Write and index
	filePath := filepath.Join(manager.GetBasePath(), todo.ID+".md")
	content := fmt.Sprintf("# Task: %s\n\n", todo.Task) // Template content would be added
	err = search.IndexTodo(todo, content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to index todo: %v\n", err)
	}

	return FormatTodoTemplateResponse(todo, filePath, template), nil
}

// HandleTodoLink links related todos
func (h *TodoHandlers) HandleTodoLink(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	parentID, err := request.RequireString("parent_id")
	if err != nil {
		return HandleError(err), nil
	}

	childID, err := request.RequireString("child_id")
	if err != nil {
		return HandleError(err), nil
	}

	linkType := request.GetString("link_type", "parent-child")

	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Create link using the factory's linker
	linker := h.factory.CreateLinker(manager)
	if linker == nil {
		return HandleError(fmt.Errorf("Linking feature not available with current manager")), nil
	}

	err = linker.LinkTodos(parentID, childID, linkType)
	if err != nil {
		return HandleError(err), nil
	}

	return FormatTodoLinkResponse(parentID, childID, linkType), nil
}

// HandleTodoStats generates statistics
func (h *TodoHandlers) HandleTodoStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get managers for the current context
	_, _, stats, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

	// Get time period
	period := request.GetString("period", "all")

	// Calculate stats with period filtering
	statsResult, err := stats.GenerateTodoStatsForPeriod(period)
	if err != nil {
		return HandleError(err), nil
	}

	return FormatTodoStatsResponse(statsResult), nil
}
