# MCP Todo Server Transport Guide

## Overview

The MCP Todo Server now supports two transport modes:
- **STDIO** (default for backward compatibility)
- **HTTP** (recommended for multi-instance support)

## Why HTTP Transport?

The original STDIO transport has a limitation: only one Claude Code instance can connect to the server at a time. This is because STDIO communication uses stdin/stdout, which can only have one active connection.

With HTTP transport, you can:
- Run multiple Claude Code instances simultaneously
- Connect to the same todo server from different projects
- Run multiple server instances on different ports
- Use standard HTTP tools for debugging
- **Automatic project-based todo creation** - todos are created in the project where Claude Code is running, not in the server's directory

## Usage

### Starting the Server

#### HTTP Mode (Recommended)
```bash
# Default port 8080
./mcp-todo-server -transport http

# Custom port
./mcp-todo-server -transport http -port 8090

# Custom host and port
./mcp-todo-server -transport http -host 0.0.0.0 -port 8090
```

#### STDIO Mode (Legacy)
```bash
# Explicitly use STDIO
./mcp-todo-server -transport stdio

# Or just run without flags (defaults to http now)
./mcp-todo-server
```

### Configuration

#### Using Claude Code CLI (Recommended)

For HTTP transport:
```bash
claude mcp add --transport http todo http://localhost:8080/mcp
```

For STDIO transport:
```bash
claude mcp add todo /path/to/mcp-todo-server --args "-transport" "stdio"
```

#### Manual Configuration Files

For HTTP Transport (.mcp.json):
```json
{
  "mcpServers": {
    "todo": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

For STDIO Transport (.mcp.json):
```json
{
  "mcpServers": {
    "todo": {
      "type": "stdio",
      "command": "/path/to/mcp-todo-server",
      "args": ["-transport", "stdio"]
    }
  }
}
```

## Multiple Instances

With HTTP transport, you can run multiple server instances:

```bash
# Terminal 1: Personal todos
./mcp-todo-server -transport http -port 8080

# Terminal 2: Work todos
./mcp-todo-server -transport http -port 8081

# Terminal 3: Project-specific todos
./mcp-todo-server -transport http -port 8082
```

Each can be configured in different projects with their own `.mcp-http.json` files.

## Context-Aware Todo Creation

When using HTTP transport, the server automatically creates todos in the correct project directory through the `X-Working-Directory` header that Claude Code sends. This means:

- Todos are created in `/your/project/.claude/todos/` not in the server's directory
- Each project maintains its own todo list
- No manual configuration needed - it just works!

For technical details, see [docs/http-headers.md](docs/http-headers.md).

## Testing

### Test HTTP Server
```bash
./test_http.sh
```

### Test Multiple Instances
```bash
./test_multi_instance.sh
```

### Manual Testing with curl
```bash
# Initialize
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"curl","version":"1.0.0"}}}'

# List tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```

## Migration from STDIO to HTTP

1. Stop any running STDIO servers
2. Update your `.mcp.json` to `.mcp-http.json` format
3. Start the server with `-transport http`
4. Restart Claude Code

## Troubleshooting

### Port Already in Use
If you get "address already in use" error:
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill <PID>
```

### Multiple Instance Prevention
**New in v2.1.0**: The server automatically prevents multiple instances from running on the same port using file locking.

If you try to start a second instance, you'll see:
```
Failed to acquire server lock: another instance is already running on this port
```

This prevents conflicts and ensures only one server instance per port.

### Multiple Zombie Processes
The HTTP transport prevents zombie processes since each instance runs on its own port, and file locking prevents accidental duplicate instances.

### Connection Refused
- Ensure the server is running: `ps aux | grep mcp-todo-server`
- Check the correct port is being used
- Verify firewall settings if connecting remotely

## Environment Variables

You can also use environment variables:
```bash
export MCP_TRANSPORT=http
export MCP_PORT=8090
./mcp-todo-server
```

## Future Enhancements

- Automatic port selection if default is busy
- Built-in authentication for HTTP transport
- WebSocket support for real-time updates
- HTTPS support with TLS certificates