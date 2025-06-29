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
					// Raw content for other sections
					sectionData[key] = sectionContent
				}
			}
			
			data["sections"] = sectionData
		}
		
		results = append(results, data)
	}
	
	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// formatSingleTodoWithContent formats a single todo with full content
func formatSingleTodoWithContent(todo *core.Todo, content string, format string) *mcp.CallToolResult {
	if format != "full" {
		// Use existing formatSingleTodo for other formats
		return formatSingleTodo(todo, format)
	}
	
	// Full format with content
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
	sections := extractSectionContents(content)
	sectionData := make(map[string]interface{})
	
	for key, sectionContent := range sections {
		if key == "checklist" {
			// Parse checklist items
			sectionData[key] = core.ParseChecklist(sectionContent)
		} else {
			// Raw content for other sections
			sectionData[key] = sectionContent
		}
	}
	
	data["sections"] = sectionData
	
	jsonData, _ := json.MarshalIndent(data, "", "  ")
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
	if len(todos) == 0 {
		return mcp.NewToolResultText("No todos found")
	}
	
	var result strings.Builder
	
	// Check if we have any parent-child relationships
	hasHierarchy := false
	for _, todo := range todos {
		if todo.ParentID != "" || todo.Type == "multi-phase" || todo.Type == "phase" || todo.Type == "subtask" {
			hasHierarchy = true
			break
		}
	}
	
	// If we have hierarchical relationships, show tree view first
	if hasHierarchy {
		roots, orphans := core.BuildTodoHierarchy(todos)
		
		// Only show tree if we have actual hierarchy
		if len(roots) > 0 || len(orphans) > 0 {
			formatter := core.NewTreeFormatter()
			treeView := formatter.FormatHierarchy(roots, orphans)
			
			if treeView != "" {
				result.WriteString("HIERARCHICAL VIEW:\n")
				result.WriteString(treeView)
				result.WriteString("\n")
			}
		}
	}
	
	// Then show traditional grouped view
	result.WriteString("\nGROUPED BY STATUS:")
	
	// Group by status
	statusGroups := make(map[string][]*core.Todo)
	for _, todo := range todos {
		statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
	}
	
	// Format by status
	for _, status := range []string{"in_progress", "blocked", "completed"} {
		if todos, ok := statusGroups[status]; ok && len(todos) > 0 {
			result.WriteString(fmt.Sprintf("\n\n%s (%d):\n", strings.ToUpper(status), len(todos)))
			for _, todo := range todos {
				result.WriteString(formatTodoSummaryLine(todo) + "\n")
			}
		}
	}
	
	return mcp.NewToolResultText(result.String())
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

// TodoUpdateResponse represents the enriched response after updating a todo
type TodoUpdateResponse struct {
	Message  string                 `json:"message"`
	Todo     *TodoSummary          `json:"todo"`
	Progress map[string]interface{} `json:"progress,omitempty"`
}

// TodoSummary represents a summary of a todo with parsed content
type TodoSummary struct {
	ID        string                    `json:"id"`
	Task      string                    `json:"task"`
	Status    string                    `json:"status"`
	Priority  string                    `json:"priority"`
	Type      string                    `json:"type"`
	ParentID  string                    `json:"parent_id,omitempty"`
	Checklist []core.ChecklistItem      `json:"checklist,omitempty"`
	Sections  map[string]SectionSummary `json:"sections,omitempty"`
}

// SectionSummary represents a summary of a section
type SectionSummary struct {
	Title      string `json:"title"`
	HasContent bool   `json:"hasContent"`
	WordCount  int    `json:"wordCount,omitempty"`
}

// FormatTodoUpdateResponse formats the response for todo_update
func FormatTodoUpdateResponse(todoID string, section string, operation string) *mcp.CallToolResult {
	message := fmt.Sprintf("Todo '%s' updated successfully", todoID)
	if section != "" {
		message = fmt.Sprintf("Todo '%s' %s section updated (%s)", todoID, section, operation)
	}
	
	return mcp.NewToolResultText(message)
}

// FormatEnrichedTodoUpdateResponse formats an enriched response with full todo data
func FormatEnrichedTodoUpdateResponse(todo *core.Todo, content string, section string, operation string) *mcp.CallToolResult {
	message := fmt.Sprintf("Todo '%s' updated successfully", todo.ID)
	if section != "" {
		message = fmt.Sprintf("Todo '%s' %s section updated (%s)", todo.ID, section, operation)
	}
	
	summary := &TodoSummary{
		ID:       todo.ID,
		Task:     todo.Task,
		Status:   todo.Status,
		Priority: todo.Priority,
		Type:     todo.Type,
		ParentID: todo.ParentID,
	}
	
	// Parse sections from content
	sections := make(map[string]SectionSummary)
	sectionContents := extractSectionContents(content)
	
	// Check for checklist content
	if checklistContent, exists := sectionContents["checklist"]; exists {
		summary.Checklist = core.ParseChecklist(checklistContent)
	}
	
	// Add section summaries
	for key, sectionContent := range sectionContents {
		sections[key] = SectionSummary{
			Title:      fmt.Sprintf("## %s", strings.Title(strings.ReplaceAll(key, "_", " "))),
			HasContent: len(strings.TrimSpace(sectionContent)) > 0,
			WordCount:  len(strings.Fields(sectionContent)),
		}
	}
	summary.Sections = sections
	
	// Calculate progress
	progress := make(map[string]interface{})
	if len(summary.Checklist) > 0 {
		completed := 0
		inProgress := 0
		pending := 0
		
		for _, item := range summary.Checklist {
			switch item.Status {
			case "completed":
				completed++
			case "in_progress":
				inProgress++
			case "pending":
				pending++
			}
		}
		
		total := len(summary.Checklist)
		completionPercentage := 0
		if total > 0 {
			completionPercentage = (completed * 100) / total
		}
		
		progress["checklist"] = fmt.Sprintf("%d/%d completed (%d%%)", completed, total, completionPercentage)
		progress["checklist_breakdown"] = map[string]int{
			"pending":     pending,
			"in_progress": inProgress,
			"completed":   completed,
			"total":       total,
		}
	}
	
	// Count sections with content
	sectionsWithContent := 0
	totalSections := len(sections)
	for _, section := range sections {
		if section.HasContent {
			sectionsWithContent++
		}
	}
	if totalSections > 0 {
		progress["sections"] = fmt.Sprintf("%d/%d sections have content", sectionsWithContent, totalSections)
	}
	
	response := TodoUpdateResponse{
		Message:  message,
		Todo:     summary,
		Progress: progress,
	}
	
	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}

// extractSectionContents extracts content for each section from the full todo content
func extractSectionContents(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")
	
	currentSection := ""
	currentContent := []string{}
	
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.Join(currentContent, "\n")
			}
			
			// Start new section
			title := strings.TrimPrefix(line, "## ")
			currentSection = normalizeKey(title)
			currentContent = []string{}
		} else if currentSection != "" {
			currentContent = append(currentContent, line)
		}
	}
	
	// Save last section
	if currentSection != "" {
		sections[currentSection] = strings.Join(currentContent, "\n")
	}
	
	return sections
}

// normalizeKey converts a section title to a normalized key
func normalizeKey(title string) string {
	// Handle common mappings
	switch title {
	case "Findings & Research":
		return "findings"
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

