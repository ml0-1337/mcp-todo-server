# MCP Todo Server API Reference

This document provides detailed API documentation for all MCP Todo Server tools, including input/output schemas, examples, and error handling.

## Table of Contents

1. [todo_create](#todo_create) - Create new todos
2. [todo_read](#todo_read) - Read todos with filtering
3. [todo_update](#todo_update) - Update todo sections
4. [todo_search](#todo_search) - Full-text search
5. [todo_archive](#todo_archive) - Archive completed todos
6. [todo_template](#todo_template) - Create from templates
7. [todo_link](#todo_link) - Link related todos
8. [todo_stats](#todo_stats) - Analytics and metrics
9. [todo_clean](#todo_clean) - Bulk operations

## Common Response Format

All tools return responses in JSON format with consistent error handling:

**Success Response:**
```json
{
  "result": {
    // Tool-specific response data
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| INVALID_PARAMS | Missing or invalid parameters |
| TODO_NOT_FOUND | Specified todo doesn't exist |
| FILE_ERROR | File system operation failed |
| INDEX_ERROR | Search index operation failed |
| TEMPLATE_ERROR | Template processing failed |
| VALIDATION_ERROR | Input validation failed |

---

## todo_create

Creates a new todo with metadata and optional template.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| task | string | Yes | - | Task description (used for ID generation) |
| priority | string | No | "high" | Priority: high, medium, low |
| type | string | No | "feature" | Type: feature, bug, refactor, research, multi-phase |
| template | string | No | - | Template name to use |
| parent_id | string | No | - | Parent todo for multi-phase projects |

### Output Schema

```json
{
  "id": "string",      // Generated todo ID
  "path": "string",    // Full file path
  "message": "string"  // Success message
}
```

### Examples

**Basic Todo Creation:**
```json
// Input
{
  "task": "Implement user authentication"
}

// Output
{
  "id": "implement-user-authentication",
  "path": "/Users/name/.claude/todos/implement-user-authentication.md",
  "message": "Todo created successfully"
}
```

**With Template:**
```json
// Input
{
  "task": "Fix login timeout bug",
  "priority": "high",
  "type": "bug",
  "template": "bug-fix"
}

// Output
{
  "id": "fix-login-timeout-bug",
  "path": "/Users/name/.claude/todos/fix-login-timeout-bug.md",
  "message": "Todo created from template: bug-fix"
}
```

**Multi-phase Project:**
```json
// Input
{
  "task": "Phase 1: Design API",
  "type": "multi-phase",
  "parent_id": "implement-rest-api"
}

// Output
{
  "id": "phase-1-design-api",
  "path": "/Users/name/.claude/todos/phase-1-design-api.md",
  "message": "Todo created with parent: implement-rest-api"
}
```

### ID Generation Rules

1. Convert to lowercase
2. Replace spaces with hyphens
3. Remove special characters
4. Limit to 50 characters
5. Append number if duplicate exists

Examples:
- "Fix Login Bug!" → "fix-login-bug"
- "Implement OAuth 2.0" → "implement-oauth-2-0"
- Duplicate task → "fix-login-bug-2"

### Error Cases

```json
// Missing required parameter
{
  "error": {
    "code": "INVALID_PARAMS",
    "message": "Required parameter 'task' is missing"
  }
}

// Invalid priority
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid priority 'urgent'. Valid values: high, medium, low"
  }
}

// Template not found
{
  "error": {
    "code": "TEMPLATE_ERROR",
    "message": "Template 'custom-template' not found"
  }
}
```

---

## todo_read

Reads single todo or lists multiple todos with filtering options.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| id | string | No | - | Specific todo ID to read |
| filter | object | No | - | Filter options |
| filter.status | string | No | "all" | Status: in_progress, completed, blocked, all |
| filter.priority | string | No | "all" | Priority: high, medium, low, all |
| filter.days | number | No | - | Todos from last N days |
| format | string | No | "summary" | Output format: full, summary, list |

### Output Schema

**Single Todo (format: full):**
```markdown
---
todo_id: implement-search
started: 2025-01-27 10:30:00
status: in_progress
priority: high
type: feature
---

# Task: Implement search functionality

[Full markdown content]
```

**Single Todo (format: summary):**
```json
{
  "id": "implement-search",
  "task": "Implement search functionality",
  "status": "in_progress",
  "priority": "high",
  "type": "feature",
  "started": "2025-01-27T10:30:00Z",
  "current_section": "Working on Test 3: Case-insensitive search"
}
```

**List Format (Native Compatible):**
```json
[
  {
    "id": "implement-search",
    "content": "Implement search functionality",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "id": "fix-auth-bug",
    "content": "Fix authentication timeout",
    "status": "completed",
    "priority": "high"
  }
]
```

### Examples

**Read Specific Todo:**
```json
// Input
{
  "id": "implement-search",
  "format": "full"
}

// Output: Complete markdown file content
```

**List In-Progress High Priority:**
```json
// Input
{
  "filter": {
    "status": "in_progress",
    "priority": "high"
  },
  "format": "list"
}

// Output: Array of matching todos
```

**Recent Todos (Last 7 Days):**
```json
// Input
{
  "filter": {
    "days": 7
  },
  "format": "summary"
}

// Output: Todos created in last week
```

### Error Cases

```json
// Todo not found
{
  "error": {
    "code": "TODO_NOT_FOUND",
    "message": "Todo 'invalid-id' not found"
  }
}

// Invalid format
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid format 'detailed'. Valid values: full, summary, list"
  }
}
```

---

## todo_update

Updates todo content or metadata with atomic operations.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| id | string | Yes | - | Todo ID to update |
| section | string | No | - | Section: findings, tests, checklist, scratchpad |
| operation | string | No | "append" | Operation: append, replace, prepend, toggle |
| content | string | No | - | Content to add/update |
| metadata | object | No | - | Metadata updates |
| metadata.status | string | No | - | New status |
| metadata.priority | string | No | - | New priority |
| metadata.current_test | string | No | - | Current test being worked on |

### Output Schema

**Simple Response (metadata updates or sections without content):**
```json
{
  "message": "Todo '{id}' updated successfully"
}
```

**Enriched Response (section updates with content):**
```json
{
  "message": "Todo '{id}' {section} section updated ({operation})",
  "todo": {
    "id": "string",
    "task": "string",
    "status": "string",
    "priority": "string",
    "type": "string",
    "checklist": [
      {
        "text": "string",
        "status": "pending|in_progress|completed"
      }
    ],
    "sections": {
      "{section_name}": {
        "title": "string",
        "hasContent": "boolean",
        "wordCount": "number"
      }
    }
  },
  "progress": {
    "checklist": "string (e.g., '2/4 completed (50%)')",
    "checklist_breakdown": {
      "completed": "number",
      "in_progress": "number",
      "pending": "number",
      "total": "number"
    },
    "sections": "string (e.g., '1/8 sections have content')"
  }
}
```

### Section Names

| Section | Description |
|---------|-------------|
| findings | Research and discoveries |
| tests | Test cases and results |
| checklist | Task items |
| scratchpad | Temporary notes |

### Examples

**Update Status:**
```json
// Input
{
  "id": "implement-search",
  "metadata": {
    "status": "completed"
  }
}

// Output
{
  "message": "Todo 'implement-search' updated successfully"
}
```

**Append Test Results:**
```json
// Input
{
  "id": "fix-login-bug",
  "section": "tests",
  "operation": "append",
  "content": "\n## Test 3: Timeout after 30 seconds\n\n[2025-01-27 14:00:00] Red Phase: Test failing as expected\n[2025-01-27 14:15:00] Green Phase: Test passing"
}

// Output
{
  "message": "Section 'tests' updated for todo 'fix-login-bug'"
}
```

**Replace Checklist:**
```json
// Input
{
  "id": "implement-api",
  "section": "checklist",
  "operation": "replace",
  "content": "- [x] Design API schema\n- [x] Implement endpoints\n- [ ] Add authentication\n- [ ] Write documentation"
}
```

**Track Current Test:**
```json
// Input
{
  "id": "add-validation",
  "metadata": {
    "current_test": "Test 5: Email format validation"
  }
}
```

**Toggle Checklist Item:**
```json
// Input
{
  "id": "implement-feature",
  "section": "checklist",
  "operation": "toggle",
  "content": "Add authentication"
}

// Output (Enriched Response)
{
  "message": "Todo 'implement-feature' checklist section updated (toggle)",
  "todo": {
    "id": "implement-feature",
    "task": "Implement user authentication",
    "status": "in_progress",
    "priority": "high",
    "type": "feature",
    "checklist": [
      {
        "text": "Design API schema",
        "status": "completed"
      },
      {
        "text": "Implement endpoints",
        "status": "completed"
      },
      {
        "text": "Add authentication",
        "status": "in_progress"
      },
      {
        "text": "Write documentation",
        "status": "pending"
      }
    ],
    "sections": {
      "checklist": {
        "title": "## Checklist",
        "hasContent": true,
        "wordCount": 20
      }
    }
  },
  "progress": {
    "checklist": "2/4 completed (50%)",
    "checklist_breakdown": {
      "completed": 2,
      "in_progress": 1,
      "pending": 1,
      "total": 4
    },
    "sections": "1/8 sections have content"
  }
}
```

### Enriched Response Format

When updating sections with content (especially checklists), the response includes full todo data:

- **todo**: Complete todo information including parsed checklist items
- **checklist**: Array of items with text and status (pending, in_progress, completed)
- **sections**: Summary of all sections with content status and word count
- **progress**: Checklist completion metrics with status breakdown

This eliminates the need for a separate read call after updates.

### Operation Behaviors

| Operation | Behavior |
|-----------|----------|
| append | Adds content to end of section |
| replace | Replaces entire section content |
| prepend | Adds content to beginning of section |
| toggle | Toggles checklist item status (checklist sections only) |

### Toggle Operation

The toggle operation cycles checklist items through three states:
- `[ ]` pending → `[>]` in_progress → `[x]` completed → `[ ]` pending

Supported in-progress markers: `[>]`, `[-]`, `[~]`

To toggle an item, use the item's text as the content parameter.

### Error Cases

```json
// Todo not found
{
  "error": {
    "code": "TODO_NOT_FOUND",
    "message": "Todo 'invalid-id' not found"
  }
}

// Invalid section
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid section 'notes'. Valid sections: findings, tests, checklist, scratchpad"
  }
}

// File write error
{
  "error": {
    "code": "FILE_ERROR",
    "message": "Failed to update todo: permission denied"
  }
}
```

---

## todo_search

Full-text search across all todos with advanced filtering.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| query | string | Yes | - | Search query (supports phrases) |
| scope | string[] | No | ["all"] | Fields: task, findings, tests, all |
| filters | object | No | - | Additional filters |
| filters.status | string | No | - | Filter by status |
| filters.date_from | string | No | - | Start date (YYYY-MM-DD) |
| filters.date_to | string | No | - | End date (YYYY-MM-DD) |
| limit | number | No | 20 | Maximum results (max: 100) |

### Query Syntax

- **Simple search**: `authentication`
- **Phrase search**: `"user authentication"`
- **Multiple terms**: `login timeout bug`

### Output Schema

```json
[
  {
    "id": "string",
    "task": "string",
    "score": 0.95,
    "snippet": "string with <mark>highlighted</mark> matches"
  }
]
```

### Examples

**Basic Search:**
```json
// Input
{
  "query": "authentication"
}

// Output
[
  {
    "id": "implement-oauth",
    "task": "Implement OAuth 2.0 authentication",
    "score": 0.95,
    "snippet": "...implement <mark>authentication</mark> using OAuth 2.0..."
  },
  {
    "id": "fix-auth-timeout",
    "task": "Fix authentication timeout issue",
    "score": 0.87,
    "snippet": "...when <mark>authentication</mark> expires after..."
  }
]
```

**Scoped Search:**
```json
// Input
{
  "query": "test failure",
  "scope": ["tests"],
  "limit": 10
}

// Output: Results only from test sections
```

**Filtered Search:**
```json
// Input
{
  "query": "bug",
  "filters": {
    "status": "in_progress",
    "date_from": "2025-01-20",
    "date_to": "2025-01-27"
  }
}

// Output: In-progress bugs from last week
```

### Field Boosting

Search relevance scoring:
- Task: 2.0x weight
- Findings: 1.5x weight
- Tests: 1.3x weight
- Other content: 1.0x weight

### Error Cases

```json
// Missing query
{
  "error": {
    "code": "INVALID_PARAMS",
    "message": "Required parameter 'query' is missing"
  }
}

// Index unavailable
{
  "error": {
    "code": "INDEX_ERROR",
    "message": "Search index temporarily unavailable"
  }
}
```

---

## todo_archive

Archives completed todos to daily archive folders.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| id | string | Yes | - | Todo ID to archive |
| quarter | string | No | Current | Override quarter (ignored for daily) |

### Archive Path Structure

```
archive/YYYY/MM/DD/todo-id.md
```

Example: `archive/2025/01/27/fix-login-bug.md`

### Examples

**Basic Archive:**
```json
// Input
{
  "id": "fix-login-bug"
}

// Output
{
  "id": "fix-login-bug",
  "archive_path": "/Users/name/.claude/archive/2025/01/27/fix-login-bug.md",
  "message": "Todo archived successfully"
}
```

### Validation Rules

1. Todo must exist
2. Status should be "completed"
3. No active child todos
4. Updates completed timestamp

### Error Cases

```json
// Has active children
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Cannot archive todo with 3 active child todos"
  }
}

// Not completed
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Todo status is 'in_progress'. Only completed todos can be archived"
  }
}
```

---

## todo_template

Creates todos from predefined templates with variable substitution.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| template | string | No | - | Template name or "list" |
| task | string | No | - | Task description |
| priority | string | No | Template | Override priority |
| type | string | No | Template | Override type |

### Template Variables

| Variable | Description |
|----------|-------------|
| {{.Task}} | Task description |
| {{.ID}} | Generated todo ID |
| {{.Started}} | Current timestamp |
| {{.Priority}} | Task priority |
| {{.Date}} | Current date |

### Examples

**List Templates:**
```json
// Input
{
  "template": "list"
}

