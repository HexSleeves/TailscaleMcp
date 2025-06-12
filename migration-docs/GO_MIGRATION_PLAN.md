# Go Migration Plan - Tailscale MCP Server

## Overview

This document outlines the migration plan from TypeScript to Go for the Tailscale MCP Server project, maintaining all existing functionality while improving performance, distribution, and maintainability.

## Current State Analysis

**TypeScript Codebase:**

- ~5,700 lines of code
- Modular architecture with tool registry pattern
- 5 main tool modules (ACL, Admin, Device, Network, Core)
- Dual server modes (stdio/http)
- Comprehensive testing (unit + integration)
- Docker + NPM distribution

**Key Dependencies:**

- `@modelcontextprotocol/sdk` - MCP protocol implementation
- `axios` - HTTP client for Tailscale API
- `zod` - Runtime validation
- `express` - HTTP server
- `dotenv` - Environment configuration
- `jest` - Testing framework

## Go Project Structure

```bash
tailscale-mcp-server/
├── cmd/
│   └── tailscale-mcp-server/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Environment configuration
│   ├── logger/
│   │   └── logger.go              # Structured logging
│   ├── server/
│   │   ├── server.go              # Main server implementation
│   │   ├── stdio.go               # Stdio MCP server
│   │   └── http.go                # HTTP MCP server
│   ├── tailscale/
│   │   ├── api.go                 # REST API client
│   │   ├── cli.go                 # CLI command wrapper
│   │   └── types.go               # Tailscale data structures
│   ├── tools/
│   │   ├── registry.go            # Tool registry system
│   │   ├── acl.go                 # ACL management tools
│   │   ├── admin.go               # Administrative tools
│   │   ├── device.go              # Device management tools
│   │   ├── network.go             # Network operation tools
│   │   └── types.go               # Tool interfaces
│   └── validation/
│       └── schemas.go             # Input validation
├── pkg/
│   └── mcp/                       # MCP protocol types (if needed)
├── scripts/
│   ├── build.sh                   # Build script
│   ├── test.sh                    # Testing script
│   └── docker-build.sh            # Docker build
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
├── test/
│   ├── integration/               # Integration tests
│   └── fixtures/                  # Test data
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Dependency Mapping

| TypeScript | Go Equivalent | Purpose |
|------------|---------------|---------|
| `@modelcontextprotocol/sdk` | Custom implementation + `encoding/json` | MCP protocol |
| `axios` | `net/http` + custom client | HTTP requests |
| `zod` | Custom validation + `reflect` | Input validation |
| `express` | `net/http` + `gorilla/mux` | HTTP server |
| `dotenv` | `os` + `github.com/joho/godotenv` | Environment vars |
| `jest` | `testing` + `testify/assert` | Testing |
| `child_process` | `os/exec` | CLI execution |
| Built-in `flag` | `github.com/spf13/cobra` | CLI framework |

**Key Go Dependencies:**

```go
require (
    github.com/spf13/cobra v1.8.0          // CLI framework and commands
    github.com/gorilla/mux v1.8.0          // HTTP routing
    github.com/joho/godotenv v1.5.1        // Environment loading
    github.com/stretchr/testify v1.8.4     // Testing assertions
    github.com/go-playground/validator/v10  // Struct validation
    github.com/sirupsen/logrus v1.9.3      // Structured logging
)
```

## Migration Phases

### Phase 1: Foundation (Week 1)

**Goal:** Set up Go project structure and core infrastructure

**Tasks:**

- [x] Initialize Go module and project structure
- [x] Implement configuration system with environment variables
- [x] Create structured logging system
- [x] Set up basic MCP protocol types and interfaces
- [x] Implement CLI argument parsing with `flag` package
- [x] Create Makefile with build/test/lint targets
- [ ] **Migrate CLI to Cobra framework** (improved UX and subcommands)

**Deliverables:**

- Working Go project with proper module structure
- Basic CLI that can parse arguments and load configuration
- Logging system with appropriate levels

### Phase 2: Core Infrastructure (Week 1-2)

**Goal:** Implement server foundations and Tailscale integrations

**Tasks:**

- [ ] Implement MCP server interfaces (stdio/http modes)
- [ ] Create Tailscale API client using `net/http`
- [ ] Implement Tailscale CLI wrapper with `os/exec`
- [ ] Design tool registry system with reflection-based registration
- [ ] Implement input validation system
- [ ] Set up error handling patterns

**Deliverables:**

- Functional MCP server that can handle basic requests
- Working Tailscale API client
- CLI command execution with proper error handling

### Phase 3: Tool Migration (Week 2-3)

**Goal:** Port all existing tools to Go

**Priority Order:**

1. **Network Tools** (`network.go`)
   - `get_network_status`
   - `connect_network`
   - `disconnect_network`
   - `ping_peer`

2. **Device Tools** (`device.go`)
   - `list_devices`
   - `device_action`
   - `manage_routes`

3. **Admin Tools** (`admin.go`)
   - `get_version`
   - `get_tailnet_info`

4. **ACL Tools** (`acl.go`)
   - ACL management operations

**Tasks:**

- [ ] Port tool implementations with proper Go patterns
- [ ] Implement JSON schema validation for tool inputs
- [ ] Add comprehensive error handling
- [ ] Ensure security validations (CLI injection prevention)

**Deliverables:**

- All tools functional and tested
- Input validation working correctly
- Proper error messages and logging

### Phase 4: Testing & Quality (Week 3)

**Goal:** Comprehensive testing and quality assurance

**Tasks:**

- [ ] Port unit tests to Go testing framework
- [ ] Implement integration tests with real Tailscale CLI
- [ ] Add test coverage reporting
- [ ] Implement security tests (input validation, CLI safety)
- [ ] Performance benchmarking
- [ ] Memory usage optimization

**Deliverables:**

- Complete test suite with >90% coverage
- Security validation tests
- Performance benchmarks

### Phase 5: Packaging & Distribution (Week 3-4)

**Goal:** Build and distribution system

**Tasks:**

- [ ] Cross-compilation setup for multiple platforms
- [ ] Docker image optimization (multi-stage build)
- [ ] GitHub Actions CI/CD pipeline
- [ ] Release automation
- [ ] Documentation updates
- [ ] Migration guide for users

**Deliverables:**

- Automated builds for Linux, macOS, Windows
- Optimized Docker images
- Release pipeline
- Updated documentation

## Implementation Strategy

### 1. CLI Framework with Cobra

Using Cobra for enhanced CLI experience over basic `flag` package:

```go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "tailscale-mcp-server",
    Short: "Tailscale MCP Server",
    Long:  "A Model Context Protocol server for Tailscale integration",
}

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the MCP server",
    Run:   runServer,
}
```

**Benefits:**

- **Better UX**: Professional help formatting and command structure
- **Subcommands**: Future extensibility (e.g., `serve`, `version`, `config`)
- **Auto-completion**: Shell completion support
- **Validation**: Built-in flag validation and required parameters
- **Industry Standard**: Used by kubectl, helm, docker, etc.

### 2. MCP Protocol Implementation

Since there's no official Go MCP SDK, we'll implement the core protocol:

```go
type MCPServer interface {
    ListTools(ctx context.Context) ([]Tool, error)
    CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolResult, error)
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

