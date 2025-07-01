package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
)

// TodoUpdateResponse represents the enriched response after updating a todo
type TodoUpdateResponse struct {
	Message  string                 `json:"message"`
	Todo     *TodoSummary           `json:"todo"`
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

// FormatTodoSectionsResponse formats the response for todo_sections
func FormatTodoSectionsResponse(todo *core.Todo) *mcp.CallToolResult {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Todo: %s\n", todo.ID))
	sb.WriteString("Sections:\n")

	// Handle legacy todos with no sections
	if todo.Sections == nil || len(todo.Sections) == 0 {
		sb.WriteString("  No section metadata defined (legacy todo)\n")
		return mcp.NewToolResultText(sb.String())
	}

	// Sort sections by key for consistent output
	var keys []string
	for key := range todo.Sections {
		keys = append(keys, key)
	}
	// Simple sorting for now
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Format each section
	for _, key := range keys {
		section := todo.Sections[key]
		if section == nil {
			continue
		}
		
		sb.WriteString(fmt.Sprintf("%s:\n", key))
		sb.WriteString(fmt.Sprintf("  title: %s\n", section.Title))
		sb.WriteString(fmt.Sprintf("  order: %d\n", section.Order))
		sb.WriteString(fmt.Sprintf("  schema: %s\n", section.Schema))
		sb.WriteString(fmt.Sprintf("  required: %v\n", section.Required))
		
		if section.Custom {
			sb.WriteString("  custom: true\n")
		}
		
		// Handle metadata if present
		if section.Metadata != nil && len(section.Metadata) > 0 {
			sb.WriteString("  metadata:\n")
			for mKey, mValue := range section.Metadata {
				sb.WriteString(fmt.Sprintf("    %s: %v\n", mKey, mValue))
			}
		}
	}

	return mcp.NewToolResultText(sb.String())
}

// FormatTodoSectionsResponseWithContent formats sections with content status
func FormatTodoSectionsResponseWithContent(todo *core.Todo, content string) *mcp.CallToolResult {
	// Handle legacy todos with no sections
	if todo.Sections == nil || len(todo.Sections) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No sections found for todo '%s'", todo.ID))
	}
	
	sections := make(map[string]interface{})

	// Parse content to check which sections have content
	sectionContents := extractSectionContents(content)

	// Add existing sections with content status
	for key, section := range todo.Sections {
		sectionData := map[string]interface{}{
			"title":    section.Title,
			"schema":   section.Schema,
			"required": section.Required,
			"order":    section.Order,
		}

		// Check if section has content
		if sectionContent, exists := sectionContents[key]; exists {
			hasContent := len(strings.TrimSpace(sectionContent)) > 0
			sectionData["hasContent"] = hasContent
			if hasContent {
				sectionData["wordCount"] = len(strings.Fields(sectionContent))
			}
		} else {
			sectionData["hasContent"] = false
		}

		sections[key] = sectionData
	}

	response := map[string]interface{}{
		"id":       todo.ID,
		"task":     todo.Task,
		"sections": sections,
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(jsonData))
}