package handlers

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
)

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

	// Validate format
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

	// Optional parameters
	if section, ok := args["section"].(string); ok {
		params.Section = section
	}

	params.Operation = "append"
	if op, ok := args["operation"].(string); ok {
		params.Operation = op
	}

	if content, ok := args["content"].(string); ok {
		params.Content = content
	}

	// Extract metadata if provided
	if metadataObj, ok := args["metadata"].(map[string]interface{}); ok {
		if status, ok := metadataObj["status"].(string); ok {
			params.Metadata.Status = status
		}
		if priority, ok := metadataObj["priority"].(string); ok {
			params.Metadata.Priority = priority
		}
		if currentTest, ok := metadataObj["current_test"].(string); ok {
			params.Metadata.CurrentTest = currentTest
		}
	}

	// Validate operation
	if !isValidOperation(params.Operation) {
		return nil, fmt.Errorf("invalid operation '%s', must be one of: append, replace, prepend, toggle", params.Operation)
	}

	// Validate enum values in metadata
	if params.Metadata.Priority != "" && !isValidPriority(params.Metadata.Priority) {
		return nil, fmt.Errorf("invalid priority '%s' in metadata", params.Metadata.Priority)
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
	if !ok {
		return nil, fmt.Errorf("missing required parameter 'query'")
	}
	params.Query = query

	// Optional scope (array of strings)
	if scopeInterface, ok := args["scope"].([]interface{}); ok {
		for _, s := range scopeInterface {
			if str, ok := s.(string); ok {
				params.Scope = append(params.Scope, str)
			}
		}
	}

	// Extract filters if provided
	if filtersObj, ok := args["filters"].(map[string]interface{}); ok {
		if status, ok := filtersObj["status"].(string); ok {
			params.Filters.Status = status
		}
		if dateFrom, ok := filtersObj["date_from"].(string); ok {
			params.Filters.DateFrom = dateFrom
		}
		if dateTo, ok := filtersObj["date_to"].(string); ok {
			params.Filters.DateTo = dateTo
		}
	}

	// Limit with default
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

	// Optional quarter override
	if quarter, ok := args["quarter"].(string); ok {
		params.Quarter = quarter
	}

	return params, nil
}

// ExtractTodoCreateMultiParams extracts and validates todo_create_multi parameters
func ExtractTodoCreateMultiParams(request mcp.CallToolRequest) (*TodoCreateMultiParams, error) {
	params := &TodoCreateMultiParams{}

	args := request.GetArguments()

	// Extract parent
	parentObj, ok := args["parent"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing required parameter 'parent'")
	}

	// Parse parent task
	if task, ok := parentObj["task"].(string); ok && task != "" {
		params.Parent.Task = task
	} else {
		return nil, fmt.Errorf("parent must have a task")
	}

	// Parse parent priority with default
	params.Parent.Priority = "high"
	if priority, ok := parentObj["priority"].(string); ok {
		params.Parent.Priority = priority
	}

	// Parse parent type with default
	params.Parent.Type = "multi-phase"
	if todoType, ok := parentObj["type"].(string); ok {
		params.Parent.Type = todoType
	}

	// Extract children
	childrenInterface, ok := args["children"].([]interface{})
	if !ok || len(childrenInterface) == 0 {
		return nil, fmt.Errorf("missing or empty 'children' parameter")
	}

	// Parse each child
	for i, childInterface := range childrenInterface {
		childObj, ok := childInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid child at index %d", i)
		}

		child := TodoCreateInfo{}

		// Parse child task
		if task, ok := childObj["task"].(string); ok && task != "" {
			child.Task = task
		} else {
			return nil, fmt.Errorf("child at index %d must have a task", i)
		}

		// Parse child priority with default
		child.Priority = "medium"
		if priority, ok := childObj["priority"].(string); ok {
			child.Priority = priority
		}

		// Parse child type with default
		child.Type = "phase"
		if todoType, ok := childObj["type"].(string); ok {
			child.Type = todoType
		}

		params.Children = append(params.Children, child)
	}

	// Validate parent priority
	if !isValidPriority(params.Parent.Priority) {
		return nil, fmt.Errorf("invalid parent priority '%s'", params.Parent.Priority)
	}

	// Validate parent type
	if !isValidTodoType(params.Parent.Type) {
		return nil, fmt.Errorf("invalid parent type '%s'", params.Parent.Type)
	}

	// Validate each child
	for i, child := range params.Children {
		if !isValidPriority(child.Priority) {
			return nil, fmt.Errorf("invalid priority '%s' for child at index %d", child.Priority, i)
		}
		if !isValidTodoType(child.Type) {
			return nil, fmt.Errorf("invalid type '%s' for child at index %d", child.Type, i)
		}
	}

	return params, nil
}