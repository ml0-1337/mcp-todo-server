package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// Integration test to verify todo and template path discovery work together
func TestPathDiscoveryIntegration(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Test 1: Both paths should resolve to project directories
	todoPath, err := ResolveTodoPath()
	if err != nil {
		t.Errorf("ResolveTodoPath() error = %v; want nil", err)
	}

	templatePath, err := ResolveTemplatePath()
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	// Verify paths are within project
	expectedTodoPath := filepath.Join(tempDir, ".claude", "todos")
	expectedTemplatePath := filepath.Join(tempDir, ".claude", "templates")

	resolvedTodoPath, _ := filepath.EvalSymlinks(todoPath)
	resolvedExpectedTodoPath, _ := filepath.EvalSymlinks(expectedTodoPath)

	if resolvedTodoPath != resolvedExpectedTodoPath {
		t.Errorf("ResolveTodoPath() = %s; want %s", todoPath, expectedTodoPath)
	}

	resolvedTemplatePath, _ := filepath.EvalSymlinks(templatePath)
	resolvedExpectedTemplatePath, _ := filepath.EvalSymlinks(expectedTemplatePath)

	if resolvedTemplatePath != resolvedExpectedTemplatePath {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, expectedTemplatePath)
	}

	// Test 2: Todo directory should be created
	if !IsDirectory(todoPath) {
		t.Errorf("Todo directory was not created at %s", todoPath)
	}

	// Test 3: Environment overrides should work
	customTodoPath := "/tmp/custom-todos"
	customTemplatePath := "/tmp/custom-templates"

	os.Setenv("CLAUDE_TODO_PATH", customTodoPath)
	os.Setenv("CLAUDE_TEMPLATE_PATH", customTemplatePath)
	defer os.Unsetenv("CLAUDE_TODO_PATH")
	defer os.Unsetenv("CLAUDE_TEMPLATE_PATH")

	todoPath2, _ := ResolveTodoPath()
	templatePath2, _ := ResolveTemplatePath()

	if todoPath2 != customTodoPath {
		t.Errorf("ResolveTodoPath() with env = %s; want %s", todoPath2, customTodoPath)
	}

	if templatePath2 != customTemplatePath {
		t.Errorf("ResolveTemplatePath() with env = %s; want %s", templatePath2, customTemplatePath)
	}
}
