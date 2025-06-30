#!/usr/bin/env bash
#
# test_context_complete.sh - Comprehensive test for context-aware todo operations
#
# This script tests all todo operations (create, read, update, archive) to ensure
# they correctly use the X-Working-Directory header for project isolation.
#
# Usage: ./test_context_complete.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "Testing all context-aware todo operations..."

# Test directory
TEST_DIR="/tmp/mcp-context-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

echo "Test directory: $TEST_DIR"

# Initialize session
echo -e "\n1. Initializing session..."
INIT_RESPONSE=$(curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}')

SESSION_ID=$(echo "$INIT_RESPONSE" | grep -i "Mcp-Session-Id:" | cut -d' ' -f2 | tr -d '\r')
echo "Session ID: $SESSION_ID"

# Test 1: Create todo
echo -e "\n2. Testing CREATE..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
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
        "task": "Context test todo",
        "priority": "high",
        "type": "feature"
      }
    }
  }')

TODO_ID=$(echo "$CREATE_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.id')
echo "Created todo: $TODO_ID"

# Test 2: Read todos
echo -e "\n3. Testing READ (list)..."
READ_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "todo_read",
      "arguments": {
        "format": "list"
      }
    }
  }')

if echo "$READ_RESPONSE" | jq -r '.result.content[0].text' | grep -q "$TODO_ID"; then
    echo "✓ READ works: Found todo in project directory"
else
    echo "✗ READ failed: Todo not found"
fi

# Test 3: Update todo
echo -e "\n4. Testing UPDATE..."
UPDATE_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 4,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"todo_update\",
      \"arguments\": {
        \"id\": \"$TODO_ID\",
        \"section\": \"findings\",
        \"operation\": \"append\",
        \"content\": \"This is a test update from context-aware operations\"
      }
    }
  }")

if echo "$UPDATE_RESPONSE" | jq -e '.result.content[0].text' > /dev/null 2>&1; then
    echo "✓ UPDATE works: Successfully updated todo"
else
    echo "✗ UPDATE failed: $(echo "$UPDATE_RESPONSE" | jq -r '.error.message // "Unknown error"')"
fi

# Test 4: Read specific todo
echo -e "\n5. Testing READ (specific todo)..."
READ_SINGLE_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 5,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"todo_read\",
      \"arguments\": {
        \"id\": \"$TODO_ID\",
        \"format\": \"full\"
      }
    }
  }")

if echo "$READ_SINGLE_RESPONSE" | jq -r '.result.content[0].text' | grep -q "test update from context-aware"; then
    echo "✓ READ (single) works: Found our update in content"
else
    echo "✗ READ (single) failed: Update not found"
fi

# Test 5: Archive todo
echo -e "\n6. Testing ARCHIVE..."
ARCHIVE_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 6,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"todo_archive\",
      \"arguments\": {
        \"id\": \"$TODO_ID\"
      }
    }
  }")

if echo "$ARCHIVE_RESPONSE" | jq -e '.result.content[0].text' > /dev/null 2>&1; then
    echo "✓ ARCHIVE works: Todo archived successfully"
    
    # Check if archived file exists
    ARCHIVE_PATH="$TEST_DIR/.claude/archive/$(date +%Y/%m/%d)/$TODO_ID.md"
    if [ -f "$ARCHIVE_PATH" ]; then
        echo "✓ Archive file exists at: $ARCHIVE_PATH"
    else
        echo "✗ Archive file not found at expected path"
    fi
else
    echo "✗ ARCHIVE failed: $(echo "$ARCHIVE_RESPONSE" | jq -r '.error.message // "Unknown error"')"
fi

# Test 6: Verify todo removed from active list
echo -e "\n7. Verifying todo removed from active list..."
VERIFY_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
      "name": "todo_read",
      "arguments": {
        "format": "list"
      }
    }
  }')

if echo "$VERIFY_RESPONSE" | jq -r '.result.content[0].text' | grep -q "$TODO_ID"; then
    echo "✗ Todo still in active list after archive"
else
    echo "✓ Todo correctly removed from active list"
fi

echo -e "\n=== Summary ==="
echo "Project directory: $TEST_DIR"
echo "Active todos: $(ls -la $TEST_DIR/.claude/todos/ 2>/dev/null | wc -l)"
echo "Archived todos: $(find $TEST_DIR/.claude/archive -name "*.md" 2>/dev/null | wc -l)"

echo -e "\nTest complete!"