// Output
{
  "templates": [
    "bug-fix",
    "feature",
    "research",
    "refactor"
  ]
}
```

**Create from Template:**
```json
// Input
{
  "template": "bug-fix",
  "task": "Fix memory leak in search engine",
  "priority": "high"
}

// Output
{
  "id": "fix-memory-leak-in-search-engine",
  "path": "/Users/name/.claude/todos/fix-memory-leak-in-search-engine.md",
  "message": "Todo created from template: bug-fix"
}
```

### Error Cases

```json
// Template not found
{
  "error": {
    "code": "TEMPLATE_ERROR",
    "message": "Template 'custom' not found. Use 'list' to see available templates"
  }
}
```

---

## todo_link

Creates relationships between todos for project management.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| parent_id | string | Yes | - | Parent todo ID |
| child_id | string | Yes | - | Child todo ID |
| link_type | string | No | "parent-child" | Link type (currently only parent-child) |

### Link Types

| Type | Description | Status |
|------|-------------|--------|
| parent-child | Hierarchical relationship | Implemented |
| blocks | Dependency relationship | Planned |
| relates-to | Related tasks | Planned |

### Examples

**Create Parent-Child Link:**
```json
// Input
{
  "parent_id": "implement-api",
  "child_id": "implement-auth-endpoint"
}

