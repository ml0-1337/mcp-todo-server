package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
	"time"
)

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

// formatSingleTodoWithContent formats a single todo with its full content
func formatSingleTodoWithContent(todo *core.Todo, content string, format string) *mcp.CallToolResult {
	if format == "full" {
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
				// Regular section content - just the string
				sectionData[key] = strings.TrimSpace(sectionContent)
			}
		}
		data["sections"] = sectionData

		jsonData, _ := json.MarshalIndent(data, "", "  ")
		// For full JSON format, return pure JSON without guidance
		return mcp.NewToolResultText(string(jsonData))
	}

	// For non-full formats, use regular formatting
	return formatSingleTodo(todo, format)
}

// formatSingleTodo formats a single todo based on format
func formatSingleTodo(todo *core.Todo, format string) *mcp.CallToolResult {
	if format == "full" {
		// For full format, we'd need to read the entire file content
		// For now, return structured data
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

		jsonData, _ := json.MarshalIndent(data, "", "  ")
		// For full JSON format, return pure JSON without guidance
		return mcp.NewToolResultText(string(jsonData))
	}

	// Summary format - add guidance for non-JSON formats
	summaryLine := formatTodoSummaryLine(todo)
	return addSingleTodoGuidance(summaryLine, todo)
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
	if len(todos) == 0 {
		return mcp.NewToolResultText("No todos found")
	}

	var lines []string
	for _, todo := range todos {
		lines = append(lines, fmt.Sprintf("- %s: %s", todo.ID, todo.Task))
	}

	return mcp.NewToolResultText(strings.Join(lines, "\n"))
}

// formatTodosSummary formats todos with detailed summary
func formatTodosSummary(todos []*core.Todo) *mcp.CallToolResult {
	if len(todos) == 0 {
		return mcp.NewToolResultText("No todos found")
	}

	// Check if any todos have parent relationships
	hasHierarchy := false
	for _, todo := range todos {
		if todo.ParentID != "" {
			hasHierarchy = true
			break
		}
	}

	if hasHierarchy {
		// Use hierarchical view
		roots, orphans := core.BuildTodoHierarchy(todos)
		
		var lines []string
		lines = append(lines, "HIERARCHICAL VIEW:")
		lines = append(lines, "")
		
		// Format root todos with their children using tree structure
		for _, root := range roots {
			// Format the root todo
			lines = append(lines, formatTodoSummaryLineWithOptions(root.Todo, false, true))
			
			// Format children with tree branches
			for i, child := range root.Children {
				isLast := i == len(root.Children)-1
				childLines := formatTodoNodeTreeInternal(child, "", isLast, true)
				for _, line := range strings.Split(childLines, "\n") {
					lines = append(lines, line)
				}
			}
		}
		
		// Show orphaned todos if any
		if len(orphans) > 0 {
			lines = append(lines, "")
			lines = append(lines, "ORPHANED PHASES/SUBTASKS:")
			for _, todo := range orphans {
				// Show type for orphans but add "not found" to parent
				line := formatTodoSummaryLineWithOptions(todo, true, true)
				// Add "not found" to parent reference for orphans
				if todo.ParentID != "" {
					line = strings.Replace(line, "(parent: "+todo.ParentID+")", "(parent: "+todo.ParentID+" not found)", 1)
				}
				lines = append(lines, line)
			}
		}
		
		// Also show grouped view
		lines = append(lines, "")
		lines = append(lines, "GROUPED BY STATUS:")
		
		// Group todos by status
		statusGroups := make(map[string][]*core.Todo)
		for _, todo := range todos {
			statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
		}
		
		// Show each status group
		statusOrder := []string{"in_progress", "completed", "blocked"}
		for _, status := range statusOrder {
			if todosInStatus, exists := statusGroups[status]; exists && len(todosInStatus) > 0 {
				lines = append(lines, "")
				lines = append(lines, fmt.Sprintf("%s (%d):", strings.ToUpper(status), len(todosInStatus)))
				for _, todo := range todosInStatus {
					lines = append(lines, "  "+formatTodoSummaryLineWithOptions(todo, true, true))
				}
			}
		}
		
		result := strings.Join(lines, "\n")
		return addMultiTodoGuidance(result, todos)
	}

	// No hierarchy - use grouped view
	var lines []string
	lines = append(lines, "GROUPED BY STATUS:")
	
	// Group todos by status
	statusGroups := make(map[string][]*core.Todo)
	for _, todo := range todos {
		statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
	}
	
	// Show each status group
	statusOrder := []string{"in_progress", "completed", "blocked"}
	for _, status := range statusOrder {
		if todosInStatus, exists := statusGroups[status]; exists && len(todosInStatus) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("%s (%d):", strings.ToUpper(status), len(todosInStatus)))
			for _, todo := range todosInStatus {
				lines = append(lines, "  "+formatTodoSummaryLine(todo))
			}
		}
	}

	result := strings.Join(lines, "\n")
	return addMultiTodoGuidance(result, todos)
}

