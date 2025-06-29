# Section Metadata Feature Guide

The MCP Todo Server now supports rich section metadata, allowing you to define custom sections with validation schemas, ordering, and requirements. This guide explains how to use these powerful new features.

## Overview

Todos can now have structured sections defined in their YAML frontmatter. Each section can have:
- A unique key identifier
- Custom title
- Display order
- Validation schema
- Required/optional status
- Custom metadata

## Section Schemas

The following section schemas are available:

### 1. **research** - Free-form research notes
- Accepts any text content
- Tracks word count metric
- No validation constraints

### 2. **checklist** - Checkbox task lists
- Validates checkbox syntax: `- [ ]` or `- [x]`
- Tracks completion metrics (completed/total)
- Rejects invalid checkbox formats

### 3. **test_cases** - Code test cases
- Requires code blocks (triple backticks)
- Counts number of code blocks
- Supports multiple programming languages

### 4. **results** - Timestamped log entries
- Requires entries to start with `[timestamp]`
- Tracks number of entries
- Example: `[2025-01-29 10:00:00] Test passed`

### 5. **strategy** - Structured planning
- Accepts structured content with subsections
- Counts `###` subsections
- No specific validation

### 6. **freeform** - Unstructured content
- Accepts any content
- Tracks content length
- Default for custom sections

## Defining Sections

Sections are defined in the YAML frontmatter:

```yaml
---
todo_id: implement-feature
status: in_progress
sections:
  findings:
    title: "## Findings & Research"
    order: 1
    schema: research
    required: true
  
  test_list:
    title: "## Test List"
    order: 2
    schema: checklist
    required: true
    metadata:
      min_items: 3
  
  implementation:
    title: "## Implementation Notes"
    order: 3
    schema: freeform
    custom: true
---
```

## API Usage

### Discover Sections
```json
{
  "tool": "mcp__todo__todo_sections",
  "parameters": {
    "id": "todo-id"
  }
}
```

Returns all sections with their metadata, including content status:
```
findings:
  title: ## Findings & Research
  order: 1
  schema: research
  required: true
  hasContent: true
  contentLength: 245
```

### Add Custom Section
```json
{
  "tool": "mcp__todo__todo_add_section",
  "parameters": {
    "id": "todo-id",
    "key": "risks",
    "title": "## Risks & Mitigations",
    "schema": "freeform",
    "order": 5,
    "required": false
  }
}
```

### Reorder Sections
```json
{
  "tool": "mcp__todo__todo_reorder_sections",
  "parameters": {
    "id": "todo-id",
    "orders": {
      "findings": 3,
      "test_list": 1,
      "implementation": 2
    }
  }
}
```

### Update Section with Validation
When updating sections, content is validated against the schema:

```json
{
  "tool": "mcp__todo__todo_update",
  "parameters": {
    "id": "todo-id",
    "section": "checklist",
    "operation": "replace",
    "content": "- [x] Task 1\n- [ ] Task 2"
  }
}
```

Invalid content will be rejected with an error message.

## Backwards Compatibility

Legacy todos without section metadata continue to work:
- Sections are inferred from `##` headings in markdown
- Standard sections are mapped to appropriate schemas
- Custom sections default to freeform schema

## Examples

### Creating a Todo with Custom Sections

```yaml
---
todo_id: api-design
sections:
  requirements:
    title: "## Requirements"
    order: 1
    schema: checklist
    required: true
  
  api_spec:
    title: "## API Specification"
    order: 2
    schema: freeform
    required: true
  
  test_scenarios:
    title: "## Test Scenarios"
    order: 3
    schema: test_cases
  
  performance:
    title: "## Performance Metrics"
    order: 4
    schema: results
---

# Task: Design REST API for user management

## Requirements

- [ ] Support CRUD operations
- [ ] Include authentication
- [ ] Rate limiting

## API Specification

POST /api/users
GET /api/users/:id
PUT /api/users/:id
DELETE /api/users/:id

## Test Scenarios

```javascript
test('create user', async () => {
  const user = await api.createUser({
    name: 'Test User',
    email: 'test@example.com'
  });
  expect(user.id).toBeDefined();
});
```

## Performance Metrics

[2025-01-29 14:30:00] Baseline: 50ms response time
[2025-01-29 15:00:00] After optimization: 30ms response time
```

### Migrating Legacy Todos

Legacy todos are automatically compatible. To add section metadata:

1. Use `mcp__todo__todo_sections` to discover existing sections
2. Add metadata using `mcp__todo__todo_add_section` for custom sections
3. The system preserves all existing content

## Best Practices

1. **Define sections upfront** - Include section definitions when creating todos
2. **Use appropriate schemas** - Choose schemas that match your content type
3. **Mark critical sections as required** - Ensure important sections aren't omitted
4. **Order sections logically** - Use the order field to arrange sections
5. **Validate before updates** - Schema validation prevents malformed content

## Troubleshooting

### "Section key already exists"
Each section must have a unique key. Use different keys for different sections.

### "Invalid schema type"
Only the predefined schemas (research, checklist, test_cases, results, strategy, freeform) are supported.

### "Validation failed"
Check that your content matches the schema requirements:
- Checklist: Use `- [ ]` or `- [x]` format
- Test cases: Include code blocks with triple backticks
- Results: Start entries with `[timestamp]`

### Legacy todos show "No sections found"
This happens when a legacy todo has no recognizable `##` headings. Add section headings to enable section discovery.