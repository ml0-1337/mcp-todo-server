package errors_test

import (
	"fmt"
	"testing"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

func TestTodoError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name      string
		todoError *interrors.TodoError
		want      string
	}{
		{
			name: "full todo error",
			todoError: interrors.NewTodoError(
				"test-123",
				"update",
				"validation failed",
				interrors.CategoryValidation,
				fmt.Errorf("invalid status"),
			),
			want: "operation update: todo test-123: validation failed: invalid status",
		},
		{
			name: "todo error without cause",
			todoError: interrors.NewTodoError(
				"test-456",
				"create",
				"duplicate ID",
				interrors.CategoryConflict,
				nil,
			),
			want: "operation create: todo test-456: duplicate ID",
		},
		{
			name: "todo error without operation",
			todoError: &interrors.TodoError{
				ID:       "test-789",
				Message:  "not found",
				Category: interrors.CategoryNotFound,
			},
			want: "todo test-789: not found",
		},
		{
			name: "todo error with only message",
			todoError: &interrors.TodoError{
				Message: "general error",
			},
			want: "general error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.todoError.Error(); got != tt.want {
				t.Errorf("TodoError.Error() = %v, want %v", got, tt.want)
			}

			// Test Unwrap
			if tt.todoError.Cause != nil {
				if unwrapped := tt.todoError.Unwrap(); unwrapped != tt.todoError.Cause {
					t.Errorf("TodoError.Unwrap() = %v, want %v", unwrapped, tt.todoError.Cause)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		err  *interrors.ValidationError
		want string
	}{
		{
			name: "validation error with field",
			err:  interrors.NewValidationError("priority", "invalid", "must be high, medium, or low"),
			want: "validation error for field 'priority' with value 'invalid': must be high, medium, or low",
		},
		{
			name: "validation error without field",
			err: &interrors.ValidationError{
				Message: "invalid input format",
				Cause:   interrors.ErrValidation,
			},
			want: "validation error: invalid input format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}

			// Check that it unwraps to ErrValidation
			if unwrapped := tt.err.Unwrap(); !interrors.Is(unwrapped, interrors.ErrValidation) {
				t.Errorf("ValidationError should unwrap to ErrValidation")
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		err  *interrors.NotFoundError
		want string
	}{
		{
			name: "not found with type and ID",
			err:  interrors.NewNotFoundError("todo", "test-123"),
			want: "todo 'test-123' not found",
		},
		{
			name: "not found with custom message",
			err: &interrors.NotFoundError{
				Message: "template not found in registry",
			},
			want: "template not found in registry",
		},
		{
			name: "not found with only resource type",
			err: &interrors.NotFoundError{
				ResourceType: "archive",
			},
			want: "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("NotFoundError.Error() = %v, want %v", got, tt.want)
			}

			// Check that it unwraps to ErrNotFound
			if unwrapped := tt.err.Unwrap(); !interrors.Is(unwrapped, interrors.ErrNotFound) {
				t.Errorf("NotFoundError should unwrap to ErrNotFound")
			}
		})
	}
}

func TestOperationError(t *testing.T) {
	t.Helper()

	baseErr := fmt.Errorf("disk full")

	tests := []struct {
		name string
		err  *interrors.OperationError
		want string
	}{
		{
			name: "operation error with all fields",
			err: interrors.NewOperationError(
				"save",
				"todo file",
				"insufficient disk space",
				baseErr,
			),
			want: "operation 'save' failed: on todo file: insufficient disk space: disk full",
		},
		{
			name: "operation error without cause",
			err: interrors.NewOperationError(
				"delete",
				"archive",
				"permission denied",
				nil,
			),
			want: "operation 'delete' failed: on archive: permission denied",
		},
		{
			name: "operation error with only operation",
			err: &interrors.OperationError{
				Operation: "index",
			},
			want: "operation 'index' failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("OperationError.Error() = %v, want %v", got, tt.want)
			}

			// Check unwrap behavior
			unwrapped := tt.err.Unwrap()
			if tt.err.Cause != nil {
				if unwrapped != tt.err.Cause {
					t.Errorf("OperationError.Unwrap() = %v, want %v", unwrapped, tt.err.Cause)
				}
			} else {
				if !interrors.Is(unwrapped, interrors.ErrOperation) {
					t.Errorf("OperationError without cause should unwrap to ErrOperation")
				}
			}
		})
	}
}

func TestMultiError(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		errors []error
		want   string
	}{
		{
			name:   "no errors",
			errors: []error{},
			want:   "no errors",
		},
		{
			name:   "single error",
			errors: []error{fmt.Errorf("single error")},
			want:   "single error",
		},
		{
			name: "multiple errors",
			errors: []error{
				fmt.Errorf("error 1"),
				fmt.Errorf("error 2"),
				fmt.Errorf("error 3"),
			},
			want: "multiple errors: [error 1; error 2; error 3]",
		},
		{
			name: "errors with nil",
			errors: []error{
				fmt.Errorf("error 1"),
				nil,
				fmt.Errorf("error 2"),
			},
			want: "multiple errors: [error 1; error 2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiErr := interrors.NewMultiError()
			for _, err := range tt.errors {
				multiErr.Add(err)
			}

			if got := multiErr.Error(); got != tt.want {
				t.Errorf("MultiError.Error() = %v, want %v", got, tt.want)
			}

			// Test HasErrors
			hasErrors := len(tt.errors) > 0
			for _, err := range tt.errors {
				if err == nil {
					hasErrors = len(tt.errors) > 1 // Only if there are non-nil errors
				}
			}

			if multiErr.HasErrors() != hasErrors {
				t.Errorf("MultiError.HasErrors() = %v, want %v", multiErr.HasErrors(), hasErrors)
			}

			// Test Unwrap
			if len(tt.errors) > 0 && tt.errors[0] != nil {
				if multiErr.Unwrap() != tt.errors[0] {
					t.Errorf("MultiError.Unwrap() should return first error")
				}
			}
		})
	}
}
