# MCP Todo Server - Development Guidelines

## Critical Guidelines for STDIO Mode Compatibility

### Log Output Interference Prevention

**CRITICAL**: In STDIO mode, the server communicates via JSON-RPC protocol over stdin/stdout. ANY output to stdout that is not valid JSON-RPC will corrupt the protocol and cause connection timeouts.

#### Rules for Logging

1. **NEVER use stdout for logging** - No `fmt.Println()`, `log.Print()`, or any stdout writes
2. **Always redirect logs to stderr** - Use `fmt.Fprintf(os.Stderr, ...)` or `log.SetOutput(os.Stderr)`
3. **Check all packages** - Ensure utils, core, handlers, and server packages follow these rules
4. **Init-time logging** - Be especially careful with initialization code that runs before main()

#### Safe Logging Patterns

```go
// ✅ GOOD - Logs to stderr
fmt.Fprintf(os.Stderr, "Debug message: %s\n", value)
log.SetOutput(os.Stderr)
log.Printf("This goes to stderr")

// ❌ BAD - Corrupts STDIO protocol
fmt.Println("This breaks STDIO mode")
log.Print("This also breaks STDIO mode")
fmt.Printf("Don't do this either")
```

#### Testing STDIO Mode

When testing STDIO mode:
1. Always check stderr for logs: `./mcp-todo-server -transport=stdio 2>stderr.log`
2. Ensure stdout only contains JSON-RPC messages
3. Test with proper initialization: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}' | ./mcp-todo-server -transport=stdio`

## Project-Specific Guidelines

### Archive Directory Structure

- Archives are stored in `.claude/archive/YYYY/MM/DD/` within the project directory
- When `CLAUDE_TODO_PATH` is set, archives remain within that directory structure
- Never use `filepath.Dir(basePath)` for archive paths - always use `filepath.Join(basePath, ".claude", "archive")`

### Testing Guidelines

- Run `make test` before committing
- Use `make test-verbose` for detailed output
- Edge case tests are important - fix them when they fail
- Archive operations should be atomic (use rename, not copy+delete)

### Build Process

- Use `make build` to create binaries with version info
- Binary is output to `./build/mcp-todo-server`
- If STDIO mode hangs with no output, check if binary is corrupted (compare with build/ version)

## Development Workflow

1. Make changes
2. Run tests: `make test`
3. Build: `make build`
4. Test STDIO mode: `./build/mcp-todo-server -transport=stdio 2>stderr.log`
5. Test HTTP mode: `./build/mcp-todo-server -transport=http`
6. Copy binary: `cp build/mcp-todo-server .`
7. Commit changes with semantic commit messages

## Common Issues and Solutions

### STDIO Mode Timeout
- **Symptom**: "MCP server for mcp-todo-server failed to respond within 30000ms"
- **Cause**: Log output to stdout corrupting JSON-RPC protocol
- **Solution**: Check all code for stdout writes, redirect to stderr

### Binary Hangs on Startup
- **Symptom**: No output even to stderr, process in uninterruptible sleep (UE)
- **Cause**: Corrupted binary or macOS quarantine attributes
- **Solution**: Rebuild with `make build` and copy fresh binary from build/

### Archive Files in Wrong Location
- **Symptom**: Archives created outside project directory
- **Cause**: Using `filepath.Dir(basePath)` in archive path calculation
- **Solution**: Use `filepath.Join(basePath, ".claude", "archive", ...)`