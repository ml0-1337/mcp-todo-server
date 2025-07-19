# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP Todo Server is a Go-based Model Context Protocol (MCP) server that replaces Claude Code's native todo system with a comprehensive task management solution. It maintains full compatibility with the existing `.claude/todos/` markdown file structure while providing enhanced capabilities like search, templates, and analytics.

## Critical Guidelines

### STDIO Mode Compatibility

**CRITICAL**: In STDIO mode, the server communicates via JSON-RPC protocol over stdin/stdout. ANY output to stdout that is not valid JSON-RPC will corrupt the protocol and cause connection timeouts.

#### Logging Rules
1. **NEVER use stdout for logging** - No `fmt.Println()`, `log.Print()`, or any stdout writes
2. **Always redirect logs to stderr** - Use `fmt.Fprintf(os.Stderr, ...)` or `log.SetOutput(os.Stderr)`
3. **Check all packages** - Ensure utils, core, handlers, and server packages follow these rules
4. **Init-time logging** - Be especially careful with initialization code that runs before main()

### Archive Directory Structure

- Archives are stored in `.claude/archive/YYYY/MM/DD/` within the project directory
- When `CLAUDE_TODO_PATH` is set, archives remain within that directory structure
- Never use `filepath.Dir(basePath)` for archive paths - always use `filepath.Join(basePath, ".claude", "archive")`
- Daily archives optimize for high-volume usage (20-50 todos/day)

### Build Process

- Always use `make build` to create binaries with version info
- Binary outputs to `./build/mcp-todo-server`
- If STDIO mode hangs, check if binary is corrupted (compare with build/ version)
- Copy binary from build/ after successful compilation: `cp build/mcp-todo-server .`

## Development Commands

### Building
```bash
make build              # Build server binary with version info
make build-all          # Build for multiple platforms
make install            # Build and install to $GOPATH/bin
```

### Testing
```bash
make test-claude        # Best for Claude Code - clean output, fast execution (20s timeout)
make test               # Standard test suite with race detector (30s timeout)
make test-quick         # Quick tests without race detector (20s timeout)
make test-coverage      # Generate HTML coverage report
make test-e2e           # Run comprehensive end-to-end tests
```

### Running
```bash
make server-http        # Start HTTP server on port 8080 (recommended)
make server-stdio       # Start STDIO server (legacy)
make run                # Build and run in HTTP mode
```

### Development
```bash
make lint               # Run golangci-lint
make fmt                # Format code with go fmt
make vet                # Run go vet for static analysis
make check              # Run all checks (fmt, vet, lint, test)
make mod-tidy           # Clean up go.mod dependencies
```

## High-Level Architecture

### Transport Layer
- **HTTP Mode** (recommended): Supports multiple instances, context-aware todo creation
- **STDIO Mode**: Legacy single-instance mode, requires careful stdout handling
- Both modes use the MCP protocol for communication

### Handler Layer (`handlers/`)
- Processes MCP tool requests
- Parameter extraction and validation
- Business logic orchestration
- Error conversion to MCP responses

Key interfaces:
- `TodoManager` - Core todo operations
- `SearchEngine` - Full-text search with Bleve
- `StatsEngine` - Analytics operations
- `TemplateManager` - Template operations

### Core Layer (`core/`)
- Todo lifecycle management
- Search indexing with Bleve (sub-50ms for 2400+ todos)
- Statistics calculation
- Template processing
- Multi-phase project linking

### Storage Layer (`storage/`)
- File system operations
- Path resolution and validation
- Archive management
- Directory structure maintenance

## Code Organization

### Package Structure
```
server/         # MCP server setup and tool registration
handlers/       # MCP tool handlers with clean interfaces
core/           # Business logic implementation
internal/       # Shared utilities
  errors/       # Structured error types
  validation/   # Centralized validation
  testutil/     # Test helpers and fixtures
storage/        # File system operations
```

### Interface Conventions
Following Go idioms, interfaces don't have "Interface" suffix:
- `TodoManager` (not TodoManagerInterface)
- `SearchEngine` (not SearchEngineInterface)

### Error Handling
```go
// In core layer
if os.IsNotExist(err) {
    return interrors.NewNotFoundError("todo", id)
}

// In handler layer
if interrors.IsNotFound(err) {
    return mcp.NewToolResultError("Todo not found")
}
```

### Testing Strategy
- Unit tests with mocked interfaces
- Integration tests with real file system
- Test utilities in `internal/testutil`
- Use temporary directories to avoid conflicts

## Transport Modes

### HTTP Mode (Recommended)
```bash
# Start server
./mcp-todo-server -transport http -port 8080

# Configure Claude Code
claude mcp add --transport http todo http://localhost:8080/mcp
```

