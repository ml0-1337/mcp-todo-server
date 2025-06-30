#!/usr/bin/env bash
#
# test_server.sh - Basic STDIO transport mode test for MCP Todo Server
#
# This script tests basic STDIO functionality including initialization,
# tool listing, and todo creation.
#
# Usage: ./test_server.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Test script for MCP Todo Server

echo "Testing MCP Todo Server..."

# Create a temporary file for server output
TMPFILE=$(mktemp)

# Start the server in background and capture output
(
    # Send initialize request
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'
    
    # Wait a bit for response
    sleep 0.5
    
    # Send tools/list request to see available tools
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
    
    # Wait for response
    sleep 0.5
    
    # Send a test todo_create request
    echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"todo_create","arguments":{"task":"Test todo from script","priority":"high"}}}'
    
    # Wait and then kill the server
    sleep 1
) | ./mcp-todo-server 2>&1 | tee "$TMPFILE" &

# Get the PID and wait a bit
SERVER_PID=$!
sleep 3

# Kill the server
kill $SERVER_PID 2>/dev/null

echo -e "\n\n=== Server Output ==="
cat "$TMPFILE"
rm -f "$TMPFILE"

echo -e "\n\nServer test complete!"