package handlers

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"github.com/mark3labs/mcp-go/mcp"
)

// Common error types
var (
	ErrTodoNotFound   = errors.New("todo not found")
	ErrInvalidID      = errors.New("invalid todo ID")
	ErrSearchFailed   = errors.New("search operation failed")
	ErrArchiveFailed  = errors.New("archive operation failed")
	ErrTemplateFailed = errors.New("template operation failed")
)

// HandleError converts errors to appropriate MCP tool results
func HandleError(err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}
	
	// Handle common error patterns
	errStr := err.Error()
	switch {
	case os.IsNotExist(err):
		return mcp.NewToolResultError("Todo not found")
		
	case strings.Contains(errStr, "not found"):
		return mcp.NewToolResultError("Todo not found")
		
	case strings.Contains(errStr, "invalid"):
		return mcp.NewToolResultError("Invalid parameter or ID format")
		
	case strings.Contains(errStr, "search"):
		return mcp.NewToolResultError("Search operation failed")
		
	case strings.Contains(errStr, "archive"):
		return mcp.NewToolResultError("Archive operation failed")
		
	default:
		// Generic error with details
		return mcp.NewToolResultError(fmt.Sprintf("Operation failed: %v", err))
	}
}

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
	return fmt.Errorf("invalid %s '%s', must be one of: %v", paramName, value, allowed)
}