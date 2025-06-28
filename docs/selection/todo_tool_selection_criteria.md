# Todo Tool Selection Criteria

This document defines the evaluation criteria and decision matrix for selecting between Claude's native todo tools and the MCP Todo Server. The framework enables intelligent, context-aware tool selection while maintaining user override capabilities.

## Executive Summary

Claude Code can use either native todo tools (session-based, simple) or MCP Todo Server (persistent, feature-rich). This framework provides a weighted scoring system to automatically select the appropriate tool based on task requirements.

**Quick Decision**: 
- Score < 40: Use Native Todo Tools
- Score ≥ 40: Use MCP Todo Server
- User explicitly requests: Override automatic selection

## Decision Matrix

### Evaluation Categories and Weights

| Category | Weight | Description |
|----------|--------|-------------|
| Session Persistence | 25% | Need for data to survive session restarts |
| Task Complexity | 20% | Project structure and workflow requirements |
| Feature Requirements | 20% | Advanced capabilities needed |
| Data Volume | 15% | Number of todos and update frequency |
| Performance Needs | 10% | Response time and resource constraints |
| Integration Requirements | 10% | External tool and system integration |

### Scoring Guidelines

Each criterion is scored on a 0-10 scale:
- 0-3: Low requirement (favors native tools)
- 4-6: Medium requirement (neutral)
- 7-10: High requirement (favors MCP server)

## Detailed Criteria

### 1. Session Persistence (Weight: 25%)

**Scoring Factors:**

| Score | Criteria | Example |
|-------|----------|---------|
| 0-1 | Single session only, no persistence needed | Quick syntax check |
| 2-3 | Same-day completion expected | Minor bug fix |
| 4-5 | Might span 2-3 sessions | Small feature |
| 6-7 | Multi-day project | API implementation |
| 8-9 | Week-long development | Major feature |
| 10 | Long-term project (weeks/months) | Architecture redesign |

**Key Questions:**
- Will this task span multiple Claude sessions?
- Do I need to reference this todo tomorrow?
- Is this part of a larger project timeline?

### 2. Task Complexity (Weight: 20%)

**Scoring Factors:**

| Score | Criteria | Example |
|-------|----------|---------|
| 0-1 | Single, atomic task | Fix typo |
| 2-3 | 2-3 related subtasks | Update config |
| 4-5 | Multiple independent tasks | Refactor module |
| 6-7 | Requires task dependencies | Feature with tests |
| 8-9 | Multi-phase project | Full TDD cycle |
| 10 | Complex hierarchical project | Microservices migration |

**Key Questions:**
- Does this need parent-child relationships?
- Will I follow TDD/RGRC workflow?
- Are there multiple phases or milestones?

### 3. Feature Requirements (Weight: 20%)

**Scoring Factors:**

| Feature | Points | Description |
|---------|--------|-------------|
| Search needed | +3 | Need to search across todos |
| Templates useful | +2 | Repeated task patterns |
| Analytics wanted | +2 | Progress tracking |
| Archive required | +2 | Historical reference |
| Sections needed | +1 | Structured content |

**Scoring Scale:**
- 0-2 points: Score 0-3
- 3-4 points: Score 4-6  
- 5-7 points: Score 7-8
- 8+ points: Score 9-10

### 4. Data Volume (Weight: 15%)

**Scoring Factors:**

| Score | Todo Count | Update Frequency | Example |
|-------|------------|------------------|---------|
| 0-1 | 1-2 todos | Rarely updated | Single task |
| 2-3 | 3-5 todos | Daily updates | Small project |
| 4-5 | 6-10 todos | Multiple daily | Feature work |
| 6-7 | 11-25 todos | Frequent updates | Sprint work |
| 8-9 | 26-50 todos | Continuous updates | Major project |
| 10 | 50+ todos | High frequency | Product development |

### 5. Performance Needs (Weight: 10%)

**Scoring Factors:**

| Score | Requirement | Use Case |
|-------|-------------|----------|
| 0-3 | Standard performance acceptable | Normal development |
| 4-6 | Some operations time-sensitive | Live debugging |
| 7-10 | Critical real-time needs | Production fixes |

**Note**: Native tools have faster response for simple operations. MCP server adds ~10-50ms overhead but provides caching for repeated access.

### 6. Integration Requirements (Weight: 10%)

**Scoring Factors:**

| Score | Integration Needs | Example |
|-------|-------------------|---------|
| 0-1 | No external integration | Standalone task |
| 2-3 | Basic file references | Code review |
| 4-5 | Version control awareness | Git workflow |
| 6-7 | CI/CD pipeline integration | Build automation |
| 8-9 | Multi-tool workflow | IDE + CLI + Web |
| 10 | Full ecosystem integration | Enterprise toolchain |

