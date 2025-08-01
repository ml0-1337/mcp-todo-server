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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}

	// Add contextual prompts
	prompt := getCreatePrompts(todo.Type, false)
	result := string(jsonData) + "\n\n" + prompt

	return mcp.NewToolResultText(result)
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

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}

	// Add contextual prompts
	prompt := getCreatePrompts(todo.Type, false)
	result := string(jsonData) + "\n\n" + prompt

	return mcp.NewToolResultText(result)
}

// FormatTodoCreateMultiResponse formats the response for todo_create_multi
func FormatTodoCreateMultiResponse(parent *core.Todo, children []*core.Todo) *mcp.CallToolResult {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Created multi-phase project: %s\n\n", parent.ID))
	sb.WriteString("📋 Project Structure:\n")

	// Parent
	sb.WriteString(fmt.Sprintf("[→] %s: %s [%s] [%s]\n", parent.ID, parent.Task,
		strings.ToUpper(parent.Priority), parent.Type))

	// Children
	for _, child := range children {
		sb.WriteString(fmt.Sprintf("  └─ %s: %s [%s] [%s]\n", child.ID, child.Task,
			strings.ToUpper(child.Priority), child.Type))
	}

	childCount := len(children)
	sb.WriteString(fmt.Sprintf("\n✅ Successfully created %d todos (1 parent, %d children)\n",
		childCount+1, childCount))

	sb.WriteString("\n💡 TIP: Use `todo_read` to see the full hierarchy or `todo_update` to modify individual todos.")

	// Add contextual prompts for multi-phase projects
	prompt := getCreatePrompts(parent.Type, true)
	sb.WriteString("\n\n" + prompt)

	return mcp.NewToolResultText(sb.String())
}

// getCreatePrompts returns contextual prompts based on todo type
func getCreatePrompts(todoType string, isMultiPhase bool) string {
	if isMultiPhase {
		return "Multi-phase project created. To maintain project momentum:\n\n" +
			"- Which phase should you start with first?\n" +
			"- Are there any dependencies between phases to consider?\n" +
			"- Do you need to adjust priorities based on current context?\n\n" +
			"Start with the first phase using todo_read to see the full hierarchy."
	}

	switch todoType {
	case "feature":
		return "Feature todo created. To ensure successful implementation:\n\n" +
			"- What are the key behaviors this feature should exhibit?\n" +
			"- Which test scenarios will validate the implementation?\n" +
			"- Are there any edge cases or error conditions to consider?\n\n" +
			"Consider using todo_update to add your Test List before implementation."

	case "bug":
		return "Bug todo created. To effectively resolve this issue:\n\n" +
			"- Can you reproduce the bug consistently?\n" +
			"- What is the expected behavior versus actual behavior?\n" +
			"- Which test would catch this bug in the future?\n\n" +
			"Start by using todo_update to document reproduction steps."

	case "research":
		return "Research todo created. To guide your investigation:\n\n" +
			"- What are the key questions you need to answer?\n" +
			"- Which resources or documentation should you consult?\n" +
			"- What would constitute a successful research outcome?\n\n" +
			"Use todo_update to document findings as you research."

	case "refactor":
		return "Refactor todo created. To ensure safe code improvements:\n\n" +
			"- What is the primary goal of this refactoring?\n" +
			"- Are there existing tests to ensure behavior doesn't change?\n" +
			"- Which files or components will be affected?\n\n" +
			"Consider running existing tests first to establish a baseline."

	default:
		return "Todo created successfully. To make progress:\n\n" +
			"- What is the first concrete step to take?\n" +
			"- Are there any blockers or dependencies to address?\n" +
			"- How will you know when this todo is complete?\n\n" +
			"Use todo_update to track progress and findings."
	}
}

// getTemplatePrompts returns contextual prompts based on template type
func getTemplatePrompts(template string) string {
	switch template {
	case "bug-fix":
		return "Bug fix template applied. To get started:\n\n" +
			"- Update the 'Steps to Reproduce' section with specific details\n" +
			"- Add the failing test that reproduces the bug\n" +
			"- Document your fix approach in the findings section\n\n" +
			"Begin by using todo_update to fill in the template sections."

	case "feature":
		return "Feature template applied. To develop effectively:\n\n" +
			"- Define clear acceptance criteria in the checklist\n" +
			"- List all test scenarios in the Test List section\n" +
			"- Consider edge cases and error handling\n\n" +
			"Start with todo_update to customize the template for your specific feature."

	case "research":
		return "Research template applied. To conduct thorough research:\n\n" +
			"- List specific questions to answer in the checklist\n" +
			"- Document sources and references as you find them\n" +
			"- Summarize key findings and recommendations\n\n" +
			"Use todo_update to capture research findings as you progress."

	case "refactor":
		return "Refactor template applied. To refactor safely:\n\n" +
			"- Document the current state and problems\n" +
			"- List all affected files and components\n" +
			"- Ensure test coverage before making changes\n\n" +
			"Begin by using todo_update to detail the refactoring plan."

	case "tdd-cycle":
		return "TDD cycle template applied. To follow test-driven development:\n\n" +
			"- Start with the Test List - what behaviors need testing?\n" +
			"- Pick one test and follow Red-Green-Refactor cycle\n" +
			"- Document each phase in the test results section\n\n" +
			"Use todo_update to track your TDD progress through each cycle."

	default:
		return "Template applied successfully. To make it your own:\n\n" +
			"- Review the template sections and customize as needed\n" +
			"- Add specific details relevant to your task\n" +
			"- Remove any sections that don't apply\n\n" +
			"Start customizing with todo_update."
	}
}

// FormatTodoTemplateResponse formats the response for template creation
func FormatTodoTemplateResponse(todo *core.Todo, filePath string, template string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"id":       todo.ID,
		"path":     filePath,
		"message":  fmt.Sprintf("Todo created from template successfully: %s", todo.ID),
		"template": fmt.Sprintf("Applied %s template", template),
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}

	// Add contextual prompts for template
	prompt := getTemplatePrompts(template)
	result := string(jsonData) + "\n\n" + prompt

	return mcp.NewToolResultText(result)
}
