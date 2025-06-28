# Tool Selection Framework

This directory contains the complete framework for choosing between native todo tools and the MCP Todo Server.

## Contents

- **[todo_tool_selection_criteria.md](./todo_tool_selection_criteria.md)** - Selection Criteria
  - Decision matrix with 6 weighted categories
  - Scoring methodology (0-100 scale)
  - Selection algorithm
  - Real-world scoring examples
  - Quick reference card

- **[todo_tool_use_cases.md](./todo_tool_use_cases.md)** - Use Cases & Scenarios
  - 13 real-world scenarios
  - Native tool scenarios
  - MCP server scenarios
  - Hybrid usage patterns
  - Migration triggers
  - Edge cases

- **[todo_tool_implementation_guide.md](./todo_tool_implementation_guide.md)** - Implementation Guide
  - Technical architecture
  - Automatic selection logic
  - Request analysis algorithms
  - User override detection
  - Fallback strategies
  - Performance optimization
  - Monitoring and adaptation

## Quick Decision Guide

### Use Native Tools When:
- Single session task (Score: 0-20)
- < 5 todos needed
- No search required
- Simple status tracking
- Quick fixes or checks

### Use MCP Server When:
- Multi-session project (Score: 60+)
- TDD/RGRC workflow
- Need search capability
- Want templates
- Tracking many todos

### Decision Threshold:
- **Score < 40**: Use Native Tools
- **Score ≥ 40**: Use MCP Server
- **User Override**: Always respected

## Implementation Phases

1. **Phase 1**: Basic selection logic
2. **Phase 2**: Performance optimization
3. **Phase 3**: Advanced features
4. **Phase 4**: Production ready

## Quick Links

- For API details → [API Documentation](../api/)
- For comparisons → [Analysis](../analysis/)
- For examples → [Overview](../overview/)