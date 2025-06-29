package core

import (
	"fmt"
	"strings"
)

// TreeFormatter formats hierarchical todos as ASCII trees
type TreeFormatter struct {
	// Configuration
	ShowStatus   bool
	ShowPriority bool
	ShowType     bool
	IndentSize   int

	// Tree characters
	Branch     string // ├──
	LastBranch string // └──
	Vertical   string // │
	Space      string // "   " (same width as vertical)
}

// NewTreeFormatter creates a new tree formatter with default settings
func NewTreeFormatter() *TreeFormatter {
	return &TreeFormatter{
		ShowStatus:   true,
		ShowPriority: true,
		ShowType:     true,
		IndentSize:   4,
		Branch:       "├── ",
		LastBranch:   "└── ",
		Vertical:     "│   ",
		Space:        "    ",
	}
}

// FormatHierarchy formats a hierarchical todo structure as an ASCII tree
func (tf *TreeFormatter) FormatHierarchy(roots []*TodoNode, orphans []*Todo) string {
	var result strings.Builder

	// Format root nodes
	for i, root := range roots {
		isLast := i == len(roots)-1
		tf.formatNode(&result, root, "", isLast)
	}

	// Add orphaned phases/subtasks section if any exist
	if len(orphans) > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("\nORPHANED PHASES/SUBTASKS (need parent assignment):\n")
		for _, orphan := range orphans {
			result.WriteString(fmt.Sprintf("  %s %s [parent: %s not found]\n",
				tf.formatTodoLine(orphan), orphan.ID, orphan.ParentID))
		}
	}

	return result.String()
}

// formatNode recursively formats a node and its children
func (tf *TreeFormatter) formatNode(result *strings.Builder, node *TodoNode, prefix string, isLast bool) {
	// Format current node
	line := tf.formatTodoLine(node.Todo)

	// Add tree branch
	if prefix != "" {
		if isLast {
			result.WriteString(prefix + tf.LastBranch)
		} else {
			result.WriteString(prefix + tf.Branch)
		}
	}

	result.WriteString(line)
	result.WriteString("\n")

	// Format children
	if len(node.Children) > 0 {
		// Calculate child prefix
		childPrefix := prefix
		if isLast {
			childPrefix += tf.Space
		} else {
			childPrefix += tf.Vertical
		}

		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1
			tf.formatNode(result, child, childPrefix, isLastChild)
		}
	}
}

// formatTodoLine formats a single todo line with status, ID, task, and optional priority
func (tf *TreeFormatter) formatTodoLine(todo *Todo) string {
	var parts []string

	// Status indicator
	if tf.ShowStatus {
		status := tf.getStatusIndicator(todo.Status)
		parts = append(parts, status)
	}

	// ID and task
	parts = append(parts, fmt.Sprintf("%s: %s", todo.ID, todo.Task))

	// Priority
	if tf.ShowPriority && todo.Priority != "medium" {
		priority := tf.getPriorityIndicator(todo.Priority)
		if priority != "" {
			parts = append(parts, priority)
		}
	}

	// Type (for special types)
	if tf.ShowType && (todo.Type == "phase" || todo.Type == "subtask" || todo.Type == "multi-phase") {
		parts = append(parts, fmt.Sprintf("[%s]", todo.Type))
	}

	return strings.Join(parts, " ")
}

