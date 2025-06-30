package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/utils"
)

// TestE2EMultiPhaseProject tests a complete multi-phase project workflow
func TestE2EMultiPhaseProject(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "e2e-multiphase-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize all components
	todoManager := core.NewTodoManager(tempDir)
	searchEngine, err := core.NewSearchEngine(
		filepath.Join(tempDir, ".claude", "index", "todos.bleve"),
		tempDir,
	)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	statsEngine := core.NewStatsEngine(todoManager)
	templateManager := core.NewTemplateManager(utils.GetEnv("CLAUDE_TODO_TEMPLATES_PATH", "templates"))
	todoLinker := &core.TodoLinker{Manager: todoManager}
	
	handlers := NewTodoHandlersWithDependencies(
		todoManager,
		searchEngine,
		statsEngine,
		templateManager,
		todoLinker,
	)

	// Phase 1: Create main project
	t.Run("CreateMainProject", func(t *testing.T) {
		result := handlers.HandleTodoCreate(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"task":     "Build E-commerce Platform",
				"priority": "high",
				"type":     "multi-phase",
			},
		})

		if result.IsError {
			t.Fatalf("Failed to create main project: %v", result.Content)
		}

		// Extract project ID
		content, ok := result.Content.([]interface{})
		if !ok || len(content) == 0 {
			t.Fatal("Invalid response format")
		}

		todoMap, ok := content[0].(map[string]interface{})
		if !ok {
			t.Fatal("Invalid todo format")
		}

		projectID, ok := todoMap["id"].(string)
		if !ok || projectID == "" {
			t.Fatal("Failed to get project ID")
		}

		// Store for later phases
		t.Setenv("PROJECT_ID", projectID)
	})

	projectID := os.Getenv("PROJECT_ID")

	// Phase 2: Create multiple sub-phases
	phases := []struct {
		name     string
		priority string
		tasks    []string
	}{
		{
			name:     "Phase 1: Backend API Development",
			priority: "high",
			tasks: []string{
				"Design REST API endpoints",
				"Implement authentication system",
				"Create database schema",
				"Add payment integration",
			},
		},
		{
			name:     "Phase 2: Frontend Development",
			priority: "high",
			tasks: []string{
				"Design UI mockups",
				"Implement product catalog",
				"Create shopping cart",
				"Add checkout flow",
			},
		},
		{
			name:     "Phase 3: Testing and Deployment",
			priority: "medium",
			tasks: []string{
				"Write unit tests",
				"Perform integration testing",
				"Setup CI/CD pipeline",
				"Deploy to production",
			},
		},
	}

	phaseIDs := make(map[string]string)

	t.Run("CreatePhases", func(t *testing.T) {
		for _, phase := range phases {
			result := handlers.HandleTodoCreate(context.Background(), &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":      phase.name,
					"priority":  phase.priority,
					"type":      "phase",
					"parent_id": projectID,
				},
			})

			if result.IsError {
				t.Errorf("Failed to create phase %s: %v", phase.name, result.Content)
				continue
			}

			// Extract phase ID
			if content, ok := result.Content.([]interface{}); ok && len(content) > 0 {
				if todoMap, ok := content[0].(map[string]interface{}); ok {
					if id, ok := todoMap["id"].(string); ok {
						phaseIDs[phase.name] = id
					}
				}
			}
		}

		if len(phaseIDs) != len(phases) {
			t.Fatalf("Failed to create all phases: got %d, want %d", len(phaseIDs), len(phases))
		}
	})

	// Phase 3: Create tasks under each phase
	t.Run("CreatePhaseTasks", func(t *testing.T) {
		for _, phase := range phases {
			phaseID, ok := phaseIDs[phase.name]
			if !ok {
				t.Errorf("Missing phase ID for %s", phase.name)
				continue
			}

			for _, task := range phase.tasks {
				result := handlers.HandleTodoCreate(context.Background(), &MockCallToolRequest{
					Arguments: map[string]interface{}{
						"task":      task,
						"priority":  "medium",
						"type":      "task",
						"parent_id": phaseID,
					},
				})

				if result.IsError {
					t.Errorf("Failed to create task '%s': %v", task, result.Content)
				}
			}
		}
	})

	// Phase 4: Test search functionality
	t.Run("SearchFunctionality", func(t *testing.T) {
		// Wait a bit for indexing
		time.Sleep(100 * time.Millisecond)

		// Search for API-related tasks
		result := handlers.HandleTodoSearch(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"query": "API",
				"limit": 10,
			},
		})

		if result.IsError {
			t.Fatalf("Search failed: %v", result.Content)
		}

		results, ok := result.Content.([]interface{})
		if !ok {
			t.Fatal("Invalid search results format")
		}

		if len(results) == 0 {
			t.Error("Expected to find API-related tasks")
		}

		// Search for testing tasks
		result = handlers.HandleTodoSearch(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"query": "test",
				"limit": 10,
			},
		})

		if result.IsError {
			t.Fatalf("Search for 'test' failed: %v", result.Content)
		}

		results, ok = result.Content.([]interface{})
		if !ok || len(results) == 0 {
			t.Error("Expected to find testing-related tasks")
		}
	})

	// Phase 5: Update task progress
	t.Run("UpdateTaskProgress", func(t *testing.T) {
		// Search for a specific task to update
		result := handlers.HandleTodoSearch(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"query": "authentication",
				"limit": 1,
			},
		})

		if result.IsError {
			t.Fatalf("Failed to find authentication task: %v", result.Content)
		}

		results, ok := result.Content.([]interface{})
		if !ok || len(results) == 0 {
			t.Fatal("No authentication task found")
		}

		todoMap, ok := results[0].(map[string]interface{})
		if !ok {
			t.Fatal("Invalid todo format")
		}

		todoID, ok := todoMap["id"].(string)
		if !ok {
			t.Fatal("Missing todo ID")
		}

		// Update the task
		updateResult := handlers.HandleTodoUpdate(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        todoID,
				"section":   "checklist",
				"operation": "append",
				"content": `- [x] Design authentication flow
- [x] Implement JWT tokens
- [ ] Add refresh token mechanism
- [ ] Implement OAuth providers`,
			},
		})

		if updateResult.IsError {
			t.Errorf("Failed to update task: %v", updateResult.Content)
		}
	})

	// Phase 6: Get project statistics
	t.Run("ProjectStatistics", func(t *testing.T) {
		result := handlers.HandleTodoStats(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"period": "all",
			},
		})

		if result.IsError {
			t.Fatalf("Failed to get stats: %v", result.Content)
		}

		// Verify we have meaningful stats
		if _, ok := result.Content.([]interface{}); !ok {
			t.Error("Invalid stats format")
		}
	})

	// Phase 7: Complete and archive a task
	t.Run("CompleteAndArchiveTask", func(t *testing.T) {
		// Find a task to complete
		result := handlers.HandleTodoSearch(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"query": "mockups",
				"limit": 1,
			},
		})

		if result.IsError {
			t.Fatalf("Failed to find task: %v", result.Content)
		}

		results, ok := result.Content.([]interface{})
		if !ok || len(results) == 0 {
			t.Fatal("No task found")
		}

		todoMap, ok := results[0].(map[string]interface{})
		if !ok {
			t.Fatal("Invalid todo format")
		}

		todoID, ok := todoMap["id"].(string)
		if !ok {
			t.Fatal("Missing todo ID")
		}

		// Mark as completed
		updateResult := handlers.HandleTodoUpdate(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        todoID,
				"operation": "set_metadata",
				"metadata": map[string]interface{}{
					"status":    "completed",
					"completed": time.Now().Format("2006-01-02 15:04:05"),
				},
			},
		})

		if updateResult.IsError {
			t.Errorf("Failed to complete task: %v", updateResult.Content)
		}

		// Archive it
		archiveResult := handlers.HandleTodoArchive(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": todoID,
			},
		})

		if archiveResult.IsError {
			t.Errorf("Failed to archive task: %v", archiveResult.Content)
		}

		// Verify it's archived
		readResult := handlers.HandleTodoRead(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"filter": map[string]interface{}{
					"status": "completed",
				},
				"include_archived": true,
			},
		})

		if readResult.IsError {
			t.Errorf("Failed to read archived todos: %v", readResult.Content)
		}
	})

	// Phase 8: Generate final project tree
	t.Run("ProjectTreeView", func(t *testing.T) {
		result := handlers.HandleTodoRead(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"format": "tree",
			},
		})

		if result.IsError {
			t.Fatalf("Failed to get tree view: %v", result.Content)
		}

		// Should have a tree structure
		if content, ok := result.Content.([]interface{}); ok && len(content) > 0 {
			treeStr := fmt.Sprintf("%v", content[0])
			if !strings.Contains(treeStr, "└──") && !strings.Contains(treeStr, "├──") {
				t.Error("Tree view doesn't contain expected tree characters")
			}
		}
	})
}

