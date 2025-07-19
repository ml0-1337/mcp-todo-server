# Product Requirements Document: MCP Todo Server
**Version**: 1.0.0  
**Date**: January 2025  
**Status**: In Development

## 1. Executive Summary

The MCP Todo Server is a Go-based Model Context Protocol server that replaces Claude Code's native todo system with a comprehensive task management solution. It maintains full compatibility with the existing `.claude/todos/` markdown file structure while providing enhanced capabilities for task creation, management, search, and analytics.

## 2. Problem Statement

### Current Pain Points
1. **Limited Native Tools**: Current TodoRead/TodoWrite only handle simple JSON arrays with minimal metadata
2. **No Persistence**: Native todos are lost after session ends
3. **No Search Capability**: With 2400+ existing todos, finding specific tasks is difficult
4. **Manual Process**: Creating proper todo files requires multiple bash commands
5. **No Analytics**: Cannot track productivity patterns or task completion metrics
6. **No Templates**: Repetitive task structures must be recreated each time

### Impact
- Lost research and documentation when sessions end
- Time wasted recreating task structures
- Difficulty tracking multi-phase projects
- Reduced productivity from manual file operations
- Inability to learn from past task patterns

## 3. Goals and Objectives

### Primary Goals
1. **Preserve CLAUDE.md Workflow**: Maintain 100% compatibility with existing todo file format
2. **Enhance Productivity**: Reduce todo creation time by 80% through automation
3. **Enable Discovery**: Make all 2400+ todos searchable and analyzable
4. **Automate Workflows**: Handle timestamps, archiving, and formatting automatically

### Success Metrics
- Todo creation time: <5 seconds (from current ~30 seconds)
- Search response time: <100ms for 10,000 todos
- Zero data loss during migration
- 100% format compatibility with existing files

## 4. User Stories and Use Cases

### User Stories

**As Claude Code, I want to:**
1. Create a todo with full metadata in one command
2. Search all todos by content, status, or date
3. Update todo progress without manual file editing
4. Archive completed todos automatically
5. Track test progress through RGRC cycles
6. Link related todos for multi-phase projects
7. Generate analytics on task completion patterns
8. Use templates for common task types

### Detailed Use Cases

#### UC1: Create New Todo
```
Actor: Claude Code
Trigger: User requests new task
Flow:
1. Call todo_create with task description
2. Server generates unique ID and timestamp
3. Server creates markdown file with full structure
4. Returns todo ID for reference
Outcome: Complete todo file created in <1 second
```

#### UC2: Search Historical Todos
```
Actor: Claude Code
Trigger: Need to find similar past work
Flow:
1. Call todo_search with query terms
2. Server searches content, findings, test cases
3. Returns ranked results with snippets
4. Can filter by status, date, priority
Outcome: Find relevant todos from 2400+ in <100ms
```

#### UC3: Track RGRC Test Progress
```
Actor: Claude Code
Trigger: Working through TDD cycle
Flow:
1. Update test status to "Red phase"
2. Add test failure output
3. Update to "Green phase" with implementation
4. Mark test complete, move to next
5. Server tracks all state changes
Outcome: Complete test history preserved
```

## 5. Functional Requirements

### 5.1 Core Todo Operations

#### FR1: todo_create
**Description**: Create new todo with full metadata structure
**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "task": {
      "type": "string",
      "description": "Task description"
    },
    "priority": {
      "type": "string",
      "enum": ["high", "medium", "low"],
      "default": "high"
    },
    "type": {
      "type": "string",
      "enum": ["feature", "bug", "refactor", "research", "multi-phase"],
      "default": "feature"
    },
    "template": {
      "type": "string",
      "description": "Optional template name"
    },
    "parent_id": {
      "type": "string",
      "description": "Parent todo for multi-phase projects"
    }
  },
  "required": ["task"]
}
```
**Output**: Todo ID and file path
**File Creation**: Full markdown with YAML frontmatter

#### FR2: todo_read
**Description**: Read single todo or list all todos
**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Specific todo ID"
    },
    "filter": {
      "type": "object",
      "properties": {
        "status": {
          "type": "string",
          "enum": ["in_progress", "completed", "blocked", "all"]
        },
        "priority": {
          "type": "string",
          "enum": ["high", "medium", "low", "all"]
        },
        "days": {
          "type": "integer",
          "description": "Todos from last N days"
        }
      }
    },
    "format": {
      "type": "string",
      "enum": ["full", "summary", "list"],
      "default": "summary"
    }
  }
}
```
**Output**: Todo content or filtered list

#### FR3: todo_update
**Description**: Update todo content or metadata
**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Todo ID to update"
    },
    "section": {
      "type": "string",
      "enum": ["status", "findings", "tests", "checklist", "scratchpad"]
    },
    "operation": {
      "type": "string",
      "enum": ["append", "replace", "prepend"]
    },
    "content": {
      "type": "string",
      "description": "Content to add/update"
    },
    "metadata": {
      "type": "object",
      "properties": {
        "status": {"type": "string"},
        "priority": {"type": "string"},
        "current_test": {"type": "string"}
      }
    }
  },
  "required": ["id"]
}
```

#### FR4: todo_search
**Description**: Full-text search across all todos
**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search terms"
    },
    "scope": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["task", "findings", "tests", "all"]
      },
      "default": ["all"]
    },
    "filters": {
      "type": "object",
      "properties": {
        "status": {"type": "string"},
        "date_from": {"type": "string", "format": "date"},
        "date_to": {"type": "string", "format": "date"}
      }
    },
    "limit": {
      "type": "integer",
      "default": 20,
      "maximum": 100
    }
  },
  "required": ["query"]
}
```
**Output**: Ranked search results with snippets

