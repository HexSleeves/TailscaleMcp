# Go Migration Tasks - Tailscale MCP Server

## Migration Status Summary

### ✅ Completed Components

- **Basic server architecture** (`internal/server/server.go`)
- **Configuration management** (`internal/config/config.go`)
- **Logging system** (`internal/logger/logger.go`)
- **MCP protocol types** (`pkg/mcp/types.go`)
- **Basic MCP server** (`pkg/mcp/server.go`)
- **Tailscale API client** (`internal/tailscale/api.go`)
- **Tailscale CLI wrapper** (`internal/tailscale/cli.go`)
- **CLI validation schemas** (`pkg/cli/schema.go`)
- **Command infrastructure** (`internal/cmd/`)
- **Cross-platform builds** (`Makefile`)

### ❌ Missing Critical Components

- **Tool registry system** (❗ CRITICAL)
- **Individual tool implementations** (❗ CRITICAL)
- **HTTP/Stdio server implementations** (❗ CRITICAL)
- **Comprehensive test coverage**
- **Integration with MCP protocol**

## Phase 1: Core Infrastructure (HIGH PRIORITY)

### Task 1.1: Implement Tool Registry System

**Priority**: CRITICAL
**Estimated Time**: 4-6 hours
**Files to Create/Modify**:

- `internal/tools/registry.go` - Tool registry implementation
- `internal/tools/types.go` - Tool interfaces and types
- `internal/tools/context.go` - Tool execution context

**Requirements**:

```go
type ToolRegistry struct {
    tools map[string]Tool
    api   *tailscale.APIClient
    cli   *tailscale.TailscaleCLI
}

type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResponse, error)
}

type ToolContext struct {
    API *tailscale.APIClient
    CLI *tailscale.TailscaleCLI
}
```

**Implementation Notes**:

- Use composition over inheritance pattern
- Implement tool registration system similar to TypeScript version
- Support for tool validation and error handling
- Context-aware tool execution

### Task 1.2: Complete HTTP/Stdio Server Implementations

**Priority**: CRITICAL
**Estimated Time**: 3-4 hours
**Files to Create/Modify**:

- `internal/server/stdio.go` - Complete stdio MCP server
- `internal/server/http.go` - Complete HTTP MCP server

**Requirements**:

- Full MCP protocol implementation
- Integration with tool registry
- Proper request/response handling
- Error propagation and logging
- Context cancellation support

### Task 1.3: Restructure MCP Server Interface

**Priority**: HIGH
**Estimated Time**: 2-3 hours
**Files to Modify**:

- `pkg/mcp/server.go` - Remove BasicMCPServer, implement proper interface
- `pkg/mcp/types.go` - Ensure all MCP types are complete

**Requirements**:

- Remove placeholder BasicMCPServer implementation
- Create proper MCP server interface
- Integrate with tool registry
- Support for tool discovery and execution

## Phase 2: Device Management Tools (HIGH PRIORITY)

### Task 2.1: Device Tools Implementation

**Priority**: HIGH
**Estimated Time**: 4-5 hours
**Files to Create**:

- `internal/tools/device/list_devices.go`
- `internal/tools/device/device_action.go`
- `internal/tools/device/manage_routes.go`

**Tool Specifications**:

#### list_devices

```go
type ListDevicesInput struct {
    IncludeRoutes bool `json:"includeRoutes" description:"Include route information for each device"`
}
```

- Use API client to fetch devices
- Format output with device metadata
- Optional route information display
- Handle authorization status indicators

#### device_action

```go
type DeviceActionInput struct {
    DeviceID string `json:"deviceId" description:"The ID of the device to act on"`
    Action   string `json:"action" description:"The action to perform (authorize|deauthorize|delete|expire-key)"`
}
```

- Implement all four actions using API client methods
- Proper error handling for each action type
- Confirmation for destructive operations

#### manage_routes

```go
type ManageRoutesInput struct {
    DeviceID string   `json:"deviceId" description:"The ID of the device"`
    Routes   []string `json:"routes" description:"Array of CIDR routes to manage"`
    Action   string   `json:"action" description:"Whether to enable or disable the routes"`
}
```

- CIDR validation using Go's net package
- Batch route operations
- Route conflict detection

### Task 2.2: Network Tools Implementation

**Priority**: HIGH
**Estimated Time**: 4-5 hours
**Files to Create**:

- `internal/tools/network/get_status.go`
- `internal/tools/network/connect.go`
- `internal/tools/network/disconnect.go`
- `internal/tools/network/ping.go`
- `internal/tools/network/version.go`

**Implementation Requirements**:

- Use CLI wrapper for network operations
- Proper input validation for all parameters
- Structured output formatting
- Error handling with CLI stderr capture

## Phase 3: Access Control & Security Tools (MEDIUM PRIORITY)

### Task 3.1: ACL and DNS Management

**Priority**: MEDIUM
**Estimated Time**: 5-6 hours
**Files to Create**:

- `internal/tools/acl/manage_acl.go`
- `internal/tools/acl/manage_dns.go`
- `internal/tools/acl/manage_keys.go`
- `internal/tools/acl/network_lock.go`
- `internal/tools/acl/policy_file.go`

**Implementation Requirements**:

- HuJSON format support for ACLs
- DNS configuration validation
- Auth key management with capabilities
- Network lock operations
- Policy testing functionality

## Phase 4: Administrative Tools (MEDIUM PRIORITY)

### Task 4.1: Admin Tools Implementation

