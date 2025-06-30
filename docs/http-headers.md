# HTTP Header-Based Working Directory Resolution

## Overview

The MCP Todo Server supports dynamic working directory resolution through HTTP headers. This allows Claude Code and other MCP clients to specify where todos should be created, ensuring todos are created in the correct project directory rather than the server's directory.

## How It Works

When Claude Code connects to the MCP server over HTTP, it can send an `X-Working-Directory` header containing the path where it's currently running. The server uses this information to create todos in that project's `.claude/todos` directory.

### Header Format

```
X-Working-Directory: /path/to/your/project
```

## Implementation Details

### 1. Middleware Layer

The server uses HTTP middleware to extract headers from incoming requests:

```go
// server/middleware.go
func HTTPMiddleware(sessionManager *SessionManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract working directory from header
            workingDir := r.Header.Get("X-Working-Directory")
            
            // Extract session ID for persistent connections
            sessionID := r.Header.Get("Mcp-Session-Id")
            
            // Store in context for handlers to use
            ctx := context.WithValue(r.Context(), ctxkeys.WorkingDirectoryKey, workingDir)
            r = r.WithContext(ctx)
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2. Context-Aware Todo Manager

The `ContextualTodoManager` dynamically creates and caches todo managers for different working directories:

```go
// handlers/context_manager.go
func (c *ContextualTodoManager) GetManagerForContext(ctx context.Context) *core.TodoManager {
    // Extract working directory from context
    if workingDir, ok := ctx.Value(ctxkeys.WorkingDirectoryKey).(string); ok && workingDir != "" {
        // Resolve todo path for this working directory
        todoPath, err := utils.ResolveTodoPathFromWorkingDir(workingDir)
        if err == nil {
            // Return cached or create new manager for this path
            return c.getOrCreateManager(todoPath)
        }
    }
    // Fall back to default manager
    return c.defaultManager
}
```

### 3. Session Management

The server maintains sessions to persist working directories across multiple requests in the same connection:

```go
type SessionInfo struct {
    ID               string
    WorkingDirectory string
}
```

Sessions are managed through the `Mcp-Session-Id` header, ensuring consistent behavior throughout a connection.

## Known Issues (Fixed)

Previously, only `todo_create` operations respected the `X-Working-Directory` header. Other operations like `todo_read`, `todo_update`, and `todo_archive` would always operate on the server's directory. This has been fixed - all operations now properly use the context-aware manager.

## Testing

### Manual Test Script

```bash
#!/bin/bash
# Test HTTP header working directory

TEST_DIR="/tmp/mcp-header-test"
mkdir -p $TEST_DIR

# Initialize session
INIT_RESPONSE=$(curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}')

# Extract session ID
SESSION_ID=$(echo "$INIT_RESPONSE" | grep -i "Mcp-Session-Id:" | cut -d' ' -f2 | tr -d '\r')

# Create todo with working directory
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Test todo",
        "priority": "high"
      }
    }
  }'

# Check if todo was created in correct directory
ls -la $TEST_DIR/.claude/todos/
```

## Claude Code Integration

When Claude Code is configured to use the MCP Todo Server over HTTP, it automatically sends the `X-Working-Directory` header with each request:

```bash
# Configure Claude Code to use HTTP transport
claude mcp add todo-server http://localhost:8080/mcp
```

Claude Code will then include headers like:
- `X-Working-Directory: /path/to/current/project`
- `Mcp-Session-Id: <session-id>`

## Benefits

1. **Project Isolation**: Todos are created in the project where Claude Code is running, not in the server's directory
2. **Multiple Projects**: Can work on multiple projects simultaneously without todo conflicts
3. **Backward Compatible**: Falls back to default behavior if no header is provided
4. **Session Persistence**: Working directory is remembered for the duration of a session

## Troubleshooting

### Todos Still Created in Server Directory

1. Ensure the server was restarted after code changes
2. Verify the `X-Working-Directory` header is being sent
3. Check server logs for header processing
4. Confirm the handler is using `ContextualTodoManagerWrapper`

### "Unauthorized" Errors

The StreamableHTTPServer requires proper session management:
1. Send an `initialize` request first
2. Extract the `Mcp-Session-Id` from the response
3. Include this session ID in all subsequent requests

### Path Resolution Issues

- The server automatically creates the `.claude/todos` directory structure
- Paths must be absolute (e.g., `/Users/username/project`, not `./project`)
- The server validates and sanitizes paths for security

## Security Considerations

1. **Path Validation**: The server validates that provided paths are absolute and accessible
2. **Directory Creation**: Only creates directories under `.claude/todos` within the specified path
3. **No Path Traversal**: Paths are sanitized to prevent directory traversal attacks