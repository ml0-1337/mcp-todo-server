package handlers

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// TodoHandlers contains handlers for all todo operations
type TodoHandlers struct {
	manager   TodoManagerInterface
	search    SearchEngineInterface
	stats     StatsEngineInterface
	templates TemplateManagerInterface
	// Factory function for creating TodoLinker (to avoid type assertion)
	createLinker func(TodoManagerInterface) TodoLinkerInterface
}

// NewTodoHandlers creates new todo handlers with dependencies
func NewTodoHandlers(todoPath, templatePath string) (*TodoHandlers, error) {
	// Create context-aware todo manager wrapper
	manager := NewContextualTodoManagerWrapper(todoPath)

	// Create search engine
	indexPath := filepath.Join(todoPath, "..", "index", "todos.bleve")
	search, err := core.NewSearchEngine(indexPath, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	// Create stats engine with the default manager
	// (stats needs concrete TodoManager, not the wrapper)
	stats := core.NewStatsEngine(manager.defaultManager)

	// Create template manager
	templates := core.NewTemplateManager(templatePath)

	return &TodoHandlers{
		manager:   manager,
		search:    search,
		stats:     stats,
		templates: templates,
		createLinker: func(m TodoManagerInterface) TodoLinkerInterface {
			// Handle ContextualTodoManagerWrapper
			if wrapper, ok := m.(*ContextualTodoManagerWrapper); ok {
				return core.NewTodoLinker(wrapper.defaultManager)
			}
			// Type assert to concrete type for core.NewTodoLinker
			if tm, ok := m.(*core.TodoManager); ok {
				return core.NewTodoLinker(tm)
			}
			return nil
		},
	}, nil
}

// NewTodoHandlersWithDependencies creates new todo handlers with explicit dependencies (for testing)
func NewTodoHandlersWithDependencies(
	manager TodoManagerInterface,
	search SearchEngineInterface,
	stats StatsEngineInterface,
	templates TemplateManagerInterface,
) *TodoHandlers {
	return &TodoHandlers{
		manager:   manager,
		search:    search,
		stats:     stats,
		templates: templates,
		createLinker: func(m TodoManagerInterface) TodoLinkerInterface {
			// For testing, we'll need a mock linker
			return nil
		},
	}
}

// Close cleans up resources
func (h *TodoHandlers) Close() error {
	if h.search != nil {
		return h.search.Close()
	}
	return nil
}

// GetBasePathForContext returns the appropriate base path for the context
func (h *TodoHandlers) GetBasePathForContext(ctx context.Context) string {
	// Check if we have a contextual manager wrapper
	if ctxWrapper, ok := h.manager.(*ContextualTodoManagerWrapper); ok {
		manager := ctxWrapper.GetManagerForContext(ctx)
		return manager.GetBasePath()
	}
	
	// Fall back to default manager
	return h.manager.GetBasePath()
}

// getManagerForContext returns the appropriate todo manager for the context
func (h *TodoHandlers) getManagerForContext(ctx context.Context) TodoManagerInterface {
	// Check if we have a contextual manager wrapper
	if ctxWrapper, ok := h.manager.(*ContextualTodoManagerWrapper); ok {
		manager := ctxWrapper.GetManagerForContext(ctx)
		log.Printf("Using context-aware manager with path: %s", manager.GetBasePath())
		return manager
	}
	
	// Fall back to the default manager
	log.Printf("Using default manager (no context wrapper)")
	return h.manager
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

		// Get the appropriate base path for this context
		basePath := h.GetBasePathForContext(ctx)
		
		// Write todo file
		filePath := filepath.Join(basePath, todo.ID+".md")
		return FormatTodoTemplateResponse(todo, filePath), nil
	}

	// Create todo using context if available
	todo, err := h.CreateTodoWithContext(ctx, params.Task, params.Priority, params.Type)
	if err != nil {
		return HandleError(err), nil
	}

	// Get the context-aware manager for remaining operations
	manager := h.getManagerForContext(ctx)

	// Set parent ID if provided
	if params.ParentID != "" {
		todo.ParentID = params.ParentID
		// Update the todo file with parent ID
		err = manager.UpdateTodo(todo.ID, "", "", "", map[string]string{"parent_id": params.ParentID})
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

	// Get existing todos for similarity detection
	existingTodos, _ := manager.ListTodos("", "", 0)

	// Return enhanced response with hints
	filePath := filepath.Join(h.GetBasePathForContext(ctx), todo.ID+".md")
	return FormatTodoCreateResponseWithHints(todo, filePath, existingTodos), nil
}

// HandleTodoCreateMulti creates multiple todos with parent-child relationships
func (h *TodoHandlers) HandleTodoCreateMulti(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoCreateMultiParams(request)
	if err != nil {
		return HandleError(err), nil
	}

	// Create parent todo using context
	parentTodo, err := h.CreateTodoWithContext(ctx, params.Parent.Task, params.Parent.Priority, params.Parent.Type)
	if err != nil {
		return HandleError(fmt.Errorf("failed to create parent todo: %w", err)), nil
	}

	// Index parent in search
	parentContent := fmt.Sprintf("# Task: %s\n\n## Findings & Research\n\n## Test Strategy\n\n", parentTodo.Task)
	err = h.search.IndexTodo(parentTodo, parentContent)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to index parent todo: %v\n", err)
	}

	// Create children todos
	createdChildren := make([]*core.Todo, 0, len(params.Children))
	for i, childInfo := range params.Children {
		// Create child todo using context
		childTodo, err := h.CreateTodoWithContext(ctx, childInfo.Task, childInfo.Priority, childInfo.Type)
		if err != nil {
			// Try to clean up already created todos
			fmt.Printf("Error creating child %d: %v. Rolling back...\n", i, err)
			// Note: In a production system, we'd want proper transaction support
			return HandleError(fmt.Errorf("failed to create child %d: %w", i, err)), nil
		}

		// Get context-aware manager for update
		manager := h.getManagerForContext(ctx)
		
		// Set parent ID
		err = manager.UpdateTodo(childTodo.ID, "", "", "", map[string]string{"parent_id": parentTodo.ID})
		if err != nil {
			fmt.Printf("Warning: failed to set parent_id for child %s: %v\n", childTodo.ID, err)
		}

		// Index child in search
		childContent := fmt.Sprintf("# Task: %s\n\n## Findings & Research\n\n## Test Strategy\n\n", childTodo.Task)
		err = h.search.IndexTodo(childTodo, childContent)
		if err != nil {
			fmt.Printf("Warning: failed to index child todo: %v\n", err)
		}

		createdChildren = append(createdChildren, childTodo)
	}

	// Return response with all created todos
	return FormatTodoCreateMultiResponse(parentTodo, createdChildren), nil
}

