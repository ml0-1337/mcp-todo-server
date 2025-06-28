# MCP Todo Server Documentation

This directory contains comprehensive documentation for the MCP Todo Server and its integration with Claude Code.

## 📁 Documentation Structure

### 📘 Overview
Core product documentation and high-level guides.

- **[Product Requirements Document](./overview/PRD.md)** - Complete product specification
- **[MCP Todo Server Documentation](./overview/mcp_todo_server_documentation.md)** - System overview, architecture, features
- **[Real-World Use Scenario](./overview/use_scenario_example.md)** - Week-long developer story with examples

### 🔧 API Documentation
Technical reference and migration guides.

- **[API Reference](./api/api_reference.md)** - All 9 MCP tools with schemas and examples
- **[Migration Guide](./api/migration_guide.md)** - Step-by-step migration from native tools

### 📊 Analysis & Research
Comparative analysis and technical research.

- **[Native Todo Tools Analysis](./analysis/native_todo_tools_analysis.md)** - Deep dive into Claude's built-in tools
- **[Go MCP Server Research](./analysis/go_mcp_server_research.md)** - mark3labs/mcp-go implementation study

### 🎯 Tool Selection Framework
Complete framework for choosing between native and MCP tools.

- **[Selection Criteria](./selection/todo_tool_selection_criteria.md)** - Decision matrix with weighted scoring
- **[Use Cases](./selection/todo_tool_use_cases.md)** - 13 real-world scenarios
- **[Implementation Guide](./selection/todo_tool_implementation_guide.md)** - Technical implementation details

## 🚀 Quick Start Guides

### For Different Audiences

| Audience | Start Here | Then Read |
|----------|------------|-----------|
| **New Users** | [Use Cases](./selection/todo_tool_use_cases.md) | [Selection Criteria](./selection/todo_tool_selection_criteria.md) |
| **Developers** | [API Reference](./api/api_reference.md) | [Implementation Guide](./selection/todo_tool_implementation_guide.md) |
| **Migrating Users** | [Migration Guide](./api/migration_guide.md) | [MCP Documentation](./overview/mcp_todo_server_documentation.md) |
| **Decision Makers** | [PRD](./overview/PRD.md) | [Native Tools Analysis](./analysis/native_todo_tools_analysis.md) |

### By Task

- **"Which tool should I use?"** → [Selection Criteria](./selection/todo_tool_selection_criteria.md)
- **"Show me examples"** → [Use Cases](./selection/todo_tool_use_cases.md) or [Use Scenario](./overview/use_scenario_example.md)
- **"How do I implement this?"** → [API Reference](./api/api_reference.md)
- **"How do I migrate?"** → [Migration Guide](./api/migration_guide.md)
- **"What are the differences?"** → [Native Tools Analysis](./analysis/native_todo_tools_analysis.md)

## Key Concepts

### Native Todo Tools
- Session-based, in-memory storage
- Simple 4-field structure (content, status, priority, id)
- Fast for quick tasks
- No persistence between sessions

### MCP Todo Server
- File-based persistent storage
- 15+ metadata fields with YAML frontmatter
- Full-text search with Bleve
- Templates, analytics, and archival
- Parent-child relationships

### Selection Framework
The new selection framework helps Claude Code automatically choose the right tool:

1. **Automatic Selection**: Based on 6 weighted criteria
2. **User Override**: Explicit requests always honored
3. **Hybrid Usage**: Use both tools for different phases
4. **Smart Migration**: Automatic suggestions when tasks evolve

### Decision Threshold
- **Score < 40**: Use Native Todo Tools
- **Score ≥ 40**: Use MCP Todo Server
- **User Override**: Always respected regardless of score

## Document Relationships

```
Selection Criteria
    ↓
Use Cases → Implementation Guide
    ↓
API Reference ← Migration Guide
    ↓
MCP Documentation
```

## Version History

- v1.0.0: Initial documentation (native tools analysis, MCP implementation)
- v1.1.0: Added comprehensive tool selection framework
- v1.2.0: Enhanced with real-world scenarios and implementation guide

## Contributing

To improve documentation:
1. Follow existing document structure
2. Include real examples
3. Update version history
4. Cross-reference related documents