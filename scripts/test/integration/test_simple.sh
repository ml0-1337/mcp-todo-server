#!/usr/bin/env bash
#
# test_simple.sh - Simple HTTP header and session management test
#
# This script tests session management and working directory isolation
# using HTTP headers including X-Working-Directory and Mcp-Session-Id.
#
# Usage: ./test_simple.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "Testing HTTP header-based working directory..."

# Test directory
TEST_DIR="/tmp/mcp-header-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

echo "Test directory: $TEST_DIR"

# Initialize session
echo -e "\n1. Initializing session..."
INIT_RESPONSE=$(curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}')

# Extract session ID from headers
SESSION_ID=$(echo "$INIT_RESPONSE" | grep -i "Mcp-Session-Id:" | cut -d' ' -f2 | tr -d '\r')

echo "Session ID: $SESSION_ID"

# Create todo
echo -e "\n2. Creating todo..."
RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
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
        "task": "Test todo with HTTP header working directory",
        "priority": "high",
        "type": "feature"
      }
    }
  }')

echo "Response:"
echo "$RESPONSE"

# Try to parse with jq, but don't fail if it doesn't work
echo "$RESPONSE" | jq . 2>/dev/null || echo "Raw response: $RESPONSE"

# Extract file path from response - the path is in the JSON text
FILE_PATH=$(echo "$RESPONSE" | jq -r '.result.content[0].text' 2>/dev/null | jq -r '.path' 2>/dev/null || echo "")

echo -e "\n3. Checking results..."
if [ -z "$FILE_PATH" ]; then
    echo "✗ Could not extract file path from response"
else
    echo "Todo created at: $FILE_PATH"
    
    # Check if file exists in test directory
    if [[ "$FILE_PATH" == "$TEST_DIR"* ]]; then
        echo "✓ SUCCESS: Todo created in test directory!"
        
        if [ -f "$FILE_PATH" ]; then
            echo -e "\nTodo contents:"
            head -n 15 "$FILE_PATH"
        fi
    else
        echo "✗ FAILED: Todo created in wrong directory: $FILE_PATH"
    fi
fi

# Also check directory listing
echo -e "\n4. Directory listing for $TEST_DIR/.claude/todos/:"
if ls -la $TEST_DIR/.claude/todos/*.md 2>/dev/null; then
    echo "✓ Found todos in test directory"
else
    echo "✗ No todos found in test directory"
fi

echo -e "\nTest complete!"