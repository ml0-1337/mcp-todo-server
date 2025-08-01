package core

import (
	"io/ioutil"
	"strings"
	"testing"
)

// Test 1: web_searches section is included in standardSectionMappings
func TestWebSearchesSectionInStandardMappings(t *testing.T) {
	// Check if "## Web Searches" is in the standard mappings
	mapping, exists := standardSectionMappings["## Web Searches"]

	if !exists {
		t.Fatal("web_searches section not found in standardSectionMappings")
	}

	// Verify the key is correct
	if mapping.Key != "web_searches" {
		t.Errorf("Expected key 'web_searches', got '%s'", mapping.Key)
	}

	// Verify it uses research schema
	if mapping.Schema != SchemaResearch {
		t.Errorf("Expected schema SchemaResearch, got '%s'", mapping.Schema)
	}
}

// Test 2: web_searches section uses SchemaResearch validation
func TestWebSearchesUsesResearchValidation(t *testing.T) {
	// Get validator for research schema
	validator := GetValidator(SchemaResearch)

	// Test that it accepts any text content
	testContent := `[2025-06-29] Query: "test driven development best practices"
Results: TDD involves writing tests first...
	
[2025-06-29] Query: "golang testing patterns"
Key findings:
- Table-driven tests are idiomatic
- Use subtests for better organization`

	err := validator.Validate(testContent)
	if err != nil {
		t.Errorf("Research validator should accept any content, got error: %v", err)
	}

	// Check metrics
	metrics := validator.GetMetrics(testContent)
	wordCount, ok := metrics["word_count"].(int)
	if !ok {
		t.Fatal("Expected word_count metric")
	}
	if wordCount == 0 {
		t.Error("Expected non-zero word count")
	}
}

// Test 3: web_searches section has correct order (after findings, before test_strategy)
func TestWebSearchesSectionOrder(t *testing.T) {
	// Create markdown content with sections in expected order
	content := `---
todo_id: test-todo
---
# Test Todo

## Findings & Research
Some findings here

## Web Searches
[2025-06-29] Query: "test"
Results: test results

## Test Strategy
Test strategy content

## Test List
- [ ] Test 1
`

	// Infer sections from markdown
	sections := InferSectionsFromMarkdown(content)

	// Get ordered sections
	ordered := GetOrderedSections(sections)

	// Verify we have the expected sections
	if len(ordered) != 4 {
		t.Fatalf("Expected 4 sections, got %d", len(ordered))
	}

	// Check the order
	expectedOrder := []string{"findings", "web_searches", "test_strategy", "test_list"}
	for i, expected := range expectedOrder {
		if ordered[i].Key != expected {
			t.Errorf("Position %d: expected '%s', got '%s'", i, expected, ordered[i].Key)
		}
	}

	// Verify specific orders
	if sections["findings"].Order >= sections["web_searches"].Order {
		t.Error("web_searches should come after findings")
	}
	if sections["web_searches"].Order >= sections["test_strategy"].Order {
		t.Error("web_searches should come before test_strategy")
	}
}

// Test 4: New todos include web_searches section by default
func TestNewTodosIncludeWebSearchesSection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a TodoManager
	tm := NewTodoManager(tempDir)

	// Create a new todo
	todo, err := tm.CreateTodo("Test task with web searches", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Read the todo file using ResolveTodoPath to handle date-based structure
	filePath, err := ResolveTodoPath(tempDir, todo.ID)
	if err != nil {
		t.Fatalf("Failed to resolve todo path: %v", err)
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read todo file: %v", err)
	}

	// Check that web searches section is included
	if !strings.Contains(string(content), "## Web Searches") {
		t.Error("New todo should include ## Web Searches section")
	}

	// Verify order: should be after Findings and before Test Strategy
	contentStr := string(content)
	findingsIdx := strings.Index(contentStr, "## Findings & Research")
	webSearchesIdx := strings.Index(contentStr, "## Web Searches")
	testStrategyIdx := strings.Index(contentStr, "## Test Strategy")

	if findingsIdx == -1 || webSearchesIdx == -1 || testStrategyIdx == -1 {
		t.Fatal("Expected sections not found")
	}

	if webSearchesIdx <= findingsIdx {
		t.Error("Web Searches should come after Findings & Research")
	}
	if webSearchesIdx >= testStrategyIdx {
		t.Error("Web Searches should come before Test Strategy")
	}
}