// getStatusIndicator returns the status indicator symbol
func (tf *TreeFormatter) getStatusIndicator(status string) string {
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

// getPriorityIndicator returns the priority indicator
func (tf *TreeFormatter) getPriorityIndicator(priority string) string {
	switch priority {
	case "high":
		return "[HIGH]"
	case "low":
		return "[LOW]"
	default:
		return ""
	}
}

// FormatSimpleTree formats todos as a simple indented tree (no box drawing)
func (tf *TreeFormatter) FormatSimpleTree(roots []*TodoNode) string {
	var result strings.Builder

	for _, root := range roots {
		tf.formatSimpleNode(&result, root, 0)
	}

	return result.String()
}

// formatSimpleNode formats a node with simple indentation
func (tf *TreeFormatter) formatSimpleNode(result *strings.Builder, node *TodoNode, depth int) {
	// Add indentation
	indent := strings.Repeat(" ", depth*tf.IndentSize)

	// Format line
	line := tf.formatTodoLine(node.Todo)
	result.WriteString(indent + line + "\n")

	// Format children
	for _, child := range node.Children {
		tf.formatSimpleNode(result, child, depth+1)
	}
}

// FormatCompactTree formats todos in a compact tree format
func (tf *TreeFormatter) FormatCompactTree(roots []*TodoNode) string {
	var result strings.Builder

	for i, root := range roots {
		if i > 0 {
			result.WriteString("\n")
		}
		tf.formatCompactNode(&result, root, "", i == len(roots)-1)
	}

	return result.String()
}

// formatCompactNode formats nodes in compact format (minimal spacing)
func (tf *TreeFormatter) formatCompactNode(result *strings.Builder, node *TodoNode, prefix string, isLast bool) {
	// Compact format: just ID and task with status
	status := tf.getCompactStatus(node.Todo.Status)

	if prefix != "" {
		if isLast {
			result.WriteString(prefix + "└─ ")
		} else {
			result.WriteString(prefix + "├─ ")
		}
	}

	result.WriteString(fmt.Sprintf("%s %s: %s\n", status, node.Todo.ID, node.Todo.Task))

	// Children with minimal prefix
	if len(node.Children) > 0 {
		childPrefix := prefix
		if isLast {
			childPrefix += "  "
		} else {
			childPrefix += "│ "
		}

		for i, child := range node.Children {
			tf.formatCompactNode(result, child, childPrefix, i == len(node.Children)-1)
		}
	}
}

// getCompactStatus returns a compact status indicator
func (tf *TreeFormatter) getCompactStatus(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "in_progress":
		return "→"
	case "blocked":
		return "✗"
	default:
		return "○"
	}
}

// FormatHierarchyWithStats formats hierarchy with statistics
func (tf *TreeFormatter) FormatHierarchyWithStats(roots []*TodoNode, orphans []*Todo, todos []*Todo) string {
	var result strings.Builder

	// Calculate stats
	stats := GetHierarchyStats(todos)

	// Add stats header
	result.WriteString(fmt.Sprintf("Todo Hierarchy (%d total, %d roots, depth: %d)\n",
		len(todos), stats.TotalRoots, stats.MaxDepth))
	result.WriteString(strings.Repeat("─", 50) + "\n\n")

	// Format hierarchy
	hierarchyStr := tf.FormatHierarchy(roots, orphans)
	result.WriteString(hierarchyStr)

	// Add stats footer if there are interesting stats
	if stats.TotalWithParent > 0 || len(orphans) > 0 {
		result.WriteString("\n" + strings.Repeat("─", 50) + "\n")
		result.WriteString(fmt.Sprintf("Summary: %d with parents, %d orphaned\n",
			stats.TotalWithParent, stats.TotalOrphans))
	}

	return result.String()
}

// FormatFlatWithIndication formats todos in a flat list but indicates parent-child relationships
func (tf *TreeFormatter) FormatFlatWithIndication(todos []*Todo) string {
	var result strings.Builder

	// Group by status as usual
	statusGroups := make(map[string][]*Todo)
	for _, todo := range todos {
		statusGroups[todo.Status] = append(statusGroups[todo.Status], todo)
	}

	// Format by status
	for _, status := range []string{"in_progress", "blocked", "completed"} {
		if todos, ok := statusGroups[status]; ok && len(todos) > 0 {
			result.WriteString(fmt.Sprintf("\n%s (%d):\n", strings.ToUpper(status), len(todos)))
			for _, todo := range todos {
				line := tf.formatTodoLine(todo)

				// Add parent indication
				if todo.ParentID != "" {
					line += fmt.Sprintf(" [parent: %s]", todo.ParentID)
				}

				// Add children count if applicable
				childCount := countChildren(todo.ID, todos)
				if childCount > 0 {
					line += fmt.Sprintf(" [%d children]", childCount)
				}

				result.WriteString(line + "\n")
			}
		}
	}

	return result.String()
}

// countChildren counts direct children of a todo
func countChildren(parentID string, todos []*Todo) int {
	count := 0
	for _, todo := range todos {
		if todo.ParentID == parentID {
			count++
		}
	}
	return count
}
