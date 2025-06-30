#!/usr/bin/env bash
#
# test_fix.sh - Test context-aware working directory fix
#
# This script verifies that todos are created in the correct directory
# when using the X-Working-Directory header, not in the server's directory.
#
# Usage: ./test_fix.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "Testing context-aware todo creation fix..."

# Create test directory
TEST_DIR="/tmp/mcp-fix-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

echo "Test directory: $TEST_DIR"

# Start server in background
echo "Starting server..."
./mcp-todo-server -transport http -port 8081 &
SERVER_PID=$!

# Give server time to start
sleep 2

# Create a todo with custom working directory
echo -e "\nCreating todo with X-Working-Directory header..."
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "0.1.0",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }' | jq .

echo -e "\nCreating todo..."
RESPONSE=$(curl -s -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Test todo with fixed context awareness",
        "priority": "high",
        "type": "feature"
      }
    }
  }')

echo "$RESPONSE" | jq .

# Give time for file to be written
sleep 1

# Check results
echo -e "\n=== Checking results ==="
echo "Looking for todos in $TEST_DIR/.claude/todos/..."
if ls $TEST_DIR/.claude/todos/*.md 2>/dev/null; then
    echo -e "\n✓ SUCCESS: Todo created in the correct directory!"
    echo -e "\nTodo contents:"
    cat $TEST_DIR/.claude/todos/*.md | head -n 20
else
    echo -e "\n✗ FAILED: No todos found in $TEST_DIR/.claude/todos/"
    echo "Checking server's directory..."
    if ls .claude/todos/*.md 2>/dev/null | grep -v "create-amp-proxy"; then
        echo "ERROR: Todos are still being created in server directory!"
    fi
fi

# Clean up
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\nTest complete!"