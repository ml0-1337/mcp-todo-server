#!/usr/bin/env bash
#
# test_comprehensive.sh - Comprehensive end-to-end test suite for MCP Todo Server
#
# This script tests the full functionality of the MCP Todo Server in STDIO mode,
# including initialization, tool listing, todo CRUD operations, search, stats,
# and templates.
#
# Usage: ./test_comprehensive.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Comprehensive test script for MCP Todo Server

echo "=== MCP Todo Server Comprehensive Test ==="
echo

# Create a temporary file for server output
TMPFILE=$(mktemp)

# Function to send JSON-RPC request
send_request() {
    local id=$1
    local method=$2
    local params=$3
    echo "{\"jsonrpc\":\"2.0\",\"id\":$id,\"method\":\"$method\",\"params\":$params}"
}

# Start the test
(
    # 1. Initialize the server
    echo "1. Initializing server..."
    send_request 1 "initialize" '{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}'
    sleep 0.5
    
    # 2. List available tools
    echo -e "\n2. Listing available tools..."
    send_request 2 "tools/list" '{}'
    sleep 0.5
    
    # 3. Create a todo
    echo -e "\n3. Creating a test todo..."
    send_request 3 "tools/call" '{"name":"todo_create","arguments":{"task":"Complete unit tests","priority":"high","type":"feature"}}'
    sleep 0.5
    
    # 4. Read all todos
    echo -e "\n4. Reading all todos..."
    send_request 4 "tools/call" '{"name":"todo_read","arguments":{"format":"summary"}}'
    sleep 0.5
    
    # 5. Update the todo
    echo -e "\n5. Updating todo status..."
    send_request 5 "tools/call" '{"name":"todo_update","arguments":{"id":"complete-unit-tests","metadata":{"status":"in_progress"}}}'
    sleep 0.5
    
    # 6. Search for todos
    echo -e "\n6. Searching for 'unit' in todos..."
    send_request 6 "tools/call" '{"name":"todo_search","arguments":{"query":"unit"}}'
    sleep 0.5
    
    # 7. Get todo statistics
    echo -e "\n7. Getting todo statistics..."
    send_request 7 "tools/call" '{"name":"todo_stats","arguments":{}}'
    sleep 0.5
    
    # 8. List templates
    echo -e "\n8. Listing available templates..."
    send_request 8 "tools/call" '{"name":"todo_template","arguments":{}}'
    sleep 1
    
) | ./mcp-todo-server 2>&1 | tee "$TMPFILE" &

# Get the PID and wait
SERVER_PID=$!
sleep 5

# Kill the server
kill $SERVER_PID 2>/dev/null

echo -e "\n\n=== Server Output ==="
echo "===================="
cat "$TMPFILE" | jq -r 'select(.result != null) | .result' 2>/dev/null || cat "$TMPFILE"

# Clean up
rm -f "$TMPFILE"

echo -e "\n\n=== Test Complete ==="
echo "Check ~/.claude/todos/ for created todos"