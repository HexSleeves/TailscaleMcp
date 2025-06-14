# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go Build and Development

```bash
make build              # Build the Go binary
make run-dev            # Run with `go run` for fast feedback
make build-all          # Cross-compile for all target platforms
make test               # Run all Go tests
make test-unit          # Go unit tests only
make test-integration   # Go integration tests (requires Tailscale CLI)
make lint               # Run Go linter (golangci-lint)
make fmt                # Format Go code
make check              # Run all quality checks
make clean              # Clean build artifacts
```

### TypeScript (Legacy)

```bash
npm test                # Run all TS tests
npm run test:unit       # TS unit tests only
npm run test:integration # TS integration tests (requires Tailscale CLI)
npm run test:watch      # Watch mode for all tests
npm run test:coverage   # Generate coverage reports
npm run qa              # Quick QA: typecheck + unit tests + lint
npm run qa:full         # Full QA: all tests + checks
npm run typecheck       # TypeScript validation
npm run lint            # ESLint
npm run format          # Format code with ESLint + Prettier
npm run inspector       # Test with MCP Inspector
npm run clean           # Clean dist directory
```

## Architecture Overview

This is a **Model Context Protocol (MCP) server** that provides Tailscale integration through both CLI commands and REST API operations. The project is currently undergoing migration from TypeScript to Go while maintaining full backward compatibility.

### Project Status: TypeScript → Go Migration

- **Current State**: Hybrid codebase with both TypeScript and Go implementations
- **Go Implementation**: Primary development focus in `cmd/`, `internal/`, `pkg/` directories
- **TypeScript (Legacy)**: Original implementation in `src/` directory, maintained for compatibility
- **Migration Branch**: `go-migration-plan` contains the transition strategy

### Go Architecture (Primary)

**Entry Points:**

- `cmd/tailscale-mcp-server/main.go` - CLI binary entry point using Cobra framework
- `internal/cmd/` - Command definitions (root, serve, version)

**Core Components:**

- `internal/server/server.go` - Main server implementation with functional options pattern
- `internal/server/stdio.go` - Stdio MCP server implementation
- `internal/server/http.go` - HTTP MCP server implementation
- `internal/tailscale/` - Tailscale integration (api.go for REST, cli.go for CLI)
- `internal/tools/` - Tool registry and individual tool implementations
- `pkg/mcp/` - MCP protocol types and server interfaces

**Infrastructure:**

- `internal/config/` - Configuration management with validation
- `internal/logger/` - Structured logging using standard library
- `internal/validation/` - Input validation schemas

### TypeScript Architecture (Legacy)

**Entry Points:**

- `src/index.ts` → `src/cli.ts` - CLI interface and signal handling
- `src/server.ts` - Main server class with dual-mode support

**Core Components:**

- `src/servers/` - Server implementations (stdio, http)
- `src/tools/` - Modular tool registry with Zod validation
- `src/tailscale/` - Tailscale integrations (API client, CLI wrapper)
- `src/types.ts` - Core type definitions
- `src/logger.ts` - Centralized logging

### Key Patterns

- **Go Idioms**: Functional options, composition over inheritance, interface-based design
- **Tool Registry Pattern**: Consistent tool interfaces across both implementations
- **Dual Integration**: Both CLI and API operations supported for different use cases
- **Input Validation**: Go uses custom validation, TypeScript uses Zod schemas
- **Error Handling**: Structured error responses with proper MCP format

### Server Modes

The server supports two modes:

- **stdio**: Standard MCP communication (default)
- **http**: HTTP server for testing/development

## Testing Strategy

### Go Testing (Primary)

**Test Structure:**

- **Unit Tests**: `internal/` and `pkg/` directories with `*_test.go` files
- **Integration Tests**: `test/integration/` directory with build tag `integration`
- **Go Testing**: Uses standard `testing` package with `testify/assert` for assertions

**Test Commands:**

- `make test-unit` - Run Go unit tests with race detection and coverage
- `make test-integration` - Run Go integration tests (requires Tailscale CLI)
- `make test-coverage` - Generate HTML coverage report

### TypeScript Testing (Legacy)

**Test Structure:**

- **Unit Tests**: `src/__test__/**/*.test.ts` - Isolated component testing
- **Integration Tests**: `src/__test__/**/*.integration.test.ts` - End-to-end with Tailscale CLI
- **Separate Configs**: Different Jest configs for unit vs integration testing

**Test Environment:**

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

## Migration Context

### Current Development Focus

- **Primary Development**: Go implementation (`cmd/`, `internal/`, `pkg/`)
- **Maintenance Mode**: TypeScript implementation (`src/`)
- **Migration Status**: Active development on `go-migration-plan` branch
- **Compatibility**: Both implementations maintained during transition

### When to Use Which Implementation

- **New Features**: Implement in Go first, then backport to TypeScript if needed
- **Bug Fixes**: Fix in both implementations if the bug exists in both
- **Tool Development**: Follow Go patterns for new tools
- **Testing**: Prioritize Go test coverage for new functionality

## Development Guidelines

### Go Development (Primary)

- Follow Go idioms: composition over inheritance, interfaces for abstraction
- Use functional options pattern for configuration
- Implement proper error wrapping with `fmt.Errorf`
- Use context.Context for cancellation and timeouts
- Follow the existing tool registry pattern for new tools
- Use custom validation structs instead of external schema libraries

### TypeScript Development (Legacy)

- Follow the modular tool system pattern when adding new tools
- Use Zod schemas for all input validation
- Implement both CLI and API versions where applicable
- Keep tools focused and atomic

### Error Handling

- **Go**: Use error wrapping and structured logging
- **TypeScript**: Return proper `CallToolResult` format with structured error messages
- Log errors appropriately based on severity in both implementations

### Security Considerations

- CLI commands are validated to prevent injection attacks in both implementations
- API keys are required for REST operations
- Input validation through custom validation (Go) or Zod schemas (TypeScript)

## Cursor Rules Integration

This project uses comprehensive Cursor rules focusing on:

- **Systematic Implementation**: Step-by-step analysis, planning, and coding
- **Modular Design**: Break complex logic into smaller, atomic components
- **Comprehensive Testing**: Unit and integration tests for all functionality
- **Code Preservation**: Maintain working components, minimize disruption
- **Incremental Changes**: One logical feature at a time with full dependency resolution

## Memory Context

- You are allowed to search hidden files such as `.cursor` or `.taskmaster`
