# Test Utilities Documentation

This document describes the test utilities provided by the `internal/testutil` package and how to use them effectively.

## Overview

The `testutil` package provides a comprehensive set of test helpers, fixtures, and assertions designed to reduce boilerplate and improve test readability. All utilities follow modern Go testing practices including proper use of `t.Helper()`, `t.Cleanup()`, and `t.TempDir()`.

## Package Structure

```
internal/testutil/
├── helpers.go     # Test setup and environment helpers
├── fixtures.go    # Test data builders and generators
└── assertions.go  # Custom assertion functions
```

## Test Helpers (`helpers.go`)

### SetupTestTodoManager

Creates a fully configured `TodoManager` with a temporary directory:

```go
func TestTodoOperations(t *testing.T) {
    manager, tempDir := testutil.SetupTestTodoManager(t)
    // tempDir is automatically cleaned up
    
    todo, err := manager.CreateTodo("Test task", "high", "feature")
    testutil.RequireNoError(t, err, "create todo")
}
```

### SetupTestDir

Creates a temporary directory structure for testing:

```go
func TestFileOperations(t *testing.T) {
    tempDir, cleanup := testutil.SetupTestDir(t)
    defer cleanup() // No-op with t.TempDir(), kept for compatibility
    
    // Directory structure is created:
    // tempDir/
    //   └── .claude/
    //       └── todos/
}
```

### CreateTestTodo

Quick helper to create a todo for testing:

```go
func TestTodoUpdate(t *testing.T) {
    manager, _ := testutil.SetupTestTodoManager(t)
    
    todo := testutil.CreateTestTodo(t, manager, "Test task", "medium", "bug")
    // todo is guaranteed to be created successfully
}
```

### CreateTestTodoWithDate

Creates a todo with a specific started date:

```go
func TestArchiveByDate(t *testing.T) {
    manager, _ := testutil.SetupTestTodoManager(t)
    
    pastDate := time.Now().AddDate(0, -3, 0) // 3 months ago
    todo := testutil.CreateTestTodoWithDate(t, manager, "Old task", pastDate)
}
```

### File Assertions

```go
// Check if file exists
testutil.AssertFileExists(t, "/path/to/file.md")

// Check if file doesn't exist
testutil.AssertFileNotExists(t, "/path/to/deleted.md")

// Check file contains substring
testutil.AssertFileContains(t, "/path/to/file.md", "expected content")

// Require file contains (fails test immediately)
testutil.RequireFileContains(t, "/path/to/file.md", "critical content")
```

### Path Helpers

```go
// Check if any path exists (file or directory)
testutil.AssertPathExists(t, "/path/to/resource")

// Wait for file to appear (useful for async operations)
testutil.WaitForFile(t, "/path/to/file.md", 5*time.Second)
```

## Test Fixtures (`fixtures.go`)

### TodoBuilder

Fluent builder for creating test todos with specific configurations:

```go
func TestComplexTodo(t *testing.T) {
    todo := testutil.NewTodoBuilder("Complex task").
        WithPriority("high").
        WithType("feature").
        WithStatus("in_progress").
        WithParentID("parent-123").
        WithTags("urgent", "backend", "api").
        WithSection("findings", "## Findings\n\nDetailed findings here").
        WithSection("tests", "## Tests\n\nTest cases here").
        Build()
    
    // Use todo in tests
    testutil.AssertEqual(t, todo.Priority, "high", "priority")
    testutil.AssertLen(t, todo.Tags, 3, "tags count")
}
```

### Content Generators

Generate realistic test content:

```go
// Generate section content
content := testutil.GenerateSectionContent("findings", 3) // 3 paragraphs

// Generate long-form content
longContent := testutil.GenerateLongContent(500) // ~500 words

// Generate code block
codeContent := testutil.GenerateCodeContent("go", `
func main() {
    fmt.Println("Hello, test!")
}
`)

// Generate test checklist
checklist := testutil.GenerateChecklist(5) // 5 items

// Generate web search results
searchResults := testutil.GenerateWebSearches(3) // 3 search queries
```

