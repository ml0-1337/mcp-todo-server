# MCP Todo Server Migration Guide

This guide helps you migrate from Claude's native todo tools to the MCP Todo Server while maintaining compatibility and taking advantage of enhanced features.

## Table of Contents

1. [Overview](#overview)
2. [Pre-Migration Checklist](#pre-migration-checklist)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Compatibility Mode](#compatibility-mode)
6. [Feature Migration](#feature-migration)
7. [Data Migration](#data-migration)
8. [Troubleshooting](#troubleshooting)
9. [Rollback Plan](#rollback-plan)

## Overview

The MCP Todo Server is designed as a drop-in replacement for Claude's native todo tools with these advantages:

- **100% API Compatible**: Existing workflows continue working
- **Persistent Storage**: Todos saved as files, survive session restarts
- **Enhanced Features**: Access advanced features gradually
- **Graceful Degradation**: Falls back safely if issues occur

### Migration Benefits

| Native Tools | MCP Server |
|--------------|------------|
| Session-only memory | Permanent file storage |
| 4 basic fields | 15+ metadata fields |
| No search | Full-text search |
| No templates | Template system |
| No analytics | Comprehensive stats |
| No relationships | Parent-child linking |

## Pre-Migration Checklist

Before migrating, ensure:

- [ ] Go 1.21 or higher installed
- [ ] Access to `~/.claude/` directory
- [ ] Claude Code with MCP support
- [ ] Backup of any important session todos
- [ ] Understanding of native tool limitations

## Installation

### Step 1: Clone or Download

```bash
# Option 1: Clone from repository
git clone https://github.com/user/mcp-todo-server
cd mcp-todo-server

# Option 2: Download and extract
wget https://github.com/user/mcp-todo-server/archive/main.zip
unzip main.zip
cd mcp-todo-server-main
```

### Step 2: Build the Server

```bash
# Install dependencies
go mod tidy

# Build the binary
go build -o mcp-todo-server

# Verify build
./mcp-todo-server --version
```

### Step 3: Create Directory Structure

```bash
# Create todo directories
mkdir -p ~/.claude/todos
mkdir -p ~/.claude/archive
mkdir -p ~/.claude/templates

# Set permissions
chmod 755 ~/.claude/todos
```

## Configuration

### Claude Code Settings

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "todo-server": {
      "command": "/path/to/mcp-todo-server",
      "args": [],
      "env": {
        "CLAUDE_BASE_PATH": "$HOME/.claude"
      }
    }
  }
}
```

### Environment Variables

Optional configuration:

```bash
# Custom base path
export CLAUDE_BASE_PATH=/custom/path

# Debug logging
export MCP_DEBUG=true

# Custom index location
export MCP_TODO_INDEX=/path/to/index
```

### Verify Installation

Test the server:

```bash
# Start server manually
./mcp-todo-server

# In another terminal, check logs
tail -f ~/.claude/logs/mcp-todo-server.log
```

## Compatibility Mode

The MCP Server maintains full compatibility with native tools:

### Native Tool Format

```json
// Native TodoWrite format - still works!
{
  "todos": [
    {
      "content": "Implement feature",
      "status": "in_progress",
      "priority": "high",
      "id": "implement-feature"
    }
  ]
}
```

### Enhanced Format

The server automatically enhances todos with metadata:

```markdown
---
todo_id: implement-feature
started: 2025-01-27 14:30:00
completed: ""
status: in_progress
priority: high
type: feature
---

# Task: Implement feature

## Findings & Research
## Test Strategy
## Test List
## Checklist
## Working Scratchpad
```

### Transparent Enhancement

1. **First Write**: Creates basic todo
2. **First Update**: Adds full structure
3. **Preservation**: Never loses native fields
4. **Backward Compatible**: Native tools can still read

## Feature Migration

### Phase 1: Drop-in Replacement (Day 1)

Use exactly like native tools:

```javascript
// Continue using TodoWrite as before
await TodoWrite({
  todos: [
    {
      content: "Fix bug",
      status: "pending",
      priority: "high",
      id: "fix-bug"
    }
  ]
});

// TodoRead works identically
const todos = await TodoRead();
```

### Phase 2: Persistent Storage (Week 1)

Todos now persist between sessions:

```javascript
// Create todo in one session
await TodoWrite({
  todos: [{
    content: "Research API",
    status: "in_progress",
    priority: "medium",
    id: "research-api"
  }]
});

// Still there after restart!
const todos = await TodoRead();
// Returns: [{ id: "research-api", ... }]
```

### Phase 3: Enhanced Features (Week 2)

Start using advanced features:

```javascript
// Use todo_create for templates
await todo_create({
  task: "Fix authentication bug",
  template: "bug-fix",
  priority: "high"
});

// Search across all todos
const results = await todo_search({
  query: "authentication",
  scope: ["task", "findings"]
});

// Update specific sections
await todo_update({
  id: "fix-auth-bug",
  section: "tests",
  operation: "append",
  content: "Test 1: Login timeout after 30 seconds"
});
```

### Phase 4: Full Migration (Month 1)

Leverage all capabilities:

```javascript
// Create linked todos
await todo_link({
  parent_id: "implement-api",
  child_id: "implement-auth-endpoint"
});

// Get analytics
const stats = await todo_stats({
  period: "month"
});

// Archive old todos
await todo_clean({
  operation: "archive_old",
  days: 30
});
```

## Data Migration

### Migrating Existing Session Todos

If you have todos in an active session:

1. **Export Current Todos**:
```javascript
// In native session
const todos = await TodoRead();
console.log(JSON.stringify(todos));
```

2. **Import to MCP Server**:
```javascript
// With MCP Server
for (const todo of todos) {
  await todo_create({
    task: todo.content,
    priority: todo.priority,
    type: "migrated"
  });
}
```

### Bulk Import Script

Create `migrate-todos.js`:

```javascript
const fs = require('fs');
const { execSync } = require('child_process');

// Read exported todos
const todos = JSON.parse(fs.readFileSync('exported-todos.json'));

// Import each todo
todos.forEach(todo => {
  const cmd = `mcp-cli todo_create --task "${todo.content}" --priority ${todo.priority}`;
  execSync(cmd);
  console.log(`Imported: ${todo.id}`);
});
```

### Handling Large Migrations

For many todos:

1. **Batch Processing**: Import in groups of 50
2. **Verify Imports**: Check todo count matches
3. **Update IDs**: Map old IDs to new ones
4. **Test Search**: Ensure all todos are indexed

## Troubleshooting

### Common Issues

#### Server Not Starting

```bash
# Check logs
tail -f ~/.claude/logs/mcp-todo-server.log

# Verify permissions
ls -la ~/.claude/todos

# Test server directly
./mcp-todo-server --debug
```

#### Todos Not Persisting

```bash
# Check file creation
ls ~/.claude/todos/

# Verify write permissions
touch ~/.claude/todos/test.md

# Check environment
echo $CLAUDE_BASE_PATH
```

#### Search Not Working

```bash
# Check index exists
ls ~/.claude/index/todos.bleve/

# Rebuild index
rm -rf ~/.claude/index/todos.bleve/
# Restart server to rebuild
```

### Debug Mode

Enable detailed logging:

```json
{
  "mcpServers": {
    "todo-server": {
      "command": "/path/to/mcp-todo-server",
      "args": ["--debug"],
      "env": {
        "MCP_DEBUG": "true"
      }
    }
  }
}
```

### Performance Issues

If experiencing slow responses:

1. **Archive Old Todos**: Use `todo_clean`
2. **Rebuild Index**: Delete and recreate
3. **Check File Count**: Limit active todos
4. **Monitor Resources**: Check CPU/memory

## Rollback Plan

If you need to revert to native tools:

### Step 1: Disable MCP Server

Remove from `settings.json`:

```json
{
  "mcpServers": {
    // Remove or comment out todo-server
  }
}
```

### Step 2: Export Critical Todos

```bash
# Export current todos
mcp-cli todo_read --format full > todos-backup.txt

# Or as JSON
mcp-cli todo_read --format list > todos-backup.json
```

### Step 3: Restart Claude Code

Native tools will resume working immediately.

### Step 4: Re-import if Needed

If returning to MCP Server later:
- Todos remain in `~/.claude/todos/`
- Search index rebuilds automatically
- No data loss occurs

## Best Practices

### Gradual Adoption

1. **Week 1**: Use as drop-in replacement
2. **Week 2**: Try search and updates
3. **Week 3**: Implement templates
4. **Month 2**: Full feature adoption

### Backup Strategy

```bash
# Daily backup
tar -czf claude-todos-$(date +%Y%m%d).tar.gz ~/.claude/todos/

# Before major changes
cp -r ~/.claude/todos ~/.claude/todos.backup
```

### Monitoring

Track adoption success:

```javascript
// Check todo count
const stats = await todo_stats();
console.log(`Total todos: ${stats.total_todos}`);
console.log(`Completion rate: ${stats.completion_rate}%`);
```

## Migration Timeline

### Day 1
- Install MCP Server
- Configure settings.json
- Test basic operations

### Week 1
- Migrate active todos
- Start using persistence
- Enable search

### Week 2-4
- Adopt templates
- Implement linking
- Use analytics

### Month 2+
- Full feature utilization
- Customize templates
- Optimize workflow

## Support Resources

### Documentation
- [API Reference](./api_reference.md)
- [Main Documentation](./mcp_todo_server_documentation.md)
- [Native Tools Analysis](./native_todo_tools_analysis.md)

### Community
- GitHub Issues: Report problems
- Discussions: Share workflows
- Wiki: User contributions

### Troubleshooting Checklist

- [ ] Server running? Check `ps aux | grep mcp-todo`
- [ ] Correct path? Verify `$CLAUDE_BASE_PATH`
- [ ] Files created? Check `~/.claude/todos/`
- [ ] Index working? Test with `todo_search`
- [ ] Permissions OK? Run `ls -la ~/.claude/`

## Conclusion

The MCP Todo Server provides a seamless migration path from native tools with:

1. **Zero Breaking Changes**: Existing code continues working
2. **Gradual Enhancement**: Adopt features at your pace
3. **Safe Rollback**: Can revert anytime without data loss
4. **Clear Benefits**: Persistence, search, analytics

Start with compatibility mode and gradually adopt enhanced features as your workflow evolves.