# Running Multiple Claude Code Instances with MCP Todo Server

## Quick Answer: The Server Supports Multiple Connections!

The MCP Todo Server **fully supports** multiple Claude Code instances connecting simultaneously. We've tested and confirmed:
- Multiple sessions from the same project directory ✓
- Multiple sessions from different directories ✓
- No connection limits or rejections ✓

## If Your Second Claude Code Can't Connect

The issue is with Claude Code's client-side configuration, not the server. Here's how to fix it:

### Solution 1: Check Claude Code MCP Configuration

Each Claude Code instance needs its own MCP configuration. Check your `.mcp.json` files:

```bash
# In each project directory, check:
cat .mcp.json

# Or globally:
cat ~/.claude/.mcp.json
```

Make sure:
1. The URL is correct: `http://localhost:8080/mcp`
2. Each Claude Code instance isn't using conflicting client IDs

### Solution 2: Use Different MCP Names

If you have a global MCP configuration, try using different names for each instance:

```bash
# In first Claude Code terminal:
claude mcp add todo1 http://localhost:8080/mcp

# In second Claude Code terminal:
claude mcp add todo2 http://localhost:8080/mcp
```

### Solution 3: Check for Port Conflicts

Make sure only ONE server instance is running:

```bash
# Check what's using port 8080
lsof -i :8080

# Should show only one mcp-todo-server process
```

### Solution 4: Enable Debug Logging

Run Claude Code with debug logging to see the exact error:

```bash
# Set debug environment variable
export CLAUDE_DEBUG=1

# Then run Claude Code
claude
```

## Testing Multiple Connections

You can verify the server works with multiple connections:

```bash
# Terminal 1: Start the server
./mcp-todo-server -transport http -port 8080

# Terminal 2: Test multiple connections
curl -X POST http://localhost:8080/mcp \
  -H "Mcp-Session-Id: test-1" \
  -H "X-Working-Directory: /path/to/project1" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Terminal 3: Another connection
curl -X POST http://localhost:8080/mcp \
  -H "Mcp-Session-Id: test-2" \
  -H "X-Working-Directory: /path/to/project2" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# Check active sessions
curl http://localhost:8080/debug/sessions | jq .
```

## Debug Endpoints

The server provides debug endpoints to monitor connections:

- `http://localhost:8080/health` - Server health status
- `http://localhost:8080/debug/sessions` - Active sessions
- `http://localhost:8080/debug/connections` - Connection statistics

## Common Issues and Solutions

### "Connection refused"
- Server isn't running on the expected port
- Firewall blocking the connection

### "Session already exists" 
- Claude Code is reusing session IDs
- Clear Claude Code's cache/state

### "MCP not found"
- Claude Code can't find the MCP configuration
- Re-add the MCP: `claude mcp add todo http://localhost:8080/mcp`

## Summary

The MCP Todo Server **does** support multiple Claude Code instances. If you're having connection issues:

1. It's a client-side (Claude Code) configuration issue
2. Check Claude Code's MCP configuration
3. Ensure each instance has unique settings
4. Use the debug endpoints to monitor connections
5. Check Claude Code's logs for the actual error

The server is working correctly - the issue is with how Claude Code is configured or managing connections.