**Context-Aware Todo Creation**: HTTP mode automatically creates todos in the project where Claude Code is running through the `X-Working-Directory` header.

### STDIO Mode Considerations
```bash
# Start server (with logging to stderr)
./mcp-todo-server -transport stdio 2>server.log

# Configure Claude Code
claude mcp add todo /path/to/mcp-todo-server --args "-transport" "stdio"
```

## Common Development Tasks

### Adding a New MCP Tool
1. Define handler method in `handlers/todo_handlers_*.go`
2. Add parameter extraction in `handlers/params_*.go`
3. Add response formatting in `handlers/responses_*.go`
4. Register tool in `server/server.go`
5. Add tests for handler, params, and responses

### Running Specific Tests
```bash
# Single test
go test -run TestSpecificFunction -v

# Pattern matching
go test -run "TestTodo.*" -v

# Package tests
go test ./handlers -v
```

### Debugging Test Failures
1. Use `make test-verbose` for detailed output
2. Check for blocking operations or missing mocks
3. Ensure temp directories are used in tests
4. Verify timeouts are appropriate

### Working with Todos
- Todos stored in `.claude/todos/` directory
- Archives in `.claude/archive/YYYY/MM/DD/`
- Use MCP tools, never direct file manipulation
- Test with real file operations for integration

## CI/CD Integration

### GitHub Actions Workflow
- **Lint**: golangci-lint on all code
- **Test Matrix**: Ubuntu/macOS/Windows × Go 1.22/1.23/1.24
- **Build**: Multi-platform binaries
- **Integration**: End-to-end tests
- **Security**: Trivy vulnerability scanning

### Running CI Locally
```bash
# Test workflows with act
make workflow-test

# Lint workflow files
make workflow-lint
```

## Quick Reference

### Key Files
- `Makefile` - All build/test/run commands
- `docs/guides/transport-guide.md` - Transport mode details
- `docs/testing.md` - Comprehensive testing guide
- `docs/development/architecture.md` - Detailed architecture

### Performance Targets
- Response time: <100ms for all operations
- Search latency: <50ms for 2400+ todos
- Startup time: <500ms with full index load
- Memory usage: ~20MB base + 10KB per 1000 todos

### Dependencies
- `github.com/mark3labs/mcp-go` - MCP protocol implementation ([Documentation](https://mcp-go.dev))
- `github.com/blevesearch/bleve/v2` - Full-text search engine
- Go 1.21+ required (uses Go 1.23 in CI)

## MCP-Go Library Documentation

The mcp-todo-server is built on the mark3labs/mcp-go library. Consult the official documentation at https://mcp-go.dev for detailed implementation guidance.

### Core MCP Concepts
- **[Core Concepts](https://mcp-go.dev/core-concepts/)** - Understanding Resources, Tools, Prompts, and Transports
- **[Getting Started](https://mcp-go.dev/getting-started/)** - MCP fundamentals and first server

### Server Implementation
- **[Server Basics](https://mcp-go.dev/servers/basics/)** - Server lifecycle, configuration, and middleware
- **[Tools](https://mcp-go.dev/servers/tools/)** - Implementing MCP tools with schemas, handlers, and streaming
- **[Resources](https://mcp-go.dev/servers/resources/)** - Creating resource endpoints for read-only data
- **[Prompts](https://mcp-go.dev/servers/prompts/)** - Building reusable prompt templates
- **[Advanced Features](https://mcp-go.dev/servers/advanced/)** - Typed tools, session management, hooks, sampling

### Transport Modes
- **[Transport Options](https://mcp-go.dev/transports/)** - Comparison of STDIO, HTTP, SSE transports
- **[STDIO Transport](https://mcp-go.dev/transports/stdio/)** - Subprocess communication (critical for STDIO mode issues)
- **[StreamableHTTP](https://mcp-go.dev/transports/http/)** - HTTP transport implementation details

### Testing & Development
- **[Client Basics](https://mcp-go.dev/clients/basics/)** - Creating MCP clients for testing
- **[Client Operations](https://mcp-go.dev/clients/operations/)** - Testing tool calls and responses
- **[Examples](https://github.com/mark3labs/mcp-go/tree/main/examples)** - Complete implementation examples

### When to Consult MCP-Go Docs

**Adding New Tools**: See [Tools documentation](https://mcp-go.dev/servers/tools/) for:
- Parameter schema definitions
- Handler implementation patterns
- Error handling best practices
- Streaming results for long operations

**Transport Issues**: See [Transport documentation](https://mcp-go.dev/transports/) for:
- STDIO logging and stdout corruption issues
- HTTP header handling for context awareness
- Choosing appropriate transport for deployment

**Testing**: See [Client documentation](https://mcp-go.dev/clients/) for:
- Creating test clients
- Validating tool responses
- Integration testing patterns