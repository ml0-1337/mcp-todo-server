package core

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// extractSection extracts content from a specific section in the markdown
func extractSection(content, sectionHeader string) string {
	lines := strings.Split(content, "\n")
	sectionStart := -1
	sectionEnd := len(lines)

	// Find section start
	for i, line := range lines {
		if strings.TrimSpace(line) == "## "+sectionHeader {
			sectionStart = i + 1
			// Skip the standard empty line after section header if present
			if sectionStart < len(lines) && strings.TrimSpace(lines[sectionStart]) == "" {
				sectionStart++
			}
			break
		}
	}

	if sectionStart == -1 {
		return ""
	}

	// Find next section
	for i := sectionStart; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			sectionEnd = i
			// Remove trailing empty lines before next section
			for sectionEnd > sectionStart && strings.TrimSpace(lines[sectionEnd-1]) == "" {
				sectionEnd--
			}
			break
		}
	}

	// Extract content
	result := strings.Join(lines[sectionStart:sectionEnd], "\n")

	// Trim trailing newlines
	result = strings.TrimRight(result, "\n")

	return result
}

// TestAppendToSection_NormalContent tests appending to a section with normal content
func TestAppendToSection_NormalContent(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test normal append", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add initial content
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "Initial content", nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Append new content
	err = manager.UpdateTodo(todo.ID, "findings", "append", "Appended content", nil)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")
	expectedContent := "Initial content\n\nAppended content"
	if findings != expectedContent {
		t.Errorf("Expected findings to be:\n%s\n\nBut got:\n%s", expectedContent, findings)
	}
}

