# CONTEXT.md

This file provides essential context for AI coding agents working on the Tailscale MCP Server project.

## Build/Test/Lint Commands

**Go (Primary):**
- `make build` - Build Go binary (always run after changes)
- `make test` - Run all tests (unit + integration)
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests (requires Tailscale CLI)
- `make lint` - Run golangci-lint or go vet
- `make fmt` - Format code with gofmt and goimports
- `make check` - Run all quality checks (fmt + lint + test)
- `go test -v ./internal/tailscale -run TestSpecificFunction` - Run single test

**TypeScript (Legacy):**
- `npm run qa` - Quick QA: typecheck + unit tests + lint
- `npm run typecheck` - TypeScript validation
- `npm run lint` - ESLint
- `npm run test:unit` - Unit tests only

## Code Style Guidelines

**Go Conventions:**
- Use Go standard naming: camelCase for private, PascalCase for exported
- Package names: lowercase, match directory names
- Error handling: explicit returns with `fmt.Errorf("message: %w", err)`
- Imports: group stdlib, external, internal with blank lines
- Comments: start with function/type name, end with period
- Interfaces: small and focused, define where used not implemented
- Use functional options pattern for configuration (see `ServerOption`)
- Composition over inheritance, dependency injection via constructors

**Project Patterns:**
- Tool Registry Pattern: modular tools with Zod validation (TypeScript side)
- Dual Integration: both CLI and API operations supported
- Structured logging with context: `logger.Info("message", "key", value)`
- Generic API responses: `APIResponse[T]` with Success/Data/Error fields
- Context propagation: pass `context.Context` as first parameter

**Error Handling:**
- Return errors explicitly, don't panic
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use custom error types for API errors with status codes
- Log errors at appropriate levels before returning

This is a hybrid Go/TypeScript MCP server project in migration from TypeScript to Go. Follow Go idioms for new code, maintain existing TypeScript patterns where needed. Always run `make build` to verify changes.