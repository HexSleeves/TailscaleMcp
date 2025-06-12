# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Development

```bash
npm run build           # Build the project
npm run build:watch     # Build with file watching
npm run dev:direct      # Fast development with tsx
npm run dev:watch       # Auto-rebuild on changes
```

### Testing

```bash
npm test                # Run all tests
npm run test:unit       # Unit tests only
npm run test:integration # Integration tests (requires Tailscale CLI)
npm run test:watch      # Watch mode for all tests
npm run test:coverage   # Generate coverage reports
```

### Quality Assurance

```bash
npm run qa              # Quick QA: typecheck + unit tests + lint
npm run qa:full         # Full QA: all tests + checks
npm run typecheck       # TypeScript validation
npm run lint            # ESLint
npm run format          # Format code with ESLint + Prettier
```

### Tools

```bash
npm run inspector       # Test with MCP Inspector
npm run clean           # Clean dist directory
```

## Architecture Overview

This is a **Model Context Protocol (MCP) server** that provides Tailscale integration through both CLI commands and REST API operations.

### Core Components

- **Entry Point**: `src/index.ts` → `src/cli.ts` - CLI interface and signal handling
- **Server Core**: `src/server.ts` - Main server implementation with dual-mode support (stdio/http)
- **Tool System**: `src/tools/` - Modular tool registry with Zod validation
  - `device-tools.ts` - Device management (list, authorize, routes)
  - `network-tools.ts` - Network operations (connect, disconnect, ping)
  - `acl-tools.ts` - Access control lists
  - `admin-tools.ts` - Administrative functions
- **Tailscale Integration**: `src/tailscale/`
  - `tailscale-api.ts` - REST API client using Axios
  - `tailscale-cli.ts` - CLI command wrapper
- **Infrastructure**:
  - `src/logger.ts` - Centralized logging
  - `src/types.ts` - Core type definitions

### Key Patterns

- **Tool Registry Pattern**: All tools are registered through `ToolRegistry` class with consistent interfaces
- **Zod Validation**: Input schemas defined with Zod for runtime type safety
- **Dual Integration**: Both CLI and API operations supported for different use cases
- **Error Handling**: Structured error responses with proper MCP format

### Server Modes

The server supports two modes:

- **stdio**: Standard MCP communication (default)
- **http**: HTTP server for testing/development

## Testing Strategy

### Test Structure

- **Unit Tests**: `src/__test__/**/*.test.ts` - Isolated component testing
- **Integration Tests**: `src/__test__/**/*.integration.test.ts` - End-to-end with Tailscale CLI
- **Separate Configs**: Different Jest configs for unit vs integration testing

### Test Environment

- Integration tests require Tailscale CLI installation
- Use `npm run test:setup` to configure testing environment
- Security tests verify CLI command injection protection

## Environment Configuration

### Required Variables

- `TAILSCALE_API_KEY` - Tailscale API key for REST operations
- `TAILSCALE_TAILNET` - Tailnet name for API operations

### Optional Variables

- `TAILSCALE_API_BASE_URL` - Custom API base URL (defaults to `https://api.tailscale.com`)
- `LOG_LEVEL` - Logging level 0-3 (0=debug, 1=info, 2=warn, 3=error)
- `MCP_SERVER_LOG_FILE` - Enable file logging

## Development Guidelines

### Code Organization

- Follow the modular tool system pattern when adding new tools
- Use Zod schemas for all input validation
- Implement both CLI and API versions where applicable
- Keep tools focused and atomic

### Error Handling

- All tool handlers should return proper `CallToolResult` format
- Use structured error messages with context
- Log errors appropriately based on severity

### Security Considerations

- CLI commands are validated to prevent injection attacks
- API keys are required for REST operations
- Input validation through Zod schemas

## Cursor Rules Integration

This project uses comprehensive Cursor rules focusing on:

- **Systematic Implementation**: Step-by-step analysis, planning, and coding
- **Modular Design**: Break complex logic into smaller, atomic components
- **Comprehensive Testing**: Unit and integration tests for all functionality
- **Code Preservation**: Maintain working components, minimize disruption
- **Incremental Changes**: One logical feature at a time with full dependency resolution