**Priority**: MEDIUM
**Estimated Time**: 6-7 hours
**Files to Create**:

- `internal/tools/admin/tailnet_info.go`
- `internal/tools/admin/file_sharing.go`
- `internal/tools/admin/exit_nodes.go`
- `internal/tools/admin/webhooks.go`
- `internal/tools/admin/device_tags.go`

**Implementation Requirements**:

- Comprehensive tailnet information gathering
- File sharing toggle implementation
- Exit node management with route advertisement
- Webhook CRUD operations
- Device tagging with ACL integration

## Phase 5: Testing & Quality Assurance (HIGH PRIORITY)

### Task 5.1: Unit Test Implementation

**Priority**: HIGH
**Estimated Time**: 8-10 hours
**Files to Create**:

- `internal/tools/registry_test.go`
- `internal/tools/device/*_test.go`
- `internal/tools/network/*_test.go`
- `internal/tools/acl/*_test.go`
- `internal/tools/admin/*_test.go`

**Requirements**:

- Mock API client for testing
- Mock CLI for testing
- Input validation testing
- Error condition testing
- Tool registry testing

### Task 5.2: Integration Test Enhancement

**Priority**: MEDIUM
**Estimated Time**: 4-5 hours
**Files to Create/Modify**:

- `test/integration/tools_integration_test.go`
- `test/integration/server_integration_test.go`

**Requirements**:

- End-to-end tool testing with real Tailscale CLI
- Server startup/shutdown testing
- MCP protocol compliance testing

## Phase 6: Security & Performance Optimization (MEDIUM PRIORITY)

### Task 6.1: Security Hardening

**Priority**: MEDIUM
**Estimated Time**: 3-4 hours
**Files to Review/Modify**:

- All tool implementations
- `internal/tailscale/cli.go` validation
- Input sanitization across all tools

**Requirements**:

- Command injection prevention validation
- Input length limits enforcement
- Buffer overflow protection
- Credential sanitization in logs

### Task 6.2: Performance Optimization

**Priority**: LOW
**Estimated Time**: 2-3 hours
**Activities**:

- Memory usage optimization
- Concurrent tool execution
- Response time optimization
- Large device list handling

## Phase 7: Documentation & Deployment (LOW PRIORITY)

### Task 7.1: Documentation Updates

**Priority**: LOW
**Estimated Time**: 2-3 hours
**Files to Update**:

- `README.md`
- `CLAUDE.md`
- API documentation

### Task 7.2: Docker & Deployment

**Priority**: LOW
**Estimated Time**: 2-3 hours
**Files to Review**:

- `Dockerfile`
- `docker-compose.yml`
- Build scripts

## Critical Migration Path

### Immediate Priority (Week 1)

1. **Task 1.1**: Tool Registry System (CRITICAL)
2. **Task 1.2**: HTTP/Stdio Server Implementations (CRITICAL)
3. **Task 1.3**: MCP Server Interface (HIGH)
4. **Task 2.1**: Device Management Tools (HIGH)

### Secondary Priority (Week 2)

5. **Task 2.2**: Network Management Tools (HIGH)
6. **Task 5.1**: Unit Test Implementation (HIGH)
7. **Task 3.1**: ACL and DNS Management (MEDIUM)

### Final Priority (Week 3)

8. **Task 4.1**: Administrative Tools (MEDIUM)
9. **Task 5.2**: Integration Tests (MEDIUM)
10. **Task 6.1**: Security Hardening (MEDIUM)

## Migration Validation Checklist

### Functional Parity

- [ ] All 18 tools implemented and working
- [ ] Input/output schemas match TypeScript version
- [ ] Error handling behavior identical
- [ ] MCP protocol compliance verified

### Performance Targets

- [ ] CLI operations complete within 30s timeout
- [ ] API operations complete within 30s timeout
- [ ] Memory usage optimized for large device lists
- [ ] Concurrent tool execution capability

### Security Compliance

- [ ] Command injection prevention validated
- [ ] Input sanitization comprehensive
- [ ] Credential handling secure
- [ ] No sensitive information in logs

### Testing Coverage

- [ ] Unit tests for all tools (>80% coverage)
- [ ] Integration tests pass with real Tailscale CLI
- [ ] Security tests validate injection prevention
- [ ] Performance benchmarks meet targets

## Implementation Guidelines

### Code Organization

```
internal/tools/
├── registry.go          # Tool registry implementation
├── types.go            # Common tool interfaces
├── context.go          # Tool execution context
├── device/             # Device management tools
│   ├── list_devices.go
│   ├── device_action.go
│   └── manage_routes.go
├── network/            # Network management tools
│   ├── get_status.go
│   ├── connect.go
│   ├── disconnect.go
│   ├── ping.go
│   └── version.go
├── acl/               # Access control tools
│   ├── manage_acl.go
│   ├── manage_dns.go
│   ├── manage_keys.go
│   ├── network_lock.go
│   └── policy_file.go
└── admin/             # Administrative tools
    ├── tailnet_info.go
    ├── file_sharing.go
    ├── exit_nodes.go
    ├── webhooks.go
    └── device_tags.go
```

### Development Process

1. Implement tool interface and registration
2. Create individual tool implementations
3. Add comprehensive unit tests
4. Integrate with MCP server
5. Test with integration suite
6. Security review and hardening
7. Performance optimization
8. Documentation updates

This migration plan ensures systematic completion of the Go implementation while maintaining compatibility with the existing TypeScript version and achieving production readiness.
