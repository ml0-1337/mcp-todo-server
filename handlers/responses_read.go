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
		return mcp.NewToolResultText(string(jsonData))
	}

	// Summary format
	summary := formatTodoSummaryLine(todo)

	// Add single todo prompt
	prompt := getSingleTodoPrompt(todo)
	return mcp.NewToolResultText(summary + prompt)
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
		prompt := getReadPrompts(todos, "summary")
		return mcp.NewToolResultText("No todos found" + prompt)
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

		prompt := getReadPrompts(todos, "summary")
		return mcp.NewToolResultText(strings.Join(lines, "\n") + prompt)
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

	prompt := getReadPrompts(todos, "summary")
	return mcp.NewToolResultText(strings.Join(lines, "\n") + prompt)
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

// getReadPrompts returns contextual prompts based on todo status and count
func getReadPrompts(todos []*core.Todo, format string) string {
	// Don't add prompts to JSON or list formats
	if format == "list" || format == "full" {
		return ""
	}

	if len(todos) == 0 {
		return "\n\nNo todos found. To get started:\n\n" +
			"- What task or project would you like to begin?\n" +
			"- Is this a feature, bug fix, research, or refactor?\n" +
			"- Are there multiple phases that need planning?\n\n" +
			"Use todo_create to start tracking your work."
	}

	// Count todos by status
	statusCounts := make(map[string]int)
	for _, todo := range todos {
		statusCounts[todo.Status]++
	}

	// Generate status-specific prompts
	if statusCounts["blocked"] > 0 {
		return fmt.Sprintf("\n\nYou have %d blocked todo(s). To unblock progress:\n\n"+
			"- What specific issues are preventing completion?\n"+
			"- Are there alternative approaches to consider?\n"+
			"- Do you need to create separate todos for blockers?\n\n"+
			"Review blocked todos and update with findings or create linked dependencies.",
			statusCounts["blocked"])
	}

	if statusCounts["in_progress"] > 0 {
		return fmt.Sprintf("\n\nYou have %d todo(s) in progress. To maintain momentum:\n\n"+
			"- Which todo should be your primary focus?\n"+
			"- Are you following the test list or implementation plan?\n"+
			"- Any blockers emerging that need attention?\n\n"+
			"Use todo_update to track progress on your active tasks.",
			statusCounts["in_progress"])
	}

	// All completed
	return "\n\nAll todos are completed! To continue being productive:\n\n" +
		"- Are there follow-up tasks from completed work?\n" +
		"- Any technical debt or improvements to address?\n" +
		"- New features or bugs reported?\n\n" +
		"Consider archiving completed todos or creating new ones."
}

// getSingleTodoPrompt returns contextual prompts for a single todo
func getSingleTodoPrompt(todo *core.Todo) string {
	switch todo.Status {
	case "in_progress":
		return fmt.Sprintf("\n\nThis %s is in progress. To continue effectively:\n\n"+
			"- What's the next concrete step to take?\n"+
			"- Are you following your test list or implementation plan?\n"+
			"- Any obstacles or questions that need addressing?\n\n"+
			"Use todo_update to document progress and findings.",
			todo.Type)

	case "blocked":
		return fmt.Sprintf("\n\nThis %s is blocked. To resolve the blocker:\n\n"+
			"- What specifically is preventing progress?\n"+
			"- Have you documented the blocker details?\n"+
			"- Is there an alternative approach to consider?\n\n"+
			"Update the todo with blocker details or create a linked todo for the dependency.",
			todo.Type)

	case "completed":
		return fmt.Sprintf("\n\nThis %s is completed. To wrap up:\n\n"+
			"- Are there any lessons learned to document?\n"+
			"- Did this reveal any follow-up tasks?\n"+
			"- Should this be archived to keep your list clean?\n\n"+
			"Consider using todo_archive or creating follow-up todos.",
			todo.Type)

	default:
		return "\n\nNext steps for this todo:\n\n" +
			"- Review the current status and update if needed\n" +
			"- Check if any sections need documentation\n" +
			"- Determine the next action to take\n\n" +
			"Use todo_update to modify status or add content."
	}
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
		prompt := getSearchPrompts(0)
		return mcp.NewToolResultText("No matching todos found" + prompt)
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
	prompt := getSearchPrompts(len(results))
	return mcp.NewToolResultText(header + strings.Join(lines, "\n\n") + prompt)
}

// getSearchPrompts returns contextual prompts based on search results
func getSearchPrompts(resultCount int) string {
	if resultCount == 0 {
		return "\n\nNo matches found. To find what you're looking for:\n\n" +
			"- Try different search terms or keywords\n" +
			"- Check if the todo might be archived\n" +
			"- Consider creating a new todo if this is a new task\n\n" +
			"Use broader search terms or create a new todo for this work."
	}

	if resultCount == 1 {
		return "\n\nFound one matching todo. Next steps:\n\n" +
			"- Review the todo details with todo_read\n" +
			"- Check if this is the task you were looking for\n" +
			"- Update progress if this is your current work\n\n" +
			"Use todo_read to view the complete todo content."
	}

	// Multiple results
	return fmt.Sprintf("\n\nFound %d matching todos. To work effectively:\n\n"+
		"- Which todo best matches your current need?\n"+
		"- Are these related tasks that could be linked?\n"+
		"- Should any completed ones be archived?\n\n"+
		"Review the results and use todo_read to examine specific todos.",
		resultCount)
}
