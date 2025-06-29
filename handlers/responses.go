package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
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

// FormatTodoReadResponse formats the response for todo_read based on format
func FormatTodoReadResponse(todos []*core.Todo, format string, singleTodo bool) *mcp.CallToolResult {
	if singleTodo && len(todos) > 0 {
		return formatSingleTodo(todos[0], format)
	}
	
	switch format {
	case "full":
		return formatTodosFull(todos)
	case "list":
		return formatTodosList(todos)
	default: // summary
		return formatTodosSummary(todos)
	}
}

// formatSingleTodo formats a single todo based on format
func formatSingleTodo(todo *core.Todo, format string) *mcp.CallToolResult {
	if format == "full" {
		// For full format, we'd need to read the entire file content
		// For now, return structured data
		data := map[string]interface{}{
			"id":        todo.ID,
			"task":      todo.Task,
			"status":    todo.Status,
			"priority":  todo.Priority,
			"type":      todo.Type,
			"started":   todo.Started.Format(time.RFC3339),
			"tags":      todo.Tags,
		}
		if !todo.Completed.IsZero() {
			data["completed"] = todo.Completed.Format(time.RFC3339)
		}
		if todo.ParentID != "" {
			data["parent_id"] = todo.ParentID
		}
		
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		return mcp.NewToolResultText(string(jsonData))
	}
	
	// Summary format
	return mcp.NewToolResultText(formatTodoSummaryLine(todo))
}

// formatTodosFull formats multiple todos in full format
func formatTodosFull(todos []*core.Todo) *mcp.CallToolResult {
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
		results = append(results, data)
	}
	
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// formatTodosList formats todos as a simple list
func formatTodosList(todos []*core.Todo) *mcp.CallToolResult {
	var lines []string
	for _, todo := range todos {
		lines = append(lines, fmt.Sprintf("- %s: %s", todo.ID, todo.Task))
	}
	
	if len(lines) == 0 {
		return mcp.NewToolResultText("No todos found")
	}
	
	return mcp.NewToolResultText(strings.Join(lines, "\n"))
}

// formatTodosSummary formats todos in summary format
func formatTodosSummary(todos []*core.Todo) *mcp.CallToolResult {
	var lines []string
	
	// Group by status
	statusGroups := make(map[string][]*core.Todo)
	for _, todo := range todos {
		statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
	}
	
	// Format by status
	for _, status := range []string{"in_progress", "blocked", "completed"} {
		if todos, ok := statusGroups[status]; ok && len(todos) > 0 {
			lines = append(lines, fmt.Sprintf("\n%s (%d):", strings.ToUpper(status), len(todos)))
			for _, todo := range todos {
				lines = append(lines, formatTodoSummaryLine(todo))
			}
		}
	}
	
	if len(lines) == 0 {
		return mcp.NewToolResultText("No todos found")
	}
	
	return mcp.NewToolResultText(strings.Join(lines, "\n"))
}

// formatTodoSummaryLine formats a single todo as a summary line
func formatTodoSummaryLine(todo *core.Todo) string {
	status := "[ ]"
	if todo.Status == "completed" {
		status = "[✓]"
	} else if todo.Status == "in_progress" {
		status = "[→]"
	} else if todo.Status == "blocked" {
		status = "[✗]"
	}
	
	priority := ""
	if todo.Priority == "high" {
		priority = " [HIGH]"
	} else if todo.Priority == "low" {
		priority = " [LOW]"
	}
	
	return fmt.Sprintf("%s %s: %s%s", status, todo.ID, todo.Task, priority)
}

// FormatTodoUpdateResponse formats the response for todo_update
func FormatTodoUpdateResponse(todoID string, section string, operation string) *mcp.CallToolResult {
	message := fmt.Sprintf("Todo '%s' updated successfully", todoID)
	if section != "" {
		message = fmt.Sprintf("Todo '%s' %s section updated (%s)", todoID, section, operation)
	}
	
	return mcp.NewToolResultText(message)
}

// FormatSearchResult represents a search result
type FormatSearchResult struct {
	ID       string  `json:"id"`
	Task     string  `json:"task"`
	Score    float64 `json:"score"`
	Snippet  string  `json:"snippet,omitempty"`
}

// FormatTodoSearchResponse formats the response for todo_search
func FormatTodoSearchResponse(results []core.SearchResult) *mcp.CallToolResult {
	if len(results) == 0 {
		return mcp.NewToolResultText("No todos found matching your search")
	}
	
	var formatted []FormatSearchResult
	for _, r := range results {
		formatted = append(formatted, FormatSearchResult{
			ID:       r.ID,
			Task:     r.Task,
			Score:    r.Score,
			Snippet:  r.Snippet,
		})
	}
	
	// Add summary
	summary := fmt.Sprintf("Found %d todos matching your search:\n", len(results))
	
	jsonData, _ := json.MarshalIndent(formatted, "", "  ")
	return mcp.NewToolResultText(summary + string(jsonData))
}

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

