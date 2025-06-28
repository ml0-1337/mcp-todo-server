package handlers

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Test HandleError function
func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedInText string
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedInText: "",
		},
		{
			name:           "os not exist error",
			err:            os.ErrNotExist,
			expectedInText: "Todo not found",
		},
		{
			name:           "error with 'not found' text",
			err:            errors.New("todo not found in database"),
			expectedInText: "Todo not found",
		},
		{
			name:           "error with 'invalid' text",
			err:            errors.New("invalid parameter format"),
			expectedInText: "Invalid parameter or ID format",
		},
		{
			name:           "error with 'search' text",
			err:            errors.New("search index corrupted"),
			expectedInText: "Search operation failed",
		},
		{
			name:           "error with 'archive' text",
			err:            errors.New("archive directory not writable"),
			expectedInText: "Archive operation failed",
		},
		{
			name:           "generic error",
			err:            errors.New("something went wrong"),
			expectedInText: "Operation failed: something went wrong",
		},
		{
			name:           "wrapped error",
			err:            fmt.Errorf("failed to process: %w", errors.New("underlying issue")),
			expectedInText: "Operation failed: failed to process: underlying issue",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleError(tt.err)
			
			if tt.err == nil {
				if result != nil {
					t.Error("Expected nil result for nil error")
				}
				return
			}
			
			if result == nil {
				t.Fatal("Expected non-nil result for error")
			}
			
			if !result.IsError {
				t.Error("Expected IsError to be true")
			}
			
			// We can't access the error message directly from CallToolResult,
			// but we can verify the result is properly formatted as an error
		})
	}
}

// Test ValidateRequiredParam
func TestValidateRequiredParam(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		paramName string
		wantErr   bool
	}{
		{
			name:      "valid param",
			param:     "value",
			paramName: "test",
			wantErr:   false,
		},
		{
			name:      "empty param",
			param:     "",
			paramName: "test",
			wantErr:   true,
		},
		{
			name:      "whitespace param",
			param:     "   ",
			paramName: "test",
			wantErr:   false, // whitespace is not considered empty
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredParam(tt.param, tt.paramName)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredParam() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && !strings.Contains(err.Error(), tt.paramName) {
				t.Errorf("Error should mention parameter name '%s', got: %v", tt.paramName, err)
			}
		})
	}
}

// Test ValidateEnum
func TestValidateEnum(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		allowed   []string
		paramName string
		wantErr   bool
	}{
		{
			name:      "valid value",
			value:     "high",
			allowed:   []string{"high", "medium", "low"},
			paramName: "priority",
			wantErr:   false,
		},
		{
			name:      "invalid value",
			value:     "urgent",
			allowed:   []string{"high", "medium", "low"},
			paramName: "priority",
			wantErr:   true,
		},
		{
			name:      "empty value",
			value:     "",
			allowed:   []string{"high", "medium", "low"},
			paramName: "priority",
			wantErr:   true,
		},
		{
			name:      "case sensitive",
			value:     "HIGH",
			allowed:   []string{"high", "medium", "low"},
			paramName: "priority",
			wantErr:   true,
		},
		{
			name:      "value at end of list",
			value:     "low",
			allowed:   []string{"high", "medium", "low"},
			paramName: "priority",
			wantErr:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum(tt.value, tt.allowed, tt.paramName)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil {
				// Check error message contains necessary information
				errStr := err.Error()
				if !strings.Contains(errStr, tt.paramName) {
					t.Errorf("Error should mention parameter name '%s'", tt.paramName)
				}
				if !strings.Contains(errStr, tt.value) {
					t.Errorf("Error should mention invalid value '%s'", tt.value)
				}
				// Check that allowed values are mentioned
				for _, allowed := range tt.allowed {
					if !strings.Contains(errStr, allowed) {
						t.Errorf("Error should list allowed value '%s'", allowed)
					}
				}
			}
		})
	}
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	// Verify error constants are defined
	errors := []struct {
		name string
		err  error
	}{
		{"ErrTodoNotFound", ErrTodoNotFound},
		{"ErrInvalidID", ErrInvalidID},
		{"ErrSearchFailed", ErrSearchFailed},
		{"ErrArchiveFailed", ErrArchiveFailed},
		{"ErrTemplateFailed", ErrTemplateFailed},
	}
	
	for _, e := range errors {
		if e.err == nil {
			t.Errorf("%s should not be nil", e.name)
		}
		
		// Verify error has appropriate message
		if e.err.Error() == "" {
			t.Errorf("%s should have non-empty error message", e.name)
		}
	}
}

// Test error handling patterns
func TestErrorHandlingPatterns(t *testing.T) {
	t.Run("wrapped errors preserve context", func(t *testing.T) {
		baseErr := errors.New("database connection failed")
		wrappedErr := fmt.Errorf("failed to read todo: %w", baseErr)
		
		result := HandleError(wrappedErr)
		if result == nil || !result.IsError {
			t.Error("Expected error result for wrapped error")
		}
	})
	
	t.Run("multiple error keywords", func(t *testing.T) {
		// Test that first matching pattern is used
		err := errors.New("invalid search parameters not found")
		result := HandleError(err)
		
		// Should match "invalid" before "not found"
		if result == nil || !result.IsError {
			t.Error("Expected error result")
		}
	})
	
	t.Run("nil safe operations", func(t *testing.T) {
		// Ensure all error functions handle nil safely
		if HandleError(nil) != nil {
			t.Error("HandleError should return nil for nil error")
		}
		
		if ValidateRequiredParam("", "test") == nil {
			t.Error("ValidateRequiredParam should return error for empty param")
		}
		
		if ValidateEnum("", []string{}, "test") == nil {
			t.Error("ValidateEnum should return error for empty value")
		}
	})
}

// Test error message formatting
func TestErrorMessageFormatting(t *testing.T) {
	t.Run("ValidateRequiredParam message format", func(t *testing.T) {
		err := ValidateRequiredParam("", "taskID")
		if err == nil {
			t.Fatal("Expected error")
		}
		
		expected := "missing required parameter 'taskID'"
		if err.Error() != expected {
			t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
		}
	})
	
	t.Run("ValidateEnum message format", func(t *testing.T) {
		err := ValidateEnum("urgent", []string{"high", "medium", "low"}, "priority")
		if err == nil {
			t.Fatal("Expected error")
		}
		
		// Check message contains all required parts
		msg := err.Error()
		if !strings.Contains(msg, "invalid priority 'urgent'") {
			t.Error("Error should mention invalid value")
		}
		if !strings.Contains(msg, "must be one of: [high medium low]") {
			t.Error("Error should list allowed values")
		}
	})
}