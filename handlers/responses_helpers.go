package handlers

import (
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
	"time"
)

// extractSectionContents parses markdown content and extracts sections
func extractSectionContents(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	currentSection := ""
	var sectionContent strings.Builder

	for _, line := range lines {
		// Check if it's a section header
		if strings.HasPrefix(line, "## ") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = sectionContent.String()
			}

			// Start new section
			currentSection = strings.TrimSpace(strings.TrimPrefix(line, "##"))
			sectionContent.Reset()
		} else if currentSection != "" {
			// Add to current section
			sectionContent.WriteString(line + "\n")
		}
	}

	// Save last section
	if currentSection != "" {
		sections[currentSection] = sectionContent.String()
	}

	// Normalize section keys
	normalized := make(map[string]string)
	for title, content := range sections {
		key := normalizeKey(title)
		normalized[key] = content
	}

	return normalized
}

// normalizeKey converts a section title to a normalized key
func normalizeKey(title string) string {
	// Handle known sections
	switch title {
	case "Findings & Research", "Findings":
		return "findings"
	case "Web Searches":
		return "web_searches"
	case "Test Strategy":
		return "test_strategy"
	case "Test List":
		return "test_list"
	case "Test Cases":
		return "tests"
	case "Maintainability Analysis":
		return "maintainability"
	case "Test Results Log":
		return "test_results"
	case "Checklist":
		return "checklist"
	case "Working Scratchpad":
		return "scratchpad"
	default:
		// Generic normalization
		return strings.ToLower(strings.ReplaceAll(title, " ", "_"))
	}
}

// getStatusIcon returns an icon for the todo status
func getStatusIcon(status string) string {
	switch status {
	case "completed":
		return "[✓]"
	case "in_progress":
		return "[→]"
	case "blocked":
		return "[✗]"
	default:
		return "[ ]"
	}
}

// getPriorityLabel returns a label for the priority
func getPriorityLabel(priority string) string {
	switch priority {
	case "high":
		return "[HIGH]"
	case "low":
		return "[LOW]"
	default:
		return ""
	}
}

// formatTodosFullWithContent formats multiple todos with full content
func formatTodosFullWithContent(todos []*core.Todo, contents map[string]string) *mcp.CallToolResult {
	var results []map[string]interface{}

	for _, todo := range todos {
		data := map[string]interface{}{
			"id":       todo.ID,
			"task":     todo.Task,
			"status":   todo.Status,
			"priority": todo.Priority,
			"type":     todo.Type,
			"started":  todo.Started.Format(time.RFC3339),
			"tags":     todo.Tags,
		}
		if !todo.Completed.IsZero() {
			data["completed"] = todo.Completed.Format(time.RFC3339)
		}
		if todo.ParentID != "" {
			data["parent_id"] = todo.ParentID
		}

		// Add sections with content
		if content, exists := contents[todo.ID]; exists {
			sections := extractSectionContents(content)
			sectionData := make(map[string]interface{})

			for key, sectionContent := range sections {
				if key == "checklist" {
					// Parse checklist items
					sectionData[key] = core.ParseChecklist(sectionContent)
				} else {
					// Regular section content - just the string
					sectionData[key] = strings.TrimSpace(sectionContent)
				}
			}
			data["sections"] = sectionData
		}

		results = append(results, data)
	}

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}
