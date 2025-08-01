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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// getUpdatePrompts returns contextual prompts based on section, operation, and todo type
func getUpdatePrompts(section string, operation string, todoType string) string {
	// Handle different sections
	switch section {
	case "findings":
		return getFindingsPrompts(operation)
	case "tests":
		return getTestsPrompts(operation, todoType)
	case "checklist":
		return getChecklistPrompts(operation)
	case "scratchpad":
		return getScratchpadPrompts()
	default:
		// Generic section update
		return getGenericSectionPrompts(operation)
	}
}

// getFindingsPrompts returns prompts for findings section updates
func getFindingsPrompts(operation string) string {
	if operation == "replace" {
		return "Findings updated. To ensure clarity:\n\n" +
			"- Does this capture the essential insights?\n" +
			"- Are the implications for next steps clear?\n" +
			"- Have you documented any unexpected discoveries?\n\n" +
			"Review the updated findings for completeness."
	}

	// append/prepend
	return "Content added to findings. To maximize value:\n\n" +
		"- Have you captured the \"why\" behind your discoveries?\n" +
		"- Are there implications for other parts of the system?\n" +
		"- What follow-up questions emerged?\n\n" +
		"Continue documenting insights that will help future work."
}

// getTestsPrompts returns prompts for test section updates
func getTestsPrompts(operation string, todoType string) string {
	if todoType == "bug" {
		if operation == "replace" {
			return "Test updated. To ensure the bug is properly caught:\n\n" +
				"- Does this test fail without the fix?\n" +
				"- Will it catch similar bugs in the future?\n" +
				"- Is the assertion specific to the bug behavior?\n\n" +
				"Verify the test captures the exact failure condition."
		}
		return "Test added. To prevent regression:\n\n" +
			"- Does this test reproduce the bug reliably?\n" +
			"- Is the expected vs actual behavior clear?\n" +
			"- Have you considered related edge cases?\n\n" +
			"What other scenarios might trigger this bug?"
	}

	// Feature/general tests
	if operation == "replace" {
		return "Test updated. To maintain test quality:\n\n" +
			"- Is the test name descriptive of what it validates?\n" +
			"- Are the assertions clear and specific?\n" +
			"- Does it test behavior, not implementation?\n\n" +
			"Review the test for clarity and coverage."
	}

	return "Test added. To ensure comprehensive coverage:\n\n" +
		"- Does this test verify the expected behavior clearly?\n" +
		"- Are edge cases and error conditions considered?\n" +
		"- Is the test independent and repeatable?\n\n" +
		"What's the next test to implement?"
}

// getChecklistPrompts returns prompts for checklist updates
func getChecklistPrompts(operation string) string {
	if operation == "toggle" {
		return "Checklist item toggled. To maintain progress:\n\n" +
			"- Does completing this reveal new tasks?\n" +
			"- Are remaining items still relevant?\n" +
			"- Should priorities be adjusted?\n\n" +
			"Focus on the next highest priority item."
	}

	if operation == "replace" {
		return "Checklist updated. To stay organized:\n\n" +
			"- Are all items specific and actionable?\n" +
			"- Is the order logical (priority/dependency)?\n" +
			"- Do you need to break down any large items?\n\n" +
			"Review the updated checklist for clarity."
	}

	// append/prepend
	return "Checklist updated. To maintain momentum:\n\n" +
		"- Is each item specific and actionable?\n" +
		"- Are tasks ordered by priority or dependency?\n" +
		"- Do completed items reveal new tasks?\n\n" +
		"Focus on the next unchecked item."
}

// getScratchpadPrompts returns prompts for scratchpad updates
func getScratchpadPrompts() string {
	return "Scratchpad updated. Quick notes captured:\n\n" +
		"- Should any of these notes move to findings?\n" +
		"- Are there action items to add to the checklist?\n" +
		"- Do these notes suggest new test cases?\n\n" +
		"The scratchpad is great for rapid capture - refine when ready."
}

// getGenericSectionPrompts returns prompts for custom sections
func getGenericSectionPrompts(operation string) string {
	if operation == "replace" {
		return "Section updated. To ensure quality:\n\n" +
			"- Is the content clear and well-organized?\n" +
			"- Does it serve the section's purpose?\n" +
			"- Are there gaps to address?\n\n" +
			"Review the section for completeness."
	}

	return "Content added. To maintain quality:\n\n" +
		"- Does this enhance the section's value?\n" +
		"- Is the information well-structured?\n" +
		"- Are there related updates needed?\n\n" +
		"Continue building comprehensive documentation."
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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}
