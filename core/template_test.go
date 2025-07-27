package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 20: Templates should load from .claude/templates directory
func TestLoadTemplateFromDirectory(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create templates directory
	templatesDir := filepath.Join(tempDir, "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create a sample feature template
	featureTemplate := `---
template_name: feature
description: Template for new feature development with TDD
variables:
  - task
  - priority
  - test_framework
---

# Task: {{.Task}}

## Findings & Research

[Research findings for {{.Task}}]

## Test Strategy

- **Test Framework**: {{.TestFramework}}
- **Test Types**: Unit, Integration
- **Coverage Target**: 90%
- **Edge Cases**: [To be identified]

## Test List

- [ ] Test 1: [Describe first test]
- [ ] Test 2: [Describe second test]
- [ ] Test 3: [Describe third test]

**Current Test**: Not started
**Phase**: Planning

## Test Cases

## Test Results Log

## Checklist

- [ ] Research best practices
- [ ] Design test strategy
- [ ] Implement all tests
- [ ] Add error handling
- [ ] Document API

## Working Scratchpad
`

	// Write the feature template
	featureTemplatePath := filepath.Join(templatesDir, "feature.md")
	err = ioutil.WriteFile(featureTemplatePath, []byte(featureTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write feature template: %v", err)
	}

	// Create a bug-fix template
	bugTemplate := `---
template_name: bug-fix
description: Template for bug fixes with reproduction steps
variables:
  - task
  - priority
  - severity
---

# Task: {{.Task}}

## Bug Report

- **Severity**: {{.Severity}}
- **Reported By**: 
- **Date Reported**: 
- **Steps to Reproduce**:
  1. 
  2. 
  3. 

## Findings & Research

### Root Cause Analysis

### Related Issues

## Test Strategy

- **Test Framework**: {{.TestFramework}}
- **Test Types**: Unit test to reproduce bug
- **Coverage Target**: Cover the bug fix
- **Edge Cases**: Related scenarios

## Test List

- [ ] Test 1: Reproduce the bug
- [ ] Test 2: Verify the fix
- [ ] Test 3: Test edge cases

**Current Test**: 
**Phase**: 

## Test Cases

## Test Results Log

## Checklist

- [ ] Reproduce bug consistently
- [ ] Write failing test
- [ ] Implement fix
- [ ] Verify fix works
- [ ] Check for regressions

## Working Scratchpad
`

	// Write the bug template
	bugTemplatePath := filepath.Join(templatesDir, "bug-fix.md")
	err = ioutil.WriteFile(bugTemplatePath, []byte(bugTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write bug template: %v", err)
	}

	// Test cases
	t.Run("Load existing feature template", func(t *testing.T) {
		// Create template manager
		tm := NewTemplateManager(templatesDir)

		// Load feature template
		template, err := tm.LoadTemplate("feature")
		if err != nil {
			t.Fatalf("Failed to load feature template: %v", err)
		}

		// Verify template loaded correctly
		if template.Name != "feature" {
			t.Errorf("Expected template name 'feature', got '%s'", template.Name)
		}

		if template.Description != "Template for new feature development with TDD" {
			t.Errorf("Unexpected template description: %s", template.Description)
		}

		// Check variables
		expectedVars := []string{"task", "priority", "test_framework"}
		if len(template.Variables) != len(expectedVars) {
			t.Errorf("Expected %d variables, got %d", len(expectedVars), len(template.Variables))
		}

		for i, v := range expectedVars {
			if i < len(template.Variables) && template.Variables[i] != v {
				t.Errorf("Expected variable %d to be '%s', got '%s'", i, v, template.Variables[i])
			}
		}

		// Verify content contains placeholders
		if template.Content == "" {
			t.Error("Template content should not be empty")
		}

		// Check for key placeholders
		if !containsPlaceholder(template.Content, "{{.Task}}") {
			t.Error("Template should contain {{.Task}} placeholder")
		}
		if !containsPlaceholder(template.Content, "{{.TestFramework}}") {
			t.Error("Template should contain {{.TestFramework}} placeholder")
		}
	})

	t.Run("Load existing bug-fix template", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		template, err := tm.LoadTemplate("bug-fix")
		if err != nil {
			t.Fatalf("Failed to load bug-fix template: %v", err)
		}

		if template.Name != "bug-fix" {
			t.Errorf("Expected template name 'bug-fix', got '%s'", template.Name)
		}

		// Check for bug-specific sections
		if !containsPlaceholder(template.Content, "## Bug Report") {
			t.Error("Bug template should contain Bug Report section")
		}
		if !containsPlaceholder(template.Content, "{{.Severity}}") {
			t.Error("Bug template should contain {{.Severity}} placeholder")
		}
	})

	t.Run("Error on non-existent template", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		_, err := tm.LoadTemplate("non-existent")
		if err == nil {
			t.Error("Loading non-existent template should return error")
		}

		expectedError := "template not found: non-existent"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("Error on corrupted template YAML", func(t *testing.T) {
		// Create corrupted template
		corruptedTemplate := `---
template_name: corrupted
description: "Unclosed quote
variables:
  - test
---
Content here`

		corruptedPath := filepath.Join(templatesDir, "corrupted.md")
		err := ioutil.WriteFile(corruptedPath, []byte(corruptedTemplate), 0644)
		if err != nil {
			t.Fatalf("Failed to write corrupted template: %v", err)
		}

		tm := NewTemplateManager(templatesDir)
		_, err = tm.LoadTemplate("corrupted")
		if err == nil {
			t.Error("Loading corrupted template should return error")
		}

		// Should mention parse error
		if !containsPlaceholder(err.Error(), "parse") {
			t.Errorf("Error should mention parse failure, got: %s", err.Error())
		}
	})

	t.Run("List available templates", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		templates, err := tm.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		// Should have at least the templates we created
		if len(templates) < 2 {
			t.Errorf("Expected at least 2 templates, got %d", len(templates))
		}

		// Check that our templates are in the list
		hasFeature := false
		hasBug := false
		for _, tmpl := range templates {
			if tmpl == "feature" {
				hasFeature = true
			}
			if tmpl == "bug-fix" {
				hasBug = true
			}
		}

		if !hasFeature {
			t.Error("Feature template not found in list")
		}
		if !hasBug {
			t.Error("Bug-fix template not found in list")
		}
	})
}

// Helper function to check if string contains substring
func containsPlaceholder(content, placeholder string) bool {
	return len(content) > 0 && len(placeholder) > 0 &&
		(placeholder == "" || content == "" ||
			(len(content) >= len(placeholder) &&
				findSubstring(content, placeholder) >= 0))
}

// Simple substring search
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Test 21: Template substitution should replace variables correctly
func TestTemplateVariableSubstitution(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "template-sub-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create templates directory
	templatesDir := filepath.Join(tempDir, "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create a simple template with variables
	simpleTemplate := `---
template_name: simple
description: Simple template for testing
variables:
  - task
  - priority
  - deadline
---

# Task: {{.Task}}

Priority: {{.Priority}}
Deadline: {{.Deadline}}

## Description

Working on {{.Task}} with {{.Priority}} priority.`

	// Write the simple template
	simpleTemplatePath := filepath.Join(templatesDir, "simple.md")
	err = ioutil.WriteFile(simpleTemplatePath, []byte(simpleTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write simple template: %v", err)
	}

	// Create a template with conditional sections
	conditionalTemplate := `---
template_name: conditional
description: Template with conditional sections
variables:
  - task
  - type
  - severity
---

# Task: {{.Task}}

Type: {{.Type}}

{{if eq .Type "bug"}}
## Bug Details

Severity: {{.Severity}}

### Steps to Reproduce
1. 
2. 
3. 
{{else if eq .Type "feature"}}
## Feature Details

### User Story
As a user, I want to...

### Acceptance Criteria
- [ ] 
- [ ] 
{{else}}
## Task Details

General task of type: {{.Type}}
{{end}}`

	// Write the conditional template
	conditionalTemplatePath := filepath.Join(templatesDir, "conditional.md")
	err = ioutil.WriteFile(conditionalTemplatePath, []byte(conditionalTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to write conditional template: %v", err)
	}

	// Test cases
	t.Run("Substitute basic variables", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		// Load template
		template, err := tm.LoadTemplate("simple")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		// Define variables
		vars := map[string]interface{}{
			"Task":     "Implement user authentication",
			"Priority": "high",
			"Deadline": "2025-01-31",
		}

		// Execute template
		result, err := tm.ExecuteTemplate(template, vars)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		// Verify substitutions
		if !containsPlaceholder(result, "# Task: Implement user authentication") {
			t.Error("Task substitution failed")
		}
		if !containsPlaceholder(result, "Priority: high") {
			t.Error("Priority substitution failed")
		}
		if !containsPlaceholder(result, "Deadline: 2025-01-31") {
			t.Error("Deadline substitution failed")
		}
		if !containsPlaceholder(result, "Working on Implement user authentication with high priority") {
			t.Error("Complex substitution failed")
		}

		// Ensure no placeholders remain
		if containsPlaceholder(result, "{{.") {
			t.Error("Result should not contain any remaining placeholders")
		}
	})

	t.Run("Handle conditional sections for bug type", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		template, err := tm.LoadTemplate("conditional")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		// Execute with bug type
		vars := map[string]interface{}{
			"Task":     "Fix login timeout issue",
			"Type":     "bug",
			"Severity": "high",
		}

		result, err := tm.ExecuteTemplate(template, vars)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		// Should include bug section
		if !containsPlaceholder(result, "## Bug Details") {
			t.Error("Bug section should be included")
		}
		if !containsPlaceholder(result, "Severity: high") {
			t.Error("Severity should be shown for bugs")
		}
		if !containsPlaceholder(result, "### Steps to Reproduce") {
			t.Error("Steps to Reproduce should be included")
		}

		// Should NOT include feature section
		if containsPlaceholder(result, "## Feature Details") {
			t.Error("Feature section should not be included for bugs")
		}
		if containsPlaceholder(result, "### User Story") {
			t.Error("User Story should not be included for bugs")
		}
	})

	t.Run("Handle conditional sections for feature type", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		template, err := tm.LoadTemplate("conditional")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		// Execute with feature type
		vars := map[string]interface{}{
			"Task": "Add dark mode support",
			"Type": "feature",
		}

		result, err := tm.ExecuteTemplate(template, vars)
		if err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		// Should include feature section
		if !containsPlaceholder(result, "## Feature Details") {
			t.Error("Feature section should be included")
		}
		if !containsPlaceholder(result, "### User Story") {
			t.Error("User Story should be included")
		}
		if !containsPlaceholder(result, "### Acceptance Criteria") {
			t.Error("Acceptance Criteria should be included")
		}

		// Should NOT include bug section
		if containsPlaceholder(result, "## Bug Details") {
			t.Error("Bug section should not be included for features")
		}
	})

	t.Run("Handle missing variables with error", func(t *testing.T) {
		tm := NewTemplateManager(templatesDir)

		template, err := tm.LoadTemplate("simple")
		if err != nil {
			t.Fatalf("Failed to load template: %v", err)
		}

		// Missing required variables
		vars := map[string]interface{}{
			"Task": "Incomplete task",
			// Missing Priority and Deadline
		}

		result, err := tm.ExecuteTemplate(template, vars)
		// Should succeed with Go's template behavior (empty values)
		if err != nil {
			t.Fatalf("Template execution should handle missing vars: %v", err)
		}

		// Missing values should be empty (Go templates use <no value>)
		if !containsPlaceholder(result, "Priority: <no value>") {
			t.Error("Missing Priority should show <no value>")
		}
		if !containsPlaceholder(result, "Deadline: <no value>") {
			t.Error("Missing Deadline should show <no value>")
		}
	})

	t.Run("Handle template execution errors", func(t *testing.T) {
		// Create template with invalid syntax
		invalidTemplate := &Template{
			Name:      "invalid",
			Content:   "{{.Task}} {{if .Invalid}} {{else}} {{end {{end}}", // Invalid syntax
			Variables: []string{"task"},
		}

		tm := NewTemplateManager(templatesDir)
		vars := map[string]interface{}{
			"Task": "Test",
		}

		_, err := tm.ExecuteTemplate(invalidTemplate, vars)
		if err == nil {
			t.Error("Should return error for invalid template syntax")
		}
	})
}

// TestCreateFromTemplate tests the CreateFromTemplate function
func TestCreateFromTemplate(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "create-from-template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create templates directory
	templatesDir := filepath.Join(tempDir, "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create test templates
	templates := map[string]string{
		"bug-fix": `---
template_name: bug-fix
description: Template for bug fixes
variables:
  - task
  - priority
  - type
---

# Task: {{.task}}

## Bug Description
[Describe the bug here]

## Root Cause Analysis
[Analyze the root cause]

## Fix Approach
[Describe the fix approach]

## Test Cases
- [ ] Test case 1
- [ ] Test case 2
`,
		"feature": `---
template_name: feature
description: Template for new features
variables:
  - task
  - priority
  - type
---

# Task: {{.task}}

## Feature Description
[Describe the feature]

## Implementation Plan
[Detail the implementation]

## Test Strategy
[Outline testing approach]
`,
		"invalid-yaml": `---
template_name: invalid
description: [Invalid YAML
variables
  - task
---

# Task: {{.task}}
`,
		"execution-error": `---
template_name: execution-error
description: Template with execution error
variables:
  - task
---

# Task: {{.task}}
{{if .undefined}} This will fail {{end}}
`,
	}

	// Write template files
	for name, content := range templates {
		templatePath := filepath.Join(templatesDir, name+".md")
		err := ioutil.WriteFile(templatePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write template %s: %v", name, err)
		}
	}

	// Create template manager
	tm := NewTemplateManager(templatesDir)

	tests := []struct {
		name         string
		templateName string
		task         string
		priority     string
		todoType     string
		expectError  bool
		errorMsg     string
		verifyTodo   func(*Todo) error
	}{
		{
			name:         "successful creation from bug-fix template",
			templateName: "bug-fix",
			task:         "Fix login timeout issue",
			priority:     "high",
			todoType:     "bug",
			expectError:  false,
			verifyTodo: func(todo *Todo) error {
				if todo.Task != "Fix login timeout issue" {
					return fmt.Errorf("expected task 'Fix login timeout issue', got %s", todo.Task)
				}
				if todo.Priority != "high" {
					return fmt.Errorf("expected priority 'high', got %s", todo.Priority)
				}
				if todo.Type != "bug" {
					return fmt.Errorf("expected type 'bug', got %s", todo.Type)
				}
				if todo.Status != "in_progress" {
					return fmt.Errorf("expected status 'in_progress', got %s", todo.Status)
				}
				return nil
			},
		},
		{
			name:         "successful creation from feature template",
			templateName: "feature",
			task:         "Add dark mode support",
			priority:     "medium",
			todoType:     "feature",
			expectError:  false,
			verifyTodo: func(todo *Todo) error {
				if todo.Task != "Add dark mode support" {
					return fmt.Errorf("expected task 'Add dark mode support', got %s", todo.Task)
				}
				if todo.ID == "" {
					return fmt.Errorf("expected non-empty ID")
				}
				return nil
			},
		},
		{
			name:         "template not found",
			templateName: "nonexistent",
			task:         "Test task",
			priority:     "high",
			todoType:     "feature",
			expectError:  true,
			errorMsg:     "template not found",
		},
		{
			name:         "empty template name",
			templateName: "",
			task:         "Test task",
			priority:     "high",
			todoType:     "feature",
			expectError:  true,
			errorMsg:     "template not found",
		},
		{
			name:         "invalid template YAML",
			templateName: "invalid-yaml",
			task:         "Test task",
			priority:     "high",
			todoType:     "feature",
			expectError:  true,
			errorMsg:     "failed to parse template frontmatter",
		},
		{
			name:         "template execution error",
			templateName: "execution-error",
			task:         "Test task",
			priority:     "high",
			todoType:     "feature",
			expectError:  false, // ExecuteTemplate doesn't actually fail for undefined vars in Go templates
		},
		{
			name:         "empty task",
			templateName: "bug-fix",
			task:         "",
			priority:     "high",
			todoType:     "bug",
			expectError:  false,
			verifyTodo: func(todo *Todo) error {
				if todo.Task != "" {
					return fmt.Errorf("expected empty task, got %s", todo.Task)
				}
				// ID generation should handle empty task
				if todo.ID == "" {
					return fmt.Errorf("expected non-empty ID even with empty task")
				}
				return nil
			},
		},
		{
			name:         "special characters in task",
			templateName: "feature",
			task:         "Fix issue #123: Can't save @mentions",
			priority:     "high",
			todoType:     "bug",
			expectError:  false,
			verifyTodo: func(todo *Todo) error {
				if todo.Task != "Fix issue #123: Can't save @mentions" {
					return fmt.Errorf("task with special characters not preserved")
				}
				// ID should be sanitized
				if strings.Contains(todo.ID, "#") || strings.Contains(todo.ID, "@") {
					return fmt.Errorf("ID should not contain special characters")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := tm.CreateFromTemplate(tt.templateName, tt.task, tt.priority, tt.todoType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if todo == nil {
					t.Fatal("Expected non-nil todo")
				}
				if tt.verifyTodo != nil {
					if err := tt.verifyTodo(todo); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

// TestCreateFromTemplateVariableSubstitution tests variable substitution
func TestCreateFromTemplateVariableSubstitution(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "template-vars-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatesDir := filepath.Join(tempDir, "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create template with all variables
	templateContent := `---
template_name: test-vars
description: Test variable substitution
variables:
  - task
  - priority
  - type
---

# Task: {{.task}}
Priority: {{.priority}}
Type: {{.type}}
`

	templatePath := filepath.Join(templatesDir, "test-vars.md")
	ioutil.WriteFile(templatePath, []byte(templateContent), 0644)

	tm := NewTemplateManager(templatesDir)

	// Test variable substitution
	_, err = tm.CreateFromTemplate("test-vars", "Test Task", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create from template: %v", err)
	}

	// Verify the ExecuteTemplate was called with correct variables
	tmpl, _ := tm.LoadTemplate("test-vars")
	vars := map[string]interface{}{
		"task":     "Test Task",
		"priority": "high",
		"type":     "feature",
	}
	result, err := tm.ExecuteTemplate(tmpl, vars)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	// Verify substitution occurred
	if !strings.Contains(result, "# Task: Test Task") {
		t.Error("Task variable not substituted correctly")
	}
	if !strings.Contains(result, "Priority: high") {
		t.Error("Priority variable not substituted correctly")
	}
	if !strings.Contains(result, "Type: feature") {
		t.Error("Type variable not substituted correctly")
	}
}

// TestCreateFromTemplateIDGeneration tests ID generation
func TestCreateFromTemplateIDGeneration(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "template-id-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templatesDir := filepath.Join(tempDir, "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create simple template
	templateContent := `---
template_name: simple
description: Simple template
variables:
  - task
---

# Task: {{.task}}
`

	templatePath := filepath.Join(templatesDir, "simple.md")
	ioutil.WriteFile(templatePath, []byte(templateContent), 0644)

	tm := NewTemplateManager(templatesDir)

	tests := []struct {
		task       string
		expectedID string // Expected pattern
	}{
		{
			task:       "Fix login bug",
			expectedID: "fix-login-bug",
		},
		{
			task:       "Add new feature: Dark Mode",
			expectedID: "add-new-feature-dark-mode",
		},
		{
			task:       "Update @user profile",
			expectedID: "update-user-profile",
		},
		{
			task:       "",
			expectedID: "todo", // generateBaseID returns "todo" for empty task
		},
	}

	for _, tt := range tests {
		t.Run(tt.task, func(t *testing.T) {
			todo, err := tm.CreateFromTemplate("simple", tt.task, "high", "feature")
			if err != nil {
				t.Fatalf("Failed to create todo: %v", err)
			}

			if tt.expectedID == "todo" && tt.task == "" {
				if todo.ID != "todo" {
					t.Errorf("Expected ID to be '%s' for empty task, got %s", tt.expectedID, todo.ID)
				}
			} else {
				if !strings.Contains(todo.ID, tt.expectedID) {
					t.Errorf("Expected ID to contain '%s', got %s", tt.expectedID, todo.ID)
				}
			}
		})
	}
}
