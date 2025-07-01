package handlers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// HandleTodoTemplate creates a todo from template
func (h *TodoHandlers) HandleTodoTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract template parameter
	template := request.GetString("template", "")
	if template == "" {
		// List available templates
		templates, err := h.templates.ListTemplates()
		if err != nil {
			return HandleError(err), nil
		}

		response := "Available templates:\n"
		for _, t := range templates {
			response += fmt.Sprintf("- %s\n", t)
		}
		return mcp.NewToolResultText(response), nil
	}

	// Create from template
	task, _ := request.RequireString("task")
	priority := request.GetString("priority", "high")
	todoType := request.GetString("type", "feature")

	todo, err := h.templates.CreateFromTemplate(template, task, priority, todoType)
	if err != nil {
		return HandleError(err), nil
	}

	// Write and index
	filePath := filepath.Join(h.manager.GetBasePath(), todo.ID+".md")
	content := fmt.Sprintf("# Task: %s\n\n", todo.Task) // Template content would be added
	err = h.search.IndexTodo(todo, content)
	if err != nil {
		fmt.Printf("Warning: failed to index todo: %v\n", err)
	}

	return FormatTodoTemplateResponse(todo, filePath), nil
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

	// Create link
	if h.baseManager == nil {
		return HandleError(fmt.Errorf("Linking feature not available")), nil
	}
	linker := core.NewTodoLinker(h.baseManager)
	err = linker.LinkTodos(parentID, childID, linkType)
	if err != nil {
		return HandleError(err), nil
	}

	return FormatTodoLinkResponse(parentID, childID, linkType), nil
}

// HandleTodoStats generates statistics
func (h *TodoHandlers) HandleTodoStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get time period
	period := request.GetString("period", "all")

	// Calculate stats with period filtering
	stats, err := h.stats.GenerateTodoStatsForPeriod(period)
	if err != nil {
		return HandleError(err), nil
	}

	return FormatTodoStatsResponse(stats), nil
}