// formatTodoSummaryLine formats a single todo as a summary line
func formatTodoSummaryLine(todo *core.Todo) string {
	return formatTodoSummaryLineWithOptions(todo, true, true)
}

// formatTodoSummaryLineWithOptions formats with control over parent and type display
func formatTodoSummaryLineWithOptions(todo *core.Todo, showParent bool, showType bool) string {
	status := getStatusIcon(todo.Status)
	priority := getPriorityLabel(todo.Priority)
	
	// Format: [status] id: task [priority] [type]
	line := fmt.Sprintf("%s %s: %s", status, todo.ID, todo.Task)
	
	// Add priority label if not medium (medium is default, so omitted)
	if priority != "" {
		line += " " + priority
	}
	
	// Add type if requested and not empty
	if showType && todo.Type != "" {
		line += fmt.Sprintf(" [%s]", todo.Type)
	}
	
	// Add parent if requested and not empty
	if showParent && todo.ParentID != "" {
		line += fmt.Sprintf(" (parent: %s)", todo.ParentID)
	}
	
	return line
}

// formatTodoNodeTree formats a todo node and its children as a tree
func formatTodoNodeTree(node *core.TodoNode, prefix string, isLast bool) string {
	return formatTodoNodeTreeInternal(node, prefix, isLast, false)
}

// formatTodoNodeTreeInternal formats a todo node with special handling for first level
func formatTodoNodeTreeInternal(node *core.TodoNode, prefix string, isLast bool, forceTreeSymbol bool) string {
	var lines []string
	
	// Format current node
	line := prefix
	if prefix != "" || forceTreeSymbol {
		if isLast {
			line += "└── "
		} else {
			line += "├── "
		}
	}
	// Don't show parent in tree view since it's already shown by structure
	line += formatTodoSummaryLineWithOptions(node.Todo, false, true)
	lines = append(lines, line)
	
	// Format children  
	childPrefix := prefix
	if prefix != "" || forceTreeSymbol {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}
	
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		lines = append(lines, formatTodoNodeTreeInternal(child, childPrefix, childIsLast, false))
	}
	
	return strings.Join(lines, "\n")
}

// FormatSearchResult represents a search result
type FormatSearchResult struct {
	ID      string  `json:"id"`
	Task    string  `json:"task"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet,omitempty"`
}

// FormatTodoSearchResponse formats search results
func FormatTodoSearchResponse(results []core.SearchResult) *mcp.CallToolResult {
	if len(results) == 0 {
		return mcp.NewToolResultText("No matching todos found")
	}

	var lines []string
	for _, result := range results {
		score := fmt.Sprintf("%.0f%%", result.Score*100)
		snippet := result.Snippet
		if snippet == "" {
			snippet = "No content preview available"
		}
		
		lines = append(lines, fmt.Sprintf("• %s (relevance: %s)\n  %s", 
			result.ID, score, snippet))
	}

	header := fmt.Sprintf("Found %d matching todos:\n", len(results))
	return mcp.NewToolResultText(header + strings.Join(lines, "\n\n"))
}

