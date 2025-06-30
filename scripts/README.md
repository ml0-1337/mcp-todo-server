# MCP Todo Server Scripts

This directory contains all executable scripts for the MCP Todo Server project, organized by purpose.

## Directory Structure

```
scripts/
├── setup/          # Setup and installation scripts
├── test/           # Test scripts organized by type
│   ├── e2e/       # End-to-end test suites
│   ├── http/      # HTTP transport-specific tests
│   ├── stdio/     # STDIO transport-specific tests
│   └── integration/ # Integration and feature tests
└── dev/           # Development utilities (future)
```

## Setup Scripts

### setup/setup.sh
Initial setup script that helps users:
- Build the MCP Todo Server
- Choose between HTTP or STDIO transport mode
- Get instructions for Claude Code integration

**Usage:**
```bash
./scripts/setup/setup.sh
```

## Test Scripts

All test scripts assume the server binary exists at the project root. Run `make build` before running tests.

### End-to-End Tests (test/e2e/)

#### test_comprehensive.sh
Comprehensive test suite that exercises all server functionality in STDIO mode:
- Server initialization
- Tool listing
- Todo CRUD operations
- Search functionality
- Statistics
- Template operations

**Usage:**
```bash
./scripts/test/e2e/test_comprehensive.sh
# or via Makefile
make test-e2e
```

#### test_multi_instance.sh
Tests running multiple server instances simultaneously using HTTP transport on different ports.

**Usage:**
```bash
./scripts/test/e2e/test_multi_instance.sh
```

### HTTP Transport Tests (test/http/)

#### test_http.sh
Basic HTTP transport functionality test:
- Server startup on port 8080
- JSON-RPC communication
- Basic todo operations

**Usage:**
```bash
./scripts/test/http/test_http.sh
# or via Makefile
make test-http-quick
```

#### test_http_headers.sh
Tests HTTP header-based working directory isolation:
- X-Working-Directory header functionality
- Project isolation between different directories
- Session management

**Usage:**
```bash
./scripts/test/http/test_http_headers.sh
```

### STDIO Transport Tests (test/stdio/)

#### test_server.sh
Basic STDIO transport test:
- Server initialization
- Tool listing
- Todo creation

**Usage:**
```bash
./scripts/test/stdio/test_server.sh
# or via Makefile
make test-stdio-quick
```

### Integration Tests (test/integration/)

#### test_fix.sh
Tests the context-aware working directory fix:
- Verifies todos are created in the correct directory
- Tests X-Working-Directory header behavior
- Ensures server directory isolation

**Usage:**
```bash
./scripts/test/integration/test_fix.sh
```

#### test_simple.sh
Simple HTTP header and session management test:
- Session ID handling
- Working directory header processing
- Response parsing

**Usage:**
```bash
./scripts/test/integration/test_simple.sh
```

## Common Test Patterns

All test scripts follow these patterns:

1. **Error Handling**: Use `set -euo pipefail` for strict error handling
2. **Cleanup**: Kill server processes and clean up temporary files
3. **Binary Path**: Reference the server binary at `../../../mcp-todo-server`
4. **Output**: Use temporary files to capture server output for analysis

## Running Tests

### Quick Test Commands
```bash
# Run all unit tests
make test

# Run specific test suites
make test-e2e          # End-to-end tests
make test-http-quick   # Quick HTTP test
make test-stdio-quick  # Quick STDIO test

# Run scripts directly
./scripts/test/http/test_http.sh
./scripts/test/e2e/test_comprehensive.sh
```

### Prerequisites
- Go 1.21+ installed
- Server binary built (`make build`)
- Port 8080 available for HTTP tests
- Write permissions in test directories

## Adding New Scripts

When adding new scripts:

1. Place in the appropriate subdirectory
2. Use `#!/usr/bin/env bash` shebang
3. Add `set -euo pipefail` for error handling
4. Include descriptive header comments
5. Update this README with usage information
6. Add corresponding Makefile target if appropriate

## Script Standards

All scripts should follow these standards:

- **Shebang**: `#!/usr/bin/env bash` for portability
- **Error Handling**: `set -euo pipefail` at the top
- **Comments**: Clear header explaining purpose and usage
- **Variables**: UPPERCASE for constants, lowercase for locals
- **Functions**: Use for repeated code blocks
- **Cleanup**: Always clean up processes and temp files
- **Exit Codes**: Use meaningful exit codes (0=success, 1=error)