### 2. Tool Registry Pattern

```go
type ToolRegistry struct {
    tools map[string]Tool
    context *ToolContext
}

type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error)
}
```

### 3. Validation System

Instead of Zod, use struct tags with go-playground/validator:

```go
type ListDevicesRequest struct {
    Filter string `json:"filter,omitempty" validate:"omitempty,alpha"`
    Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}
```

### 4. Error Handling

Structured error types with proper context:

```go
type MCPError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}
```

## Performance Expectations

**Memory Usage:**

- TypeScript: ~50-100MB runtime
- Go: ~10-20MB runtime (5x improvement)

**Binary Size:**

- TypeScript: Node.js + dependencies (~100MB)
- Go: Single binary (~15-25MB)

**Startup Time:**

- TypeScript: ~500ms-1s
- Go: ~50-100ms (10x improvement)

**Runtime Performance:**

- JSON processing: 2-3x faster
- HTTP requests: Similar performance
- CLI operations: Similar performance

## Risk Mitigation

### Technical Risks

1. **MCP Protocol Compatibility**
   - Risk: Custom implementation may have subtle incompatibilities
   - Mitigation: Thorough testing against MCP Inspector and Claude Desktop

2. **Tool Behavior Changes**
   - Risk: Subtle differences in tool behavior
   - Mitigation: Comprehensive integration tests, side-by-side testing

3. **Performance Regressions**
   - Risk: Some operations may be slower than expected
   - Mitigation: Benchmarking against TypeScript version

### Business Risks

1. **User Migration**
   - Risk: Users may resist migration or experience issues
   - Mitigation: Maintain backward compatibility, provide clear migration guide

2. **Development Velocity**
   - Risk: Team productivity may decrease during transition
   - Mitigation: Gradual migration, maintain TypeScript version until Go is stable

## Success Criteria

### Functional Requirements

- [ ] All existing tools work identically to TypeScript version
- [ ] MCP protocol compatibility maintained
- [ ] Both stdio and HTTP server modes functional
- [ ] Docker and binary distribution working

### Non-Functional Requirements

- [ ] <20MB binary size
- [ ] <100ms startup time
- [ ] >90% test coverage
- [ ] Memory usage <20MB
- [ ] Zero security regressions

### User Experience

- [ ] Drop-in replacement for existing configurations
- [ ] Improved installation experience (single binary)
- [ ] Better error messages and logging
- [ ] No breaking changes in tool APIs

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Phase 1 | Week 1 | Go project foundation, basic CLI |
| Phase 2 | Week 1-2 | MCP server, Tailscale integration |
| Phase 3 | Week 2-3 | All tools migrated and functional |
| Phase 4 | Week 3 | Testing and quality assurance |
| Phase 5 | Week 3-4 | Packaging and distribution |

**Total Estimated Time:** 3-4 weeks for complete migration

## Next Steps

1. **Immediate:** Begin Phase 1 implementation
2. **Week 1:** Complete foundation and start core infrastructure
3. **Week 2:** Focus on tool migration
4. **Week 3:** Testing and quality assurance
5. **Week 4:** Release preparation and documentation

This migration plan ensures a systematic approach to converting the Tailscale MCP Server to Go while maintaining all functionality and improving performance characteristics.
