#!/usr/bin/env bash
#
# test_http.sh - Basic HTTP transport mode test for MCP Todo Server
#
# This script tests basic HTTP functionality including server startup,
# initialization, tool listing, and basic todo operations.
#
# Usage: ./test_http.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Test script for MCP Todo Server in HTTP mode

echo "=== MCP Todo Server HTTP Test ==="
echo

# Start the server in HTTP mode
echo "Starting server in HTTP mode on port 8080..."
./mcp-todo-server -transport http -port 8080 > /tmp/mcp-server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "❌ Server failed to start"
    cat /tmp/mcp-server.log
    exit 1
fi

echo "✅ Server started (PID: $SERVER_PID)"

# Function to make JSON-RPC request
make_request() {
    local method=$1
    local params=$2
    local id=$3
    
    echo -e "\n📤 Request: $method"
    response=$(curl -s -X POST http://localhost:8080/mcp \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"id\":$id,\"method\":\"$method\",\"params\":$params}")
    
    echo "📥 Response:"
    echo "$response" | jq . 2>/dev/null || echo "$response"
}

# Test 1: Initialize
make_request "initialize" '{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"http-test","version":"1.0.0"}}' 1

# Test 2: List tools
make_request "tools/list" '{}' 2

# Test 3: Create a todo
make_request "tools/call" '{"name":"todo_create","arguments":{"task":"Test HTTP transport","priority":"high"}}' 3

# Test 4: Read todos
make_request "tools/call" '{"name":"todo_read","arguments":{"format":"summary"}}' 4

# Clean up
echo -e "\n🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null
rm -f /tmp/mcp-server.log

echo -e "\n✅ HTTP test complete!"