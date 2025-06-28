# MCP Todo Server

A Go-based Model Context Protocol (MCP) server that replaces Claude Code's native todo system with a comprehensive task management solution.

## Overview

This MCP server maintains full compatibility with the existing `.claude/todos/` markdown file structure while providing enhanced capabilities:

- 🔍 **Full-text search** across 2400+ todos
- 📝 **Automated todo creation** with rich metadata
- 📊 **Analytics and reporting** on task completion
- 🎯 **Template system** for common workflows
- 🔗 **Todo linking** for multi-phase projects
- ⚡ **<100ms response time** for all operations
- 🌐 **HTTP transport** for multi-instance support

## 🆕 New: HTTP Transport Support

The server now supports both STDIO and HTTP transports. HTTP is recommended as it allows:
- Multiple Claude Code instances to connect simultaneously
- Running multiple server instances on different ports
- Better debugging with standard HTTP tools

See [TRANSPORT_GUIDE.md](TRANSPORT_GUIDE.md) for details.

## Current Status

**Development Phase**: Phase 3 - Archive Operations  
**Tests Complete**: 7 of 23  
**Following**: Test-Driven Development (TDD) with RGRC cycle

### Completed Tests
- ✅ Test 1-4: Server initialization, validation, ID generation
- ✅ Test 5: Markdown file creation with YAML frontmatter
- ✅ Test 6: Filesystem error handling
- ✅ Test 7: ReadTodo functionality
- ✅ Archive: Daily archive structure (YYYY/MM/DD)

## Project Structure

```
mcp-todo-server/
├── .claude/
│   └── todos/
│       └── implement-mcp-todo-server.md  # Implementation todo with test progress
├── docs/
│   ├── PRD.md                    # Product Requirements Document
│   └── go_mcp_server_research.md # MCP protocol research
├── server/
│   ├── server.go                 # MCP server implementation
│   ├── server_test.go            # Server initialization tests
│   └── validation_test.go        # Parameter validation tests
├── core/
│   ├── todo.go                   # Todo business logic
│   └── todo_test.go              # Core functionality tests
├── handlers/                     # (coming soon) Tool handlers
├── storage/                      # (coming soon) File operations
├── templates/                    # (coming soon) Template engine
├── go.mod
├── go.sum
└── README.md
```

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Access to `.claude/todos/` directory
- Claude Code with MCP support

### Installation

```bash
# Clone or navigate to the project
cd /Users/macbook/Programming/go_projects/mcp-todo-server

# Download dependencies
go mod tidy

# Build the server
go build -o mcp-todo-server

# Test the installation
./test_server.sh
```

### Quick Start

#### HTTP Mode (Recommended)
```bash
# Start server on default port 8080
./mcp-todo-server -transport http

# Configure Claude Code with .mcp-http.json:
{
  "mcpServers": {
    "todo": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

#### STDIO Mode (Legacy)
```bash
# Use existing .mcp.json configuration
./mcp-todo-server -transport stdio
```

### Running Tests

```bash
# Run all tests with verbose output
go test ./... -v

# Run specific test
go test ./server -v -run TestServerInitialization

# Check test coverage
go test ./... -cover
```

## MCP Tools

The server implements 9 MCP tools:

### Core Operations
- `todo_create` - Create new todo with metadata
- `todo_read` - Read todo(s) with filtering
- `todo_update` - Update todo sections
- `todo_search` - Full-text search
- `todo_archive` - Archive completed todos

### Advanced Features
- `todo_template` - Create from templates
- `todo_link` - Link related todos
- `todo_stats` - Analytics and metrics
- `todo_clean` - Bulk management

## Todo File Format

```markdown
---
todo_id: descriptive-task-name
started: 2025-01-27 14:30:00
completed: 
status: in_progress
priority: high
type: feature
---

# Task: Description

## Findings & Research
[Documentation of research and findings]

## Test Strategy
[Test approach and coverage goals]

## Test List
[TDD test scenarios]

## Checklist
[Task items to complete]

## Working Scratchpad
[Notes and temporary content]
```

### Archive Structure

Todos are archived in a daily directory structure based on their started date:
```
.claude/archive/
└── 2025/
    ├── 01/
    │   ├── 27/    # 20-50 todos per day
    │   ├── 28/
    │   └── 29/
    └── 02/
        └── 01/
```

This structure optimizes for high-volume usage (20-50 todos/day) while maintaining good filesystem performance.

## Development Approach

Following strict TDD with RGRC (Red-Green-Refactor-Commit) cycle:

1. **Red**: Write failing test
2. **Green**: Minimal implementation to pass
3. **Refactor**: Improve code quality
4. **Commit**: Preserve progress

Each test cycle is tracked in `.claude/todos/implement-mcp-todo-server.md`.

## Configuration

Will be configured in Claude Code's `settings.json`:

```json
{
  "mcpServers": {
    "todo-server": {
      "command": "go",
      "args": ["run", "/path/to/mcp-todo-server"],
      "env": {
        "CLAUDE_TODO_PATH": "$HOME/.claude/todos"
      }
    }
  }
}
```

## Contributing

This project follows TDD principles. To contribute:

1. Check `.claude/todos/implement-mcp-todo-server.md` for next test
2. Write the failing test
3. Implement minimal solution
4. Refactor if needed
5. Update test progress
6. Commit with descriptive message

## Documentation

- **[Product Requirements Document](docs/PRD.md)** - Full specification
- **[Implementation Todo](.claude/todos/implement-mcp-todo-server.md)** - Test progress and details
- **[MCP Research](docs/go_mcp_server_research.md)** - Protocol documentation

## License

Internal project for Claude Code enhancement.

---

**Note**: To work on this project with Claude Code, launch it directly in this directory:
```bash
cd /Users/macbook/Programming/go_projects/mcp-todo-server
# Then launch Claude Code
```