# Testing Guide for MCP Todo Server

This guide covers how to run tests effectively, especially when using Claude Code or other AI assistants.

## Quick Start

```bash
# Run all tests with Claude Code-friendly output
make test-claude

# Run standard test suite
make test

# Run quick tests without race detector
make test-quick
```

## Test Commands Overview

### For AI Assistants (Claude Code)

#### `make test-claude`
**Best for**: Claude Code and other AI assistants that need clean, predictable output
- **Timeout**: 20 seconds
- **Output**: Minimal, success/fail only
- **Exit codes**: 0 (success) or 1 (failure)
- **Example output**:
  ```
  ✓ All tests passed!
  ```
  or
  ```
  ✗ Tests failed! Output:
  FAIL: TestSomeFunction
  Error: expected X but got Y
  ```

#### `make test-quick`
**Best for**: Quick verification during development
- **Timeout**: 20 seconds  
- **Output**: Verbose test names and results
- **No race detector**: Faster execution

### Standard Test Suite

#### `make test`
**Best for**: Comprehensive testing before commits
- **Timeout**: 30 seconds
- **Features**: Race detector enabled
- **Output**: Verbose with test results

#### `make test-short`
**Best for**: Quick unit tests only
- **Timeout**: 20 seconds
- **Excludes**: Integration tests
- **Features**: Race detector enabled

### Advanced Testing

#### `make test-race`
**Best for**: Detecting race conditions
- **Timeout**: 60 seconds (extended for race detector)
- **Features**: Full race condition analysis
- **Note**: Slower but more thorough

#### `make test-coverage`
**Best for**: Code coverage analysis
- **Timeout**: 30 seconds
- **Output**: HTML coverage report
- **Location**: `./coverage/coverage.html`

#### `make test-integration`
**Best for**: Integration tests only
- **Timeout**: 60 seconds
- **Requires**: `-tags=integration` flag

#### `make test-bench`
**Best for**: Performance benchmarking
- **Features**: Runs all benchmark tests
- **Output**: Performance metrics

## Timeout Configuration

All test commands include timeouts to prevent hanging:

| Command | Timeout | Purpose |
|---------|---------|---------|
| `test-claude` | 20s | Quick AI-friendly output |
| `test-quick` | 20s | Fast development testing |
| `test-short` | 20s | Unit tests only |
| `test` | 30s | Full test suite |
| `test-verbose` | 30s | Detailed output |
| `test-coverage` | 30s | With coverage analysis |
| `test-race` | 60s | Race detection (slower) |
| `test-integration` | 60s | Integration tests |

## Common Issues and Solutions

### Tests Timing Out

If tests are timing out:

1. **Check for blocking operations**:
   - Server initialization without temp directories
   - Infinite loops in tests
   - Missing mock responses

2. **Use shorter timeouts for development**:
   ```bash
   make test-quick  # 20s timeout, no race detector
   ```

3. **Skip long-running tests**:
   ```bash
   make test-short  # Excludes integration tests
   ```

### Claude Code Getting Stuck

Claude Code has a 2-minute timeout for commands. Our test timeouts ensure tests complete well before this limit:

```bash
# Recommended for Claude Code
make test-claude  # Clean output, 20s timeout

# Alternative with more detail
make test-quick   # Verbose output, 20s timeout
```

### Test Failures

When tests fail with `make test-claude`, you'll see only the essential error information:

```
✗ Tests failed! Output:
FAIL: TestHandleTodoUpdate
    todo_test.go:42: Expected status 200, got 404
FAIL: TestSearchTodos  
    search_test.go:91: Index not initialized
```

For more detail, run:
```bash
make test-verbose
```

## Best Practices

### 1. Use Temporary Directories
Always use temp directories in tests to avoid conflicts:

```go
tempDir, err := os.MkdirTemp("", "test-")
if err != nil {
    t.Fatal(err)
}
defer os.RemoveAll(tempDir)

os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
```

### 2. Set Appropriate Timeouts
Individual test timeouts can be set:

```go
func TestSlowOperation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Test code here
}
```

### 3. Isolate Test State
Reset shared state between test cases:

```go
for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        // Reset any shared state
        currentState = initialState
        
        // Run test
    })
}
```

### 4. Use Test Flags
Control test behavior with flags:

```go
// Skip in short mode
if testing.Short() {
    t.Skip("Skipping integration test in short mode")
}
```

## Continuous Integration

For CI pipelines, use:

```bash
# GitHub Actions example
- name: Run tests
  run: make test
  timeout-minutes: 5

# Or for faster CI
- name: Quick tests
  run: make test-quick
  timeout-minutes: 2
```

## Debugging Test Failures

### 1. Run Specific Tests
```bash
# Run a single test
go test -run TestSpecificFunction -v

# Run tests matching a pattern
go test -run "TestTodo.*" -v
```

### 2. Enable Verbose Logging
```bash
# Maximum verbosity
go test -v -count=1 ./...

# Or use make
make test-verbose
```

### 3. Disable Test Caching
```bash
# Force fresh test runs
go test -count=1 ./...
```

### 4. Check Race Conditions
```bash
# Run with race detector
make test-race
```

## Writing AI-Friendly Tests

When writing tests that AI assistants will run:

1. **Clear test names**: Use descriptive names that explain what's being tested
2. **Good error messages**: Include expected vs actual values
3. **Fast execution**: Keep individual tests under 1 second
4. **Proper cleanup**: Always clean up resources in `defer` statements

Example:
```go
func TestTodoCreation(t *testing.T) {
    // Setup
    tempDir := setupTestDir(t)
    defer os.RemoveAll(tempDir)
    
    // Test
    todo, err := CreateTodo("Test task")
    
    // Assertions with clear messages
    if err != nil {
        t.Fatalf("CreateTodo failed: %v", err)
    }
    if todo.Task != "Test task" {
        t.Errorf("Expected task 'Test task', got '%s'", todo.Task)
    }
}
```

## Summary

- Use `make test-claude` for AI assistants - clean output, fast execution
- Use `make test` for comprehensive testing before commits  
- All commands have timeouts to prevent hanging
- Follow best practices for test isolation and cleanup
- Write clear, fast tests with good error messages