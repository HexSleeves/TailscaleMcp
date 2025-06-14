# Tailscale MCP Server - Production Requirements

## Overview

The Tailscale MCP Server is a production-ready Model Context Protocol (MCP) server that provides comprehensive Tailscale network management capabilities through both CLI commands and REST API operations. This document outlines the complete functionality that must be implemented in the Go migration.

## Core Architecture Requirements

### 1. Server Modes

- **stdio MCP Server**: Primary mode for MCP client communication
- **HTTP MCP Server**: Development/testing mode with REST endpoints
- **Dual Protocol Support**: Both CLI and API operations for maximum flexibility

### 2. Configuration Management

- Environment variable configuration with validation
- Support for custom Tailscale CLI paths
- Configurable logging levels (0-3: debug, info, warn, error)
- Optional file logging with MCP_SERVER_LOG_FILE
- API credentials validation (TAILSCALE_API_KEY, TAILSCALE_TAILNET)

### 3. Error Handling & Security

- Comprehensive input validation and sanitization
- Command injection prevention for CLI operations
- Structured error responses in MCP format
- Graceful degradation when API credentials unavailable
- Timeout handling for all operations
- Buffer size limits for CLI output

## Tool Registry Requirements

### Device Management Tools

#### 1. list_devices

- **Purpose**: List all devices in the Tailscale network
- **Input**: `includeRoutes` (boolean, optional) - Include route information
- **Features**:
  - Display device metadata (name, hostname, ID, OS, addresses)
  - Show authorization status with visual indicators
  - Include advertised/enabled routes when requested
  - Handle large device lists efficiently

#### 2. device_action

- **Purpose**: Perform administrative actions on devices
- **Actions**: authorize, deauthorize, delete, expire-key
- **Input**: deviceId (string), action (enum)
- **Features**:
  - Atomic operations with rollback capability
  - Confirmation for destructive actions
  - Detailed success/failure reporting

#### 3. manage_routes

- **Purpose**: Enable/disable subnet routes for devices
- **Input**: deviceId, routes (array), action (enable/disable)
- **Features**:
  - CIDR validation
  - Batch route operations
  - Route conflict detection

### Network Management Tools

#### 4. get_network_status

- **Purpose**: Get current network status via CLI
- **Input**: format (json/summary, optional)
- **Features**:
  - JSON format for programmatic access
  - Human-readable summary format
  - Peer connectivity information
  - Exit node status indication
  - MagicDNS status

#### 5. connect_network

- **Purpose**: Connect to Tailscale network with options
- **Input**: acceptRoutes, acceptDNS, hostname, advertiseRoutes, authKey, loginServer
- **Features**:
  - Flexible connection options
  - Secure auth key handling via environment variables
  - Custom coordination server support
  - Route advertisement on connection

#### 6. disconnect_network

- **Purpose**: Safely disconnect from Tailscale network
- **Features**:
  - Graceful shutdown
  - State cleanup

#### 7. ping_peer

- **Purpose**: Test connectivity to peer devices
- **Input**: target (hostname/IP), count (1-100)
- **Features**:
  - Target validation (hostname, IP, node name)
  - Configurable ping count
  - Latency reporting

#### 8. get_version

- **Purpose**: Get Tailscale version information
- **Features**:
  - Version compatibility checking
  - Feature availability detection

### Access Control & Security Tools

#### 9. manage_acl

- **Purpose**: Manage Access Control Lists
- **Operations**: get, update, validate
- **Features**:
  - HuJSON format support
  - ACL validation before application
  - Backup/restore capability
  - Group and tag owner management

#### 10. manage_dns

- **Purpose**: Manage DNS configuration
- **Operations**: get_nameservers, set_nameservers, get_preferences, set_preferences, get_searchpaths, set_searchpaths
- **Features**:
  - Custom DNS server configuration
  - MagicDNS toggle
  - Search domain management
  - DNS validation

#### 11. manage_keys

- **Purpose**: Authentication key management
- **Operations**: list, create, delete
- **Features**:
  - Key capability configuration (reusable, ephemeral, preauthorized)
  - Tag-based key creation
  - Expiry management
  - Secure key display

#### 12. manage_network_lock

- **Purpose**: Network lock (key authority) management
- **Operations**: status, enable, disable, add_key, remove_key, list_keys
- **Features**:
  - Enhanced security through key authority
  - Public key management
  - Trust relationship configuration

#### 13. manage_policy_file

