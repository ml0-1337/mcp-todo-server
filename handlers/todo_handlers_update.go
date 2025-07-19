package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

	// Get managers for the current context
	manager, search, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

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

		// Check if status is being set to completed for auto-archive
		if newStatus, hasStatus := metadataMap["status"]; hasStatus && newStatus == "completed" && !h.noAutoArchive {
			// Read todo to get its metadata for archive path
			todo, readErr := manager.ReadTodo(params.ID)
			
			// Perform auto-archive
			archiveErr := manager.ArchiveTodo(params.ID)
			if archiveErr != nil {
				// Log the error but don't fail the update
				fmt.Fprintf(os.Stderr, "Warning: failed to auto-archive todo: %v\n", archiveErr)
			}
			
			// Construct archive path
			var archivePath string
			if readErr == nil && todo != nil && archiveErr == nil {
				// Use the todo's started date for archive path
				dayPath := core.GetDailyPath(todo.Started)
				archivePath = filepath.Join(".claude", "archive", dayPath, params.ID+".md")
			}
			
			// Remove from search index if archived successfully
			if archiveErr == nil && search != nil {
				deleteErr := search.DeleteTodo(params.ID)
				if deleteErr != nil {
					// Log but don't fail
					fmt.Fprintf(os.Stderr, "Warning: failed to remove from search index: %v\n", deleteErr)
				}
			}
			
			// Create response with archive information
			if archiveErr == nil && archivePath != "" {
				return mcp.NewToolResultText(fmt.Sprintf(
					"Todo '%s' has been completed and archived to %s.\n\n"+
					"Task completed successfully. To maintain project momentum and capitalize on fresh insights, "+
					"take a moment to reflect on this completion and plan your next steps.\n\n"+
					"Consider the following:\n"+
					"- What specific knowledge or insights were gained from completing this task?\n"+
					"- Are there any immediate follow-up tasks that would benefit from your current context?\n"+
					"- Should any learnings be documented to prevent knowledge loss or help future similar work?\n"+
					"- Who might benefit from knowing about this completion or its outcomes?\n\n"+
					"Based on your reflection, what is the single most valuable action to take next? "+
					"Be specific and actionable in your recommendation.",
					params.ID, archivePath)), nil
			}
		}
		
		// Re-index if status changed (for non-completed statuses)
		if _, hasStatus := metadataMap["status"]; hasStatus && search != nil {
			todo, _ := manager.ReadTodo(params.ID)
			if todo != nil {
				content, _ := manager.ReadTodoContent(params.ID)
				search.IndexTodo(todo, content)
			}
		}

		// Create response
		updates := []string{}
		for key, value := range metadataMap {
			updates = append(updates, fmt.Sprintf("%s: %s", key, value))
		}
		
		// Check if status was set to completed (when auto-archive is disabled)
		if newStatus, hasStatus := metadataMap["status"]; hasStatus && newStatus == "completed" {
			return mcp.NewToolResultText(fmt.Sprintf(
				"Todo '%s' metadata updated: %s\n\n"+
				"Task marked as completed. This is an opportunity to leverage your current momentum and context. "+
				"Please analyze what you've accomplished and determine the most effective next action.\n\n"+
				"Key questions for your analysis:\n"+
				"- What concrete outcomes or deliverables resulted from this task?\n"+
				"- Which related tasks could benefit from immediate attention while context is fresh?\n"+
				"- What specific learnings should be captured for future reference?\n\n"+
				"Provide a clear, actionable recommendation for the next step that maximizes the value of this completion.",
				params.ID, strings.Join(updates, ", "))), nil
		}
		
		// Create enhanced response for non-completion metadata updates
		baseMsg := fmt.Sprintf("Todo '%s' metadata updated: %s", params.ID, strings.Join(updates, ", "))
		
		// Add context-specific guidance
		var guidance strings.Builder
		guidance.WriteString(baseMsg)
		guidance.WriteString("\n\n")
		
		// Check what was updated and provide relevant guidance
		if priority, hasPriority := metadataMap["priority"]; hasPriority {
			guidance.WriteString(fmt.Sprintf("Priority adjusted to %s. This change affects your work sequence:\n", priority))
			guidance.WriteString("- Review your current task list to ensure you're working on the highest priority items\n")
			guidance.WriteString("- Consider if any dependencies need to be re-evaluated\n\n")
			guidance.WriteString("Does this priority change require immediate task switching?")
		} else if status, hasStatus := metadataMap["status"]; hasStatus && status == "blocked" {
			guidance.WriteString("Task marked as blocked. To resolve this efficiently:\n")
			guidance.WriteString("- Document the specific blocker and its impact in the findings section\n")
			guidance.WriteString("- Identify who can help unblock or what information is needed\n")
			guidance.WriteString("- Consider working on alternative tasks while waiting\n\n")
			guidance.WriteString("What is the specific blocker preventing progress?")
		} else if currentTest, hasTest := metadataMap["current_test"]; hasTest {
			guidance.WriteString(fmt.Sprintf("Current test focus: %s\n\n", currentTest))
			guidance.WriteString("Following TDD workflow:\n")
			guidance.WriteString("- Ensure the test is written and failing before implementation\n")
			guidance.WriteString("- Keep changes minimal to make this specific test pass\n")
			guidance.WriteString("- Refactor only after the test is green\n\n")
			guidance.WriteString("Is the test currently in red (failing) state?")
		} else {
			guidance.WriteString("Metadata updated. Continue with your current workflow.")
		}
		
		return mcp.NewToolResultText(guidance.String()), nil
	}

	// Handle section updates
	if params.Section != "" {
		err = manager.UpdateTodo(params.ID, params.Section, params.Operation, params.Content, nil)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to update section")
		}

		// Re-index after content update
		if search != nil {
			todo, _ := manager.ReadTodo(params.ID)
			if todo != nil {
				content, _ := manager.ReadTodoContent(params.ID)
				search.IndexTodo(todo, content)
			}
		}

		// Create response
		opDesc := params.Operation
		if opDesc == "" {
			opDesc = "updated"
		}
		
		// Create enhanced response for section updates
		baseMsg := fmt.Sprintf("Todo '%s' %s section %s", params.ID, params.Section, opDesc)
		
		var guidance strings.Builder
		guidance.WriteString(baseMsg)
		guidance.WriteString("\n\n")
		
		// Provide section-specific guidance
		switch params.Section {
		case "findings":
			guidance.WriteString("Research captured. To maximize value from these findings:\n")
			guidance.WriteString("- Synthesize key insights into actionable conclusions\n")
			guidance.WriteString("- Identify patterns or recurring themes\n")
			guidance.WriteString("- Consider how these findings affect your implementation approach\n\n")
			guidance.WriteString("What key insight from your research will most impact the implementation?")
			
		case "tests", "test_cases":
			guidance.WriteString("Test cases documented. Following TDD practices:\n")
			guidance.WriteString("- Start with the simplest test that will fail\n")
			guidance.WriteString("- Write just enough code to make it pass\n")
			guidance.WriteString("- Refactor only after achieving green status\n\n")
			guidance.WriteString("Which test case should be implemented first to drive the design?")
			
		case "checklist":
			if params.Operation == "toggle" {
				guidance.WriteString("Checklist item toggled. To maintain progress:\n")
				guidance.WriteString("- Ensure completed items are truly done before checking them off\n")
				guidance.WriteString("- Update findings if you discovered anything while completing this item\n")
				guidance.WriteString("- Move to the next unchecked item systematically\n\n")
				guidance.WriteString("What's the next checklist item to tackle?")
			} else {
				guidance.WriteString("Checklist updated. For effective task tracking:\n")
				guidance.WriteString("- Break down large items into smaller, actionable steps\n")
				guidance.WriteString("- Order items by dependency and priority\n")
				guidance.WriteString("- Focus on completing one item fully before moving to the next\n\n")
				guidance.WriteString("Is the checklist ordered for optimal workflow?")
			}
			
		case "scratchpad", "notes":
			guidance.WriteString("Notes captured. Remember to:\n")
			guidance.WriteString("- Transfer important information to appropriate sections\n")
			guidance.WriteString("- Convert rough ideas into actionable items\n")
			guidance.WriteString("- Clean up the scratchpad periodically\n\n")
			guidance.WriteString("Should any of these notes be formalized into findings or test cases?")
			
		default:
			guidance.WriteString("Section updated. Continue documenting your progress systematically.")
		}
		
		return mcp.NewToolResultText(guidance.String()), nil
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

	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

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

	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

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

	// Get managers for the current context
	manager, _, _, _, err := h.factory.GetManagers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get context-aware managers: %w", err)
	}

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