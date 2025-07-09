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

## 🆕 New Features

### HTTP Transport Support

The server now supports both STDIO and HTTP transports. HTTP is recommended as it allows:
- Multiple Claude Code instances to connect simultaneously
- Running multiple server instances on different ports
- Better debugging with standard HTTP tools
- **Context-aware todo creation** - todos are created in the project where Claude Code is running

### Configurable Session & Manager Timeouts

New in v2.1.0: Configure how long sessions and manager sets stay in memory:
- **Session timeout**: Controls HTTP session cleanup (default: 7 days)
- **Manager timeout**: Controls manager set cleanup (default: 24 hours)
- Set to `0` to disable cleanup entirely

### HTTP Header-Based Working Directory

When using HTTP transport, the server automatically detects where Claude Code is running and creates todos in that project's directory instead of the server's directory. This is done through the `X-Working-Directory` header.

See [TRANSPORT_GUIDE.md](TRANSPORT_GUIDE.md) for transport details and [docs/http-headers.md](docs/http-headers.md) for working directory resolution.

## Current Status

**Version**: 2.1.0  
**Test Coverage**: 85-90% across most packages  
**Production Ready**: Yes - All critical functionality tested and working

### Major Improvements in v2.1.0
- ✅ Configurable session and manager timeouts
- ✅ Support for long-running Claude Code sessions
- ✅ Health check endpoint for monitoring

### Major Improvements in v2.0.0
- ✅ Complete codebase refactoring following Go best practices
- ✅ Clean architecture with Domain-Driven Design
- ✅ Fixed critical UpdateTodo operations (replace/prepend)
- ✅ Enhanced timestamp handling with multiple format support
- ✅ Improved error handling and validation
- ✅ Better test coverage and documentation

## Project Structure

```
mcp-todo-server/
├── .claude/
│   └── todos/                            # Active todo tasks
├── docs/
│   ├── overview/
│   │   └── PRD.md                # Product Requirements Document
│   └── analysis/
│       └── go_mcp_server_research.md # MCP protocol research
├── scripts/                      # All executable scripts
│   ├── setup/                    # Installation and setup
│   │   └── setup.sh             # Guided setup script
│   ├── test/                     # Test scripts by category
│   │   ├── e2e/                 # End-to-end tests
│   │   ├── http/                # HTTP transport tests
│   │   ├── stdio/               # STDIO transport tests
│   │   └── integration/         # Integration tests
│   └── README.md                # Script documentation
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

# Run the setup script for guided installation
./scripts/setup/setup.sh

# Or manually:
# Download dependencies
go mod tidy

# Build the server
go build -o mcp-todo-server

# Test the installation
./scripts/test/stdio/test_server.sh
```

### Quick Start

#### HTTP Mode (Recommended)
```bash
# 1. Start server on default port 8080
./mcp-todo-server -transport http

# 2. Add to Claude Code
claude mcp add --transport http todo http://localhost:8080/mcp
```

#### STDIO Mode (Legacy)
```bash
# 1. Configure with absolute path
claude mcp add todo /Users/macbook/Programming/go_projects/mcp-todo-server/mcp-todo-server --args "-transport" "stdio"

# 2. Server will start automatically when Claude Code connects
```

### Session Management & Timeouts

The server manages two types of resources with configurable timeouts:

#### Sessions (HTTP only)
- Lightweight connection tracking (~100 bytes each)
- Track Claude Code instances and their working directories
- Default timeout: 7 days
- Cleaned up every 5 minutes

#### Manager Sets
- Heavyweight service instances (5-20MB each)
- Contains TodoManager, SearchEngine, StatsEngine, TemplateManager
- One set per unique working directory
- Default timeout: 24 hours
- Cleaned up every 10 minutes

#### Timeout Configuration