// TestE2ETemplateWorkflow tests the complete template workflow
func TestE2ETemplateWorkflow(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "e2e-template-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test templates directory
	templatesDir := filepath.Join(tempDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create a test template
	bugTemplate := `---
todo_id: "{{.ID}}"
started: "{{.Started}}"
completed: ""
status: "in_progress"
priority: "{{.Priority}}"
type: "bug"
---

# Task: {{.Task}}

## Bug Report

**Reported by**: {{.ReportedBy}}
**Severity**: {{.Severity}}
**Component**: {{.Component}}

### Description
{{.Description}}

### Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

### Expected Behavior
[Describe expected behavior]

### Actual Behavior
[Describe actual behavior]

### Environment
- OS: {{.OS}}
- Browser: {{.Browser}}
- Version: {{.Version}}

## Investigation

### Root Cause Analysis
[To be filled during investigation]

### Proposed Solution
[To be filled after analysis]

## Test Cases
- [ ] Verify bug is reproducible
- [ ] Test proposed fix
- [ ] Regression testing
- [ ] Edge case testing

## Checklist
- [ ] Bug reproduced
- [ ] Root cause identified
- [ ] Solution implemented
- [ ] Tests written
- [ ] Code reviewed
- [ ] Documentation updated
`

	if err := os.WriteFile(filepath.Join(templatesDir, "bug-report.md"), []byte(bugTemplate), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Initialize components
	todoManager := core.NewTodoManager(tempDir)
	searchEngine, err := core.NewSearchEngine(
		filepath.Join(tempDir, ".claude", "index", "todos.bleve"),
		tempDir,
	)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	statsEngine := core.NewStatsEngine(todoManager)
	templateManager := core.NewTemplateManager(templatesDir)
	
	handlers := NewTodoHandlersWithDependencies(
		todoManager,
		searchEngine,
		statsEngine,
		templateManager,
		&core.TodoLinker{Manager: todoManager},
	)

	// Test template listing
	t.Run("ListTemplates", func(t *testing.T) {
		result := handlers.HandleTodoTemplate(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"template": "list",
			},
		})

		if result.IsError {
			t.Fatalf("Failed to list templates: %v", result.Content)
		}

		templates, ok := result.Content.([]interface{})
		if !ok {
			t.Fatal("Invalid templates format")
		}

		if len(templates) == 0 {
			t.Error("Expected at least one template")
		}

		// Check for our bug-report template
		found := false
		for _, tmpl := range templates {
			if tmplMap, ok := tmpl.(map[string]interface{}); ok {
				if name, ok := tmplMap["name"].(string); ok && name == "bug-report" {
					found = true
					break
				}
			}
		}

		if !found {
			t.Error("bug-report template not found")
		}
	})

	// Test creating todo from template
	t.Run("CreateFromTemplate", func(t *testing.T) {
		result := handlers.HandleTodoTemplate(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"template": "bug-report",
				"task":     "Login button not working on mobile",
				"priority": "high",
				"variables": map[string]interface{}{
					"ReportedBy":  "QA Team",
					"Severity":    "High",
					"Component":   "Authentication",
					"Description": "Users cannot tap the login button on mobile devices",
					"OS":          "iOS 17",
					"Browser":     "Safari",
					"Version":     "1.2.3",
				},
			},
		})

		if result.IsError {
			t.Fatalf("Failed to create from template: %v", result.Content)
		}

		// Verify todo was created
		if content, ok := result.Content.([]interface{}); ok && len(content) > 0 {
			if todoMap, ok := content[0].(map[string]interface{}); ok {
				// Check that template variables were replaced
				if sections, ok := todoMap["sections"].(map[string]interface{}); ok {
					if bugReport, ok := sections["Bug Report"].(string); ok {
						if !strings.Contains(bugReport, "QA Team") {
							t.Error("Template variable 'ReportedBy' not replaced")
						}
						if !strings.Contains(bugReport, "iOS 17") {
							t.Error("Template variable 'OS' not replaced")
						}
					}
				}
			}
		}
	})
}

