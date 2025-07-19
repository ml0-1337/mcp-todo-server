# MCP Todo Server Documentation

This directory contains comprehensive documentation for the MCP Todo Server and its integration with Claude Code.

## 📁 Documentation Structure

### 📘 Overview
Core product documentation and high-level guides.

- **[Product Requirements Document](./overview/prd.md)** - Complete product specification
- **[MCP Todo Server Documentation](./overview/mcp-todo-server-documentation.md)** - System overview, architecture, features
- **[Real-World Use Scenario](./overview/use-scenario-example.md)** - Week-long developer story with examples

### 🔧 API Documentation
Technical reference and migration guides.

- **[API Reference](./api/api-reference.md)** - All 9 MCP tools with schemas and examples
- **[Migration Guide](./api/migration-guide.md)** - Step-by-step migration from native tools

### 📊 Analysis & Research
Comparative analysis and technical research.

- **[Native Todo Tools Analysis](./analysis/native-todo-tools-analysis.md)** - Deep dive into Claude's built-in tools
- **[Go MCP Server Research](./analysis/go-mcp-server-research.md)** - mark3labs/mcp-go implementation study

### 🎯 Tool Selection Framework
Complete framework for choosing between native and MCP tools.

- **[Selection Criteria](./selection/todo-tool-selection-criteria.md)** - Decision matrix with weighted scoring
- **[Use Cases](./selection/todo-tool-use-cases.md)** - 13 real-world scenarios
- **[Implementation Guide](./selection/todo-tool-implementation-guide.md)** - Technical implementation details

### 📘 User Guides
Practical guides for using and configuring the server.

- **[Transport Guide](./guides/transport-guide.md)** - STDIO vs HTTP transport modes
- **[HTTP Headers](./guides/http-headers.md)** - Working directory resolution via headers
- **[Connection Resilience](./guides/connection-resilience.md)** - Handling connection issues

### 🏗️ Development
Technical documentation for developers.

- **[Architecture](./development/architecture.md)** - System architecture and design
- **[Error Handling](./development/error-handling.md)** - Error handling patterns
- **[Error Messages](./development/error-messages.md)** - User-facing error messages
- **[Validation](./development/validation.md)** - Input validation strategies
- **[Test Utilities](./development/test-utilities.md)** - Testing infrastructure
- **[Stability Improvements](./development/stability-improvements.md)** - Performance and stability enhancements

### 🧪 Testing
- **[Testing Guide](./testing.md)** - Comprehensive testing documentation
- **[Section Metadata](./section-metadata.md)** - Todo section metadata documentation

## 🚀 Quick Start Guides

### For Different Audiences

| Audience | Start Here | Then Read |
|----------|------------|-----------|
| **New Users** | [Use Cases](./selection/todo-tool-use-cases.md) | [Selection Criteria](./selection/todo-tool-selection-criteria.md) |
| **Developers** | [API Reference](./api/api-reference.md) | [Implementation Guide](./selection/todo-tool-implementation-guide.md) |
| **Migrating Users** | [Migration Guide](./api/migration-guide.md) | [MCP Documentation](./overview/mcp-todo-server-documentation.md) |
| **Decision Makers** | [PRD](./overview/prd.md) | [Native Tools Analysis](./analysis/native-todo-tools-analysis.md) |

### By Task

- **"Which tool should I use?"** → [Selection Criteria](./selection/todo-tool-selection-criteria.md)
- **"Show me examples"** → [Use Cases](./selection/todo-tool-use-cases.md) or [Use Scenario](./overview/use-scenario-example.md)
- **"How do I implement this?"** → [API Reference](./api/api-reference.md)
- **"How do I migrate?"** → [Migration Guide](./api/migration-guide.md)
- **"What are the differences?"** → [Native Tools Analysis](./analysis/native-todo-tools-analysis.md)

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
5. Use lowercase-with-hyphens.md for file names (except README.md)