- **Purpose**: Policy file and access testing
- **Operations**: get, update, test_access
- **Features**:
  - HuJSON policy format
  - Access rule testing
  - Policy validation

### Administrative Tools

#### 14. get_tailnet_info

- **Purpose**: Get comprehensive tailnet information
- **Input**: includeDetails (boolean, optional)
- **Features**:
  - Basic tailnet metadata
  - Advanced configuration details
  - Security status overview

#### 15. manage_file_sharing

- **Purpose**: File sharing configuration
- **Operations**: get_status, enable, disable
- **Features**:
  - Network-wide file sharing toggle
  - Status reporting

#### 16. manage_exit_nodes

- **Purpose**: Exit node management and routing
- **Operations**: list, set, clear, advertise, stop_advertising
- **Features**:
  - Exit node discovery
  - Route advertisement (0.0.0.0/0, ::/0)
  - Exit node switching
  - Route management

#### 17. manage_webhooks

- **Purpose**: Webhook management for events
- **Operations**: list, create, delete, test
- **Features**:
  - Event subscription management
  - Webhook endpoint validation
  - Test delivery capability
  - Secret management

#### 18. manage_device_tags

- **Purpose**: Device tagging for organization
- **Operations**: get_tags, set_tags, add_tags, remove_tags
- **Features**:
  - Tag-based device organization
  - ACL integration
  - Bulk tagging operations

## Data Types & Schemas

### Core Types

- **TailscaleDevice**: Complete device representation with metadata
- **TailscaleNetworkStatus**: Network status from CLI
- **TailscaleCLIStatus**: Structured CLI status output
- **ACLConfig**: Access control configuration
- **DNSConfig**: DNS configuration
- **AuthKey**: Authentication key with capabilities

### Validation Requirements

- CIDR route validation using Go's net package
- Hostname/IP validation with regex patterns
- String length limits (hostnames: 253 chars, args: 1000 chars)
- Command injection prevention
- Buffer size limits (10MB output)

## Performance Requirements

### Response Times

- CLI operations: < 30 seconds timeout
- API operations: < 30 seconds timeout
- Tool execution: < 2 minutes total
- Network status: < 5 seconds

### Scalability

- Support for 1000+ devices
- Efficient JSON parsing for large responses
- Memory-conscious CLI output buffering
- Concurrent tool execution capability

## Security Requirements

### Input Validation

- Comprehensive argument sanitization
- Command whitelist enforcement
- Path traversal prevention
- Buffer overflow protection

### Credential Management

- Environment variable credential injection
- No credential logging
- Secure auth key passing via environment
- API key validation

### Error Handling

- No sensitive information in error messages
- Structured error responses
- Graceful failure modes
- Attack surface minimization

## Integration Requirements

### MCP Protocol Compliance

- Full MCP 2024-11-05 protocol support
- Tool discovery and registration
- Structured request/response handling
- Error propagation

### Tailscale Integration

- CLI command execution with Windows console hiding
- REST API integration with proper authentication
- Version compatibility checking
- Feature detection

### Development Experience

- Comprehensive test coverage (unit + integration)
- Docker containerization support
- Cross-platform compilation (Linux, macOS, Windows)
- Development tooling (linting, formatting, testing)

## Quality Assurance

### Testing Strategy

- Unit tests for all tool implementations
- Integration tests requiring Tailscale CLI
- Security tests for injection prevention
- Performance benchmarks
- Error condition testing

### Code Quality

- Go idioms and best practices
- Comprehensive error handling
- Structured logging
- Documentation coverage
- Security review compliance

## Deployment Requirements

### Packaging

- Single binary deployment
- Docker image with minimal attack surface
- Cross-platform support
- Version information embedding

### Configuration

- Environment-based configuration
- Validation on startup
- Graceful degradation
- Health check endpoints (HTTP mode)

### Monitoring

- Structured logging output
- Performance metrics
- Error rate tracking
- Connection status monitoring

## Migration Compatibility

### Backward Compatibility

- Identical tool interfaces
- Same input/output schemas
- Consistent error handling
- Feature parity with TypeScript implementation

### Transition Support

- Side-by-side operation capability
- Configuration migration utilities
- Testing harness for validation
- Documentation updates

This production requirements document serves as the definitive specification for the complete Tailscale MCP Server implementation in Go, ensuring feature parity with the TypeScript version while leveraging Go's strengths in performance, security, and maintainability.