// Output
{
  "parent_id": "implement-api",
  "child_id": "implement-auth-endpoint",
  "link_type": "parent-child",
  "message": "Todos linked successfully"
}
```

### Validation Rules

1. Both todos must exist
2. No circular relationships
3. Parent cannot be archived with active children

### Error Cases

```json
// Todo not found
{
  "error": {
    "code": "TODO_NOT_FOUND",
    "message": "Parent todo 'invalid-id' not found"
  }
}

// Circular relationship
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Cannot create circular relationship"
  }
}
```

---

## todo_stats

Generates comprehensive statistics and analytics.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| period | string | No | "all" | Period: all, week, month, quarter |

### Output Schema

```json
{
  "total_todos": 150,
  "status_breakdown": {
    "completed": 120,
    "in_progress": 25,
    "pending": 3,
    "blocked": 2
  },
  "completion_rate": 80.0,
  "average_completion_time": "3d 14h 30m",
  "by_type": {
    "feature": {
      "count": 60,
      "completed": 50,
      "completion_rate": 83.3
    },
    "bug": {
      "count": 40,
      "completed": 38,
      "completion_rate": 95.0
    }
  },
  "by_priority": {
    "high": {
      "count": 80,
      "completed": 70,
      "completion_rate": 87.5,
      "avg_completion_time": "2d 10h"
    }
  },
  "test_coverage": {
    "todos_with_tests": 100,
    "coverage_percentage": 66.7,
    "average_test_completion": 75.5
  },
  "productivity": {
    "most_productive_day": "Tuesday",
    "average_daily_completions": 4.2,
    "streak": {
      "current": 5,
      "longest": 12
    }
  }
}
```

### Examples

**All-Time Stats:**
```json
// Input
{
  "period": "all"
}