### CreateTodoFile

Create a todo file directly on disk:

```go
func TestFileReading(t *testing.T) {
    tempDir := t.TempDir()
    
    todoID := "test-todo-123"
    content := testutil.GenerateTodoContent(
        todoID,
        "Test task",
        "high",
        "bug",
    )
    
    path := testutil.CreateTodoFile(t, tempDir, todoID, content)
    // File is created at path with proper structure
}
```

## Assertions (`assertions.go`)

### Basic Assertions

```go
// Nil checks
testutil.AssertNil(t, err, "error should be nil")
testutil.AssertNotNil(t, todo, "todo should exist")

// Boolean checks
testutil.AssertTrue(t, result, "operation should succeed")
testutil.AssertFalse(t, failed, "should not fail")

// Equality
testutil.AssertEqual(t, actual, expected, "values should match")

// String contains
testutil.AssertContains(t, output, "expected text", "output check")

// Length checks
testutil.AssertLen(t, items, 5, "should have 5 items")
```

### Error Assertions

```go
// Basic error checks
testutil.AssertError(t, err, "should return error")
testutil.AssertNoError(t, err, "should succeed")

// Error type checking
testutil.AssertErrorIs(t, err, interrors.ErrNotFound)
testutil.AssertErrorAs(t, err, &validationErr)

// Error category
testutil.AssertErrorCategory(t, err, interrors.CategoryValidation)

// Error message contains
testutil.AssertErrorContains(t, err, "task cannot be empty")
```

### Todo Assertions

```go
// Compare todos
testutil.AssertTodoEqual(t, expectedTodo, actualTodo)

// This checks:
// - ID, Task, Priority, Type, Status
// - ParentID
// - All other fields
```

### Advanced Assertions

```go
// Panic checking
testutil.AssertPanic(t, func() {
    dangerousOperation()
}, "should panic")

testutil.AssertNoPanic(t, func() {
    safeOperation()
}, "should not panic")

// Eventually (for async operations)
testutil.AssertEventually(t, func() bool {
    return checkCondition()
}, 5*time.Second, 100*time.Millisecond, "condition should be met")

// Slice comparison
expected := []string{"a", "b", "c"}
actual := []string{"a", "b", "c"}
testutil.CompareSlices(t, expected, actual, "slices should match")
```

### Require vs Assert

Use `Require*` variants when the test cannot continue:

```go
// Assert: Test continues even if it fails
testutil.AssertNoError(t, err, "setup error")

// Require: Test stops immediately if it fails
testutil.RequireNoError(t, err, "critical setup")

// Also available:
testutil.RequireError(t, err, "must fail")
```

## Migration Guide

### From os.MkdirTemp to t.TempDir

Before:
```go
tempDir, err := os.MkdirTemp("", "test-*")
if err != nil {
    t.Fatalf("Failed to create temp dir: %v", err)
}
defer os.RemoveAll(tempDir)
```

After:
```go
tempDir := t.TempDir() // Automatic cleanup!
```

### From Manual Setup to Helpers

Before:
```go
tempDir := t.TempDir()
todosDir := filepath.Join(tempDir, ".claude", "todos")
err := os.MkdirAll(todosDir, 0755)
if err != nil {
    t.Fatal(err)
}
manager := core.NewTodoManager(tempDir)
```

After:
```go
manager, tempDir := testutil.SetupTestTodoManager(t)
```

### From Basic Checks to Assertions

Before:
```go
if err != nil {
    t.Errorf("Expected no error, got: %v", err)
}
if todo.Task != expected {
    t.Errorf("Task mismatch: expected %s, got %s", expected, todo.Task)
}
```

After:
```go
testutil.AssertNoError(t, err, "todo creation")
testutil.AssertEqual(t, todo.Task, expected, "task")
```

## Best Practices

### 1. Always Use t.Helper()

