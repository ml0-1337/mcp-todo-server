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
				prompts := ""
				if readErr == nil && todo != nil {
					prompts = getCompletionPrompts(todo.Type)
				} else {
					prompts = getCompletionPrompts("")
				}

				return mcp.NewToolResultText(fmt.Sprintf(
					"Todo '%s' has been completed and archived to %s.\n\n%s",
					params.ID, archivePath, prompts)), nil
			}
		}

		// Re-index if status changed (for non-completed statuses)
		if _, hasStatus := metadataMap["status"]; hasStatus && search != nil {
			todo, err := manager.ReadTodo(params.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to read todo for re-indexing %s: %v\n", params.ID, err)
			} else if todo != nil {
				content, err := manager.ReadTodoContent(params.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to read todo content for re-indexing %s: %v\n", params.ID, err)
				} else if err := search.IndexTodo(todo, content); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to re-index todo %s: %v\n", params.ID, err)
				}
			}
		}

		// Create response
		updates := []string{}
		for key, value := range metadataMap {
			updates = append(updates, fmt.Sprintf("%s: %s", key, value))
		}

		// Check if status was set to completed (when auto-archive is disabled)
		if newStatus, hasStatus := metadataMap["status"]; hasStatus && newStatus == "completed" {
			// Read todo to get type for contextual prompts
			todo, err := manager.ReadTodo(params.ID)
			todoType := ""
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to read todo for type %s: %v\n", params.ID, err)
			} else if todo != nil {
				todoType = todo.Type
			}

			return mcp.NewToolResultText(fmt.Sprintf(
				"Todo '%s' metadata updated: %s\n\n%s",
				params.ID, strings.Join(updates, ", "), getCompletionPrompts(todoType))), nil
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
		if search != nil {
			todo, err := manager.ReadTodo(params.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to read todo for re-indexing %s: %v\n", params.ID, err)
			} else if todo != nil {
				content, err := manager.ReadTodoContent(params.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to read todo content for re-indexing %s: %v\n", params.ID, err)
				} else if err := search.IndexTodo(todo, content); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to re-index todo %s: %v\n", params.ID, err)
				}
			}
		}

		// Create response
		opDesc := params.Operation
		if opDesc == "" {
			opDesc = "updated"
		}

		// Get todo type for contextual prompts
		todoType := ""
		todo, err := manager.ReadTodo(params.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read todo for type %s: %v\n", params.ID, err)
		} else if todo != nil {
			todoType = todo.Type
		}

		// Build response with contextual prompts
		baseMessage := fmt.Sprintf("Todo '%s' %s section %s", params.ID, params.Section, opDesc)
		prompts := getUpdatePrompts(params.Section, params.Operation, todoType)

		if prompts != "" {
			return mcp.NewToolResultText(baseMessage + "\n\n" + prompts), nil
		}

		return mcp.NewToolResultText(baseMessage), nil
	}

	return nil, interrors.NewValidationError("operation", "", "no update operation specified")
}

// getCompletionPrompts returns contextual prompts based on todo type
func getCompletionPrompts(todoType string) string {
	switch todoType {
	case "feature":
		return "Task completed successfully. To maintain momentum and capture insights:\n\n" +
			"- What went well with this implementation?\n" +
			"- Were there unexpected challenges that future features should consider?\n" +
			"- Are there follow-up improvements or related features to pursue?\n\n" +
			"What's the most valuable next action based on this completion?"

	case "bug":
		return "Bug fix completed. To prevent similar issues:\n\n" +
			"- What was the root cause of this bug?\n" +
			"- Could this issue have been caught earlier in the development process?\n" +
			"- Are there related areas of the codebase that might have similar vulnerabilities?\n\n" +
			"What preventive measure would have the highest impact?"

	case "research":
		return "Research completed. To maximize the value of your findings:\n\n" +
			"- What are the key insights or conclusions?\n" +
			"- Which findings should influence upcoming decisions?\n" +
			"- Are there areas that warrant deeper investigation?\n\n" +
			"What action should be taken based on these research findings?"

	case "refactor":
		return "Refactoring completed. To ensure code quality improvements:\n\n" +
			"- What specific improvements were achieved?\n" +
			"- Did this reveal additional technical debt to address?\n" +
			"- How can we measure the impact of these changes?\n\n" +
			"What's the next priority for code quality improvement?"

	default:
		return "Task completed successfully. To capitalize on your progress:\n\n" +
			"- What key learnings emerged from this work?\n" +
			"- Are there immediate follow-up tasks to consider?\n" +
			"- Should any insights be documented for future reference?\n\n" +
			"What's the most valuable next step to maintain momentum?"
	}
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
	orderMap, ok := orderParam.(map[string]any)
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
	if len(todo.Sections) == 0 {
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