#### FR5: todo_archive
**Description**: Archive completed todo to quarterly folder
**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Todo ID to archive"
    },
    "quarter": {
      "type": "string",
      "description": "Override quarter (YYYY-QN)",
      "pattern": "^\\d{4}-Q[1-4]$"
    }
  },
  "required": ["id"]
}
```
**Actions**: 
- Set completed timestamp
- Move to archive folder
- Update any linked todos

### 5.2 Advanced Features

#### FR6: todo_template
**Description**: Create todo from predefined template
**Templates**:
- bug-fix: Bug reproduction and fix workflow
- feature: New feature with TDD
- refactor: Code improvement workflow
- research: Investigation and documentation
- multi-phase: Project with sub-todos

#### FR7: todo_link
**Description**: Link related todos (parent-child, dependencies)
**Relationships**:
- parent-child: Multi-phase projects
- blocks: Dependencies
- relates-to: Related work

#### FR8: todo_stats
**Description**: Analytics and reporting
**Metrics**:
- Completion rates by type/priority
- Average time to complete
- Test coverage statistics
- Productivity trends

#### FR9: todo_clean
**Description**: Manage large todo collection
**Operations**:
- Archive old completed todos
- Find duplicate/similar todos
- Bulk status updates
- Export for backup

### 5.3 Test Management Features

#### FR10: test_update
**Description**: Update test list and track RGRC progress
**Phases**: Red → Green → Refactor → Commit
**Tracking**: Current test, phase, results

## 6. Non-Functional Requirements

### 6.1 Performance Requirements
- **Response Time**: All operations <100ms for up to 10,000 todos
- **Concurrent Operations**: Support parallel tool calls
- **Memory Usage**: <100MB for 10,000 todos
- **Startup Time**: <1 second

### 6.2 Reliability Requirements
- **Data Durability**: No data loss on crashes
- **Atomic Operations**: File writes must be atomic
- **Backup Strategy**: Support incremental backups
- **Error Recovery**: Graceful handling of corrupted files

### 6.3 Security Requirements
- **File Access**: Restricted to .claude directory
- **Input Validation**: Prevent path traversal attacks
- **No External Network**: Pure local file operations
- **Audit Trail**: Log all modifications

### 6.4 Compatibility Requirements
- **File Format**: 100% compatible with existing todos
- **Go Version**: 1.21+
- **Platform**: macOS, Linux, Windows
- **MCP Protocol**: Latest 2025 specification

## 7. Technical Architecture

### 7.1 System Architecture
```
┌─────────────────┐
│  Claude Code    │
│  (MCP Client)   │
└────────┬────────┘
         │ stdio
┌────────┴────────┐
│  MCP Todo       │
│  Server (Go)    │
├─────────────────┤
│ - Tool Handler  │
│ - File Manager  │
│ - Search Engine │
│ - Template Eng. │
└────────┬────────┘
         │
┌────────┴────────┐
│ .claude/todos/  │
│  (Markdown)     │
└─────────────────┘
```

### 7.2 Component Design

#### Core Components
1. **MCP Server**: Handles protocol, tool registration
2. **Todo Manager**: CRUD operations, validation
3. **File System**: Atomic file operations
4. **Search Engine**: Full-text indexing (bleve)
5. **Template Engine**: Template management
6. **Analytics Engine**: Statistics generation

#### Data Flow
1. Tool request → JSON-RPC handler
2. Parameter validation → Business logic
3. File operations → Atomic writes
4. Response formatting → JSON-RPC response

### 7.3 File Structure
```
.claude/
├── todos/
│   ├── [todo-id].md          # Active todos
│   └── main-[project].md      # Multi-phase projects
├── archive/
│   └── YYYY-QQ/
│       └── [todo-id].md       # Archived todos
├── templates/
│   └── [template-name].md     # Todo templates
└── index/
    └── todos.bleve/           # Search index
```

## 8. Data Models

### 8.1 Todo File Format
```markdown
---
todo_id: descriptive-task-name
started: 2025-01-27 14:30:00
completed: 
status: in_progress
priority: high
type: feature
parent_id: 
tags: [tdd, authentication, security]
---

# Task: Implement JWT authentication

## Findings & Research

[Research findings, WebSearch results, etc.]

## Test Strategy

- **Test Framework**: Jest
- **Test Types**: Unit, Integration
- **Coverage Target**: 90%
- **Edge Cases**: Token expiry, invalid signatures

## Test List

- [ ] Test 1: Should generate valid JWT token
- [ ] Test 2: Should reject expired tokens
- [ ] Test 3: Should validate signatures

