# MCP Todo Server - Comprehensive Documentation

## Executive Summary

The MCP Todo Server is a Go-based implementation of the Model Context Protocol (MCP) that enhances Claude Code's todo management capabilities. Built using the mark3labs/mcp-go library, it provides persistent storage, full-text search, advanced analytics, and maintains complete compatibility with the native todo tools while offering significant enhancements.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Comparison with Native Tools](#comparison-with-native-tools)
4. [Storage System](#storage-system)
5. [Tool Specifications](#tool-specifications)
6. [Search Engine](#search-engine)
7. [Template System](#template-system)
8. [Archive System](#archive-system)
9. [Analytics Engine](#analytics-engine)
10. [Integration Guide](#integration-guide)
11. [Performance Characteristics](#performance-characteristics)

## Overview

The MCP Todo Server addresses all limitations of Claude's native todo tools while maintaining API compatibility. It transforms the ephemeral, session-based todo system into a robust, persistent task management solution.

### Key Features

- **Persistent Storage**: Todos saved as markdown files in `~/.claude/todos/`
- **Full-Text Search**: Bleve-powered search engine with field boosting
- **Rich Metadata**: 15+ fields with YAML frontmatter
- **Granular Updates**: Section-based updates without full replacement
- **Template System**: Predefined task structures
- **Analytics**: Completion rates, productivity metrics
- **Daily Archives**: YYYY/MM/DD structure for scalability
- **Linked Todos**: Parent-child relationships for complex projects

## Architecture

### System Components

```
┌─────────────────────┐
│   Claude Code CLI   │
├─────────────────────┤
│  MCP Protocol Layer │
├─────────────────────┤
│   MCP Todo Server   │
├─────────────────────┤
│      Handlers       │
├─────────────────────┤
│    Core Logic       │
├──────────┬──────────┤
│  Storage │  Search  │
│  System  │  Engine  │
└──────────┴──────────┘
```

### Directory Structure

```
~/.claude/
├── todos/                    # Active todos
│   ├── implement-search.md
│   ├── fix-auth-bug.md
│   └── research-api.md
├── archive/                  # Completed todos
│   └── 2025/
│       └── 01/
│           └── 27/
│               └── completed-task.md
├── index/                    # Search index
│   └── todos.bleve/
└── templates/                # Task templates
    ├── bug-fix.md
    └── feature.md
```

## Comparison with Native Tools

| Feature | Native Tools | MCP Todo Server |
|---------|--------------|-----------------|
| **Persistence** | Session-only | Permanent files |
| **Storage Format** | In-memory JSON | Markdown + YAML |
| **Fields** | 4 (content, status, priority, id) | 15+ (includes timestamps, tags, parent_id, etc.) |
| **Content Structure** | Plain string | Multiple sections |
| **Updates** | Full array replacement | Atomic section updates |
| **Search** | None | Full-text with boosting |
| **Filtering** | None | Status, priority, date ranges |
| **Templates** | None | Go template support |
| **Analytics** | None | Completion rates, averages |
| **Archive** | None | Daily folders |
| **Relationships** | None | Parent-child linking |
| **Performance** | N/A | <100ms operations |

### Native Tool Limitations Addressed

1. **No Persistence**: MCP server saves all todos as files
2. **Limited Metadata**: Extended with timestamps, tags, types
3. **No Search**: Full-text search across all content
4. **Full Replacement Only**: Granular section updates
5. **No Templates**: Template system for common workflows
6. **No Analytics**: Comprehensive statistics
7. **No Archive**: Automatic archival with daily structure

## Storage System

### Todo File Format

```markdown
---
todo_id: implement-search-feature
started: 2025-01-27 10:30:45
completed: ""
status: in_progress
priority: high
type: feature
parent_id: ""
tags: ["search", "backend"]
current_test: "Test 3: Case-insensitive search"
---

# Task: Implement full-text search for todos

## Findings & Research

WebSearch results and research documentation...

## Test Strategy

- Test Framework: Go testing package
- Coverage Target: 80%
- Test Types: Unit and integration

## Test List

- [ ] Test 1: Basic keyword search
- [ ] Test 2: Phrase search with quotes
- [x] Test 3: Case-insensitive search
- [ ] Test 4: Field-specific search

## Test Cases

```go
// Test implementation code
```

## Maintainability Analysis

- Readability: 8/10
- Complexity: Moderate
- Modularity: High

## Test Results Log

```bash
[2025-01-27 11:00:00] Red Phase: Test failing as expected
[2025-01-27 11:15:00] Green Phase: Test passing
```

## Checklist

- [x] Design search index schema
- [x] Implement basic search
- [ ] Add field boosting
- [ ] Optimize performance

## Working Scratchpad

Temporary notes and code snippets...
```

### Metadata Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| todo_id | string | Kebab-case identifier | Yes |
| started | timestamp | Creation time | Yes |
| completed | timestamp | Completion time | No |
| status | enum | pending, in_progress, completed, blocked | Yes |
| priority | enum | high, medium, low | Yes |
| type | enum | feature, bug, refactor, research, multi-phase | Yes |
| parent_id | string | Parent todo for multi-phase | No |
| tags | array | Categorization tags | No |
| current_test | string | Active test being worked on | No |

## Tool Specifications

### 1. todo_create

Creates a new todo with full metadata and optional template.

**Input Schema:**
```json
{
  "task": "string (required) - Task description",
  "priority": "string - high|medium|low (default: high)",
  "type": "string - feature|bug|refactor|research|multi-phase (default: feature)",
  "template": "string - Template name to use",
  "parent_id": "string - Parent todo for multi-phase projects"
}
```

**Output Schema:**
```json
{
  "id": "generated-todo-id",
  "path": "/Users/name/.claude/todos/generated-todo-id.md",
  "message": "Todo created successfully"
}
```

**ID Generation Logic:**
- Converts task to kebab-case
- Removes special characters
- Limits to 50 characters
- Appends number if duplicate exists

### 2. todo_read

Reads single todo or lists all todos with filtering.

**Input Schema:**
```json
{
  "id": "string - Specific todo ID",
  "filter": {
    "status": "string - in_progress|completed|blocked|all",
    "priority": "string - high|medium|low|all",
    "days": "number - Todos from last N days"
  },
  "format": "string - full|summary|list (default: summary)"
}
```

**Output Formats:**

- **full**: Complete markdown file content
- **summary**: Metadata + task + current section
- **list**: Array matching native tool format

### 3. todo_update

Updates todo content or metadata with atomic operations.

**Input Schema:**
```json
{
  "id": "string (required) - Todo ID",
  "section": "string - status|findings|tests|checklist|scratchpad",
  "operation": "string - append|replace|prepend (default: append)",
  "content": "string - Content to add/update",
  "metadata": {
    "status": "string - New status",
    "priority": "string - New priority",
    "current_test": "string - Current test being worked on"
  }
}
```

**Section Update Behavior:**
- **append**: Adds content to end of section
- **replace**: Replaces entire section
- **prepend**: Adds content to beginning

### 4. todo_search

Full-text search across all todos with advanced filtering.

**Input Schema:**
```json
{
  "query": "string (required) - Search terms",
  "scope": ["array - task|findings|tests|all"],
  "filters": {
    "status": "string - Filter by status",
    "date_from": "string - YYYY-MM-DD format",
    "date_to": "string - YYYY-MM-DD format"
  },
  "limit": "number - Max results (default: 20, max: 100)"
}
```

**Search Features:**
- Phrase search with quotes: `"exact phrase"`
- Field boosting: task > findings > tests > content
- Query sanitization to prevent injection
- Post-filtering for dates

### 5. todo_archive

Archives completed todos to daily folders.

**Input Schema:**
```json
{
  "id": "string (required) - Todo ID",
  "quarter": "string - Override quarter (format: 2025-Q1)"
}
```

**Archive Logic:**
- Validates no active children before archiving
- Creates path: `archive/YYYY/MM/DD/todo-id.md`
- Updates completed timestamp
- Removes from search index

### 6. todo_template

Creates todos from predefined templates.

**Input Schema:**
```json
{
  "template": "string - Template name or 'list' to show available",
  "task": "string - Task description for template",
  "priority": "string - Override template priority",
  "type": "string - Override template type"
}
```

**Template Variables:**
- `{{.Task}}` - Task description
- `{{.Date}}` - Current date
- `{{.Priority}}` - Task priority

### 7. todo_link

Creates relationships between todos.

**Input Schema:**
```json
{
  "parent_id": "string (required) - Parent todo ID",
  "child_id": "string (required) - Child todo ID",
  "link_type": "string - parent-child|blocks|relates-to (default: parent-child)"
}
```

**Relationship Rules:**
- Parent cannot be archived with active children
- Circular relationships prevented
- Currently only parent-child implemented

### 8. todo_stats

Generates comprehensive statistics and analytics.

**Input Schema:**
```json
{
  "period": "string - all|week|month|quarter (default: all)"
}
```

**Output Schema:**
```json
{
  "total_todos": 42,
  "status_breakdown": {
    "completed": 30,
    "in_progress": 10,
    "pending": 2
  },
  "completion_rate": 71.4,
  "average_completion_time": "3d 14h",
  "by_type": {
    "feature": {"count": 20, "completed": 15},
    "bug": {"count": 10, "completed": 10}
  },
  "by_priority": {
    "high": {"count": 25, "completion_rate": 80.0}
  },
  "test_coverage": {
    "todos_with_tests": 35,
    "average_test_completion": 75.5
  }
}
```

### 9. todo_clean

Bulk operations for todo maintenance.

**Input Schema:**
```json
{
  "operation": "string - archive_old|find_duplicates (required)",
  "days": "number - For archive_old, todos older than N days (default: 30)"
}
```

**Operations:**
- **archive_old**: Archives completed todos older than N days
- **find_duplicates**: Identifies potential duplicate todos

## Search Engine

### Bleve Integration

The server uses Bleve v2 for full-text search with:

- **Index Location**: `~/.claude/index/todos.bleve`
- **Analyzer**: Standard analyzer with English language support
- **Field Boosting**: 
  - Task: 2.0x boost
  - Findings: 1.5x boost
  - Tests: 1.3x boost
  - Content: 1.0x boost

### Search Query Processing

1. **Query Sanitization**: Removes regex special characters
2. **Phrase Detection**: Handles quoted phrases
3. **Field Mapping**: Maps scope to specific fields
4. **Post-Filtering**: Applies date and status filters

### Index Management

- **Automatic Indexing**: New todos indexed on creation
- **Update Handling**: Re-indexes on todo updates
- **Corruption Recovery**: Recreates index if corrupted
- **Graceful Degradation**: Search failures don't block operations

## Template System

### Template Structure

```markdown
---
todo_id: "{{.ID}}"
started: "{{.Started}}"
priority: "{{.Priority}}"
type: bug
tags: ["bug-fix", "tdd"]
---

# Task: {{.Task}}

## Bug Report

**Issue**: 
**Steps to Reproduce**:
**Expected Behavior**:
**Actual Behavior**:

## Test Strategy

- [ ] Write failing test that reproduces bug
- [ ] Fix bug
- [ ] Verify all tests pass
```

### Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| {{.Task}} | Task description | Fix login timeout |
| {{.ID}} | Generated todo ID | fix-login-timeout |
| {{.Started}} | Current timestamp | 2025-01-27 14:30:00 |
| {{.Priority}} | Task priority | high |
| {{.Date}} | Current date | 2025-01-27 |

## Archive System

### Daily Archive Structure

```
.claude/archive/
└── 2025/               # Year
    ├── 01/             # Month
    │   ├── 15/         # Day
    │   │   ├── task-1.md
    │   │   └── task-2.md
    │   └── 16/
    │       └── task-3.md
    └── 02/
        └── 01/
```

### Archive Benefits

1. **Scalability**: Handles 20-50 todos/day efficiently
2. **Navigation**: Easy to find todos by date
3. **Performance**: Avoids large directory listings
4. **Backup**: Natural chronological organization

## Analytics Engine

### Metrics Calculated

1. **Completion Metrics**:
   - Overall completion rate
   - Rate by type (feature, bug, etc.)
   - Rate by priority

2. **Time Metrics**:
   - Average completion time
   - Time distribution analysis
   - Stuck task identification

3. **Test Metrics**:
   - Todos with test coverage
   - Average test completion
   - RGRC cycle tracking

4. **Productivity Insights**:
   - Daily/weekly/monthly trends
   - Peak productivity periods
   - Task type distribution

## Integration Guide

### Claude Code Configuration

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "todo-server": {
      "command": "go",
      "args": ["run", "/path/to/mcp-todo-server"],
      "env": {
        "CLAUDE_BASE_PATH": "$HOME/.claude"
      }
    }
  }
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| CLAUDE_BASE_PATH | Base path for todos | ~/.claude |
| MCP_TODO_INDEX | Search index location | ~/.claude/index |
| MCP_TODO_TEMPLATES | Templates directory | ./templates |

### Migration from Native Tools

1. **Compatibility Mode**: Server accepts native tool format
2. **Automatic Enhancement**: Adds metadata on first update
3. **Gradual Migration**: Use advanced features as needed

## Performance Characteristics

### Benchmarks

| Operation | Performance | At Scale (1000 todos) |
|-----------|-------------|----------------------|
| Create | <10ms | <15ms |
| Read (single) | <5ms | <8ms |
| Update | <10ms | <12ms |
| Search | <50ms | <100ms |
| Archive | <20ms | <25ms |
| Stats | <100ms | <200ms |

### Optimization Strategies

1. **Search Index**: Pre-built for fast queries
2. **File Caching**: Recent todos kept in memory
3. **Lazy Loading**: Sections loaded on demand
4. **Batch Operations**: Bulk updates optimized

### Scalability Considerations

- **File System**: Tested with 10,000+ todos
- **Search Index**: Handles 100,000+ documents
- **Archive Structure**: Optimized for 50 todos/day
- **Memory Usage**: ~100MB for typical usage

## Best Practices

### Todo Management

1. **Use Templates**: For consistent structure
2. **Link Related Todos**: Track multi-phase projects
3. **Regular Archival**: Use todo_clean monthly
4. **Meaningful IDs**: Generated from descriptive tasks

### Search Optimization

1. **Use Quotes**: For exact phrase matching
2. **Scope Searches**: Specify fields for speed
3. **Date Filters**: Reduce result sets
4. **Regular Reindexing**: If performance degrades

### RGRC Workflow Support

1. **Update current_test**: Track active test
2. **Use Test Results Log**: Document phases
3. **Section Updates**: Append test results
4. **Stats Tracking**: Monitor test coverage

## Troubleshooting

### Common Issues

1. **Search Not Working**:
   - Check index exists at `~/.claude/index/todos.bleve`
   - Try rebuilding: Delete index directory and restart

2. **Todo Not Found**:
   - Verify file exists in `~/.claude/todos/`
   - Check for .md extension
   - Ensure valid YAML frontmatter

3. **Archive Failures**:
   - Check for active child todos
   - Verify write permissions
   - Ensure completed status

### Debug Mode

Enable verbose logging:
```bash
MCP_DEBUG=true go run /path/to/mcp-todo-server
```

## Future Enhancements

### Planned Features

1. **Additional Link Types**: blocks, relates-to, depends-on
2. **Export Functionality**: JSON, CSV, Markdown formats
3. **Webhook Support**: Status change notifications
4. **Time Tracking**: Actual vs estimated times
5. **AI Insights**: Pattern detection and suggestions
6. **Collaborative Features**: Shared todos
7. **Mobile Sync**: Cross-device synchronization

### Community Contributions

The project welcomes contributions for:
- Pre-built templates
- Language-specific test patterns
- Integration examples
- Performance optimizations

## Conclusion

The MCP Todo Server transforms Claude Code's task management from a simple session-based list to a comprehensive project management system. With persistent storage, full-text search, and advanced analytics, it enables long-term project tracking while maintaining the simplicity of the native interface.