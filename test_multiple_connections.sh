#!/bin/bash

# Test multiple concurrent connections to MCP server

echo "Testing multiple connections to MCP server..."

# Function to make a request with specific session ID
make_request() {
    local session_id=$1
    local working_dir=$2
    
    echo "Making request with session $session_id for $working_dir"
    
    curl -s -X POST http://localhost:8080/mcp \
        -H "Content-Type: application/json" \
        -H "Mcp-Session-Id: $session_id" \
        -H "X-Working-Directory: $working_dir" \
        -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' \
        > /tmp/response_$session_id.json
    
    if [ $? -eq 0 ]; then
        echo "✓ Session $session_id connected successfully"
    else
        echo "✗ Session $session_id failed to connect"
    fi
}

# Test 1: Multiple sessions from same directory
echo -e "\n--- Test 1: Multiple sessions from same directory ---"
make_request "claude-1" "/Users/macbook/rakulog-frontend" &
make_request "claude-2" "/Users/macbook/rakulog-frontend" &
make_request "claude-3" "/Users/macbook/rakulog-frontend" &
wait

# Check debug endpoint
echo -e "\n--- Active Sessions ---"
curl -s http://localhost:8080/debug/sessions | jq .

# Test 2: Multiple sessions from different directories
echo -e "\n--- Test 2: Multiple sessions from different directories ---"
make_request "claude-4" "/Users/macbook/project-a" &
make_request "claude-5" "/Users/macbook/project-b" &
make_request "claude-6" "/Users/macbook/project-c" &
wait

# Check debug endpoint again
echo -e "\n--- Active Sessions After Test 2 ---"
curl -s http://localhost:8080/debug/sessions | jq .

# Check connection stats
echo -e "\n--- Connection Stats ---"
curl -s http://localhost:8080/debug/connections | jq .sessions