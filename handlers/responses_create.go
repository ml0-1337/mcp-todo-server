package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
)

// FormatTodoCreateResponse formats the response for todo_create
func FormatTodoCreateResponse(todo *core.Todo, filePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":      todo.ID,
		"path":    filePath,
		"message": fmt.Sprintf("Todo created successfully: %s", todo.ID),
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	jsonStr := string(jsonData)
	
	// Add context-aware guidance based on todo type and status
	var guidance strings.Builder
	guidance.WriteString(jsonStr)
	guidance.WriteString("\n\n")
	
	// Type-specific guidance
	switch todo.Type {
	case "bug":
		guidance.WriteString("Starting a bug fix. To ensure a robust solution:\n")
		guidance.WriteString("- First reproduce the issue and document it in 'findings'\n")
		guidance.WriteString("- Write a failing test that captures the bug\n")
		guidance.WriteString("- Only then implement the fix\n\n")
		guidance.WriteString("Can you describe how to reproduce this bug?")
		
	case "feature":
		guidance.WriteString(fmt.Sprintf("Starting a new %s. To build effectively:\n", todo.Type))
		guidance.WriteString("- Define clear acceptance criteria in the 'tests' section\n")
		guidance.WriteString("- Document design decisions in 'findings' as you research\n")
		guidance.WriteString("- Break down into subtasks if scope seems large\n\n")
		guidance.WriteString("What specific aspect will you tackle first?")
		
	case "phase", "subtask":
		if todo.ParentID != "" {
			guidance.WriteString("Phase created and linked to parent project. Consider:\n")
			guidance.WriteString("- Review parent todo for overall context and goals\n")
			guidance.WriteString("- Check if other phases might have dependencies\n")
			guidance.WriteString("- Update parent's checklist to track this phase\n\n")
			guidance.WriteString("Ready to begin implementation or need to create more phases first?")
		} else {
			guidance.WriteString(fmt.Sprintf("Created %s todo. Note: This type typically requires a parent_id.\n", todo.Type))
			guidance.WriteString("Consider using todo_link to connect this to a parent project.\n\n")
			guidance.WriteString("What's the broader context for this work?")
		}
		
	case "refactor":
		guidance.WriteString("Starting refactoring work. To maintain code quality:\n")
		guidance.WriteString("- Ensure comprehensive tests exist before changing structure\n")
		guidance.WriteString("- Document the current design problems in 'findings'\n")
		guidance.WriteString("- Make structural changes separate from behavioral changes\n\n")
		guidance.WriteString("What specific code smell or design issue are you addressing?")
		
	case "research":
		guidance.WriteString("Starting research task. To capture valuable insights:\n")
		guidance.WriteString("- Document all findings systematically as you explore\n")
		guidance.WriteString("- Include links and references for future access\n")
		guidance.WriteString("- Synthesize conclusions at the end\n\n")
		guidance.WriteString("What specific question are you trying to answer?")
		
	default:
		guidance.WriteString(fmt.Sprintf("Starting new %s task. To maximize effectiveness:\n", todo.Type))
		guidance.WriteString("- Define clear success criteria\n")
		guidance.WriteString("- Document progress in appropriate sections\n")
		guidance.WriteString("- Consider if this should be broken into smaller pieces\n\n")
		guidance.WriteString("What's the first concrete step to make progress?")
	}
	
	return mcp.NewToolResultText(guidance.String())
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

// FormatTodoCreateMultiResponse formats the response for todo_create_multi
func FormatTodoCreateMultiResponse(parent *core.Todo, children []*core.Todo) *mcp.CallToolResult {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("Created multi-phase project: %s\n\n", parent.ID))
	sb.WriteString("Project Structure:\n")
	
	// Parent
	sb.WriteString(fmt.Sprintf("[PARENT] %s: %s [%s] [%s]\n", parent.ID, parent.Task, 
		strings.ToUpper(parent.Priority), parent.Type))
	
	// Children
	for i, child := range children {
		prefix := "├─"
		if i == len(children)-1 {
			prefix = "└─"
		}
		sb.WriteString(fmt.Sprintf("  %s %s: %s [%s] [%s]\n", prefix, child.ID, child.Task,
			strings.ToUpper(child.Priority), child.Type))
	}
	
	childCount := len(children)
	sb.WriteString(fmt.Sprintf("\nSuccessfully created %d todos (1 parent, %d children)\n\n", 
		childCount+1, childCount))
	
	sb.WriteString("Multi-phase project initialized. To work effectively:\n")
	sb.WriteString("- Start with the first phase while keeping the full scope in mind\n")
	sb.WriteString("- Update parent's checklist as you complete each phase\n")
	sb.WriteString("- Use todo_read to see progress across all phases\n")
	sb.WriteString("- Consider dependencies between phases before starting work\n\n")
	sb.WriteString("Which phase should be tackled first based on dependencies and priority?")
	
	return mcp.NewToolResultText(sb.String())
}

// FormatTodoTemplateResponse formats the response for template creation
func FormatTodoTemplateResponse(todo *core.Todo, filePath string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":       todo.ID,
		"path":     filePath,
		"message":  fmt.Sprintf("Todo created from template successfully: %s", todo.ID),
		"template": "Applied template sections and structure",
	}

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	jsonStr := string(jsonData)
	
	var guidance strings.Builder
	guidance.WriteString(jsonStr)
	guidance.WriteString("\n\n")
	guidance.WriteString("Template applied with pre-structured sections. To leverage the template effectively:\n")
	guidance.WriteString("- Review the template structure and customize for your specific needs\n")
	guidance.WriteString("- Fill in the pre-defined sections with relevant information\n")
	guidance.WriteString("- The template provides a proven workflow - follow it for best results\n\n")
	guidance.WriteString("What information do you need to gather to complete the first template section?")
	
	return mcp.NewToolResultText(guidance.String())
}