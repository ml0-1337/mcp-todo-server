package core

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// TestReplaceSection_PreservesSections verifies that replaceSection doesn't delete other sections
func TestReplaceSection_PreservesSections(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-replace-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Replace section test", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Count initial sections
	content1, _ := manager.ReadTodoContent(todo.ID)
	initialSections := strings.Count(content1, "## ")
	
	// Replace findings
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "New findings content", nil)
	if err != nil {
		t.Fatalf("Failed to replace: %v", err)
	}

	// Read after replace
	content2, _ := manager.ReadTodoContent(todo.ID)
	afterSections := strings.Count(content2, "## ")
	
	// Check section count
	if afterSections != initialSections {
		t.Errorf("Section count changed from %d to %d", initialSections, afterSections)
	}
	
	// Check Web Searches specifically (was being deleted due to bug)
	if !strings.Contains(content2, "## Web Searches") {
		t.Error("Web Searches section was deleted")
	}
	
	// Check for duplication
	count := strings.Count(content2, "New findings content")
	if count != 1 {
		t.Errorf("Content appears %d times (expected 1)", count)
	}
}

// TestReplaceSection_EmptyContent tests replacing with empty content
func TestReplaceSection_EmptyContent(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-replace-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Empty replace test", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add some content first
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "Initial content", nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Replace with empty content
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "", nil)
	if err != nil {
		t.Fatalf("Failed to replace with empty: %v", err)
	}

	// Verify the section is empty but still exists
	content, _ := manager.ReadTodoContent(todo.ID)
	if !strings.Contains(content, "## Findings & Research") {
		t.Error("Section header was removed")
	}
	
	// Verify old content is gone
	if strings.Contains(content, "Initial content") {
		t.Error("Old content was not removed")
	}
}

// TestPrependToSection_Basic tests basic prepend functionality
func TestPrependToSection_Basic(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-prepend-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Prepend test", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add initial content
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "Existing content", nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Prepend new content
	err = manager.UpdateTodo(todo.ID, "findings", "prepend", "Prepended content", nil)
	if err != nil {
		t.Fatalf("Failed to prepend: %v", err)
	}

	// Read and verify order
	content, _ := manager.ReadTodoContent(todo.ID)
	prependIdx := strings.Index(content, "Prepended content")
	existingIdx := strings.Index(content, "Existing content")
	
	if prependIdx == -1 || existingIdx == -1 {
		t.Fatal("Content not found")
	}
	
	if prependIdx > existingIdx {
		t.Error("Prepended content appears after existing content")
	}
}