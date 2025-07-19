# Connection Resilience Guide

This guide explains the connection issues that occur after macOS sleep/resume and the solutions implemented in the MCP Todo Server.

## The Problem

When a Mac goes to sleep:
1. All TCP connections are terminated by the OS
2. The MCP server sees "connection reset by peer"
3. Claude Code maintains a stale connection reference
4. After wake, Claude Code cannot reconnect automatically
5. Users must manually restart Claude Code

This is particularly frustrating for users who frequently close their laptops (sleep mode) for travel or meetings.

## Solutions Implemented

### 1. Health Check Endpoint

**Endpoint**: `/health`

The server now provides a health check endpoint that returns:
- Server status
- Uptime information
- Active session count
- Server version
- Current time

**Usage**:
```bash
curl http://localhost:8080/health
```

**Response**:
```json
{
  "status": "healthy",
  "uptime": "2h15m30s",
  "uptimeMs": 8130000,
  "serverTime": "2025-07-02T15:30:45Z",
  "transport": "http",
  "version": "2.0.0",
  "sessions": 3
}
```

### 2. HTTP Server Timeouts

The server now configures proper timeouts to detect stale connections faster:
- **ReadTimeout**: 60 seconds
- **WriteTimeout**: 60 seconds
- **IdleTimeout**: 120 seconds

These timeouts help the server clean up dead connections more quickly.

### 3. Session Management Improvements

Sessions now track:
- Last activity timestamp
- Automatic cleanup of stale sessions (10 minute timeout)
- Cleanup routine runs every 5 minutes
- Sessions are updated on each request

### 4. macOS Auto-Restart Workaround

Since Claude Code doesn't support automatic reconnection, we provide a launchd service that:
- Detects when the Mac wakes from sleep
- Automatically restarts Claude Code
- Maintains logs of restart activity

## Installation

### Server-Side Setup

The server improvements are automatically included in v2.0.0. Just ensure you're running the latest version:

```bash
# Check version
./mcp-todo-server -version

# Start with HTTP transport
./mcp-todo-server -transport http
```

### Client-Side Workaround (macOS)

1. **Install the auto-restart service**:
   ```bash
   cd scripts/macos
   ./install-restart-service.sh
   ```

2. **Test the service**:
   ```bash
   ./claude-restart-on-wake.sh test
   ```

3. **Check logs**:
   ```bash
   tail -f ~/.claude/logs/restart-on-wake.log
   ```

### Uninstall

To remove the auto-restart service:
```bash
launchctl unload ~/Library/LaunchAgents/com.claude.restart-on-wake.plist
rm ~/Library/LaunchAgents/com.claude.restart-on-wake.plist
```

## Monitoring Connection Health

### Using the Health Check Script

```bash
# Check server health
scripts/test/health_check.sh

# Check specific server
scripts/test/health_check.sh http://localhost:9090
```

### Manual Health Monitoring

You can create a simple monitoring loop:
```bash
while true; do
  curl -s http://localhost:8080/health | jq '.uptime'
  sleep 30
done
```

## Best Practices

1. **Use HTTP Transport**: More resilient than STDIO
2. **Monitor Server Logs**: Watch for session cleanup messages
3. **Regular Health Checks**: Use monitoring tools to detect issues
4. **Enable Auto-Restart**: Install the launchd service on macOS

## Future Improvements

When Claude Code adds automatic reconnection support:
1. The server already sends proper connection state
2. Health endpoint can be used for connection validation
3. Session management will handle reconnection gracefully

## Troubleshooting

### Server Not Responding After Wake

1. Check if server is still running:
   ```bash
   ps aux | grep mcp-todo-server
   ```

2. Check server logs for errors:
   ```bash
   tail -n 50 server.log
   ```

3. Verify health endpoint:
   ```bash
   curl -v http://localhost:8080/health
   ```

### Auto-Restart Not Working

1. Check if service is loaded:
   ```bash
   launchctl list | grep claude
   ```

2. Check service logs:
   ```bash
   tail ~/.claude/logs/restart-on-wake.stderr.log
   ```

3. Test manual restart:
   ```bash
   scripts/macos/claude-restart-on-wake.sh test
   ```

## Technical Details

### Session Cleanup Algorithm

```go
// Sessions inactive for 10 minutes are removed
sessionTimeout := 10 * time.Minute

// Cleanup runs every 5 minutes
ticker := time.NewTicker(5 * time.Minute)
```

### Wake Detection Method

The launchd service watches `/var/run/resolv.conf` which is updated when:
- Network changes occur
- Mac wakes from sleep
- VPN connects/disconnects

This provides reliable wake detection without polling.

### Throttling

The service includes a 60-second throttle to prevent:
- Rapid restarts during network instability
- Multiple triggers from quick sleep/wake cycles
- Excessive resource usage