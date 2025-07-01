package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
)

// FormatTodoCreateResponse formats the response for todo_create
func FormatTodoCreateResponse(todo *core.Todo, filePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":      todo.ID,
		"path":    filePath,
		"message": fmt.Sprintf("Todo created successfully: %s", todo.ID),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatTodoCreateResponseWithHints formats the response with pattern hints
func FormatTodoCreateResponseWithHints(todo *core.Todo, filePath string, existingTodos []*core.Todo) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":      todo.ID,
		"path":    filePath,
		"message": fmt.Sprintf("Todo created successfully: %s", todo.ID),
	}

	// Detect patterns in the title
	if hint := core.DetectPattern(todo.Task); hint != nil {
		response["hint"] = map[string]interface{}{
			"pattern":       hint.Pattern,
			"suggestedType": hint.SuggestedType,
			"message":       hint.Message,
		}
	}

	// Find similar todos
	if existingTodos != nil && len(existingTodos) > 0 {
		similar := core.FindSimilarTodos(existingTodos, todo.Task)
		if len(similar) > 0 {
			response["similar_todos"] = similar
		}
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatTodoCreateMultiResponse formats the response for todo_create_multi
func FormatTodoCreateMultiResponse(parent *core.Todo, children []*core.Todo) *mcp.CallToolResult {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Created multi-phase project: %s\n\n", parent.ID))
	sb.WriteString("📋 Project Structure:\n")
	
	// Parent
	sb.WriteString(fmt.Sprintf("[→] %s: %s [%s] [%s]\n", parent.ID, parent.Task, 
		strings.ToUpper(parent.Priority), parent.Type))
	
	// Children
	for _, child := range children {
		sb.WriteString(fmt.Sprintf("  └─ %s: %s [%s] [%s]\n", child.ID, child.Task,
			strings.ToUpper(child.Priority), child.Type))
	}
	
	childCount := len(children)
	sb.WriteString(fmt.Sprintf("\n✅ Successfully created %d todos (1 parent, %d children)\n", 
		childCount+1, childCount))
	
	sb.WriteString("\n💡 TIP: Use `todo_read` to see the full hierarchy or `todo_update` to modify individual todos.")
	
	return mcp.NewToolResultText(sb.String())
}

// FormatTodoTemplateResponse formats the response for template creation
func FormatTodoTemplateResponse(todo *core.Todo, filePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":       todo.ID,
		"path":     filePath,
		"message":  fmt.Sprintf("Todo created from template successfully: %s", todo.ID),
		"template": "Applied template sections and structure",
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}