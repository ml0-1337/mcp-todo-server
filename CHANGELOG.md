# Changelog

All notable changes to the MCP Todo Server project will be documented in this file.

## [Unreleased]

### Added
- HTTP header-based working directory resolution
  - Todos are now created in the project where Claude Code is running
  - Automatic detection via `X-Working-Directory` header
  - Session management for persistent connections
  - Context-aware todo managers for different projects
  - Full backward compatibility with existing behavior

### Fixed
- Fixed initialization bug where `NewTodoHandlers` was creating regular `TodoManager` instead of `ContextualTodoManagerWrapper`
- Fixed session management for StreamableHTTPServer requiring `Mcp-Session-Id` headers

### Changed
- HTTP transport now recommended over STDIO for better project isolation
- Updated documentation to explain context-aware todo creation

## [1.0.0] - Previous Release

### Added
- Initial MCP Todo Server implementation
- Full markdown-based todo management
- HTTP and STDIO transport support
- Full-text search with Bleve
- Template system
- Todo linking for multi-phase projects
- Analytics and reporting