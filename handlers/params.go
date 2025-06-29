package handlers

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
)

// TodoCreateParams represents parameters for todo_create
type TodoCreateParams struct {
	Task     string
	Priority string
	Type     string
	Template string
	ParentID string
}

// TodoReadParams represents parameters for todo_read
type TodoReadParams struct {
	ID     string
	Filter TodoFilter
	Format string
}

// TodoFilter represents filter options for todo_read
type TodoFilter struct {
	Status   string
	Priority string
	Days     int
}

// TodoUpdateParams represents parameters for todo_update
type TodoUpdateParams struct {
	ID        string
	Section   string
	Operation string
	Content   string
	Metadata  TodoMetadata
}

// TodoMetadata represents metadata updates
type TodoMetadata struct {
	Status      string
	Priority    string
	CurrentTest string
}

// TodoSearchParams represents parameters for todo_search
type TodoSearchParams struct {
	Query   string
	Scope   []string
	Filters SearchFilters
	Limit   int
}

// SearchFilters represents search filter options
type SearchFilters struct {
	Status   string
	DateFrom string
	DateTo   string
}

// TodoArchiveParams represents parameters for todo_archive
type TodoArchiveParams struct {
	ID      string
	Quarter string
}

// ExtractTodoCreateParams extracts and validates todo_create parameters
func ExtractTodoCreateParams(request mcp.CallToolRequest) (*TodoCreateParams, error) {
	params := &TodoCreateParams{}
	
	// Get arguments
	args := request.GetArguments()
	
	// Required parameter
	task, ok := args["task"].(string)
	if !ok || task == "" {
		return nil, fmt.Errorf("missing required parameter 'task'")
	}
	params.Task = task
	
	// Optional parameters with defaults
	params.Priority = "high"
	if priority, ok := args["priority"].(string); ok {
		params.Priority = priority
	}
	
	params.Type = "feature"
	if todoType, ok := args["type"].(string); ok {
		params.Type = todoType
	}
	
	if template, ok := args["template"].(string); ok {
		params.Template = template
	}
	
	if parentID, ok := args["parent_id"].(string); ok {
		params.ParentID = parentID
	}
	
	// Validate enums
	if !isValidPriority(params.Priority) {
		return nil, fmt.Errorf("invalid priority '%s', must be one of: high, medium, low", params.Priority)
	}
	
	if !isValidTodoType(params.Type) {
		return nil, fmt.Errorf("invalid type '%s', must be one of: feature, bug, refactor, research, multi-phase, phase, subtask", params.Type)
	}
	
	// Validate that phase and subtask types require parent_id
	if (params.Type == "phase" || params.Type == "subtask") && params.ParentID == "" {
		return nil, fmt.Errorf("type '%s' requires parent_id to be specified", params.Type)
	}
	
	return params, nil
}

// ExtractTodoReadParams extracts and validates todo_read parameters
func ExtractTodoReadParams(request mcp.CallToolRequest) (*TodoReadParams, error) {
	params := &TodoReadParams{}
	
	// Get arguments map
	args := request.GetArguments()
	
	// Optional ID for single todo
	if id, ok := args["id"].(string); ok {
		params.ID = id
	}
	
	// Extract filter if provided
	if filterObj, ok := args["filter"].(map[string]interface{}); ok {
		if status, ok := filterObj["status"].(string); ok {
			params.Filter.Status = status
		}
		if priority, ok := filterObj["priority"].(string); ok {
			params.Filter.Priority = priority
		}
		if days, ok := filterObj["days"].(float64); ok {
			params.Filter.Days = int(days)
		}
	}
	
	// Format with default
	params.Format = "summary"
	if format, ok := args["format"].(string); ok {
		params.Format = format
	}
	if !isValidFormat(params.Format) {
		return nil, fmt.Errorf("invalid format '%s', must be one of: full, summary, list", params.Format)
	}
	
	return params, nil
}

// ExtractTodoUpdateParams extracts and validates todo_update parameters
func ExtractTodoUpdateParams(request mcp.CallToolRequest) (*TodoUpdateParams, error) {
	params := &TodoUpdateParams{}
	
	// Get arguments
	args := request.GetArguments()
	
	// Required ID
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("missing required parameter 'id'")
	}
	params.ID = id
	
	// Optional section update
	if section, ok := args["section"].(string); ok {
		params.Section = section
	}
	
	params.Operation = "append"
	if operation, ok := args["operation"].(string); ok {
		params.Operation = operation
	}
	
	if content, ok := args["content"].(string); ok {
		params.Content = content
	}
	
	// Extract metadata if provided
	if metaObj, ok := args["metadata"].(map[string]interface{}); ok {
		if status, ok := metaObj["status"].(string); ok {
			params.Metadata.Status = status
		}
		if priority, ok := metaObj["priority"].(string); ok {
			params.Metadata.Priority = priority
		}
		if currentTest, ok := metaObj["current_test"].(string); ok {
			params.Metadata.CurrentTest = currentTest
		}
	}
	
	// Validate operation
	if !isValidOperation(params.Operation) {
		return nil, fmt.Errorf("invalid operation '%s'", params.Operation)
	}
	
	return params, nil
}

// ExtractTodoSearchParams extracts and validates todo_search parameters
func ExtractTodoSearchParams(request mcp.CallToolRequest) (*TodoSearchParams, error) {
	params := &TodoSearchParams{}
	
	// Get arguments
	args := request.GetArguments()
	
	// Required query
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("missing required parameter 'query'")
	}
	params.Query = query
	
	// Optional scope with default
	params.Scope = []string{"all"}
	if scopeArr, ok := args["scope"].([]interface{}); ok {
		params.Scope = []string{}
		for _, s := range scopeArr {
			if str, ok := s.(string); ok {
				params.Scope = append(params.Scope, str)
			}
		}
		if len(params.Scope) == 0 {
			params.Scope = []string{"all"}
		}
	}
	
	// Extract filters if provided
	if filterObj, ok := args["filters"].(map[string]interface{}); ok {
		if status, ok := filterObj["status"].(string); ok {
			params.Filters.Status = status
		}
		if dateFrom, ok := filterObj["date_from"].(string); ok {
			params.Filters.DateFrom = dateFrom
		}
		if dateTo, ok := filterObj["date_to"].(string); ok {
			params.Filters.DateTo = dateTo
		}
	}
	
	// Limit with default and max
	params.Limit = 20
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
		if params.Limit > 100 {
			params.Limit = 100
		}
	}
	
	return params, nil
}

// ExtractTodoArchiveParams extracts and validates todo_archive parameters
func ExtractTodoArchiveParams(request mcp.CallToolRequest) (*TodoArchiveParams, error) {
	params := &TodoArchiveParams{}
	
	// Get arguments
	args := request.GetArguments()
	
	// Required ID
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("missing required parameter 'id'")
	}
	params.ID = id
	
	// Optional quarter
	if quarter, ok := args["quarter"].(string); ok {
		params.Quarter = quarter
	}
	
	return params, nil
}

// Validation helpers
func isValidPriority(p string) bool {
	return p == "high" || p == "medium" || p == "low"
}

func isValidTodoType(t string) bool {
	return t == "feature" || t == "bug" || t == "refactor" || t == "research" || t == "multi-phase" || t == "phase" || t == "subtask"
}

func isValidFormat(f string) bool {
	return f == "full" || f == "summary" || f == "list"
}

func isValidOperation(o string) bool {
	return o == "append" || o == "replace" || o == "prepend" || o == "toggle"
}