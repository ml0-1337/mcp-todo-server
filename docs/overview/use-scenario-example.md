# MCP Todo Server - Real-World Use Scenario

## Developer Story: Building a REST API with Authentication

This document follows Sarah, a backend developer, as she builds a REST API project using the MCP Todo Server. We'll see how she progresses from basic todo management to leveraging advanced features, with actual example data at each step.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Day 1: Project Planning](#day-1-project-planning)
3. [Day 2-3: API Development](#day-2-3-api-development)
4. [Day 4: Bug Discovery](#day-4-bug-discovery)
5. [Day 5: Authentication Implementation](#day-5-authentication-implementation)
6. [Day 6: Testing and Optimization](#day-6-testing-and-optimization)
7. [Day 7: Project Completion](#day-7-project-completion)
8. [Analytics and Insights](#analytics-and-insights)

## Project Overview

**Project**: E-commerce REST API with JWT Authentication
**Timeline**: 1 week sprint
**Requirements**: 
- Product CRUD operations
- User authentication
- Search functionality
- Order management
- 80% test coverage

## Day 1: Project Planning

### 9:00 AM - Creating the Main Project Todo

Sarah starts by creating a multi-phase project todo:

```json
// API Call
{
  "tool": "todo_create",
  "parameters": {
    "task": "Build E-commerce REST API with Authentication",
    "priority": "high",
    "type": "multi-phase"
  }
}

// Response
{
  "id": "build-e-commerce-rest-api",
  "path": "/Users/sarah/.claude/todos/build-e-commerce-rest-api.md",
  "message": "Todo created successfully"
}
```

**File Created**: `/Users/sarah/.claude/todos/build-e-commerce-rest-api.md`
```markdown
---
todo_id: build-e-commerce-rest-api
started: 2025-01-27 09:00:00
completed: ""
status: in_progress
priority: high
type: multi-phase
parent_id: ""
tags: []
---

# Task: Build E-commerce REST API with Authentication

## Findings & Research

## Test Strategy

## Test List

## Test Cases

## Test Results Log

## Checklist

## Working Scratchpad
```

### 9:30 AM - Creating Phase Todos

Sarah breaks down the project into phases:

```json
// Phase 1: API Design
{
  "tool": "todo_create",
  "parameters": {
    "task": "Phase 1: Design REST API endpoints and data models",
    "priority": "high",
    "type": "feature",
    "parent_id": "build-e-commerce-rest-api"
  }
}

// Response
{
  "id": "phase-1-design-rest-api",
  "path": "/Users/sarah/.claude/todos/phase-1-design-rest-api.md",
  "message": "Todo created with parent: build-e-commerce-rest-api"
}
```

```json
// Phase 2: Core Implementation
{
  "tool": "todo_create",
  "parameters": {
    "task": "Phase 2: Implement product CRUD endpoints",
    "priority": "high",
    "type": "feature",
    "parent_id": "build-e-commerce-rest-api"
  }
}

// Phase 3: Authentication
{
  "tool": "todo_create",
  "parameters": {
    "task": "Phase 3: Add JWT authentication",
    "priority": "high",
    "type": "feature",
    "parent_id": "build-e-commerce-rest-api"
  }
}

// Phase 4: Testing
{
  "tool": "todo_create",
  "parameters": {
    "task": "Phase 4: Write tests and documentation",
    "priority": "medium",
    "type": "feature",
    "parent_id": "build-e-commerce-rest-api"
  }
}
```

### 10:00 AM - Research and Planning

Sarah updates the main todo with her research:

```json
{
  "tool": "todo_update",
  "parameters": {
    "id": "build-e-commerce-rest-api",
    "section": "findings",
    "operation": "replace",
    "content": "## API Framework Research\n\n- **Framework**: Express.js with TypeScript\n- **Database**: PostgreSQL with Prisma ORM\n- **Authentication**: JWT with refresh tokens\n- **Testing**: Jest + Supertest\n- **Documentation**: OpenAPI 3.0\n\n## Best Practices Found\n\n1. RESTful naming conventions\n2. Versioned endpoints (/api/v1/)\n3. Consistent error responses\n4. Rate limiting for security\n5. Request validation with Joi"
  }
}
```

### 11:00 AM - Checking Progress

```json
{
  "tool": "todo_read",
  "parameters": {
    "filter": {
      "status": "in_progress"
    },
    "format": "list"
  }
}

// Response
[
  {
    "id": "build-e-commerce-rest-api",
    "content": "Build E-commerce REST API with Authentication",
    "status": "in_progress",
    "priority": "high"
  },
  {
    "id": "phase-1-design-rest-api",
    "content": "Phase 1: Design REST API endpoints and data models",
    "status": "in_progress",
    "priority": "high"
  }
]
```

## Day 2-3: API Development

### Day 2 - 9:00 AM - Using Templates

Sarah uses a template for her TDD workflow:

```json
// First, list available templates
{
  "tool": "todo_template",
  "parameters": {
    "template": "list"
  }
}

// Response
{
  "templates": ["bug-fix", "feature", "tdd-cycle", "research"]
}

// Create feature with TDD template
{
  "tool": "todo_template",
  "parameters": {
    "template": "tdd-cycle",
    "task": "Implement product search with filters",
    "priority": "high"
  }
}
```

**File Created with Template**: `/Users/sarah/.claude/todos/implement-product-search.md`
```markdown
---
todo_id: implement-product-search
started: 2025-01-28 09:00:00
completed: ""
status: in_progress
priority: high
type: feature
parent_id: ""
tags: ["tdd", "search", "products"]
current_test: ""
---

# Task: Implement product search with filters

## Findings & Research

WebSearch: "elasticsearch vs postgresql full text search 2025"
- PostgreSQL FTS sufficient for <100k products
- Use GIN indexes for performance
- tsvector for search optimization

## Test Strategy

- **Test Framework**: Jest + Supertest
- **Test Types**: Unit tests for search logic, Integration tests for API
- **Coverage Target**: 90%
- **Edge Cases**: Empty search, special characters, SQL injection

## Test List

- [ ] Test 1: Search returns exact matches for product names
- [ ] Test 2: Search is case-insensitive
- [ ] Test 3: Search handles partial matches
- [ ] Test 4: Filter by price range works correctly
- [ ] Test 5: Filter by category returns only matching products
- [ ] Test 6: Combined search and filters work together
- [ ] Test 7: Empty search returns all products with pagination
- [ ] Test 8: Special characters are handled safely
- [ ] Test 9: Search response time under 200ms
- [ ] Test 10: Results are properly paginated

## Test Cases

## Test Results Log

## Checklist

- [ ] Design search endpoint schema
- [ ] Set up test environment
- [ ] Write failing tests (Red phase)
- [ ] Implement search logic (Green phase)
- [ ] Refactor for performance
- [ ] Add database indexes
- [ ] Document API endpoint

## Working Scratchpad
```

### Day 2 - 10:00 AM - TDD Cycle Begins

Sarah starts working on tests:

```json
{
  "tool": "todo_update",
  "parameters": {
    "id": "implement-product-search",
    "metadata": {
      "current_test": "Test 1: Search returns exact matches"
    }
  }
}

// Update test section with first test
{
  "tool": "todo_update",
  "parameters": {
    "id": "implement-product-search",
    "section": "tests",
    "operation": "append",
    "content": "\n```typescript\n// Test 1: Search returns exact matches for product names\ndescribe('Product Search', () => {\n  it('should return exact matches for product names', async () => {\n    // Arrange\n    await createProduct({ name: 'Blue Widget', price: 29.99 });\n    await createProduct({ name: 'Red Widget', price: 34.99 });\n    \n    // Act\n    const response = await request(app)\n      .get('/api/v1/products/search?q=Blue Widget')\n      .expect(200);\n    \n    // Assert\n    expect(response.body.results).toHaveLength(1);\n    expect(response.body.results[0].name).toBe('Blue Widget');\n  });\n});\n```"
  }
}

// Log Red phase
{
  "tool": "todo_update",
  "parameters": {
    "id": "implement-product-search",
    "section": "tests",
    "operation": "append",
    "content": "\n\n**[2025-01-28 10:15:00] Red Phase**: Test failing with 'Cannot GET /api/v1/products/search'"
  }
}
```

### Day 2 - 2:00 PM - Searching for Similar Implementations

Sarah needs to find similar search implementations in the codebase:

```json
{
  "tool": "todo_search",
  "parameters": {
    "query": "search implementation",
    "scope": ["findings", "tests"],
    "limit": 5
  }
}

// Response
[
  {
    "id": "implement-user-search",
    "task": "Implement user search functionality",
    "score": 0.95,
    "snippet": "...implemented <mark>search</mark> using PostgreSQL FTS with..."
  },
  {
    "id": "optimize-search-performance",
    "task": "Optimize search query performance",
    "score": 0.87,
    "snippet": "...improved <mark>search implementation</mark> by adding GIN indexes..."
  }
]
```

### Day 3 - Progress Update

By Day 3, Sarah has made significant progress. Let's check her stats:

```json
{
  "tool": "todo_stats",
  "parameters": {
    "period": "week"
  }
}

// Response
{
  "total_todos": 12,
  "status_breakdown": {
    "completed": 3,
    "in_progress": 7,
    "pending": 2,
    "blocked": 0
  },
  "completion_rate": 25.0,
  "average_completion_time": "1d 4h 30m",
  "by_type": {
    "feature": {
      "count": 8,
      "completed": 2,
      "completion_rate": 25.0
    },
    "multi-phase": {
      "count": 1,
      "completed": 0,
      "completion_rate": 0.0
    },
    "research": {
      "count": 3,
      "completed": 1,
      "completion_rate": 33.3
    }
  },
  "by_priority": {
    "high": {
      "count": 9,
      "completed": 2,
      "completion_rate": 22.2,
      "avg_completion_time": "1d 2h"
    },
    "medium": {
      "count": 3,
      "completed": 1,
      "completion_rate": 33.3,
      "avg_completion_time": "8h"
    }
  },
  "test_coverage": {
    "todos_with_tests": 6,
    "coverage_percentage": 50.0,
    "average_test_completion": 40.0
  }
}
```

## Day 4: Bug Discovery

### 10:00 AM - Critical Bug Found

During testing, Sarah discovers a critical bug:

```json
{
  "tool": "todo_create",
  "parameters": {
    "task": "Fix SQL injection vulnerability in search endpoint",
    "priority": "high",
    "type": "bug",
    "template": "bug-fix"
  }
}
```

**Bug Todo Created**: `/Users/sarah/.claude/todos/fix-sql-injection-vulnerability.md`
```markdown
---
todo_id: fix-sql-injection-vulnerability
started: 2025-01-30 10:00:00
completed: ""
status: in_progress
priority: high
type: bug
parent_id: ""
tags: ["security", "bug-fix", "critical"]
---

# Task: Fix SQL injection vulnerability in search endpoint

## Bug Report

**Issue**: User input not properly sanitized in search query
**Steps to Reproduce**:
1. Send GET request to `/api/v1/products/search?q='; DROP TABLE products;--`
2. Observe SQL error in logs
3. Check database - products table could be dropped

**Expected Behavior**: Input should be sanitized, query should return empty results
**Actual Behavior**: Raw SQL is executed

## Test Strategy

- [ ] Write failing test that reproduces SQL injection
- [ ] Fix vulnerability using parameterized queries
- [ ] Add input validation
- [ ] Verify all tests pass
- [ ] Security audit other endpoints

## Test List

- [ ] Test 1: SQL injection attempt returns safe empty result
- [ ] Test 2: Special SQL characters are escaped
- [ ] Test 3: Valid searches still work after fix
- [ ] Test 4: Performance not degraded by sanitization

## Working Scratchpad

Found issue in ProductController.search():
```typescript
// VULNERABLE CODE:
const query = `SELECT * FROM products WHERE name LIKE '%${req.query.q}%'`;
```
```

### 11:00 AM - Updating Bug Status

```json
{
  "tool": "todo_update",
  "parameters": {
    "id": "fix-sql-injection-vulnerability",
    "section": "tests",
    "operation": "replace",
    "content": "## Test Implementation\n\n```typescript\n// Test 1: SQL injection attempt returns safe empty result\nit('should safely handle SQL injection attempts', async () => {\n  // Arrange\n  const maliciousQuery = \"'; DROP TABLE products;--\";\n  \n  // Act\n  const response = await request(app)\n    .get(`/api/v1/products/search?q=${encodeURIComponent(maliciousQuery)}`)\n    .expect(200);\n  \n  // Assert\n  expect(response.body.results).toEqual([]);\n  \n  // Verify table still exists\n  const tableExists = await db.raw(\n    \"SELECT EXISTS (SELECT FROM pg_tables WHERE tablename = 'products')\"\n  );\n  expect(tableExists.rows[0].exists).toBe(true);\n});\n```\n\n**[2025-01-30 11:00:00] Red Phase**: Test confirms vulnerability exists\n**[2025-01-30 11:30:00] Green Phase**: Fixed using parameterized query:\n\n```typescript\n// SECURE CODE:\nconst results = await db('products')\n  .where('name', 'ilike', `%${db.raw('?', [req.query.q])}%`)\n  .limit(20);\n```\n\n**[2025-01-30 11:45:00] Refactor Phase**: Added input validation layer"
  }
}
```

### 2:00 PM - Linking Bug to Main Project

```json
{
  "tool": "todo_link",
  "parameters": {
    "parent_id": "build-e-commerce-rest-api",
    "child_id": "fix-sql-injection-vulnerability",
    "link_type": "parent-child"
  }
}

// Response
{
  "parent_id": "build-e-commerce-rest-api",
  "child_id": "fix-sql-injection-vulnerability",
  "link_type": "parent-child",
  "message": "Todos linked successfully"
}
```

## Day 5: Authentication Implementation

### 9:00 AM - Starting Authentication Phase

```json
{
  "tool": "todo_update",
  "parameters": {
    "id": "phase-3-add-jwt-authentication",
    "section": "checklist",
    "operation": "replace",
    "content": "- [x] Research JWT best practices\n- [x] Design auth endpoints\n- [ ] Implement /auth/register\n- [ ] Implement /auth/login\n- [ ] Implement /auth/refresh\n- [ ] Add auth middleware\n- [ ] Protect product endpoints\n- [ ] Write auth tests"
  }
}
```

### 10:00 AM - Searching for Auth Examples

```json
{
  "tool": "todo_search",
  "parameters": {
    "query": "\"JWT authentication\" OR \"auth middleware\"",
    "scope": ["findings", "tests"],
    "filters": {
      "date_from": "2025-01-01"
    }
  }
}

// Response
[
  {
    "id": "implement-jwt-refresh",
    "task": "Implement JWT refresh token rotation",
    "score": 0.92,
    "snippet": "...<mark>JWT authentication</mark> with secure refresh token rotation..."
  },
  {
    "id": "add-auth-middleware",
    "task": "Add authentication middleware to API routes",
    "score": 0.88,
    "snippet": "...created <mark>auth middleware</mark> using express-jwt library..."
  }
]
```

### File System State Check

Let's see Sarah's todo directory structure:

```bash
~/.claude/todos/
├── build-e-commerce-rest-api.md
├── phase-1-design-rest-api.md
├── phase-2-implement-product-crud.md
├── phase-3-add-jwt-authentication.md
├── phase-4-write-tests-documentation.md
├── implement-product-search.md
├── fix-sql-injection-vulnerability.md
├── implement-user-registration.md
├── implement-login-endpoint.md
├── add-refresh-token-rotation.md
├── optimize-database-queries.md
└── write-api-documentation.md
```

## Day 6: Testing and Optimization

### 9:00 AM - Running Test Coverage Analysis

Sarah checks which todos have test coverage:

```json
{
  "tool": "todo_read",
  "parameters": {
    "filter": {
      "status": "in_progress"
    },
    "format": "summary"
  }
}
```

She notices some todos lack tests and updates them:

```json
{
  "tool": "todo_update",
  "parameters": {
    "id": "optimize-database-queries",
    "section": "tests",
    "operation": "append",
    "content": "\n## Performance Tests\n\n- [ ] Test 1: Product list loads in <100ms with 1000 products\n- [ ] Test 2: Search completes in <200ms with 10k products\n- [ ] Test 3: Concurrent requests don't cause deadlocks\n- [ ] Test 4: Memory usage stays under 512MB"
  }
}
```

### 2:00 PM - Finding Duplicates

Sarah wants to clean up any duplicate todos:

```json
{
  "tool": "todo_clean",
  "parameters": {
    "operation": "find_duplicates"
  }
}

// Response
{
  "duplicates": [
    {
      "ids": ["implement-user-search", "add-user-search-endpoint"],
      "similarity": 0.89
    }
  ],
  "message": "Found 1 potential duplicate set"
}
```

## Day 7: Project Completion

### 10:00 AM - Completing Todos

Sarah marks completed todos:

```json
// Complete the search implementation
{
  "tool": "todo_update",
  "parameters": {
    "id": "implement-product-search",
    "metadata": {
      "status": "completed"
    }
  }
}

// Complete the SQL injection fix
{
  "tool": "todo_update",
  "parameters": {
    "id": "fix-sql-injection-vulnerability",
    "metadata": {
      "status": "completed"
    }
  }
}
```

### 11:00 AM - Archiving Completed Work

```json
// Archive the completed search todo
{
  "tool": "todo_archive",
  "parameters": {
    "id": "implement-product-search"
  }
}

// Response
{
  "id": "implement-product-search",
  "archive_path": "/Users/sarah/.claude/archive/2025/02/02/implement-product-search.md",
  "message": "Todo archived successfully"
}
```

**Archived File**: `/Users/sarah/.claude/archive/2025/02/02/implement-product-search.md`
```markdown
---
todo_id: implement-product-search
started: 2025-01-28 09:00:00
completed: 2025-02-02 10:00:00
status: completed
priority: high
type: feature
parent_id: ""
tags: ["tdd", "search", "products"]
current_test: "All tests completed"
---

# Task: Implement product search with filters

## Findings & Research

WebSearch: "elasticsearch vs postgresql full text search 2025"
- PostgreSQL FTS sufficient for <100k products
- Use GIN indexes for performance
- tsvector for search optimization

## Test Strategy

- **Test Framework**: Jest + Supertest
- **Test Types**: Unit tests for search logic, Integration tests for API
- **Coverage Target**: 90%
- **Edge Cases**: Empty search, special characters, SQL injection

## Test List

- [x] Test 1: Search returns exact matches for product names
- [x] Test 2: Search is case-insensitive
- [x] Test 3: Search handles partial matches
- [x] Test 4: Filter by price range works correctly
- [x] Test 5: Filter by category returns only matching products
- [x] Test 6: Combined search and filters work together
- [x] Test 7: Empty search returns all products with pagination
- [x] Test 8: Special characters are handled safely
- [x] Test 9: Search response time under 200ms
- [x] Test 10: Results are properly paginated

## Test Cases

[... full test implementations with Red-Green-Refactor cycles ...]

## Test Results Log

**[2025-01-28 10:15:00] Red Phase**: Test 1 failing - endpoint doesn't exist
**[2025-01-28 10:45:00] Green Phase**: Test 1 passing - basic endpoint created
**[2025-01-28 11:00:00] Refactor Phase**: Extracted search logic to service
[... continue for all 10 tests ...]
**[2025-02-01 16:00:00] All tests passing - 92% coverage achieved**

## Checklist

- [x] Design search endpoint schema
- [x] Set up test environment
- [x] Write failing tests (Red phase)
- [x] Implement search logic (Green phase)
- [x] Refactor for performance
- [x] Add database indexes
- [x] Document API endpoint
```

### 3:00 PM - Bulk Archive Old Todos

```json
{
  "tool": "todo_clean",
  "parameters": {
    "operation": "archive_old",
    "days": 7
  }
}

// Response
{
  "archived_count": 3,
  "archived_todos": [
    "research-api-frameworks",
    "setup-development-environment",
    "design-database-schema"
  ],
  "message": "Archived 3 todos older than 7 days"
}
```

### Archive Structure After Week

```
~/.claude/archive/
└── 2025/
    ├── 01/
    │   ├── 27/
    │   │   ├── research-api-frameworks.md
    │   │   └── setup-development-environment.md
    │   ├── 28/
    │   │   └── design-database-schema.md
    │   └── 29/
    │       └── implement-basic-crud.md
    └── 02/
        ├── 01/
        │   ├── add-validation-layer.md
        │   └── implement-error-handling.md
        └── 02/
            ├── implement-product-search.md
            └── fix-sql-injection-vulnerability.md
```

## Analytics and Insights

### Final Project Statistics

```json
{
  "tool": "todo_stats",
  "parameters": {
    "period": "all"
  }
}

// Response
{
  "total_todos": 15,
  "status_breakdown": {
    "completed": 12,
    "in_progress": 2,
    "pending": 0,
    "blocked": 1
  },
  "completion_rate": 80.0,
  "average_completion_time": "1d 8h 45m",
  "by_type": {
    "feature": {
      "count": 10,
      "completed": 8,
      "completion_rate": 80.0,
      "avg_completion_time": "1d 10h"
    },
    "bug": {
      "count": 2,
      "completed": 2,
      "completion_rate": 100.0,
      "avg_completion_time": "4h 30m"
    },
    "multi-phase": {
      "count": 1,
      "completed": 0,
      "completion_rate": 0.0
    },
    "research": {
      "count": 2,
      "completed": 2,
      "completion_rate": 100.0,
      "avg_completion_time": "2h"
    }
  },
  "by_priority": {
    "high": {
      "count": 11,
      "completed": 9,
      "completion_rate": 81.8,
      "avg_completion_time": "1d 4h"
    },
    "medium": {
      "count": 4,
      "completed": 3,
      "completion_rate": 75.0,
      "avg_completion_time": "2d 1h"
    }
  },
  "test_coverage": {
    "todos_with_tests": 8,
    "coverage_percentage": 53.3,
    "average_test_completion": 87.5
  },
  "productivity": {
    "most_productive_day": "Wednesday",
    "average_daily_completions": 2.4,
    "peak_hours": "10:00-12:00",
    "streak": {
      "current": 7,
      "longest": 7
    }
  }
}
```

### Search Index State

The search index now contains all todo content:

```
~/.claude/index/todos.bleve/
├── index_meta.json
├── store/
│   ├── 00000000.zap  # Contains indexed todo content
│   └── root.bolt     # Index structure
```

Sample index entry:
```json
{
  "id": "implement-product-search",
  "task": "Implement product search with filters",
  "status": "completed",
  "priority": "high",
  "type": "feature",
  "started": "2025-01-28T09:00:00Z",
  "completed": "2025-02-02T10:00:00Z",
  "content": "Full markdown content...",
  "findings": "WebSearch: elasticsearch vs postgresql...",
  "tests": "Test 1: Search returns exact matches..."
}
```

## Comparison: Native Tools vs MCP Server

### What Sarah Couldn't Do with Native Tools

1. **No Persistence**: Would lose all todos on session restart
2. **No Search**: Couldn't find that auth example from last week
3. **No Relationships**: Couldn't link bug to main project
4. **No Archives**: Completed todos would clutter the list
5. **No Analytics**: No insights into productivity patterns
6. **No Templates**: Would recreate structure each time
7. **Limited Updates**: Only full array replacement

### What MCP Server Enabled

1. **Full Project History**: Every todo preserved with timestamps
2. **Quick Search**: Found similar implementations in seconds
3. **Organization**: Parent-child relationships for complex projects
4. **Clean Workspace**: Archived completed work stays accessible
5. **Data-Driven Insights**: Discovered Wednesday productivity peak
6. **Consistent Structure**: Templates ensured uniform todo format
7. **Granular Updates**: Updated specific sections during development

## Key Takeaways

### For Developers

- **TDD Support**: Track test progress with current_test field
- **Bug Tracking**: Dedicated bug type with report template
- **Knowledge Base**: Search finds solutions from past todos
- **Performance**: Sub-100ms operations even with 1000+ todos

### For Teams

- **Standardization**: Templates ensure consistent task structure
- **Metrics**: Track team velocity and completion rates
- **Audit Trail**: Complete history of changes with timestamps
- **Integration Ready**: JSON API works with other tools

### For Projects

- **Scalability**: Daily archives handle high-volume workflows
- **Flexibility**: Gradual adoption from basic to advanced features
- **Reliability**: File-based storage survives crashes
- **Extensibility**: Easy to add custom fields and templates

## Conclusion

Through Sarah's week-long project, we've seen how the MCP Todo Server transforms task management from a simple in-memory list to a comprehensive project management system. The combination of persistent storage, full-text search, templates, and analytics provides a powerful foundation for modern software development workflows.