## Selection Algorithm

```python
def select_todo_tool(criteria_scores, user_preference=None):
    """
    Select appropriate todo tool based on weighted criteria.
    
    Args:
        criteria_scores: Dict with scores for each category (0-10)
        user_preference: Optional "native" or "mcp" override
    
    Returns:
        "native" or "mcp"
    """
    # Honor user preference
    if user_preference in ["native", "mcp"]:
        return user_preference
    
    # Define weights
    weights = {
        "session_persistence": 0.25,
        "task_complexity": 0.20,
        "feature_requirements": 0.20,
        "data_volume": 0.15,
        "performance_needs": 0.10,
        "integration_requirements": 0.10
    }
    
    # Calculate weighted score
    total_score = 0
    for category, weight in weights.items():
        score = criteria_scores.get(category, 0)
        total_score += score * weight * 10  # Scale to 100
    
    # Decision threshold
    if total_score < 40:
        return "native"
    else:
        return "mcp"
```

## Quick Reference Card

### Use Native Tools When:
- ✓ Single session task
- ✓ < 5 todos needed
- ✓ No search required
- ✓ Simple status tracking
- ✓ Quick fixes or checks
- ✓ Resource constrained

### Use MCP Server When:
- ✓ Multi-session project
- ✓ TDD/RGRC workflow
- ✓ Need search capability
- ✓ Want templates
- ✓ Tracking many todos
- ✓ Long-term reference

## Scoring Examples

### Example 1: Quick Bug Fix
```yaml
Task: "Fix typo in README"
Scores:
  session_persistence: 1    # Single session
  task_complexity: 1        # Simple task
  feature_requirements: 0   # No features needed
  data_volume: 1           # One todo
  performance_needs: 2     # Standard
  integration_requirements: 2  # Basic git

Weighted Score: 11.5
Decision: Use Native Tools ✓
```

### Example 2: Feature Implementation
```yaml
Task: "Implement user authentication with TDD"
Scores:
  session_persistence: 8    # Multi-day project
  task_complexity: 8        # TDD cycle, multiple tests
  feature_requirements: 7   # Templates, sections, search
  data_volume: 6           # 10-20 todos expected
  performance_needs: 3     # Standard development
  integration_requirements: 5  # Git, testing framework

Weighted Score: 68.5
Decision: Use MCP Server ✓
```

### Example 3: Research Task
```yaml
Task: "Research API options for payment processing"
Scores:
  session_persistence: 4    # Might span sessions
  task_complexity: 3        # Single research task
  feature_requirements: 5   # Want sections for findings
  data_volume: 2           # Few todos
  performance_needs: 2     # Not time critical
  integration_requirements: 1  # Minimal

Weighted Score: 35.5
Decision: Use Native Tools ✓ (close to threshold)
```

## User Override Syntax

Users can explicitly request a specific tool:

```bash
# Force native tools
"Create a todo using native tools for..."
"Use simple todo for..."
"Quick todo: ..."

# Force MCP server
"Create a persistent todo for..."
"Use MCP todo for..."
"Create a todo file for..."
```

## Migration Triggers

Automatic migration from native to MCP should be suggested when:

1. **Session Boundary**: User returns to incomplete native todos
2. **Complexity Growth**: Simple task evolves into multi-step project
3. **Search Needed**: User asks "Which todo was about...?"
4. **Feature Request**: User wants templates or analytics
5. **Volume Threshold**: More than 10 active todos

## Performance Considerations

### Native Tools Performance
- Create/Read/Update: < 5ms
- No disk I/O
- No search overhead
- Memory only

### MCP Server Performance
- Create: ~10ms (file write)
- Read: ~5ms (cached) / ~15ms (disk)
- Update: ~10ms (file write)
- Search: ~50ms (indexed)
- Stats: ~100ms (aggregation)

### Optimization Tips
1. Use native for high-frequency temporary todos
2. Use MCP for todos needing search or persistence
3. Batch operations when using MCP
4. Cache recent todos in memory

## Implementation Notes

1. **Default Behavior**: Calculate score automatically
2. **Transparency**: Show score and decision reasoning
3. **Flexibility**: Always allow user override
4. **Learning**: Track override patterns for improvement
5. **Fallback**: If MCP unavailable, use native tools

## Conclusion

This framework provides a systematic approach to todo tool selection that:
- Maximizes efficiency by choosing the right tool
- Preserves simplicity for basic tasks
- Enables power features when needed
- Respects user preferences
- Supports gradual migration

The weighted scoring system ensures that tool selection aligns with actual task requirements while maintaining the flexibility to adapt to user preferences and changing project needs.