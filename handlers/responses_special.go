package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// FormatTodoArchiveResponse formats the response for todo_archive
func FormatTodoArchiveResponse(todoID string, archivePath string, todoType string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":           todoID,
		"archive_path": archivePath,
		"message":      fmt.Sprintf("Todo '%s' archived successfully", todoID),
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}

	// Add contextual prompts for archiving
	prompt := getArchivePrompts(todoType)
	result := string(jsonData) + "\n\n" + prompt

	return mcp.NewToolResultText(result)
}

// getArchivePrompts returns contextual prompts based on todo type after archiving
func getArchivePrompts(todoType string) string {
	basePrompt := "Todo archived successfully. "

	switch todoType {
	case "feature":
		return basePrompt + "To build on this feature:\n\n" +
			"- Were there any edge cases discovered during implementation?\n" +
			"- Are there enhancement opportunities for the future?\n" +
			"- Did users request any related functionality?\n\n" +
			"Consider creating follow-up todos for improvements or related features."

	case "bug":
		return basePrompt + "Bug fix archived. To prevent similar issues:\n\n" +
			"- What was the root cause of this bug?\n" +
			"- Are there similar patterns in the codebase to check?\n" +
			"- Should monitoring or tests be added to catch this earlier?\n\n" +
			"Document lessons learned or create todos for preventive measures."

	case "research":
		return basePrompt + "Research completed. To apply findings:\n\n" +
			"- What were the key insights from this research?\n" +
			"- Are there actionable recommendations to implement?\n" +
			"- Should findings be documented in project knowledge base?\n\n" +
			"Create implementation todos based on research outcomes."

	case "refactor":
		return basePrompt + "Refactoring complete. To maintain code quality:\n\n" +
			"- Did this reveal other areas needing refactoring?\n" +
			"- Are there new patterns to apply elsewhere?\n" +
			"- Were any performance improvements measured?\n\n" +
			"Consider documenting new patterns or creating todos for similar improvements."

	default:
		return basePrompt + "To continue productive work:\n\n" +
			"- What did you learn from completing this task?\n" +
			"- Are there any follow-up actions needed?\n" +
			"- What should be the next priority?\n\n" +
			"Review your todo list or create new tasks based on lessons learned."
	}
}

// FormatTodoStatsResponse formats the response for todo_stats
func FormatTodoStatsResponse(stats *core.TodoStats) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// FormatCleanResponse formats the response for todo_clean operations
func FormatCleanResponse(operation string, result interface{}) *mcp.CallToolResult {
	response := map[string]interface{}{
		"operation": operation,
		"result":    result,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}