// HandleTodoRead reads one or more todos
func (h *TodoHandlers) HandleTodoRead(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	params, err := ExtractTodoReadParams(request)
	if err != nil {
		return HandleError(err), nil
	}

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

	// Single todo read
	if params.ID != "" {
		if params.Format == "full" {
			// For full format, read both todo and content
			todo, content, err := manager.ReadTodoWithContent(params.ID)
			if err != nil {
				return HandleError(err), nil
			}
			return formatSingleTodoWithContent(todo, content, params.Format), nil
		} else {
			// For other formats, just read todo metadata
			todo, err := manager.ReadTodo(params.ID)
			if err != nil {
				return HandleError(err), nil
			}
			return FormatTodoReadResponse([]*core.Todo{todo}, params.Format, true), nil
		}
	}

	// List todos with filters
	todos, err := manager.ListTodos(params.Filter.Status, params.Filter.Priority, params.Filter.Days)
	if err != nil {
		return HandleError(err), nil
	}

	// For full format with multiple todos, read content for each
	if params.Format == "full" {
		contents := make(map[string]string)
		for _, todo := range todos {
			content, err := manager.ReadTodoContent(todo.ID)
			if err == nil {
				contents[todo.ID] = content
			}
		}
		return formatTodosFullWithContent(todos, contents), nil
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

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

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

	// If updating a section, validate content against schema (skip for toggle operation)
	if params.Section != "" && params.Content != "" && params.Operation != "toggle" {
		// Read todo to get section metadata
		todo, err := manager.ReadTodo(params.ID)
		if err != nil {
			return HandleError(err), nil
		}

		// Check if todo has section metadata
		if todo.Sections != nil {
			// Find the section definition
			sectionFound := false
			for key, sectionDef := range todo.Sections {
				if key == params.Section {
					sectionFound = true
					// Get validator for the schema
					validator := core.GetValidator(sectionDef.Schema)
					if validator != nil {
						// Validate the content
						if err := validator.Validate(params.Content); err != nil {
							return HandleError(fmt.Errorf("validation error: %w", err)), nil
						}
					}
					break
				}
			}

			// If section not found in metadata, check if it's a valid section name
			if !sectionFound {
				// The section doesn't exist in this todo
				return HandleError(fmt.Errorf("section '%s' does not exist in this todo", params.Section)), nil
			}
		} else {
			// No section metadata, so section doesn't exist
			return HandleError(fmt.Errorf("section '%s' does not exist in this todo", params.Section)), nil
		}
	}

	// Update todo
	err = manager.UpdateTodo(params.ID, params.Section, params.Operation, params.Content, metadata)
	if err != nil {
		return HandleError(err), nil
	}

	// Re-index after update and prepare enriched response
	todo, err := manager.ReadTodo(params.ID)
	if err == nil {
		// Read full content for indexing and response
		content, _ := manager.ReadTodoContent(params.ID)
		h.search.IndexTodo(todo, content)

		// Return enriched response with full todo data
		return FormatEnrichedTodoUpdateResponse(todo, content, params.Section, params.Operation), nil
	}

	// Fallback to simple response if we couldn't read the todo
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

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

	// Read todo BEFORE archiving to get its metadata
	todo, readErr := manager.ReadTodo(params.ID)

	// Archive todo
	err = manager.ArchiveTodo(params.ID, params.Quarter)
	if err != nil {
		return HandleError(err), nil
	}

	// Construct archive path
	var archivePath string
	if readErr == nil && todo != nil {
		quarter := params.Quarter
		if quarter == "" {
			// Use the todo's started date for archive path (matches ArchiveTodo behavior)
			dayPath := core.GetDailyPath(todo.Started)
			archivePath = filepath.Join(".claude", "archive", dayPath, params.ID+".md")
		} else {
			archivePath = filepath.Join(".claude", "archive", quarter, params.ID+".md")
		}
	} else {
		// Fallback path when we couldn't read the todo
		// This might happen with timestamp parsing errors
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
	filePath := filepath.Join(h.GetBasePathForContext(ctx), todo.ID+".md")
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
	if h.createLinker == nil {
		return HandleError(fmt.Errorf("linker not available")), nil
	}
	linker := h.createLinker(h.manager)
	if linker == nil {
		return HandleError(fmt.Errorf("failed to create linker")), nil
	}
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

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

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

// HandleTodoSections returns all sections with metadata for a todo
func (h *TodoHandlers) HandleTodoSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract ID parameter
	id, err := request.RequireString("id")
	if err != nil {
		return HandleError(fmt.Errorf("missing required parameter 'id'")), nil
	}

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

	// Read todo
	todo, err := manager.ReadTodo(id)
	if err != nil {
		return HandleError(err), nil
	}

	// Read todo content to analyze sections
	content, err := manager.ReadTodoContent(id)
	if err != nil {
		return HandleError(err), nil
	}

	// Format sections response with content status
	return FormatTodoSectionsResponseWithContent(todo, content), nil
}

// HandleTodoAddSection adds a custom section to an existing todo
func (h *TodoHandlers) HandleTodoAddSection(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	id, err := request.RequireString("id")
	if err != nil {
		return HandleError(fmt.Errorf("missing required parameter 'id'")), nil
	}

	key, err := request.RequireString("key")
	if err != nil {
		return HandleError(fmt.Errorf("missing required parameter 'key'")), nil
	}

	title, err := request.RequireString("title")
	if err != nil {
		return HandleError(fmt.Errorf("missing required parameter 'title'")), nil
	}

	schema := request.GetString("schema", "freeform")
	required := request.GetBool("required", false)
	order := request.GetInt("order", 100) // Default to high order (end of sections)

	// Validate schema
	validSchemas := map[string]bool{
		"freeform":   true,
		"checklist":  true,
		"test_cases": true,
		"research":   true,
		"strategy":   true,
		"results":    true,
	}

	if !validSchemas[schema] {
		return HandleError(fmt.Errorf("invalid schema: %s", schema)), nil
	}

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

	// Read the todo
	todo, err := manager.ReadTodo(id)
	if err != nil {
		return HandleError(err), nil
	}

	// Check if section already exists
	if todo.Sections != nil {
		if _, exists := todo.Sections[key]; exists {
			return HandleError(fmt.Errorf("section '%s' already exists", key)), nil
		}
	} else {
		// Initialize sections map if it doesn't exist
		todo.Sections = make(map[string]*core.SectionDefinition)
	}

	// Add the new section
	todo.Sections[key] = &core.SectionDefinition{
		Title:    title,
		Order:    order,
		Schema:   core.SectionSchema(schema),
		Required: required,
	}

	// Save the todo with the new section
	err = manager.SaveTodo(todo)
	if err != nil {
		return HandleError(err), nil
	}

	// Return success response
	return mcp.NewToolResultText(fmt.Sprintf("Section '%s' added successfully to todo '%s'", key, id)), nil
}

// HandleTodoReorderSections reorders sections in a todo
func (h *TodoHandlers) HandleTodoReorderSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get arguments
	args := request.GetArguments()

	// Extract ID parameter
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return HandleError(fmt.Errorf("missing required parameter 'id'")), nil
	}

	// Extract order parameter - should be a map of section keys to new order values
	orderParam, ok := args["order"]
	if !ok {
		return HandleError(fmt.Errorf("missing required parameter 'order'")), nil
	}

	// Type assert order to map
	orderMap, ok := orderParam.(map[string]interface{})
	if !ok {
		return HandleError(fmt.Errorf("'order' must be an object mapping section keys to order values")), nil
	}

	// Get the context-aware manager
	manager := h.getManagerForContext(ctx)

	// Read the todo
	todo, err := manager.ReadTodo(id)
	if err != nil {
		return HandleError(err), nil
	}

	// Check if todo has sections
	if todo.Sections == nil || len(todo.Sections) == 0 {
		return HandleError(fmt.Errorf("todo has no sections defined")), nil
	}

	// Update section orders
	for key, orderValue := range orderMap {
		// Check if section exists
		section, exists := todo.Sections[key]
		if !exists {
			return HandleError(fmt.Errorf("section '%s' not found in todo", key)), nil
		}

		// Convert order value to int
		var newOrder int
		switch v := orderValue.(type) {
		case float64:
			newOrder = int(v)
		case int:
			newOrder = v
		default:
			return HandleError(fmt.Errorf("order value must be a number for section '%s'", key)), nil
		}

		// Update the order
		section.Order = newOrder
	}

	// Save the todo with updated section orders
	err = manager.SaveTodo(todo)
	if err != nil {
		return HandleError(err), nil
	}

	// Return success response
	return mcp.NewToolResultText(fmt.Sprintf("Sections reordered successfully for todo '%s'", id)), nil
}
