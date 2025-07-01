package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// HandleTodoUpdate handles the todo_update tool
func (h *TodoHandlers) HandleTodoUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params, err := ExtractTodoUpdateParams(request)
	if err != nil {
		return nil, err
	}

	// Get the manager for this context
	manager := h.getManagerForContext(ctx)

	// Handle metadata updates
	metadataMap := make(map[string]string)
	if params.Metadata.Status != "" {
		metadataMap["status"] = params.Metadata.Status
	}
	if params.Metadata.Priority != "" {
		metadataMap["priority"] = params.Metadata.Priority
	}
	if params.Metadata.CurrentTest != "" {
		metadataMap["current_test"] = params.Metadata.CurrentTest
	}

	if len(metadataMap) > 0 {
		err = manager.UpdateTodo(params.ID, "", "", "", metadataMap)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to update metadata")
		}

		// Re-index if status changed
		if _, hasStatus := metadataMap["status"]; hasStatus && h.search != nil {
			todo, _ := manager.ReadTodo(params.ID)
			if todo != nil {
				content, _ := manager.ReadTodoContent(params.ID)
				h.search.IndexTodo(todo, content)
			}
		}

		// Create response
		updates := []string{}
		for key, value := range metadataMap {
			updates = append(updates, fmt.Sprintf("%s: %s", key, value))
		}
		
		return mcp.NewToolResultText(fmt.Sprintf("Todo '%s' metadata updated: %s", params.ID, strings.Join(updates, ", "))), nil
	}

	// Handle section updates
	if params.Section != "" {
		err = manager.UpdateTodo(params.ID, params.Section, params.Operation, params.Content, nil)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to update section")
		}

		// Re-index after content update
		if h.search != nil {
			todo, _ := manager.ReadTodo(params.ID)
			if todo != nil {
				content, _ := manager.ReadTodoContent(params.ID)
				h.search.IndexTodo(todo, content)
			}
		}

		// Create response
		opDesc := params.Operation
		if opDesc == "" {
			opDesc = "updated"
		}
		
		return mcp.NewToolResultText(fmt.Sprintf("Todo '%s' %s section %s", params.ID, params.Section, opDesc)), nil
	}

	return nil, interrors.NewValidationError("operation", "", "no update operation specified")
}

// HandleTodoSections handles the todo_sections tool
func (h *TodoHandlers) HandleTodoSections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract ID parameter
	id, err := request.RequireString("id")
	if err != nil {
		return HandleError(interrors.NewValidationError("id", "", "missing required parameter")), nil
	}

	// Get the manager for this context
	manager := h.getManagerForContext(ctx)

	// Read the todo
	todo, content, err := manager.ReadTodoWithContent(id)
	if err != nil {
		return HandleError(err), nil
	}

	// Format sections response
	return FormatTodoSectionsResponseWithContent(todo, content), nil
}

// HandleTodoAddSection handles adding a new section to a todo
func (h *TodoHandlers) HandleTodoAddSection(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	id, err := request.RequireString("id")
	if err != nil {
		return HandleError(interrors.NewValidationError("id", "", "missing required parameter")), nil
	}

	key, err := request.RequireString("key")
	if err != nil {
		return HandleError(interrors.NewValidationError("key", "", "missing required parameter")), nil
	}

	title, err := request.RequireString("title")
	if err != nil {
		return HandleError(interrors.NewValidationError("title", "", "missing required parameter")), nil
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
		return HandleError(interrors.NewValidationError("schema", schema, "invalid schema type")), nil
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
			return HandleError(interrors.NewConflictError("section", key, "section already exists")), nil
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
		return HandleError(interrors.NewValidationError("id", "", "missing required parameter")), nil
	}

	// Extract order parameter - should be a map of section keys to new order values
	orderParam, ok := args["order"]
	if !ok {
		return HandleError(interrors.NewValidationError("order", nil, "missing required parameter")), nil
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