// FormatTodoTemplateResponse formats the response for todo_template
func FormatTodoTemplateResponse(todo *core.Todo, filePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":       todo.ID,
		"path":     filePath,
		"template": "applied",
		"message":  fmt.Sprintf("Todo created from template: %s", todo.ID),
	}
	
	jsonData, _ := json.MarshalIndent(response, "", "  ")
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

// FormatTodoSectionsResponse formats todo sections response
func FormatTodoSectionsResponse(todo *core.Todo) *mcp.CallToolResult {
	var response strings.Builder
	
	response.WriteString(fmt.Sprintf("Todo: %s\n\n", todo.ID))
	response.WriteString("Sections:\n")
	
	if todo.Sections == nil || len(todo.Sections) == 0 {
		response.WriteString("No section metadata defined (legacy todo)\n")
	} else {
		// Get sections in order
		ordered := core.GetOrderedSections(todo.Sections)
		
		for _, section := range ordered {
			response.WriteString(fmt.Sprintf("\n%s:\n", section.Key))
			response.WriteString(fmt.Sprintf("  title: %s\n", section.Definition.Title))
			response.WriteString(fmt.Sprintf("  order: %d\n", section.Definition.Order))
			response.WriteString(fmt.Sprintf("  schema: %s\n", section.Definition.Schema))
			response.WriteString(fmt.Sprintf("  required: %v\n", section.Definition.Required))
			
			if section.Definition.Custom {
				response.WriteString("  custom: true\n")
			}
			
			if section.Definition.Metadata != nil && len(section.Definition.Metadata) > 0 {
				response.WriteString("  metadata:\n")
				for k, v := range section.Definition.Metadata {
					response.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
				}
			}
		}
	}
	
	return mcp.NewToolResultText(response.String())
}

// FormatTodoSectionsResponseWithContent formats todo sections response with content status
func FormatTodoSectionsResponseWithContent(todo *core.Todo, content string) *mcp.CallToolResult {
	var response strings.Builder
	
	response.WriteString(fmt.Sprintf("Todo: %s\n\n", todo.ID))
	response.WriteString("Sections:\n")
	
	// Parse content to extract sections
	sectionContents := extractSectionContents(content)
	
	if todo.Sections == nil || len(todo.Sections) == 0 {
		// Legacy todo - infer sections from markdown
		todo.Sections = core.InferSectionsFromMarkdown(content)
	}
	
	if todo.Sections == nil || len(todo.Sections) == 0 {
		response.WriteString("No sections found\n")
	} else {
		// Get sections in order
		ordered := core.GetOrderedSections(todo.Sections)
		
		for _, section := range ordered {
			response.WriteString(fmt.Sprintf("\n%s:\n", section.Key))
			response.WriteString(fmt.Sprintf("  title: %s\n", section.Definition.Title))
			response.WriteString(fmt.Sprintf("  order: %d\n", section.Definition.Order))
			response.WriteString(fmt.Sprintf("  schema: %s\n", section.Definition.Schema))
			response.WriteString(fmt.Sprintf("  required: %v\n", section.Definition.Required))
			
			// Add content status
			sectionContent, exists := sectionContents[section.Definition.Title]
			if exists {
				// Check if content is not just whitespace
				trimmed := strings.TrimSpace(sectionContent)
				hasContent := len(trimmed) > 0
				response.WriteString(fmt.Sprintf("  hasContent: %v\n", hasContent))
				response.WriteString(fmt.Sprintf("  contentLength: %d\n", len(trimmed)))
			} else {
				response.WriteString("  hasContent: false\n")
				response.WriteString("  contentLength: 0\n")
			}
			
			if section.Definition.Custom {
				response.WriteString("  custom: true\n")
			}
			
			if section.Definition.Metadata != nil && len(section.Definition.Metadata) > 0 {
				response.WriteString("  metadata:\n")
				for k, v := range section.Definition.Metadata {
					response.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
				}
			}
		}
	}
	
	return mcp.NewToolResultText(response.String())
}

// extractSectionContents parses markdown content and extracts section contents
func extractSectionContents(content string) map[string]string {
	sections := make(map[string]string)
	
	// Split content into lines
	lines := strings.Split(content, "\n")
	
	currentSection := ""
	var sectionContent strings.Builder
	inFrontmatter := false
	
	for _, line := range lines {
		// Handle frontmatter
		if line == "---" {
			if !inFrontmatter && currentSection == "" {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				continue
			}
		}
		
		if inFrontmatter {
			continue
		}
		
		// Check for section header (## Title)
		if strings.HasPrefix(line, "## ") {
			// Save previous section if any
			if currentSection != "" {
				sections[currentSection] = sectionContent.String()
			}
			
			// Start new section
			currentSection = line
			sectionContent.Reset()
		} else if currentSection != "" {
			// Add line to current section
			sectionContent.WriteString(line)
			sectionContent.WriteString("\n")
		}
	}
	
	// Save last section
	if currentSection != "" {
		sections[currentSection] = sectionContent.String()
	}
	
	return sections
}