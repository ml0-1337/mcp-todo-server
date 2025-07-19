# Todo Tool Use Cases

This document provides real-world scenarios for using native todo tools, MCP Todo Server, and hybrid approaches. Each scenario includes context, tool selection rationale, and example workflows.

## Table of Contents

1. [Native Todo Tool Scenarios](#native-todo-tool-scenarios)
2. [MCP Todo Server Scenarios](#mcp-todo-server-scenarios)
3. [Hybrid Usage Scenarios](#hybrid-usage-scenarios)
4. [Migration Scenarios](#migration-scenarios)
5. [Edge Cases](#edge-cases)

## Native Todo Tool Scenarios

### Scenario 1: Quick Code Review

**Context**: Reviewing a pull request with multiple small fixes needed.

**Tool Selection**: Native (Score: 18/100)
- Session persistence: 1 (same session)
- Task complexity: 2 (simple fixes)
- Features: 0 (basic tracking)

**Workflow**:
```javascript
// Create review todos
TodoWrite({
  todos: [
    { id: "fix-typo-line-42", content: "Fix typo on line 42", status: "pending", priority: "low" },
    { id: "add-error-handling", content: "Add error handling to API call", status: "pending", priority: "medium" },
    { id: "update-test", content: "Update test for new behavior", status: "pending", priority: "high" }
  ]
});

// Mark complete as you fix
TodoWrite({
  todos: [
    { id: "fix-typo-line-42", content: "Fix typo on line 42", status: "completed", priority: "low" },
    // ... other todos
  ]
});
```

### Scenario 2: Debugging Session

**Context**: Tracking multiple hypotheses while debugging a production issue.

**Tool Selection**: Native (Score: 25/100)
- Session persistence: 2 (urgent, same day)
- Task complexity: 3 (related tasks)
- Performance: 8 (time critical)

**Workflow**:
```javascript
// Track debugging steps
TodoWrite({
  todos: [
    { id: "check-logs", content: "Check error logs for pattern", status: "completed", priority: "high" },
    { id: "test-edge-case", content: "Test with empty input", status: "in_progress", priority: "high" },
    { id: "verify-database", content: "Verify database connection", status: "pending", priority: "high" }
  ]
});
```

### Scenario 3: Meeting Action Items

**Context**: Capturing action items during a technical discussion.

**Tool Selection**: Native (Score: 22/100)
- Session persistence: 3 (follow up same day)
- Task complexity: 1 (independent items)
- Data volume: 3 (5-8 items)

**Workflow**:
```javascript
// During meeting
TodoWrite({
  todos: [
    { id: "research-caching", content: "Research Redis vs Memcached", status: "pending", priority: "medium" },
    { id: "update-arch-diagram", content: "Update architecture diagram", status: "pending", priority: "low" },
    { id: "benchmark-api", content: "Benchmark API performance", status: "pending", priority: "high" }
  ]
});
```

## MCP Todo Server Scenarios

### Scenario 4: TDD Feature Development

**Context**: Implementing a new feature using Test-Driven Development.

**Tool Selection**: MCP Server (Score: 72/100)
- Session persistence: 8 (multi-day)
- Task complexity: 9 (TDD cycle)
- Features: 8 (templates, sections)

**Workflow**:
```javascript
// Create with template
await todo_create({
  task: "Implement user authentication",
  template: "tdd-feature",
  priority: "high"
});

// Update with test results
await todo_update({
  id: "implement-user-authentication",
  section: "tests",
  operation: "append",
  content: `
## Test 1: User can login with valid credentials
[2025-01-28 10:00:00] Red Phase: Test failing - login() not implemented
[2025-01-28 10:15:00] Green Phase: Basic implementation passing
[2025-01-28 10:30:00] Refactor: Extracted validation logic
`
});

// Track current test
await todo_update({
  id: "implement-user-authentication",
  metadata: {
    current_test: "Test 2: Invalid password returns error"
  }
});
```

### Scenario 5: Multi-Phase Project

**Context**: Large refactoring project with multiple phases.

**Tool Selection**: MCP Server (Score: 85/100)
- Session persistence: 10 (weeks-long)
- Task complexity: 10 (hierarchical)
- Data volume: 8 (many todos)

**Workflow**:
```javascript
// Create parent project
await todo_create({
  task: "Migrate to microservices architecture",
  type: "multi-phase",
  priority: "high"
});

// Create phase todos
await todo_create({
  task: "Phase 1: Extract user service",
  parent_id: "migrate-to-microservices-architecture",
  priority: "high"
});

await todo_create({
  task: "Phase 2: Extract payment service",
  parent_id: "migrate-to-microservices-architecture",
  priority: "high"
});

// Link related tasks
await todo_link({
  parent_id: "phase-1-extract-user-service",
  child_id: "design-user-service-api"
});
```

### Scenario 6: Bug Investigation with History

**Context**: Investigating a complex bug that requires maintaining investigation history.

**Tool Selection**: MCP Server (Score: 65/100)
- Session persistence: 7 (multi-day investigation)
- Features: 8 (search, sections)
- Integration: 6 (links to code)

**Workflow**:
```javascript
// Create investigation todo
await todo_create({
  task: "Investigate memory leak in search service",
  type: "bug",
  template: "bug-investigation"
});

// Document findings
await todo_update({
  id: "investigate-memory-leak-in-search-service",
  section: "findings",
  operation: "append",
  content: `
## Heap Analysis (2025-01-28)
- Heap grows by 50MB per hour
- SearchIndex objects not being garbage collected
- Reference held by global cache

## Code Review
- Found circular reference in SearchCache class
- Event listeners not being removed
`
});

// Search related issues
const related = await todo_search({
  query: "memory leak cache",
  scope: ["findings"]
});
```

## Hybrid Usage Scenarios

### Scenario 7: Sprint Planning to Execution

**Context**: Planning sprint tasks then executing them over two weeks.

**Tool Selection**: Hybrid approach
- Planning: Native tools (quick capture)
- Execution: MCP Server (persistence)

**Workflow**:
```javascript
// Day 1: Sprint planning with native tools
TodoWrite({
  todos: [
    { id: "api-endpoints", content: "Design REST API endpoints", status: "pending", priority: "high" },
    { id: "database-schema", content: "Create database schema", status: "pending", priority: "high" },
    { id: "auth-flow", content: "Implement auth flow", status: "pending", priority: "medium" }
  ]
});

// Day 2: Migrate to MCP for execution
for (const todo of todos) {
  await todo_create({
    task: todo.content,
    priority: todo.priority,
    template: "sprint-task"
  });
}

// Continue with MCP features
await todo_update({
  id: "design-rest-api-endpoints",
  section: "checklist",
  content: `
- [ ] Define user endpoints
- [ ] Define product endpoints
- [ ] Create OpenAPI spec
- [ ] Review with team
`
});
```

### Scenario 8: Daily Standup Tracking

**Context**: Track daily tasks with occasional deep dives.

**Tool Selection**: Hybrid
- Daily tasks: Native (ephemeral)
- Deep work: MCP (persistent)

**Workflow**:
```javascript
// Morning: Quick daily tasks
TodoWrite({
  todos: [
    { id: "standup-updates", content: "Prepare standup updates", status: "completed", priority: "medium" },
    { id: "review-prs", content: "Review 3 PRs", status: "in_progress", priority: "high" },
    { id: "fix-ci", content: "Fix failing CI build", status: "pending", priority: "high" }
  ]
});

// Discover CI fix is complex, switch to MCP
await todo_create({
  task: "Fix CI build - investigate flaky tests",
  template: "investigation",
  priority: "high"
});

// Document investigation
await todo_update({
  id: "fix-ci-build-investigate-flaky-tests",
  section: "findings",
  content: "Race condition in test teardown..."
});
```

## Migration Scenarios

### Scenario 9: Simple to Complex Evolution

**Context**: Task grows from simple fix to major refactor.

**Migration Trigger**: Complexity increase

**Workflow**:
```javascript
// Start: Simple native todo
TodoWrite({
  todos: [
    { id: "fix-date-format", content: "Fix date formatting bug", status: "in_progress", priority: "medium" }
  ]
});

// Discover it needs refactoring
// TRIGGER: Migrate to MCP
await todo_create({
  task: "Refactor date handling system",
  template: "refactor",
  priority: "high"
});

// Transfer context
await todo_update({
  id: "refactor-date-handling-system",
  section: "findings",
  content: `
## Original Issue
- Date formatting bug in user profile
- Discovered: Inconsistent date handling across app
- 15+ locations using different formats
`
});
```

### Scenario 10: Session Boundary Migration

**Context**: Returning to incomplete work from yesterday.

**Migration Trigger**: Session restart with pending todos

**Workflow**:
```javascript
// Day 1: Native todos (forgot to complete)
TodoWrite({
  todos: [
    { id: "implement-cache", content: "Implement Redis cache", status: "in_progress", priority: "high" }
  ]
});

// Day 2: Return to work
// TRIGGER: Suggest migration
console.log("Found incomplete todos from previous session. Migrate to persistent storage?");

// Migrate with context preservation
await todo_create({
  task: "Implement Redis cache",
  priority: "high"
});

await todo_update({
  id: "implement-redis-cache",
  section: "scratchpad",
  content: "Yesterday: Researched Redis vs Memcached, decided on Redis"
});
```

## Edge Cases

### Scenario 11: High-Volume Quick Tasks

**Context**: Code review with 50+ small issues.

**Decision**: Split approach
- Overview: MCP Server (searchable)
- Execution: Native (performance)

```javascript
// Create overview in MCP
await todo_create({
  task: "Code review: PR #1234 - 50+ issues found",
  type: "review"
});

// Work through issues with native tools (faster)
TodoWrite({
  todos: currentBatch // Work in batches of 10
});

// Update overview
await todo_update({
  id: "code-review-pr-1234",
  section: "checklist",
  content: "- [x] Fixed formatting issues (1-20)\n- [ ] Address logic issues (21-35)"
});
```

### Scenario 12: Emergency Production Fix

**Context**: Critical bug in production needs immediate fix.

**Decision**: Native tools (performance critical)

```javascript
// Use native for speed
TodoWrite({
  todos: [
    { id: "stop-bleeding", content: "Deploy hotfix to stop data loss", status: "completed", priority: "high" },
    { id: "investigate-cause", content: "Find root cause", status: "in_progress", priority: "high" },
    { id: "permanent-fix", content: "Implement permanent solution", status: "pending", priority: "high" }
  ]
});

// After emergency: Document in MCP
await todo_create({
  task: "Post-mortem: Production data loss incident",
  template: "incident-report"
});
```

### Scenario 13: Learning and Experimentation

**Context**: Learning new technology with many small experiments.

**Decision**: Native tools (disposable tasks)

```javascript
// Experimentation todos
TodoWrite({
  todos: [
    { id: "try-webpack", content: "Try webpack config", status: "completed", priority: "low" },
    { id: "test-vite", content: "Test Vite performance", status: "in_progress", priority: "low" },
    { id: "compare-results", content: "Compare build times", status: "pending", priority: "low" }
  ]
});

// If deciding to adopt: Create MCP todo
if (adopted) {
  await todo_create({
    task: "Migrate build system to Vite",
    template: "migration"
  });
}
```

## Best Practices by Scenario Type

### Short-Term Tasks (< 1 day)
- Use native tools by default
- Consider MCP if need search/templates
- Migrate if task extends beyond session

### Medium-Term Projects (1-7 days)
- Start with MCP Server
- Use templates for consistency
- Leverage search for context
- Archive completed phases

### Long-Term Projects (> 1 week)
- Always use MCP Server
- Create hierarchical structure
- Use parent-child relationships
- Regular archival of completed work
- Generate stats for progress tracking

### Mixed Workflows
- Use native for daily scratch work
- Use MCP for project todos
- Migrate when tasks evolve
- Keep overview todos in MCP

## Scenario Decision Matrix

| Scenario Type | Duration | Complexity | Features | Tool Choice |
|--------------|----------|------------|----------|-------------|
| Quick fixes | < 1 hour | Low | None | Native |
| Debugging | < 1 day | Medium | None | Native |
| Daily tasks | 1 day | Low | None | Native |
| Feature work | 2-5 days | High | Search, Templates | MCP |
| Projects | > 1 week | High | All | MCP |
| Research | Variable | Medium | Sections | MCP |
| Code review | < 1 day | Low | None | Native |
| TDD cycle | 3-7 days | High | RGRC tracking | MCP |
| Sprint work | 2 weeks | High | Analytics | MCP |
| Incidents | < 4 hours | Low | None | Native → MCP |

## Conclusion

The choice between native tools and MCP Server depends on:

1. **Task Duration**: Longer tasks benefit from persistence
2. **Complexity**: Complex workflows need advanced features
3. **Collaboration**: Shared work needs file-based storage
4. **Search Needs**: Many todos benefit from search
5. **Performance**: Critical tasks may need native speed

Most workflows benefit from a hybrid approach, using the right tool for each phase of work. The key is recognizing when to migrate from simple to advanced tools as tasks evolve.