# Claude Native Todo Tools Analysis

## Overview

This document provides a comprehensive analysis of Claude's native TodoRead and TodoWrite tools, documenting their input/output specifications, behaviors, and limitations. This analysis was conducted to understand the baseline functionality that the MCP Todo Server will replace and enhance.

## Understanding additionalProperties

The `additionalProperties` field in JSON Schema controls whether properties not explicitly defined in the schema are allowed:

- **`true`**: Any additional properties are allowed (they'll be accepted but typically ignored)
- **`false`**: Only properties explicitly defined in `properties` are allowed
- **Schema object**: Additional properties must match the specified schema

This setting is crucial for API strictness and forward compatibility.

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

**additionalProperties Behavior**:

Since `additionalProperties: true`, the tool accepts but ignores extra properties:

```json
// Valid inputs (all produce the same output):

// Recommended - empty object
{}

// Valid - extra properties are ignored
{
  "foo": "bar",
  "unused": 123
}

// Valid - even complex structures are accepted
{
  "random": {
    "nested": {
      "data": true
    }
  }
}
```

All these inputs return the same todo list because TodoRead expects no parameters.

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

**additionalProperties Behavior**:

Since `additionalProperties: false` is specified, the behavior differs by level:
- **Root level**: Extra properties are rejected with validation errors
- **Item level**: Extra properties are silently dropped (not stored)

```json
// ✅ VALID - exact schema match
{
  "todos": [
    {
      "content": "Task 1",
      "status": "pending",
      "priority": "high",
      "id": "task-1"
    }
  ]
}

// ❌ INVALID - extra property at root level
{
  "todos": [...],
  "timestamp": "2024-01-27"  // This would fail validation
}

// ⚠️ ACCEPTED but properties dropped - extra property in todo item
{
  "todos": [
    {
      "content": "Task 1",
      "status": "pending",
      "priority": "high",
      "id": "task-1",
      "tags": ["tdd", "auth"]  // Extra properties are silently dropped
    }
  ]
}
// When read back, the todo will only have the 4 standard fields

// ❌ INVALID - missing required property
{
  "todos": [
    {
      "content": "Task 1",
      "status": "pending",
      "id": "task-1"
      // Missing "priority" - would fail validation
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

## Validation Test Results

Testing revealed the actual behavior of `additionalProperties` in the native tools:

### TodoRead Tests
All tests confirmed that extra properties are accepted without errors:
- Empty object `{}` ✅
- Simple properties `{"foo": "bar", "unused": 123}` ✅
- Nested structures `{"random": {"nested": {"data": true}}}` ✅
- Various types `{"array": [1,2,3], "bool": false, "null": null}` ✅

### TodoWrite Tests
1. **Valid schema match** ✅ - Works as expected
2. **Extra root property** ❌ - Error: "An unexpected parameter `timestamp` was provided"
3. **Extra item property** ⚠️ - Accepted but silently dropped (not stored)
4. **Missing required field** ❌ - Proper validation error
5. **Invalid enum value** ❌ - Clear error with valid options

### Key Finding
TodoWrite has **hybrid validation behavior**:
- Strict at root level (rejects extra properties)
- Permissive at item level (drops extra properties)

This design likely balances error detection with backward compatibility.

## Schema Design Implications

### Why Different additionalProperties Settings?

1. **TodoRead uses `true`**:
   - Expects no input (empty object preferred)
   - Permissive to prevent client errors
   - Forward compatible - won't break if clients send extra data
   - Common pattern for read-only operations

2. **TodoWrite uses `false`** (with nuanced behavior):
   - Enforces strict validation at root level
   - Catches typos in root property names
   - Silently drops extra properties in todo items
   - Balances strictness with flexibility

### Best Practices for MCP Tool Design

1. **Use `additionalProperties: false` when**:
   - You have a well-defined data model
   - Typos could cause data loss
   - You want strict validation
   - The schema is unlikely to change

2. **Use `additionalProperties: true` when**:
   - The tool accepts no input or minimal input
   - You want forward compatibility
   - The tool ignores extra fields anyway
   - You're building a flexible API

3. **Consider the trade-offs**:
   - Strict (`false`): Better error detection, less flexible
   - Permissive (`true`): More flexible, might hide errors
   - Hybrid approach: Balance strictness where it matters most

### Implementation Note for MCP Servers

When implementing MCP tools, be aware that validation might occur at multiple levels:
- **Protocol level**: May enforce root-level schema
- **Application level**: May filter or transform data
- **Storage level**: May only persist known fields

The native TodoWrite tool demonstrates this with its hybrid behavior - strict at the root but permissive for item properties.

## Conclusion

The native todo tools provide minimal functionality suitable only for brief sessions. The MCP Todo Server addresses all limitations while maintaining compatibility, enabling Claude to manage complex, long-term projects with full persistence and advanced features. Understanding the `additionalProperties` behavior helps ensure proper schema design for robust tool implementations.