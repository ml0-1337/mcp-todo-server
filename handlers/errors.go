package handlers

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"os"
	"strings"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// Common error types - deprecated, use internal/errors package instead
var (
	ErrTodoNotFound   = interrors.ErrNotFound
	ErrInvalidID      = interrors.ErrValidation
	ErrSearchFailed   = interrors.ErrOperation
	ErrArchiveFailed  = interrors.ErrOperation
	ErrTemplateFailed = interrors.ErrOperation
)

// HandleError converts errors to appropriate MCP tool results
func HandleError(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	// Check for specific error types using our structured errors
	switch {
	case interrors.IsNotFound(err):
		return mcp.NewToolResultError("Todo not found")

	case interrors.IsValidation(err):
		// Use the full error message which includes field name
		return mcp.NewToolResultError(err.Error())

	case interrors.IsOperation(err):
		// Extract specific operation message if available
		var opErr *interrors.OperationError
		if interrors.As(err, &opErr) {
			return mcp.NewToolResultError(fmt.Sprintf("%s operation failed: %s", opErr.Operation, opErr.Message))
		}
		return mcp.NewToolResultError("Operation failed")

	case interrors.IsPermission(err):
		return mcp.NewToolResultError("Permission denied")

	case interrors.IsConflict(err):
		return mcp.NewToolResultError("Resource conflict")

	case interrors.IsInternal(err):
		return mcp.NewToolResultError("Internal server error")

	// Legacy error handling for backward compatibility
	case os.IsNotExist(err):
		return mcp.NewToolResultError("Todo not found")

	case strings.Contains(err.Error(), "validation error:"):
		// Preserve validation errors with their specific messages
		return mcp.NewToolResultError(err.Error())

	case strings.Contains(strings.ToLower(err.Error()), "not found"):
		return mcp.NewToolResultError("Todo not found")

	case strings.Contains(strings.ToLower(err.Error()), "archive"):
		return mcp.NewToolResultError("Archive operation failed")

	case strings.Contains(strings.ToLower(err.Error()), "base manager not available"):
		return mcp.NewToolResultError("Linking feature not available")

	default:
		// Generic error with details
		return mcp.NewToolResultError(fmt.Sprintf("Operation failed: %v", err))
	}
}

// ValidateRequiredParam checks if a required parameter is present
func ValidateRequiredParam(param, name string) error {
	if param == "" {
		// Return a simple error to match the expected format
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
	// Return a simple error to match the expected format
	return fmt.Errorf("invalid %s '%s': must be one of: %v", paramName, value, allowed)
}
