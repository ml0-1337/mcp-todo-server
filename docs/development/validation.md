# Validation Patterns

This document describes the validation patterns and utilities used in the MCP Todo Server project.

## Overview

The project centralizes validation logic in the `internal/validation` package to promote reuse and consistency across the codebase. This approach reduces duplication and ensures validation rules are applied uniformly.

## Package Structure

```
internal/validation/
└── todo_validators.go    # Todo-specific validation functions and constants
```

## Constants

The package defines valid values for various todo fields:

```go
// Valid priorities
var ValidPriorities = []string{"high", "medium", "low"}

// Valid todo types
var ValidTodoTypes = []string{
    "feature", "bug", "refactor", "research",
    "multi-phase", "phase", "subtask",
}

// Valid formats for todo display
var ValidFormats = []string{"full", "summary", "list"}

// Valid operations for todo updates
var ValidOperations = []string{"append", "replace", "prepend", "toggle"}

// Valid schema types for sections
var ValidSchemas = []string{
    "freeform", "checklist", "test_cases",
    "research", "strategy", "results",
}

// Valid link types for todo relationships
var ValidLinkTypes = []string{"parent-child", "blocks", "relates-to"}

// Valid cleanup operations
var ValidCleanupOperations = []string{"archive_old", "find_duplicates"}

// Valid time periods for stats
var ValidPeriods = []string{"all", "week", "month", "quarter"}

// Valid scopes for search
var ValidScopes = []string{"task", "findings", "tests", "all"}
```

## Validation Functions

### Basic Validators

```go
// IsValidPriority checks if a priority value is valid
func IsValidPriority(priority string) bool {
    return isInList(priority, ValidPriorities)
}

// IsValidTodoType checks if a todo type is valid
func IsValidTodoType(todoType string) bool {
    return isInList(todoType, ValidTodoTypes)
}

// IsValidFormat checks if a format is valid
func IsValidFormat(format string) bool {
    return isInList(format, ValidFormats)
}

// IsValidOperation checks if an operation is valid
func IsValidOperation(operation string) bool {
    return isInList(operation, ValidOperations)
}
```

### Complex Validators

```go
// ValidateTodoID checks if a todo ID is valid
func ValidateTodoID(id string) error {
    if id == "" {
        return errors.New("todo ID cannot be empty")
    }
    if !isValidIDFormat(id) {
        return fmt.Errorf("invalid todo ID format: %s", id)
    }
    return nil
}

// ValidateParentChildRelationship validates parent-child constraints
func ValidateParentChildRelationship(parentType, childType string) error {
    // Children cannot be multi-phase type
    if childType == "multi-phase" {
        return errors.New("child todos cannot be of type 'multi-phase'")
    }
    
    // Phase and subtask types require appropriate parents
    if childType == "phase" && parentType != "multi-phase" {
        return errors.New("phase todos must have multi-phase parent")
    }
    
    return nil
}

// ValidateSectionKey checks if a section key is valid
func ValidateSectionKey(key string) error {
    if key == "" {
        return errors.New("section key cannot be empty")
    }
    if !isValidSectionKeyFormat(key) {
        return fmt.Errorf("invalid section key format: %s", key)
    }
    return nil
}
```

## Usage Patterns

### In Handlers

The handlers use validation functions from both the `handlers` package and the `internal/validation` package:

```go
import (
    "github.com/user/mcp-todo-server/internal/validation"
    interrors "github.com/user/mcp-todo-server/internal/errors"
)

// In parameter extraction
func ExtractTodoCreateParams(request mcp.CallToolRequest) (*TodoCreateParams, error) {
    // ... extract parameters ...
    
    // Validate using package validators
    if !validation.IsValidPriority(params.Priority) {
        return nil, fmt.Errorf("invalid priority '%s', must be one of: %v", 
            params.Priority, validation.ValidPriorities)
    }
    
    if !validation.IsValidTodoType(params.Type) {
        return nil, fmt.Errorf("invalid type '%s', must be one of: %v", 
            params.Type, validation.ValidTodoTypes)
    }
    
    // Validate parent-child constraints
    if (params.Type == "phase" || params.Type == "subtask") && params.ParentID == "" {
        return nil, fmt.Errorf("type '%s' requires parent_id to be specified", params.Type)
    }
    
    return params, nil
}
```

### In Core Package

The core package can also use the validation utilities:

```go
import "github.com/user/mcp-todo-server/internal/validation"

func (m *TodoManager) CreateTodo(task, priority, todoType string) (*Todo, error) {
    // Validate inputs
    if !validation.IsValidPriority(priority) {
        return nil, interrors.NewValidationError("priority", priority, 
            fmt.Sprintf("must be one of: %v", validation.ValidPriorities))
    }
    
    if !validation.IsValidTodoType(todoType) {
        return nil, interrors.NewValidationError("type", todoType,
            fmt.Sprintf("must be one of: %v", validation.ValidTodoTypes))
    }
    
    // ... create todo ...
}
```

## Helper Functions in Handlers Package

The `handlers` package provides additional validation helpers that work with the MCP protocol:

```go
// ValidateRequiredParam checks if a required parameter is present
func ValidateRequiredParam(param, name string) error {
    if param == "" {
        return fmt.Errorf("missing required parameter '%s'", name)
    }
    return nil
}

// ValidateEnum checks if a value is in the allowed set
func ValidateEnum(value string, allowed []string, paramName string) error {
    for _, a := range allowed {
        if value == a {
            return nil
        }
    }
    return fmt.Errorf("invalid %s '%s': must be one of: %v", 
        paramName, value, allowed)
}
```

## Testing Validation

Use table-driven tests for validation functions:

```go
func TestIsValidPriority(t *testing.T) {
    tests := []struct {
        name     string
        priority string
        want     bool
    }{
        {"valid high", "high", true},
        {"valid medium", "medium", true},
        {"valid low", "low", true},
        {"invalid empty", "", false},
        {"invalid urgent", "urgent", false},
        {"case sensitive", "HIGH", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validation.IsValidPriority(tt.priority)
            if got != tt.want {
                t.Errorf("IsValidPriority(%q) = %v, want %v", 
                    tt.priority, got, tt.want)
            }
        })
    }
}
```

## Best Practices

1. **Centralize validation logic**: Keep all validation rules in the `internal/validation` package
2. **Use constants**: Define valid values as package constants for reuse
3. **Return specific errors**: Include the invalid value and valid options in error messages
4. **Validate early**: Check inputs at the handler layer before processing
5. **Be consistent**: Use the same validation everywhere a field is used
6. **Test thoroughly**: Include edge cases and invalid inputs in tests

## Adding New Validators

When adding new validation:

1. Add constants to `internal/validation/todo_validators.go`
2. Create validation function following naming pattern `IsValidXxx`
3. Add tests for the new validator
4. Update this documentation
5. Use the validator in all relevant code paths

## Migration from Inline Validation

When refactoring existing inline validation:

### Before
```go
// In handlers
validPriorities := []string{"high", "medium", "low"}
valid := false
for _, p := range validPriorities {
    if priority == p {
        valid = true
        break
    }
}
if !valid {
    return errors.New("invalid priority")
}
```

### After
```go
// In handlers
if !validation.IsValidPriority(priority) {
    return fmt.Errorf("invalid priority '%s', must be one of: %v",
        priority, validation.ValidPriorities)
}
```

This approach reduces code duplication and ensures consistency across the codebase.