// addMultiTodoGuidance adds contextual guidance for multiple todos view
func addMultiTodoGuidance(todoList string, todos []*core.Todo) *mcp.CallToolResult {
	var guidance strings.Builder
	guidance.WriteString(todoList)
	guidance.WriteString("\n\n")
	
	// Count todos by status
	statusCounts := make(map[string]int)
	var blockedTodos []string
	var oldestInProgress *core.Todo
	
	for _, todo := range todos {
		statusCounts[todo.Status]++
		if todo.Status == "blocked" {
			blockedTodos = append(blockedTodos, todo.ID)
		}
		if todo.Status == "in_progress" && (oldestInProgress == nil || todo.Started.Before(oldestInProgress.Started)) {
			oldestInProgress = todo
		}
	}
	
	inProgressCount := statusCounts["in_progress"]
	blockedCount := statusCounts["blocked"]
	
	guidance.WriteString(fmt.Sprintf("Current workload: %d active tasks", inProgressCount))
	if blockedCount > 0 {
		guidance.WriteString(fmt.Sprintf(", %d blocked", blockedCount))
	}
	guidance.WriteString("\n\n")
	
	guidance.WriteString("Workflow suggestions:\n")
	
	// Priority guidance based on workload
	if blockedCount > 0 {
		guidance.WriteString(fmt.Sprintf("- Address blocked items first to unblock progress: %s\n", strings.Join(blockedTodos, ", ")))
	}
	
	if inProgressCount > 3 {
		guidance.WriteString("- Consider completing existing work before starting new tasks (high WIP detected)\n")
		if oldestInProgress != nil {
			guidance.WriteString(fmt.Sprintf("- Focus on '%s' - it's been in progress the longest\n", oldestInProgress.ID))
		}
	} else if inProgressCount == 0 && len(todos) > 0 {
		guidance.WriteString("- No tasks currently in progress. Choose the highest priority item to start\n")
	}
	
	// Check for priority issues
	highPriorityCount := 0
	for _, todo := range todos {
		if todo.Priority == "high" && todo.Status == "in_progress" {
			highPriorityCount++
		}
	}
	
	if highPriorityCount > 1 {
		guidance.WriteString("- Multiple high-priority items in progress. Consider if they're all truly urgent\n")
	}
	
	guidance.WriteString("\nWhich task requires your immediate attention?")
	
	return mcp.NewToolResultText(guidance.String())
}

// addSingleTodoGuidance adds contextual guidance for single todo view
func addSingleTodoGuidance(todoInfo string, todo *core.Todo) *mcp.CallToolResult {
	var guidance strings.Builder
	guidance.WriteString(todoInfo)
	guidance.WriteString("\n\n")
	
	switch todo.Status {
	case "in_progress":
		guidance.WriteString("Task is in progress. To maintain momentum:\n")
		guidance.WriteString("- Focus on the current test or next checklist item\n")
		guidance.WriteString("- Document findings as you work\n")
		guidance.WriteString("- Avoid context switching until a logical stopping point\n\n")
		guidance.WriteString("What specific progress can you make on this task right now?")
		
	case "blocked":
		guidance.WriteString("Task is blocked. To resolve efficiently:\n")
		guidance.WriteString("- Document the specific blocker in the findings section\n")
		guidance.WriteString("- Identify who can help or what information is needed\n")
		guidance.WriteString("- Consider if there's a workaround or alternative approach\n\n")
		guidance.WriteString("Can you describe the blocker in detail?")
		
	default: // new/not started
		guidance.WriteString("Task ready to begin. To start effectively:\n")
		guidance.WriteString("- Review the task description and any existing notes\n")
		guidance.WriteString("- Add acceptance criteria or break down into concrete steps\n")
		guidance.WriteString("- Consider dependencies or prerequisites\n\n")
		guidance.WriteString("What's the first concrete action to move this forward?")
	}
	
	return mcp.NewToolResultText(guidance.String())
}