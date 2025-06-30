#!/usr/bin/env bash
#
# test_multi_instance.sh - Test multiple concurrent MCP server instances
#
# This script verifies that the HTTP transport mode allows running multiple
# server instances on different ports simultaneously.
#
# Usage: ./test_multi_instance.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Test script for multiple MCP server instances

echo "=== Testing Multiple MCP Server Instances ==="
echo

# Function to start server
start_server() {
    local port=$1
    echo "Starting server on port $port..."
    ./mcp-todo-server -transport http -port $port > /tmp/mcp-server-$port.log 2>&1 &
    echo $!
}

# Function to test server
test_server() {
    local port=$1
    echo -e "\n📍 Testing server on port $port"
    
    response=$(curl -s -X POST http://localhost:$port/mcp \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"multi-test","version":"1.0.0"}}}')
    
    if echo "$response" | grep -q "MCP Todo Server"; then
        echo "✅ Server on port $port is responding"
    else
        echo "❌ Server on port $port failed to respond"
        return 1
    fi
}

# Start multiple servers
echo "Starting 3 server instances..."
PID1=$(start_server 8080)
PID2=$(start_server 8081)
PID3=$(start_server 8082)

# Wait for servers to start
sleep 3

# Test each server
test_server 8080
test_server 8081
test_server 8082

# Show that all servers are running
echo -e "\n📊 Running processes:"
ps aux | grep mcp-todo-server | grep -v grep | grep -E "808[0-2]"

# Clean up
echo -e "\n🧹 Cleaning up..."
kill $PID1 $PID2 $PID3 2>/dev/null
rm -f /tmp/mcp-server-*.log

echo -e "\n✅ Multi-instance test complete!"
echo "This demonstrates that HTTP transport allows multiple server instances on different ports."