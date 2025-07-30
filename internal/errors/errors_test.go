package errors_test

import (
	"errors"
	"fmt"
	"testing"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

func TestWrap(t *testing.T) {
	t.Helper()

	tests := []struct {
		name    string
		err     error
		message string
		want    string
	}{
		{
			name:    "wrap simple error",
			err:     errors.New("original error"),
			message: "operation failed",
			want:    "operation failed: original error",
		},
		{
			name:    "wrap nil error returns nil",
			err:     nil,
			message: "should not appear",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interrors.Wrap(tt.err, tt.message)

			if tt.err == nil {
				if got != nil {
					t.Errorf("Wrap() with nil error = %v, want nil", got)
				}
				return
			}

			if got.Error() != tt.want {
				t.Errorf("Wrap() = %v, want %v", got.Error(), tt.want)
			}

			// Check that the original error is preserved
			if !errors.Is(got, tt.err) {
				t.Errorf("Wrapped error should match original with errors.Is")
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		err    error
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "wrap with formatted message",
			err:    errors.New("disk full"),
			format: "failed to save file %s",
			args:   []interface{}{"test.txt"},
			want:   "failed to save file test.txt: disk full",
		},
		{
			name:   "wrap nil error returns nil",
			err:    nil,
			format: "this should not appear",
			args:   []interface{}{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interrors.Wrapf(tt.err, tt.format, tt.args...)

			if tt.err == nil {
				if got != nil {
					t.Errorf("Wrapf() with nil error = %v, want nil", got)
				}
				return
			}

			if got.Error() != tt.want {
				t.Errorf("Wrapf() = %v, want %v", got.Error(), tt.want)
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		err  error
		want interrors.ErrorCategory
	}{
		{
			name: "nil error",
			err:  nil,
			want: interrors.CategoryUnknown,
		},
		{
			name: "sentinel not found error",
			err:  interrors.ErrNotFound,
			want: interrors.CategoryNotFound,
		},
		{
			name: "wrapped not found error",
			err:  fmt.Errorf("todo lookup failed: %w", interrors.ErrNotFound),
			want: interrors.CategoryNotFound,
		},
		{
			name: "validation error",
			err:  interrors.ErrValidation,
			want: interrors.CategoryValidation,
		},
		{
			name: "custom todo error",
			err:  interrors.NewTodoError("123", "update", "failed", interrors.CategoryOperation, nil),
			want: interrors.CategoryOperation,
		},
		{
			name: "not found error type",
			err:  interrors.NewNotFoundError("todo", "123"),
			want: interrors.CategoryNotFound,
		},
		{
			name: "validation error type",
			err:  interrors.NewValidationError("priority", "invalid", "must be high, medium, or low"),
			want: interrors.CategoryValidation,
		},
		{
			name: "unknown error",
			err:  interrors.New("some random error"),
			want: interrors.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interrors.GetCategory(tt.err); got != tt.want {
				t.Errorf("GetCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryHelpers(t *testing.T) {
	t.Helper()

	tests := []struct {
		name         string
		err          error
		isNotFound   bool
		isValidation bool
		isOperation  bool
		isPermission bool
		isConflict   bool
		isInternal   bool
	}{
		{
			name:       "not found error",
			err:        interrors.ErrNotFound,
			isNotFound: true,
		},
		{
			name:         "validation error",
			err:          interrors.ErrValidation,
			isValidation: true,
		},
		{
			name:        "operation error",
			err:         interrors.ErrOperation,
			isOperation: true,
		},
		{
			name:         "permission error",
			err:          interrors.ErrPermission,
			isPermission: true,
		},
		{
			name:       "conflict error",
			err:        interrors.ErrConflict,
			isConflict: true,
		},
		{
			name:       "internal error",
			err:        interrors.ErrInternal,
			isInternal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interrors.IsNotFound(tt.err); got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
			if got := interrors.IsValidation(tt.err); got != tt.isValidation {
				t.Errorf("IsValidation() = %v, want %v", got, tt.isValidation)
			}
			if got := interrors.IsOperation(tt.err); got != tt.isOperation {
				t.Errorf("IsOperation() = %v, want %v", got, tt.isOperation)
			}
			if got := interrors.IsPermission(tt.err); got != tt.isPermission {
				t.Errorf("IsPermission() = %v, want %v", got, tt.isPermission)
			}
			if got := interrors.IsConflict(tt.err); got != tt.isConflict {
				t.Errorf("IsConflict() = %v, want %v", got, tt.isConflict)
			}
			if got := interrors.IsInternal(tt.err); got != tt.isInternal {
				t.Errorf("IsInternal() = %v, want %v", got, tt.isInternal)
			}
		})
	}
}

func TestErrorsIsAndAs(t *testing.T) {
	t.Helper()

	// Test errors.Is wrapper
	baseErr := interrors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	if !interrors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should work with wrapped errors")
	}

	// Test errors.As wrapper
	todoErr := interrors.NewTodoError("123", "update", "failed", interrors.CategoryOperation, baseErr)
	wrappedTodoErr := fmt.Errorf("operation failed: %w", todoErr)

	var extractedTodoErr *interrors.TodoError
	if !interrors.As(wrappedTodoErr, &extractedTodoErr) {
		t.Error("errors.As should extract TodoError from wrapped error")
	}

	if extractedTodoErr.ID != "123" {
		t.Errorf("Extracted TodoError has wrong ID: got %s, want 123", extractedTodoErr.ID)
	}
}
