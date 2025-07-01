# Error Message Guidelines

This document provides guidelines for writing consistent, helpful, and secure error messages in the MCP Todo Server project.

## Message Types

### 1. User-Facing Messages

Messages shown to end users through the MCP interface should be:
- Clear and actionable
- Free of technical jargon
- Focused on what the user can do
- Secure (no internal details)

### 2. Technical Messages

Messages logged for debugging should include:
- Full error chain with context
- Relevant IDs and parameters
- File paths and line numbers (in development)
- Stack traces for panics

## Message Format Standards

### Basic Format

```
<action> failed: <reason>
```

Examples:
- "todo creation failed: task cannot be empty"
- "search failed: invalid query syntax"
- "archive failed: todo must be completed first"

### With Context

```
failed to <action> <resource> <id>: <reason>
```

Examples:
- "failed to read todo abc-123: file not found"
- "failed to update section 'findings' in todo xyz-789: section does not exist"
- "failed to archive todo def-456: todo is not completed"

## Layer-Specific Guidelines

### Core Layer Messages

Focus on what happened at the business logic level:

```go
// Good
return interrors.Wrap(err, fmt.Sprintf("failed to create todo with task '%s'", task))
return interrors.NewValidationError("priority", priority, "must be one of: high, medium, low")
return interrors.NewNotFoundError("todo", todoID)

// Bad
return fmt.Errorf("error") // Too generic
return err // No context
return fmt.Errorf("failed to open file /home/user/.claude/todos/abc.md") // Exposes path
```

### Handler Layer Messages

Transform technical errors into user-friendly messages:

```go
// Technical error (logged)
log.Printf("Failed to create todo: %v", err)

// User-facing message (returned)
switch {
case interrors.IsValidation(err):
    return mcp.NewToolResultError("Invalid input: task cannot be empty")
case interrors.IsNotFound(err):
    return mcp.NewToolResultError("Todo not found")
default:
    return mcp.NewToolResultError("Failed to create todo. Please try again.")
}
```

### Infrastructure Layer Messages

Include system-level context:

```go
// Good
return interrors.Wrap(err, "failed to create search index")
return interrors.Wrap(err, fmt.Sprintf("failed to save todo to file: %s", filename))

// Bad
return err // No context about what operation failed
```

## Common Error Scenarios

### Validation Errors

Be specific about what's wrong and how to fix it:

```go
// Good
"task cannot be empty"
"priority must be one of: high, medium, low"
"status 'archived' is not valid for update"
"date must be in format YYYY-MM-DD"

// Bad
"invalid input"
"bad request"
"validation failed"
```

### Not Found Errors

Include what was being looked for:

```go
// Good
"todo with ID 'abc-123' not found"
"no todos found matching 'search query'"
"template 'bug-fix' does not exist"

// Bad
"not found"
"404"
"does not exist"
```

### Permission Errors

Explain what action was denied:

```go
// Good
"cannot archive todo that is not completed"
"cannot modify archived todo"
"cannot create todo in read-only mode"

// Bad
"permission denied"
"forbidden"
"not allowed"
```

### Conflict Errors

Describe the conflict and possible resolution:

```go
// Good
"section 'findings' already exists in todo"
"cannot create duplicate todo with same task name"
"todo is already archived"

// Bad
"conflict"
"already exists"
"duplicate"
```

## MCP-Specific Formatting

### Tool Result Errors

Follow MCP conventions for tool results:

```go
// Single line error
return mcp.NewToolResultError("Todo not found")

// Multi-line error with details
return mcp.NewToolResultError(fmt.Sprintf(
    "Validation failed:\n- Task: cannot be empty\n- Priority: must be high, medium, or low"
))
```

### Parameter Validation

Use consistent format for parameter errors:

```go
// Missing required parameter
"Missing required parameter: task"

// Invalid parameter value
"Invalid value for priority: 'urgent' (must be one of: high, medium, low)"

// Invalid parameter type
"Parameter 'limit' must be a number"
```

## Error Message Examples

### Good Examples

```go
// Clear and actionable
"Cannot archive todo: mark it as completed first"

// Specific validation error
"Invalid priority 'urgent': use 'high', 'medium', or 'low'"

// Helpful not found message
"No todos found. Create one with todo_create tool"

// Context-aware operation error
"Failed to update todo: section 'tests' does not exist"
```

### Bad Examples

```go
// Too generic
"Error occurred"
"Operation failed"
"Something went wrong"

// Exposes internals
"Failed to read /Users/admin/.claude/todos/abc.md"
"Database connection timeout at 127.0.0.1:5432"
"Panic in handler.go:123"

// Not actionable
"Invalid"
"Failed"
"Error"

// Too technical for users
"ENOENT: no such file or directory"
"Unmarshal error: invalid JSON"
"Nil pointer dereference"
```

## Localization Considerations

While the current implementation uses English, structure messages for future localization:

```go
// Good: Parameterized messages
fmt.Sprintf("Todo %s not found", todoID)
fmt.Sprintf("Invalid %s: %s", fieldName, reason)

// Bad: Hard-coded concatenation
"Todo " + todoID + " not found"
```

## Error Message Checklist

Before adding an error message, verify:

- [ ] **Clear**: Can a user understand what went wrong?
- [ ] **Actionable**: Does it suggest what to do next?
- [ ] **Specific**: Does it identify the exact problem?
- [ ] **Secure**: Does it avoid exposing internal details?
- [ ] **Consistent**: Does it follow the project's format?
- [ ] **Appropriate**: Is the tone suitable for the audience?
- [ ] **Complete**: Does it include necessary context?
- [ ] **Testable**: Can you write a test for this error case?

## Testing Error Messages

Always test error messages:

```go
func TestErrorMessages(t *testing.T) {
    // Test user-facing message
    result, _ := handler.HandleTodoCreate(ctx, invalidRequest)
    testutil.AssertErrorContains(t, result.GetError(), "task cannot be empty")
    
    // Test technical logging
    var logOutput bytes.Buffer
    log.SetOutput(&logOutput)
    
    handler.HandleError(technicalError)
    
    // Verify technical details are logged
    testutil.AssertContains(t, logOutput.String(), "file not found", "log output")
    
    // Verify user message is generic
    testutil.AssertEqual(t, result.GetError(), "An error occurred", "user message")
}
```

## Summary

Good error messages are:
1. **Helpful** to users trying to fix the problem
2. **Secure** by not exposing internal details
3. **Consistent** in format and tone
4. **Specific** about what went wrong
5. **Actionable** in suggesting solutions