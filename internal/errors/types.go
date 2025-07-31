package errors

import (
	"fmt"
	"strings"
)

// TodoError represents a todo-specific error with context
type TodoError struct {
	Cause     error
	ID        string
	Operation string
	Message   string
	Category  ErrorCategory
}

// Error implements the error interface
func (e *TodoError) Error() string {
	parts := []string{}

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation %s", e.Operation))
	}

	if e.ID != "" {
		parts = append(parts, fmt.Sprintf("todo %s", e.ID))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error
func (e *TodoError) Unwrap() error {
	return e.Cause
}

// NewTodoError creates a new todo error
func NewTodoError(id, operation, message string, category ErrorCategory, cause error) *TodoError {
	return &TodoError{
		Cause:     cause,
		ID:        id,
		Operation: operation,
		Message:   message,
		Category:  category,
	}
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Cause   error
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		if e.Value != nil && e.Value != "" {
			return fmt.Sprintf("validation error for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
		}
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Cause:   ErrValidation,
	}
}

// NotFoundError represents a not found error with resource details
type NotFoundError struct {
	ResourceType string
	ResourceID   string
	Message      string
}

// Error implements the error interface
func (e *NotFoundError) Error() string {
	if e.ResourceType != "" && e.ResourceID != "" {
		return fmt.Sprintf("%s '%s' not found", e.ResourceType, e.ResourceID)
	}
	if e.Message != "" {
		return e.Message
	}
	return "resource not found"
}

// Unwrap returns the underlying error
func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resourceType, resourceID string) *NotFoundError {
	return &NotFoundError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// OperationError represents a failed operation with context
type OperationError struct {
	Operation string
	Resource  string
	Message   string
	Cause     error
}

// Error implements the error interface
func (e *OperationError) Error() string {
	parts := []string{}

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation '%s' failed", e.Operation))
	}

	if e.Resource != "" {
		parts = append(parts, fmt.Sprintf("on %s", e.Resource))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error
func (e *OperationError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return ErrOperation
}

// NewOperationError creates a new operation error
func NewOperationError(operation, resource, message string, cause error) *OperationError {
	return &OperationError{
		Operation: operation,
		Resource:  resource,
		Message:   message,
		Cause:     cause,
	}
}

// ConflictError represents a resource conflict error
type ConflictError struct {
	Resource string
	ID       string
	Message  string
}

// Error implements the error interface
func (e *ConflictError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("conflict for %s '%s': %s", e.Resource, e.ID, e.Message)
	}
	return fmt.Sprintf("conflict for %s '%s'", e.Resource, e.ID)
}

// Unwrap returns the underlying error
func (e *ConflictError) Unwrap() error {
	return ErrConflict
}

// NewConflictError creates a new ConflictError
func NewConflictError(resource, id, message string) *ConflictError {
	return &ConflictError{
		Resource: resource,
		ID:       id,
		Message:  message,
	}
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	messages := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	return fmt.Sprintf("multiple errors: [%s]", strings.Join(messages, "; "))
}

// Add adds an error to the multi error
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Unwrap returns the first error for compatibility
func (e *MultiError) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

// NewMultiError creates a new multi error
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make([]error, 0),
	}
}
