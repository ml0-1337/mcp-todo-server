# Error Handling Patterns

This document describes the error handling patterns used in the MCP Todo Server project.

## Overview

The project uses structured error handling with custom error types and consistent error wrapping to provide meaningful context at each layer of the application. This approach follows Go best practices established since Go 1.13.

## Package Structure

All error handling code is located in the `internal/errors` package:

```
internal/errors/
├── errors.go    # Sentinel errors and utility functions
└── types.go     # Custom error types
```

## Sentinel Errors

We define common sentinel errors that can be checked using `errors.Is()`:

```go
var (
    ErrNotFound   = errors.New("not found")
    ErrValidation = errors.New("validation error")
    ErrOperation  = errors.New("operation failed")
    ErrPermission = errors.New("permission denied")
    ErrConflict   = errors.New("resource conflict")
    ErrInternal   = errors.New("internal error")
)
```

### Usage Example

```go
import interrors "github.com/user/mcp-todo-server/internal/errors"

// In core package
todo, err := manager.ReadTodo(id)
if err != nil {
    if os.IsNotExist(err) {
        return interrors.NewNotFoundError("todo", id)
    }
    return interrors.Wrap(err, "failed to read todo")
}

// In handler layer
if interrors.IsNotFound(err) {
    return mcp.NewToolResultError("Todo not found")
}
```

## Custom Error Types

### TodoError

Used for todo-specific operations:

```go
type TodoError struct {
    ID        string
    Operation string
    Category  ErrorCategory
    Cause     error
    Message   string
}
```

### ValidationError

Used for input validation failures:

```go
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
    Cause   error
}
```

### NotFoundError

Used when a resource cannot be found:

```go
type NotFoundError struct {
    Resource string
    ID       string
}
```

### ConflictError

Used for resource conflicts (e.g., section already exists):

```go
type ConflictError struct {
    Resource string
    ID       string
    Message  string
}
```

## Error Wrapping

Always wrap errors with context using the `Wrap` function:

```go
func Wrap(err error, message string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", message, err)
}
```

### Wrapping Guidelines

1. **Add context at each layer**: Each function should add its own context
2. **Be specific**: Include relevant IDs, operations, or field names
3. **Keep messages concise**: One line describing what failed
4. **Preserve the original error**: Always use `%w` for wrapping

### Example

```go
// Bad: Generic error
if err != nil {
    return err
}

// Good: Wrapped with context
if err != nil {
    return interrors.Wrap(err, fmt.Sprintf("failed to update todo %s", todoID))
}
```

## Error Checking

Use `errors.Is()` and `errors.As()` for error checking:

```go
// Check for specific error
if interrors.IsNotFound(err) {
    // Handle not found case
}

// Extract error details
var validErr *interrors.ValidationError
if errors.As(err, &validErr) {
    log.Printf("Validation failed for field %s: %s", validErr.Field, validErr.Message)
}
```

## Error Categories

Errors are categorized for better handling:

```go
type ErrorCategory int

const (
    CategoryUnknown ErrorCategory = iota
    CategoryNotFound
    CategoryValidation
    CategoryOperation
    CategoryPermission
    CategoryConflict
    CategoryInternal
)
```

Use `GetCategory()` to determine error category:

```go
switch interrors.GetCategory(err) {
case interrors.CategoryNotFound:
    return http.StatusNotFound
case interrors.CategoryValidation:
    return http.StatusBadRequest
// ...
}
```

## Layer-Specific Patterns

### Core Layer

- Return specific error types (`NotFoundError`, `ValidationError`)
- Wrap system errors with context
- Never expose internal paths in error messages

```go
func (m *TodoManager) ReadTodo(id string) (*Todo, error) {
    data, err := os.ReadFile(todoPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, interrors.NewNotFoundError("todo", id)
        }
        return nil, interrors.Wrap(err, "failed to read todo file")
    }
    // ...
}
```

### Handler Layer

- Convert errors to MCP responses
- Log technical details, return user-friendly messages
- Use error categories for consistent handling

```go
func HandleError(err error) mcp.ToolResult {
    switch {
    case interrors.IsNotFound(err):
        return mcp.NewToolResultError("Todo not found")
    case interrors.IsValidation(err):
        var validErr *interrors.ValidationError
        if errors.As(err, &validErr) {
            return mcp.NewToolResultError(fmt.Sprintf("Invalid %s: %s", validErr.Field, validErr.Message))
        }
        return mcp.NewToolResultError("Validation error")
    default:
        // Log technical error
        log.Printf("Internal error: %v", err)
        // Return generic message to user
        return mcp.NewToolResultError("An error occurred processing your request")
    }
}
```

## Security Considerations

1. **Never expose internal paths**: File paths should not appear in user-facing errors
2. **Log technical details separately**: Use structured logging for debugging
3. **Sanitize user input in errors**: Prevent injection attacks
4. **Use generic messages for internal errors**: Don't leak implementation details

### Example

```go
// Bad: Exposes internal path
return fmt.Errorf("failed to read /home/user/.claude/todos/abc.md: %w", err)

// Good: Generic message with ID
return interrors.Wrap(err, fmt.Sprintf("failed to read todo %s", todoID))
```

## Testing Error Handling

Use the assertion helpers from `internal/testutil`:

```go
func TestErrorHandling(t *testing.T) {
    err := someFunction()
    
    // Check error type
    testutil.AssertErrorIs(t, err, interrors.ErrNotFound)
    
    // Check error category
    testutil.AssertErrorCategory(t, err, interrors.CategoryNotFound)
    
    // Check error message contains substring
    testutil.AssertErrorContains(t, err, "todo not found")
    
    // Extract and verify custom error
    var todoErr *interrors.TodoError
    testutil.AssertErrorAs(t, err, &todoErr)
    testutil.AssertEqual(t, todoErr.ID, "test-todo", "todo ID")
}
```

## Migration Guide

When updating existing code to use the new error handling:

1. Replace `errors.New()` with appropriate sentinel errors or custom types
2. Add error wrapping at each function boundary
3. Update error checks to use `errors.Is()` instead of string comparison
4. Convert to specific error types where appropriate

### Before

```go
if err != nil {
    return errors.New("todo not found")
}
```

### After

```go
if err != nil {
    if os.IsNotExist(err) {
        return interrors.NewNotFoundError("todo", id)
    }
    return interrors.Wrap(err, "failed to read todo")
}
```

## Best Practices Summary

1. **Always wrap errors** with context using `Wrap()`
2. **Use specific error types** for domain errors
3. **Check errors with Is/As**, not string comparison
4. **Add context at each layer** without duplicating
5. **Keep user messages generic** for security
6. **Log technical details** separately from user errors
7. **Test error paths** as thoroughly as success paths
8. **Document error conditions** in function comments