// TestE2EArchiveOperations tests archive functionality end-to-end
func TestE2EArchiveOperations(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "e2e-archive-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize components
	todoManager := core.NewTodoManager(tempDir)
	searchEngine, err := core.NewSearchEngine(
		filepath.Join(tempDir, ".claude", "index", "todos.bleve"),
		tempDir,
	)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	handlers := NewTodoHandlersWithDependencies(
		todoManager,
		searchEngine,
		core.NewStatsEngine(todoManager),
		core.NewTemplateManager("templates"),
		&core.TodoLinker{Manager: todoManager},
	)

	// Create todos with different dates
	todoIDs := make(map[string]string)
	dates := []string{
		"2025-01-15",
		"2025-01-20",
		"2025-02-01",
		"2025-02-15",
	}

	t.Run("CreateTodosWithDates", func(t *testing.T) {
		for i, date := range dates {
			result := handlers.HandleTodoCreate(context.Background(), &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     fmt.Sprintf("Task from %s", date),
					"priority": "medium",
					"type":     "task",
				},
			})

			if result.IsError {
				t.Errorf("Failed to create todo %d: %v", i, result.Content)
				continue
			}

			// Extract ID and update started date
			if content, ok := result.Content.([]interface{}); ok && len(content) > 0 {
				if todoMap, ok := content[0].(map[string]interface{}); ok {
					if id, ok := todoMap["id"].(string); ok {
						todoIDs[date] = id
						
						// Update the started date
						updateResult := handlers.HandleTodoUpdate(context.Background(), &MockCallToolRequest{
							Arguments: map[string]interface{}{
								"id":        id,
								"operation": "set_metadata",
								"metadata": map[string]interface{}{
									"started": date + " 10:00:00",
								},
							},
						})
						
						if updateResult.IsError {
							t.Errorf("Failed to update date for todo %s: %v", id, updateResult.Content)
						}
					}
				}
			}
		}
	})

	// Archive todos
	t.Run("ArchiveTodos", func(t *testing.T) {
		for date, id := range todoIDs {
			// Mark as completed first
			updateResult := handlers.HandleTodoUpdate(context.Background(), &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        id,
					"operation": "set_metadata",
					"metadata": map[string]interface{}{
						"status":    "completed",
						"completed": date + " 18:00:00",
					},
				},
			})

			if updateResult.IsError {
				t.Errorf("Failed to complete todo %s: %v", id, updateResult.Content)
				continue
			}

			// Archive it
			archiveResult := handlers.HandleTodoArchive(context.Background(), &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": id,
				},
			})

			if archiveResult.IsError {
				t.Errorf("Failed to archive todo %s: %v", id, archiveResult.Content)
			}
		}
	})

	// Verify archive structure
	t.Run("VerifyArchiveStructure", func(t *testing.T) {
		archiveDir := filepath.Join(tempDir, ".claude", "archive")
		
		// Check 2025/01/15 exists
		jan15Dir := filepath.Join(archiveDir, "2025", "01", "15")
		if _, err := os.Stat(jan15Dir); os.IsNotExist(err) {
			t.Errorf("Archive directory %s does not exist", jan15Dir)
		}

		// Check 2025/02/01 exists
		feb01Dir := filepath.Join(archiveDir, "2025", "02", "01")
		if _, err := os.Stat(feb01Dir); os.IsNotExist(err) {
			t.Errorf("Archive directory %s does not exist", feb01Dir)
		}

		// Verify todos are in correct directories
		for date, id := range todoIDs {
			parts := strings.Split(date, "-")
			if len(parts) != 3 {
				continue
			}
			
			archivePath := filepath.Join(archiveDir, parts[0], parts[1], parts[2], id+".md")
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				t.Errorf("Archived todo %s not found at expected path: %s", id, archivePath)
			}
		}
	})

	// Test reading archived todos
	t.Run("ReadArchivedTodos", func(t *testing.T) {
		result := handlers.HandleTodoRead(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"include_archived": true,
				"filter": map[string]interface{}{
					"status": "completed",
				},
			},
		})

		if result.IsError {
			t.Fatalf("Failed to read archived todos: %v", result.Content)
		}

		todos, ok := result.Content.([]interface{})
		if !ok {
			t.Fatal("Invalid response format")
		}

		if len(todos) != len(todoIDs) {
			t.Errorf("Expected %d archived todos, got %d", len(todoIDs), len(todos))
		}
	})

	// Test cleaning operation
	t.Run("CleanOperation", func(t *testing.T) {
		// Create an orphaned file in archive
		orphanDir := filepath.Join(tempDir, ".claude", "archive", "2025", "03", "01")
		if err := os.MkdirAll(orphanDir, 0755); err != nil {
			t.Fatalf("Failed to create orphan directory: %v", err)
		}

		orphanFile := filepath.Join(orphanDir, "orphaned-todo.md")
		if err := os.WriteFile(orphanFile, []byte("Invalid content"), 0644); err != nil {
			t.Fatalf("Failed to create orphan file: %v", err)
		}

		// Run clean operation
		cleanResult := handlers.HandleTodoClean(context.Background(), &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"dry_run": false,
			},
		})

		if cleanResult.IsError {
			t.Fatalf("Clean operation failed: %v", cleanResult.Content)
		}

		// Verify orphan was removed
		if _, err := os.Stat(orphanFile); !os.IsNotExist(err) {
			t.Error("Orphaned file was not cleaned up")
		}
	})
}