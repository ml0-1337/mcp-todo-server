// Package errors provides structured error types and utilities for the MCP Todo Server.
// It includes categorized errors, error wrapping utilities, and type-safe error checking.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common cases
var (
	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errors.New("not found")

	// ErrValidation indicates invalid input or parameters
	ErrValidation = errors.New("validation error")

	// ErrOperation indicates a failed operation
	ErrOperation = errors.New("operation failed")

	// ErrPermission indicates insufficient permissions
	ErrPermission = errors.New("permission denied")

	// ErrConflict indicates a resource conflict
	ErrConflict = errors.New("conflict")

	// ErrInternal indicates an internal server error
	ErrInternal = errors.New("internal error")
)

// Error categories for classification
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

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with formatted context
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New creates a new error with the given message
func New(message string) error {
	return errors.New(message)
}

// Newf creates a new error with a formatted message
func Newf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetCategory returns the category of an error
func GetCategory(err error) ErrorCategory {
	if err == nil {
		return CategoryUnknown
	}

	// Check for custom error types first
	var todoErr *TodoError
	if As(err, &todoErr) {
		return todoErr.Category
	}

	var valErr *ValidationError
	if As(err, &valErr) {
		return CategoryValidation
	}

	var notFoundErr *NotFoundError
	if As(err, &notFoundErr) {
		return CategoryNotFound
	}

	var opErr *OperationError
	if As(err, &opErr) {
		return CategoryOperation
	}

	// Check sentinel errors
	switch {
	case Is(err, ErrNotFound):
		return CategoryNotFound
	case Is(err, ErrValidation):
		return CategoryValidation
	case Is(err, ErrOperation):
		return CategoryOperation
	case Is(err, ErrPermission):
		return CategoryPermission
	case Is(err, ErrConflict):
		return CategoryConflict
	case Is(err, ErrInternal):
		return CategoryInternal
	default:
		return CategoryUnknown
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return GetCategory(err) == CategoryNotFound
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	return GetCategory(err) == CategoryValidation
}

// IsOperation checks if an error is an operation error
func IsOperation(err error) bool {
	return GetCategory(err) == CategoryOperation
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	return GetCategory(err) == CategoryPermission
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	return GetCategory(err) == CategoryConflict
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	return GetCategory(err) == CategoryInternal
}
