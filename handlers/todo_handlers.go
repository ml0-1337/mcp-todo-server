package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// TodoHandlers contains handlers for all todo operations
type TodoHandlers struct {
	manager   *core.TodoManager
	search    *core.SearchEngine
	stats     *core.StatsEngine
	templates *core.TemplateManager
}

// NewTodoHandlers creates new todo handlers with dependencies
func NewTodoHandlers(todoPath, templatePath string) (*TodoHandlers, error) {
	// Create todo manager
	manager := core.NewTodoManager(todoPath)
	
	// Create search engine
	indexPath := filepath.Join(todoPath, "..", "index", "todos.bleve")
	search, err := core.NewSearchEngine(indexPath, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}
	
	// Create stats engine
	stats := core.NewStatsEngine(manager)
	
	// Create template manager
	templates := core.NewTemplateManager(templatePath)
	
	return &TodoHandlers{
		manager:   manager,
		search:    search,
		stats:     stats,
		templates: templates,
	}, nil
}

// Close cleans up resources
func (h *TodoHandlers) Close() error {
	if h.search != nil {
		return h.search.Close()
	}
	return nil
}

// HandleTodoCreate creates a new todo
func (h *TodoHandlers) HandleTodoCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoCreateParams(request)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Check if using template
	if params.Template != "" {
		// Load and apply template
		todo, err := h.templates.CreateFromTemplate(params.Template, params.Task, params.Priority, params.Type)
		if err != nil {
			return HandleError(fmt.Errorf("template error: %w", err)), nil
		}
		
		// Write todo file
		filePath := filepath.Join(h.manager.GetBasePath(), todo.ID+".md")
		return FormatTodoTemplateResponse(todo, filePath), nil
	}
	
	// Create todo
	todo, err := h.manager.CreateTodo(params.Task, params.Priority, params.Type)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Set parent ID if provided
	if params.ParentID != "" {
		todo.ParentID = params.ParentID
		// Update the todo file with parent ID
		err = h.manager.UpdateTodo(todo.ID, "", "", "", map[string]string{"parent_id": params.ParentID})
		if err != nil {
			// Log but don't fail the creation
			fmt.Printf("Warning: failed to set parent_id: %v\n", err)
		}
	}
	
	// Index in search
	content := fmt.Sprintf("# Task: %s\n\n## Findings & Research\n\n## Test Strategy\n\n", todo.Task)
	err = h.search.IndexTodo(todo, content)
	if err != nil {
		// Log but don't fail the creation
		fmt.Printf("Warning: failed to index todo: %v\n", err)
	}
	
	// Return response
	filePath := filepath.Join(h.manager.GetBasePath(), todo.ID+".md")
	return FormatTodoCreateResponse(todo, filePath), nil
}

// HandleTodoRead reads one or more todos
func (h *TodoHandlers) HandleTodoRead(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoReadParams(request)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Single todo read
	if params.ID != "" {
		todo, err := h.manager.ReadTodo(params.ID)
		if err != nil {
			return HandleError(err), nil
		}
		return FormatTodoReadResponse([]*core.Todo{todo}, params.Format, true), nil
	}
	
	// List todos with filters
	todos, err := h.manager.ListTodos(params.Filter.Status, params.Filter.Priority, params.Filter.Days)
	if err != nil {
		return HandleError(err), nil
	}
	
	return FormatTodoReadResponse(todos, params.Format, false), nil
}

// HandleTodoUpdate updates a todo
func (h *TodoHandlers) HandleTodoUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoUpdateParams(request)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Prepare metadata map
	metadata := make(map[string]string)
	if params.Metadata.Status != "" {
		metadata["status"] = params.Metadata.Status
	}
	if params.Metadata.Priority != "" {
		metadata["priority"] = params.Metadata.Priority
	}
	if params.Metadata.CurrentTest != "" {
		metadata["current_test"] = params.Metadata.CurrentTest
	}
	
	// Update todo
	err = h.manager.UpdateTodo(params.ID, params.Section, params.Operation, params.Content, metadata)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Re-index after update
	todo, err := h.manager.ReadTodo(params.ID)
	if err == nil {
		// Read full content for indexing
		content, _ := h.manager.ReadTodoContent(params.ID)
		h.search.IndexTodo(todo, content)
	}
	
	return FormatTodoUpdateResponse(params.ID, params.Section, params.Operation), nil
}

// HandleTodoSearch searches todos
func (h *TodoHandlers) HandleTodoSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoSearchParams(request)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Prepare filters
	filters := make(map[string]string)
	if params.Filters.Status != "" {
		filters["status"] = params.Filters.Status
	}
	if params.Filters.DateFrom != "" {
		filters["date_from"] = params.Filters.DateFrom
	}
	if params.Filters.DateTo != "" {
		filters["date_to"] = params.Filters.DateTo
	}
	
	// Search
	results, err := h.search.SearchTodos(params.Query, filters, params.Limit)
	if err != nil {
		return HandleError(err), nil
	}
	
	return FormatTodoSearchResponse(results), nil
}

// HandleTodoArchive archives a todo
func (h *TodoHandlers) HandleTodoArchive(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoArchiveParams(request)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Archive todo
	err = h.manager.ArchiveTodo(params.ID, params.Quarter)
	if err != nil {
		return HandleError(err), nil
	}
	
	// Construct archive path
	todo, _ := h.manager.ReadTodo(params.ID)
	var archivePath string
	if todo != nil {
		quarter := params.Quarter
		if quarter == "" {
			// Use completion time or current time for quarter
			completionTime := todo.Completed
			if completionTime.IsZero() {
				completionTime = time.Now()
			}
			// Use daily path format
			dayPath := core.GetDailyPath(completionTime)
			archivePath = filepath.Join(".claude", "archive", dayPath, params.ID+".md")
		} else {
			archivePath = filepath.Join(".claude", "archive", quarter, params.ID+".md")
		}
	} else {
		// Fallback path
		archivePath = filepath.Join(".claude", "archive", params.ID+".md")
	}
	
	// Remove from search index
	err = h.search.DeleteTodo(params.ID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to remove from search index: %v\n", err)
	}
	
	return FormatTodoArchiveResponse(params.ID, archivePath), nil
}

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
	linker := core.NewTodoLinker(h.manager)
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
	
	// Calculate stats
	stats, err := h.stats.GenerateTodoStats()
	if err != nil {
		return HandleError(err), nil
	}
	
	// Filter by period if needed
	if period != "all" {
		// TODO: Implement period filtering
		// For now, return all stats
	}
	
	return FormatTodoStatsResponse(stats), nil
}

// HandleTodoClean performs cleanup operations
func (h *TodoHandlers) HandleTodoClean(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get operation type
	operation := request.GetString("operation", "archive_old")
	
	switch operation {
	case "archive_old":
		// Archive todos older than specified days
		days := request.GetInt("days", 90)
		count, err := h.manager.ArchiveOldTodos(days)
		if err != nil {
			return HandleError(err), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Archived %d todos older than %d days", count, days)), nil
		
	case "find_duplicates":
		// Find duplicate todos
		duplicates, err := h.manager.FindDuplicateTodos()
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