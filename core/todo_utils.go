package core

import (
	"fmt"
	"strings"
	"time"
)

// generateBaseID creates a kebab-case ID from the task description
func generateBaseID(task string) string {
	// Remove null bytes and other invalid characters first
	cleaned := strings.ReplaceAll(task, "\x00", "")

	// Convert to lowercase
	lower := strings.ToLower(cleaned)

	// Replace spaces and special characters with hyphens
	// Keep numbers and dots for version numbers
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		":", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"/", "-",
		"\\", "-",
		"\"", "",
		"'", "",
		"`", "",
		"~", "",
		"!", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"^", "",
		"&", "",
		"*", "",
		"+", "",
		"=", "",
		"|", "",
		";", "",
		",", "",
		"<", "",
		">", "",
		"?", "",
		"\n", "-",
		"\r", "-",
		"\t", "-",
	)

	id := replacer.Replace(lower)

	// Replace multiple hyphens with single hyphen
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}

	// Trim hyphens from start and end
	id = strings.Trim(id, "-")

	// Limit length
	if len(id) > 50 {
		id = id[:50]
	}

	// Ensure we have something
	if id == "" {
		id = "todo"
	}

	return id
}

// parseTimestamp parses various timestamp formats
func parseTimestamp(timestamp string) (time.Time, error) {
	// Try different formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

// extractTask extracts the task from the markdown content
func extractTask(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# Task:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "# Task:"))
		}
	}
	return ""
}

// toggleChecklistItem toggles a checklist item between states
func toggleChecklistItem(content string, itemText string) string {
	lines := strings.Split(content, "\n")
	inChecklist := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if we're in the checklist section
		if strings.HasPrefix(trimmed, "## Checklist") {
			inChecklist = true
			continue
		}
		
		// If we hit another section, we're out of the checklist
		if inChecklist && strings.HasPrefix(trimmed, "##") {
			inChecklist = false
			continue
		}
		
		// Only process checklist items in the checklist section
		if inChecklist && (strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]")) {
			// Extract the item text
			itemContent := strings.TrimSpace(trimmed[5:])
			if itemContent == itemText {
				// Toggle the checkbox
				if strings.HasPrefix(trimmed, "- [ ]") {
					lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
				} else {
					lines[i] = strings.Replace(line, "- [x]", "- [ ]", 1)
				}
				break
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

// ParseChecklist parses checklist items from content
func ParseChecklist(content string) []ChecklistItem {
	var items []ChecklistItem
	lines := strings.Split(content, "\n")
	inChecklist := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if we're in the checklist section
		if strings.HasPrefix(trimmed, "## Checklist") {
			inChecklist = true
			continue
		}
		
		// If we hit another section, we're out of the checklist
		if inChecklist && strings.HasPrefix(trimmed, "##") {
			break
		}
		
		// Parse checklist items
		if inChecklist {
			if strings.HasPrefix(trimmed, "- [ ]") {
				items = append(items, ChecklistItem{
					Text:   strings.TrimSpace(trimmed[5:]),
					Status: "pending",
				})
			} else if strings.HasPrefix(trimmed, "- [x]") {
				items = append(items, ChecklistItem{
					Text:   strings.TrimSpace(trimmed[5:]),
					Status: "completed",
				})
			}
		}
	}
	
	return items
}

// formatWithTimestamp adds a timestamp to content
func formatWithTimestamp(content string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// Handle different content types
	lines := strings.Split(strings.TrimSpace(content), "\n")
	
	// Check if it's a test result or command output
	if len(lines) > 0 && (strings.HasPrefix(lines[0], "#") || strings.Contains(lines[0], "```")) {
		// It's already formatted or is a code block
		return fmt.Sprintf("[%s] %s", timestamp, content)
	}
	
	// For regular content, just prepend timestamp
	return fmt.Sprintf("[%s] %s", timestamp, content)
}

// FindDuplicateTodos finds todos with similar tasks
func (tm *TodoManager) FindDuplicateTodos() ([][]string, error) {
	todos, err := tm.ListTodos("", "", 0)
	if err != nil {
		return nil, err
	}

	// Group by normalized task
	groups := make(map[string][]string)
	for _, todo := range todos {
		// Normalize task for comparison
		normalized := strings.ToLower(strings.TrimSpace(todo.Task))
		groups[normalized] = append(groups[normalized], todo.ID)
	}

	// Find groups with duplicates
	var duplicates [][]string
	for _, group := range groups {
		if len(group) > 1 {
			duplicates = append(duplicates, group)
		}
	}

	return duplicates, nil
}

// ArchiveOldTodos archives todos older than specified days
func (tm *TodoManager) ArchiveOldTodos(days int) (int, error) {
	todos, err := tm.ListTodos("completed", "", days)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, todo := range todos {
		if todo.Status == "completed" {
			err = tm.ArchiveTodo(todo.ID, "")
			if err == nil {
				count++
			}
		}
	}

	return count, nil
}