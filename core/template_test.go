package core

import (
	"testing"
	"os"
	"path/filepath"
	"io/ioutil"
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

## Maintainability Analysis

- **Readability**: [Score] - [Comments]
- **Complexity**: [Assessment]
- **Modularity**: [Assessment]
- **Testability**: [Assessment]

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