```bash
# Default: sessions expire after 7 days, managers after 24 hours
./mcp-todo-server -transport http

# Keep sessions forever (for long-running Claude Code)
./mcp-todo-server -transport http -session-timeout 0

# Custom timeouts
./mcp-todo-server -transport http -session-timeout 30d -manager-timeout 3d

# Show available flags
./mcp-todo-server -h
```

**Common Use Cases:**
- **Long-running development**: `-session-timeout 0 -manager-timeout 24h`
- **Memory-constrained systems**: `-session-timeout 1h -manager-timeout 30m`
- **Production servers**: Default values are recommended

### Running Tests

```bash
# Best for Claude Code - clean output, fast execution
make test-claude

# Standard test suite with race detector
make test

# Quick tests without race detector
make test-quick

# See docs/testing.md for more options
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

### ⚡ Performance
- **Response time**: <100ms for all operations
- **Search latency**: <50ms for 2400+ todos
- **Startup time**: <500ms with full index load
- **Memory usage**: ~20MB base + 10KB per 1000 todos

### 🏥 Health Monitoring

The server provides a health check endpoint for monitoring:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "uptime": "2h15m30s",
  "uptimeMs": 8130000,
  "serverTime": "2025-07-02T15:30:45Z",
  "transport": "http",
  "version": "2.1.0",
  "sessions": 3
}
```

### 🧪 Test Coverage
- **Overall**: ~88% coverage
- **Core packages**: 85-90% coverage
- **Critical paths**: 100% tested
- **Architecture**: Clean, testable design with dependency injection

## Development Approach

Following strict TDD with RGRC (Red-Green-Refactor-Commit) cycle:

1. **Red**: Write failing test
2. **Green**: Minimal implementation to pass
3. **Refactor**: Improve code quality
4. **Commit**: Preserve progress

Each test cycle is tracked in `.claude/todos/implement-mcp-todo-server.md`.

## Configuration

### Command-Line Flags

```bash
-transport string     Transport type: stdio, http (default: http)
-host string         Host for HTTP transport (default: localhost)
-port string         Port for HTTP transport (default: 8080)
-session-timeout     Session timeout duration (default: 7d, 0 to disable)
-manager-timeout     Manager set timeout duration (default: 24h, 0 to disable)
-version            Print version and exit
```

### MCP Server Configuration

#### HTTP Transport with Custom Headers (Recommended)

Use HTTP transport with custom headers to specify the working directory:

```bash
# Add server with working directory header
claude mcp add --transport http todo-server http://localhost:8080/mcp \
  --header "X-Working-Directory: /path/to/your/project"
```

This ensures todos are created in your project directory, not the server's location.

#### STDIO Transport

Configure in Claude Code's `settings.json`:

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

### Template Configuration

The server looks for templates in the following locations:
1. Path specified by `CLAUDE_TEMPLATE_PATH` environment variable
2. Default: `.claude/templates/` relative to the todos directory

Example:
```json
{
  "env": {
    "CLAUDE_TEMPLATE_PATH": "/custom/path/to/templates"
  }
}
```

## Contributing

This project follows TDD principles. To contribute:

1. Check active todos in `.claude/todos/` for current tasks
2. Write the failing test (Red phase)
3. Implement minimal solution (Green phase)
4. Refactor if needed
5. Update test progress in todo file
6. Commit with descriptive message

## Documentation

- **[Product Requirements Document](docs/overview/PRD.md)** - Full specification
- **[HTTP Headers Working Directory](docs/http-headers-working-directory.md)** - Using headers for project context
- **[Testing Guide](docs/testing.md)** - How to run tests effectively
- **[Implementation Todo](.claude/archive/2025/06/28/implement-mcp-todo-server.md)** - Original implementation (archived)
- **[MCP Research](docs/analysis/go_mcp_server_research.md)** - Protocol documentation

## License

Internal project for Claude Code enhancement.

---

**Note**: To work on this project with Claude Code, launch it directly in this directory:
```bash
cd /Users/macbook/Programming/go_projects/mcp-todo-server
# Then launch Claude Code
```