# HTTP Header-Based Working Directory

## Overview

The MCP Todo Server now supports passing the working directory via HTTP headers when using HTTP transport. This allows Claude Code to specify which project directory todos should be created in, solving the issue where todos were being created in the server's installation directory.

## How It Works

### 1. HTTP Header Extraction

When using HTTP transport, the server looks for the `X-Working-Directory` header in incoming requests. This header should contain the absolute path to the project directory where todos should be created.

### 2. Session Management

The server maintains sessions using the `Mcp-Session-Id` header:
- First request with both headers creates a session
- Subsequent requests with only session ID use the stored working directory
- Sessions can be terminated with a DELETE request

### 3. Priority Order

The server resolves todo paths in this order:
1. `X-Working-Directory` header (highest priority)
2. `CLAUDE_TODO_PATH` environment variable
3. Project root detection from server's working directory

## Configuration

### Adding the Server with Claude Code

```bash
# Add server with custom header
claude mcp add --transport http todo-server http://localhost:8080/mcp \
  --header "X-Working-Directory: /path/to/your/project"
```

### Multiple Projects

You can configure different servers for different projects:

```bash
# Project 1
claude mcp add --transport http todo-project1 http://localhost:8080/mcp \
  --header "X-Working-Directory: /home/user/project1"

# Project 2
claude mcp add --transport http todo-project2 http://localhost:8080/mcp \
  --header "X-Working-Directory: /home/user/project2"
```

## Implementation Details

### Middleware Architecture

The server uses HTTP middleware to:
1. Extract headers from incoming requests
2. Store working directory in request context
3. Manage sessions for persistent connections

### Context-Aware Managers

Todo managers are created and cached per working directory:
- Each unique path gets its own manager instance
- Managers are reused across requests for the same path
- Thread-safe implementation for concurrent requests

## Testing

### Unit Tests

Run the middleware tests:
```bash
go test ./server -run TestHTTPMiddleware -v
```

### Integration Test

Run the HTTP header test script:
```bash
./test_http_headers.sh
```

This script:
1. Starts the server in HTTP mode
2. Creates todos in different directories using headers
3. Verifies todos are created in the correct locations

## Example HTTP Requests

### Initialize with Working Directory

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: /path/to/project" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "0.1.0",
      "capabilities": {},
      "clientInfo": {
        "name": "claude-code",
        "version": "1.0.0"
      }
    }
  }'
```

### Create Todo with Header

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: /path/to/project" \
  -H "Mcp-Session-Id: session-123" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Implement new feature",
        "priority": "high",
        "type": "feature"
      }
    }
  }'
```

## Benefits

1. **Project Isolation** - Todos are created in the correct project directory
2. **Multi-Tenant Support** - One server can handle multiple projects
3. **Session Persistence** - Working directory maintained across requests
4. **No Manual Configuration** - No need to set environment variables per project
5. **Claude Code Integration** - Works seamlessly with Claude Code's header support

## Future Enhancements

1. **Dynamic Variables** - Request Claude Code to support `${workspaceFolder}` in headers
2. **MCP Roots Protocol** - Implement official roots protocol when available
3. **Auto-Detection** - Detect project from request origin or other context