**Current Test**: Working on Test 1: token generation
**Phase**: Red - Writing failing test

## Test Cases

```javascript
// Test 1: Token generation
describe('generateToken', () => {
  it('should generate valid JWT', () => {
    // Test implementation
  });
});
```

## Maintainability Analysis

- **Readability**: 8/10 - Clear function names
- **Complexity**: Low - Single responsibility
- **Modularity**: High - Separate auth module
- **Testability**: High - Pure functions

## Test Results Log

```bash
# 2025-01-27 14:35:00 - Red Phase
Expected generateToken to be defined
Received: undefined

# 2025-01-27 14:40:00 - Green Phase  
✓ should generate valid JWT (5ms)
```

## Checklist

- [x] Research JWT best practices
- [x] Design test strategy
- [ ] Implement all tests
- [ ] Add error handling
- [ ] Document API

## Working Scratchpad

Notes, code snippets, commands, etc.
```

### 8.2 Search Index Schema
```go
type TodoDocument struct {
    ID          string    `json:"id"`
    Task        string    `json:"task"`
    Status      string    `json:"status"`
    Priority    string    `json:"priority"`
    Type        string    `json:"type"`
    Started     time.Time `json:"started"`
    Completed   time.Time `json:"completed"`
    Content     string    `json:"content"`      // Full text
    Findings    string    `json:"findings"`    // Research section
    Tests       string    `json:"tests"`       // Test content
    Tags        []string  `json:"tags"`
}
```

## 9. API Design (MCP Tools)

### 9.1 Tool Registration
```go
server.AddTool(mcp.NewTool("todo_create",
    mcp.WithDescription("Create a new todo with full metadata"),
    mcp.WithObject(
        mcp.WithProperty("task", 
            mcp.PropertyString().
                WithDescription("Task description").
                Required()),
        mcp.WithProperty("priority",
            mcp.PropertyString().
                WithEnum("high", "medium", "low").
                WithDefault("high")),
    ),
), todoCreateHandler)
```

### 9.2 Error Responses
```json
{
  "error": {
    "code": "TODO_NOT_FOUND",
    "message": "Todo with ID 'xyz' not found",
    "details": {
      "id": "xyz",
      "suggestion": "Use todo_search to find todos"
    }
  }
}
```

## 10. Testing Strategy

### 10.1 Test Coverage Requirements
- Unit Tests: 90% coverage
- Integration Tests: All tool handlers
- Performance Tests: Search with 10k todos
- Stress Tests: Concurrent operations

### 10.2 Test Scenarios
1. **File Operations**: Create, read, update, archive
2. **Search**: Query parsing, ranking, filtering  
3. **Templates**: Loading, variable substitution
4. **Error Handling**: Corrupted files, missing todos
5. **Concurrency**: Parallel tool calls

## 11. Migration Plan

### 11.1 Phase 1: Preparation
1. Analyze existing 2400+ todos
2. Identify format variations
3. Build compatibility layer
4. Create backup of all todos

### 11.2 Phase 2: Deployment
1. Disable native TodoRead/TodoWrite
2. Configure MCP server in settings.json
3. Start server with existing todos
4. Verify all todos accessible

### 11.3 Phase 3: Enhancement
1. Build search index
2. Analyze todo patterns
3. Create common templates
4. Generate initial analytics

## 12. Security Considerations

### 12.1 Input Validation
- Sanitize all file paths
- Validate todo IDs (alphanumeric + dash)
- Limit content size (100KB per todo)
- Escape special characters

### 12.2 Access Control
- Restrict to .claude directory
- No symlink following
- Read-only template access
- Audit all write operations

## 13. Future Enhancements

### Phase 2 Features
1. **AI-Powered Features**:
   - Auto-generate test cases from task
   - Suggest similar past solutions
   - Predict task complexity

2. **Collaboration**:
   - Export todos for sharing
   - Import external task lists
   - Git integration for todo files

3. **Advanced Analytics**:
   - Burndown charts
   - Velocity tracking
   - Pattern recognition

### Phase 3 Features
1. **Workflow Automation**:
   - Auto-create todos from PR comments
   - Slack/Discord notifications
   - Calendar integration

2. **Enhanced Search**:
   - Semantic search with embeddings
   - Natural language queries
   - Code snippet search

## 14. Assumptions and Dependencies

### Assumptions
- .claude directory exists and is writable
- Go 1.21+ available for builds
- macOS/Linux/Windows file systems
- Claude Code continues using MCP

### Dependencies
- mark3labs/mcp-go: MCP protocol implementation
- bleve: Full-text search engine
- YAML parser for frontmatter
- Standard Go libraries only

## 15. Success Criteria

### Launch Criteria
- [ ] All existing todos readable
- [ ] Search returns results in <100ms
- [ ] No data loss during migration
- [ ] All CLAUDE.md workflows functional

### 3-Month Success Metrics
- 90% reduction in todo creation time
- 95% of todos created via templates
- Search used 50+ times per week
- Zero data corruption incidents

---

**Current Status**: Phase 2 - Todo CRUD Operations (Test 4 of 23 complete)