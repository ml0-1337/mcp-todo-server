package core

import (
	"testing"
	"time"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Test 1: Parse todo with standard timestamp format - Should parse successfully
func TestParseTodoWithStandardTimestampFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo content with standard format
	content := `---
todo_id: test-standard-timestamp
started: 2025-06-28 10:30:00
completed: 2025-06-28 11:45:00
status: completed
priority: high
type: feature
---

# Task: Test standard timestamp parsing

## Description
This todo uses the standard timestamp format.
`

	// Write test file
	filePath := filepath.Join(tempDir, "test-standard-timestamp.md")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create manager and read todo
	manager := NewTodoManager(tempDir)
	todo, err := manager.ReadTodo("test-standard-timestamp")
	
	// This test should pass with current implementation
	if err != nil {
		t.Fatalf("Failed to parse todo with standard timestamp format: %v", err)
	}

	// Verify timestamps were parsed correctly
	expectedStart, _ := time.Parse("2006-01-02 15:04:05", "2025-06-28 10:30:00")
	expectedComplete, _ := time.Parse("2006-01-02 15:04:05", "2025-06-28 11:45:00")

	if !todo.Started.Equal(expectedStart) {
		t.Errorf("Started time mismatch. Expected: %v, Got: %v", expectedStart, todo.Started)
	}

	if !todo.Completed.Equal(expectedComplete) {
		t.Errorf("Completed time mismatch. Expected: %v, Got: %v", expectedComplete, todo.Completed)
	}
}

// Test 2: Parse todo with RFC3339 timestamp format - Should parse successfully with fallback
func TestParseTodoWithRFC3339TimestampFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo content with RFC3339 format
	content := `---
todo_id: test-rfc3339-timestamp
started: 2025-01-29T02:55:00Z
completed: 2025-01-29T02:58:00Z
status: completed
priority: low
type: test
---

# Task: Test RFC3339 timestamp parsing

## Description
This todo uses RFC3339 timestamp format like the problematic test-archive-demo.md.
`

	// Write test file
	filePath := filepath.Join(tempDir, "test-rfc3339-timestamp.md")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create manager and read todo
	manager := NewTodoManager(tempDir)
	todo, err := manager.ReadTodo("test-rfc3339-timestamp")
	
	// This test should now PASS with the fix
	if err != nil {
		t.Fatalf("Failed to parse todo with RFC3339 timestamp format: %v", err)
	}

	// Verify timestamps were parsed correctly
	expectedStart, _ := time.Parse(time.RFC3339, "2025-01-29T02:55:00Z")
	expectedComplete, _ := time.Parse(time.RFC3339, "2025-01-29T02:58:00Z")

	if !todo.Started.Equal(expectedStart) {
		t.Errorf("Started time mismatch. Expected: %v, Got: %v", expectedStart, todo.Started)
	}

	if !todo.Completed.Equal(expectedComplete) {
		t.Errorf("Completed time mismatch. Expected: %v, Got: %v", expectedComplete, todo.Completed)
	}
}

// Test 3: Parse todo with RFC3339Nano timestamp format - Should parse successfully
func TestParseTodoWithRFC3339NanoTimestampFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo content with RFC3339Nano format
	content := `---
todo_id: test-rfc3339nano-timestamp
started: 2025-06-28T10:30:00.123456789Z
completed: 2025-06-28T11:45:00.987654321Z
status: completed
priority: medium
type: feature
---

# Task: Test RFC3339Nano timestamp parsing

## Description
This todo uses RFC3339Nano timestamp format with nanosecond precision.
`

	// Write test file
	filePath := filepath.Join(tempDir, "test-rfc3339nano-timestamp.md")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create manager and read todo
	manager := NewTodoManager(tempDir)
	todo, err := manager.ReadTodo("test-rfc3339nano-timestamp")
	
	// This test should now PASS with the fix
	if err != nil {
		t.Fatalf("Failed to parse todo with RFC3339Nano timestamp format: %v", err)
	}

	// Verify timestamps were parsed correctly
	expectedStart, _ := time.Parse(time.RFC3339Nano, "2025-06-28T10:30:00.123456789Z")
	expectedComplete, _ := time.Parse(time.RFC3339Nano, "2025-06-28T11:45:00.987654321Z")

	if !todo.Started.Equal(expectedStart) {
		t.Errorf("Started time mismatch. Expected: %v, Got: %v", expectedStart, todo.Started)
	}

	if !todo.Completed.Equal(expectedComplete) {
		t.Errorf("Completed time mismatch. Expected: %v, Got: %v", expectedComplete, todo.Completed)
	}
}

// Test 4: Parse todo with invalid timestamp format - Should return appropriate error
func TestParseTodoWithInvalidTimestampFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo content with invalid format
	content := `---
todo_id: test-invalid-timestamp
started: June 28, 2025 at 10:30 AM
completed: June 28, 2025 at 11:45 AM
status: completed
priority: low
type: test
---

# Task: Test invalid timestamp parsing

## Description
This todo uses an invalid timestamp format.
`

	// Write test file
	filePath := filepath.Join(tempDir, "test-invalid-timestamp.md")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create manager and read todo
	manager := NewTodoManager(tempDir)
	_, err = manager.ReadTodo("test-invalid-timestamp")
	
	// This test should always FAIL, even after implementation
	if err == nil {
		t.Fatal("Expected parsing to fail with invalid timestamp format, but it succeeded")
	}

	// Verify the error message contains expected parsing failure
	expectedError := "failed to parse started timestamp"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && containsHelper(s[1:], substr)
}

// Recursive helper for contains
func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}