// Output: Complete statistics for all todos
```

**Weekly Stats:**
```json
// Input
{
  "period": "week"
}

// Output: Statistics for last 7 days
```

### Metrics Explained

| Metric | Calculation |
|--------|-------------|
| completion_rate | (completed / total) × 100 |
| average_completion_time | Mean time from started to completed |
| test_coverage | Todos with test sections / total |
| test_completion | Checked items / total items in tests |

---

## todo_clean

Bulk operations for todo maintenance.

### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| operation | string | Yes | - | Operation: archive_old, find_duplicates |
| days | number | No | 30 | For archive_old: age threshold |

### Operations

**archive_old**: Archives completed todos older than N days
**find_duplicates**: Identifies potential duplicate todos

### Examples

**Archive Old Todos:**
```json
// Input
{
  "operation": "archive_old",
  "days": 60
}

// Output
{
  "archived_count": 15,
  "archived_todos": [
    "implement-logging",
    "fix-cors-issue",
    "update-dependencies"
  ],
  "message": "Archived 15 todos older than 60 days"
}
```

**Find Duplicates:**
```json
// Input
{
  "operation": "find_duplicates"
}

// Output
{
  "duplicates": [
    {
      "ids": ["fix-login-bug", "fix-login-issue"],
      "similarity": 0.92
    },
    {
      "ids": ["implement-search", "add-search-feature"],
      "similarity": 0.87
    }
  ],
  "message": "Found 2 potential duplicate sets"
}
```

### Error Cases

```json
// Invalid operation
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid operation 'cleanup'. Valid operations: archive_old, find_duplicates"
  }
}
```

---

## Best Practices

### Performance Tips

1. **Use specific IDs** when reading single todos
2. **Limit search results** to improve response time
3. **Use scoped searches** to search specific fields
4. **Archive regularly** to maintain performance

### Error Handling

1. **Always check for errors** in responses
2. **Validate inputs** before sending requests
3. **Handle TODO_NOT_FOUND** gracefully
4. **Retry on INDEX_ERROR** with backoff

### Workflow Integration

1. **Create → Update → Archive** lifecycle
2. **Use templates** for consistency
3. **Link related todos** for organization
4. **Monitor stats** for productivity insights

## Rate Limits

The MCP Todo Server has no built-in rate limits, but consider:

- File system performance with many concurrent operations
- Search index updates may lag slightly under heavy load
- Archive operations should be batched when possible

## Backward Compatibility

The server maintains compatibility with native todo tools:

- `todo_read` with `format: "list"` returns native format
- `todo_create` accepts minimal parameters like native tools
- Additional fields are added transparently