package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// FormatTodoArchiveResponse formats the response for todo_archive
func FormatTodoArchiveResponse(todoID string, archivePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":           todoID,
		"archive_path": archivePath,
		"message":      fmt.Sprintf("Todo '%s' archived successfully", todoID),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatTodoStatsResponse formats the response for todo_stats
func FormatTodoStatsResponse(stats *core.TodoStats) *mcp.CallToolResult {
	jsonData, _ := json.MarshalIndent(stats, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatTodoLinkResponse formats the response for todo_link
func FormatTodoLinkResponse(parentID, childID string, linkType string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"parent_id": parentID,
		"child_id":  childID,
		"link_type": linkType,
		"message":   fmt.Sprintf("Todos linked successfully: %s -> %s", parentID, childID),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatCleanResponse formats the response for todo_clean operations
func FormatCleanResponse(operation string, result interface{}) *mcp.CallToolResult {
	response := map[string]interface{}{
		"operation": operation,
		"result":    result,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// FormatTemplateListResponse formats the list of available templates
func FormatTemplateListResponse(templates []string) *mcp.CallToolResult {
	if len(templates) == 0 {
		return mcp.NewToolResultText("No templates available")
	}

	response := map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
		"message":   "Available templates",
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}