All test utilities use `t.Helper()` to ensure error messages point to the test, not the helper:

```go
func MyTestHelper(t testing.TB) {
    t.Helper() // Must be first line
    // ... helper code ...
}
```

### 2. Use testing.TB Interface

Accept `testing.TB` instead of `*testing.T` to support benchmarks:

```go
func SetupBenchmark(b testing.TB) (*TodoManager, string) {
    b.Helper()
    // Works with both *testing.T and *testing.B
}
```

### 3. Prefer t.TempDir()

Always use `t.TempDir()` instead of `os.MkdirTemp`:
- Automatic cleanup
- Parallel test safe
- No defer needed
- Unique directories

### 4. Use Builders for Complex Objects

For complex test data, use builders:

```go
todo := testutil.NewTodoBuilder("Task").
    WithPriority("high").
    WithMultipleSections(map[string]string{
        "findings": "Research results",
        "tests": "Test cases",
    }).
    Build()
```

### 5. Group Related Assertions

```go
t.Run("todo creation", func(t *testing.T) {
    todo := testutil.CreateTestTodo(t, manager, "Task", "high", "feature")
    
    testutil.AssertNotNil(t, todo, "todo")
    testutil.AssertEqual(t, todo.Status, "in_progress", "status")
    testutil.AssertEqual(t, todo.Priority, "high", "priority")
})
```

## Common Patterns

### Table-Driven Tests with Fixtures

```go
func TestTodoValidation(t *testing.T) {
    tests := []struct {
        name    string
        builder func() *domain.Todo
        wantErr error
    }{
        {
            name: "valid todo",
            builder: func() *domain.Todo {
                return testutil.NewTodoBuilder("Valid task").Build()
            },
            wantErr: nil,
        },
        {
            name: "empty task",
            builder: func() *domain.Todo {
                return testutil.NewTodoBuilder("").Build()
            },
            wantErr: interrors.ErrValidation,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            todo := tt.builder()
            err := ValidateTodo(todo)
            
            if tt.wantErr != nil {
                testutil.AssertErrorIs(t, err, tt.wantErr)
            } else {
                testutil.AssertNoError(t, err, "validation")
            }
        })
    }
}
```

### Integration Tests

```go
func TestTodoLifecycle(t *testing.T) {
    manager, _ := testutil.SetupTestTodoManager(t)
    
    // Create
    todo := testutil.CreateTestTodo(t, manager, "Lifecycle test", "medium", "task")
    
    // Update
    err := manager.UpdateTodoSection(todo.ID, "findings", "New findings")
    testutil.RequireNoError(t, err, "update section")
    
    // Read
    updated, err := manager.ReadTodo(todo.ID)
    testutil.RequireNoError(t, err, "read todo")
    testutil.AssertContains(t, updated.RawContent, "New findings", "content")
    
    // Complete
    err = manager.UpdateTodoStatus(todo.ID, "completed")
    testutil.RequireNoError(t, err, "complete todo")
    
    // Archive
    err = manager.ArchiveTodo(todo.ID)
    testutil.RequireNoError(t, err, "archive todo")
    
    // Verify archived
    testutil.AssertFileNotExists(t, manager.GetTodoPath(todo.ID))
}
```

### Cleanup Patterns

```go
func TestWithCleanup(t *testing.T) {
    manager, tempDir := testutil.SetupTestTodoManager(t)
    
    // Register additional cleanup if needed
    t.Cleanup(func() {
        // Custom cleanup code
        log.Printf("Test completed, used dir: %s", tempDir)
    })
    
    // Cleanup runs in LIFO order
}
```

## Summary

The test utilities package provides:

1. **Reduced boilerplate** through helper functions
2. **Better error messages** with t.Helper()
3. **Automatic cleanup** with t.TempDir() and t.Cleanup()
4. **Consistent patterns** across all tests
5. **Type-safe assertions** with clear failure messages
6. **Flexible fixtures** for generating test data

Following these patterns leads to more maintainable and reliable tests.