// TestAppendToSection_LeadingEmptyLines tests the bug case where content starts with empty lines
func TestAppendToSection_LeadingEmptyLines(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test append with empty lines", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add content that starts with empty lines
	// Note: replace will trim the content, so "\n\nContent" becomes just "Content"
	// with the standard empty line after the section header
	contentWithEmptyLines := "\n\nContent after empty lines"
	err = manager.UpdateTodo(todo.ID, "findings", "replace", contentWithEmptyLines, nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Append new content - it should go at the END
	err = manager.UpdateTodo(todo.ID, "findings", "append", "This should be at the end", nil)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")

	// Debug: show the raw content
	t.Logf("Raw content:\n%s", content)
	t.Logf("Extracted findings:\n%q", findings)

	// The content should be at the end, and extractSection skips the standard empty line after header
	expectedContent := "Content after empty lines\n\nThis should be at the end"
	if findings != expectedContent {
		t.Errorf("Expected findings to be:\n%q\n\nBut got:\n%q", expectedContent, findings)

		// Show where the content was actually inserted
		if strings.Contains(findings, "This should be at the end\n\nContent after empty lines") {
			t.Errorf("BUG CONFIRMED: Content was inserted in the MIDDLE instead of at the END")
		}
	}
}

// TestAppendToSection_EmptySection tests appending to an empty section
func TestAppendToSection_EmptySection(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test append to empty", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Append to empty section
	err = manager.UpdateTodo(todo.ID, "findings", "append", "First content", nil)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")
	expectedContent := "First content"
	if findings != expectedContent {
		t.Errorf("Expected findings to be:\n%s\n\nBut got:\n%s", expectedContent, findings)
	}
}

// TestAppendToSection_TrailingEmptyLines tests appending when content has trailing empty lines
func TestAppendToSection_TrailingEmptyLines(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test append with trailing empty", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add content with trailing empty lines
	contentWithTrailing := "Some content\n\n\n"
	err = manager.UpdateTodo(todo.ID, "findings", "replace", contentWithTrailing, nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Append new content
	err = manager.UpdateTodo(todo.ID, "findings", "append", "Appended after trailing", nil)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")
	expectedContent := "Some content\n\nAppended after trailing"
	if findings != expectedContent {
		t.Errorf("Expected findings to be:\n%q\n\nBut got:\n%q", expectedContent, findings)
	}
}

// TestAppendToSection_MixedEmptyLines tests appending when content has empty lines throughout
func TestAppendToSection_MixedEmptyLines(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test append with mixed empty", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add content with empty lines throughout
	mixedContent := "First paragraph\n\n\nSecond paragraph\n\nThird paragraph"
	err = manager.UpdateTodo(todo.ID, "findings", "replace", mixedContent, nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Append new content
	err = manager.UpdateTodo(todo.ID, "findings", "append", "Fourth paragraph", nil)
	if err != nil {
		t.Fatalf("Failed to append content: %v", err)
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")
	expectedContent := "First paragraph\n\n\nSecond paragraph\n\nThird paragraph\n\nFourth paragraph"
	if findings != expectedContent {
		t.Errorf("Expected findings to be:\n%q\n\nBut got:\n%q", expectedContent, findings)
	}
}

// TestAppendToSection_MultipleAppends tests multiple consecutive appends
func TestAppendToSection_MultipleAppends(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test multiple appends", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Start with content that has leading empty lines
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "\n\nFirst content", nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Multiple appends
	appends := []string{"Second append", "Third append", "Fourth append"}
	for _, content := range appends {
		err = manager.UpdateTodo(todo.ID, "findings", "append", content, nil)
		if err != nil {
			t.Fatalf("Failed to append %s: %v", content, err)
		}
	}

	// Read the todo content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	// Extract findings section
	findings := extractSection(content, "Findings & Research")

	// Check that all content appears in the correct order
	expectedOrder := []string{"First content", "Second append", "Third append", "Fourth append"}
	lastIndex := -1
	for _, expected := range expectedOrder {
		index := strings.Index(findings, expected)
		if index == -1 {
			t.Errorf("Could not find '%s' in findings", expected)
		} else if index <= lastIndex {
			t.Errorf("Content '%s' appears before previous content (index %d <= %d)", expected, index, lastIndex)
		}
		lastIndex = index
	}
}

// TestAppendToSection_BugRegressionWithLeadingEmpty ensures the append bug with leading empty lines stays fixed
func TestAppendToSection_BugRegressionWithLeadingEmpty(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-regression-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Append regression test", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Replace with content that would trigger the bug (leading empty lines)
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "\n\nContent with leading empty", nil)
	if err != nil {
		t.Fatalf("Failed to replace: %v", err)
	}

	// Append - this should go at the END
	err = manager.UpdateTodo(todo.ID, "findings", "append", "This MUST be at the end", nil)
	if err != nil {
		t.Fatalf("Failed to append: %v", err)
	}

	// Read and verify
	content, _ := manager.ReadTodoContent(todo.ID)

	// Check that content appears in correct order
	contentIdx := strings.Index(content, "Content with leading empty")
	appendIdx := strings.Index(content, "This MUST be at the end")

	if contentIdx == -1 || appendIdx == -1 {
		t.Fatal("Content not found in file")
	}

	if appendIdx < contentIdx {
		t.Error("REGRESSION: Appended content appears BEFORE original content - bug has returned!")
	}
}

// TestAppendToSection_RawFileVerification directly verifies the file content
func TestAppendToSection_RawFileVerification(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-append-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test raw file", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add content with leading empty lines
	err = manager.UpdateTodo(todo.ID, "findings", "replace", "\n\nContent with leading empty", nil)
	if err != nil {
		t.Fatalf("Failed to add initial content: %v", err)
	}

	// Append
	err = manager.UpdateTodo(todo.ID, "findings", "append", "Should be at end", nil)
	if err != nil {
		t.Fatalf("Failed to append: %v", err)
	}

	// Read the raw file
	todoPath, _ := ResolveTodoPath(tempDir, todo.ID)
	rawContent, err := ioutil.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read raw file: %v", err)
	}

	// Check the raw file content
	fileStr := string(rawContent)

	// Find the findings section
	findingsIndex := strings.Index(fileStr, "## Findings & Research")
	if findingsIndex == -1 {
		t.Fatalf("Could not find findings section in file")
	}

	// Find next section
	nextSectionIndex := strings.Index(fileStr[findingsIndex+1:], "\n## ")
	if nextSectionIndex != -1 {
		nextSectionIndex += findingsIndex + 1
	} else {
		nextSectionIndex = len(fileStr)
	}

	// Extract findings section
	findingsSection := fileStr[findingsIndex:nextSectionIndex]

	// Verify order in raw content
	contentIndex := strings.Index(findingsSection, "Content with leading empty")
	appendIndex := strings.Index(findingsSection, "Should be at end")

	if contentIndex == -1 || appendIndex == -1 {
		t.Errorf("Could not find expected content in section:\n%s", findingsSection)
	} else if appendIndex < contentIndex {
		t.Errorf("BUG: Appended content appears BEFORE original content in raw file:\n%s", findingsSection)
	}
}
