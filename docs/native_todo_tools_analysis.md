# Claude Native Todo Tools Analysis

## Overview

This document provides a comprehensive analysis of Claude's native TodoRead and TodoWrite tools, documenting their input/output specifications, behaviors, and limitations. This analysis was conducted to understand the baseline functionality that the MCP Todo Server will replace and enhance.

## Native Todo Tools

### 1. TodoRead Tool

**Purpose**: Read the current to-do list for the session

**Input Schema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "additionalProperties": true,
  "description": "No input is required, leave this field blank",
  "properties": {},
  "type": "object"
}
```

**Input Requirements**:
- No parameters required - input must be completely empty
- Do NOT include dummy objects, placeholder strings, or keys like "input" or "empty"
- Simply invoke with no arguments

**Output Format**:
```json
[
  {
    "content": "string - The task description",
    "status": "pending | in_progress | completed",
    "priority": "high | medium | low",
    "id": "string - Unique identifier"
  }
]
```

**Example Output**:
```json
[
  {
    "content": "Implement MCP Todo Server in Go",
    "status": "in_progress",
    "priority": "high",
    "id": "implement-mcp-server"
  }
]
```

### 2. TodoWrite Tool

**Purpose**: Create and manage a structured task list for the current session

**Input Schema**:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "additionalProperties": false,
  "properties": {
    "todos": {
      "description": "The updated todo list",
      "items": {
        "additionalProperties": false,
        "properties": {
          "content": {
            "minLength": 1,
            "type": "string"
          },
          "id": {
            "type": "string"
          },
          "priority": {
            "enum": ["high", "medium", "low"],
            "type": "string"
          },
          "status": {
            "enum": ["pending", "in_progress", "completed"],
            "type": "string"
          }
        },
        "required": ["content", "status", "priority", "id"],
        "type": "object"
      },
      "type": "array"
    }
  },
  "required": ["todos"],
  "type": "object"
}
```

**Key Behaviors**:
1. **Complete Replacement**: The todos array completely replaces the existing list
2. **No Partial Updates**: Cannot update individual todos - must send full array
3. **All Fields Required**: Every todo must have all four fields
4. **State Management**: Convention is to have only one "in_progress" todo at a time

**Example Input**:
```json
{
  "todos": [
    {
      "content": "Create todo_create handler",
      "status": "completed",
      "priority": "high",
      "id": "create-handler"
    },
    {
      "content": "Implement todo_read functionality",
      "status": "in_progress",
      "priority": "high",
      "id": "implement-read"
    }
  ]
}
```

## Limitations of Native Tools

### 1. Persistence
- **Session-only**: All todos are lost when the session ends
- **No file storage**: Exists only in memory
- **No backup**: No way to recover lost todos

### 2. Data Model
- **Minimal metadata**: Only 4 fields (content, status, priority, id)
- **No timestamps**: No creation or completion time tracking
- **No relationships**: Cannot link related todos
- **Simple content**: Plain string only, no structured sections

### 3. Operations
- **Full replacement only**: Must send entire array for any change
- **No search**: Cannot find todos by content
- **No filtering**: Cannot filter by status, priority, or date
- **No bulk operations**: Each update requires full array

### 4. Features
- **No templates**: Cannot use predefined structures
- **No analytics**: No completion rates or productivity metrics
- **No archival**: Completed todos remain in active list
- **No version history**: Changes are not tracked

## Comparison with MCP Todo Server

| Feature | Native Tools | MCP Todo Server |
|---------|--------------|-----------------|
| **Persistence** | Session-only | Markdown files |
| **Fields** | 4 basic fields | 15+ fields with YAML frontmatter |
| **Content** | Plain string | Structured markdown sections |
| **Updates** | Full replacement | Atomic section updates |
| **Search** | None | Full-text search |
| **Filtering** | None | By status, date, priority, tags |
| **Templates** | None | Predefined task structures |
| **Analytics** | None | Completion rates, trends |
| **Archive** | None | Quarterly folders |
| **Relationships** | None | Parent-child, dependencies |
| **Test Tracking** | None | RGRC cycle support |
| **Timestamps** | None | Created, updated, completed |

## MCP Todo Server Advantages

The MCP Todo Server provides significant enhancements:

1. **Persistent Storage**: Todos saved as `.claude/todos/*.md` files
2. **Rich Metadata**: YAML frontmatter with extensive fields
3. **Structured Content**: Multiple sections (findings, tests, checklist, etc.)
4. **Granular Updates**: Update specific sections without replacing entire file
5. **Advanced Search**: Query across all todo content and metadata
6. **Workflow Support**: Templates, RGRC tracking, multi-phase projects
7. **Analytics**: Track productivity patterns and completion metrics
8. **Integration**: Works with existing CLAUDE.md workflow

## Implementation Notes

When implementing MCP todo tools, they must:

1. **Match Input/Output Format**: For drop-in replacement compatibility
2. **Enhance Behind the Scenes**: Provide persistence while maintaining API
3. **Add New Tools**: Extend functionality with search, templates, etc.
4. **Maintain Simplicity**: Keep basic operations as simple as native tools

## Conclusion

The native todo tools provide minimal functionality suitable only for brief sessions. The MCP Todo Server addresses all limitations while maintaining compatibility, enabling Claude to manage complex, long-term projects with full persistence and advanced features.