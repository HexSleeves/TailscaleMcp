# Integration Tests

This directory contains integration tests for the Tailscale MCP server that test end-to-end functionality with the Tailscale CLI.

## Running Integration Tests

### Run All Integration Tests

```bash
go test -tags=integration ./test/integration/...
```

### Run Integration Tests with Verbose Output

```bash
go test -tags=integration -v ./test/integration/...
```

### Run Specific Integration Test

```bash
go test -tags=integration -v ./test/integration/ -run TestCLIMethodFunctionality
```

### Skip Integration Tests (Default)

Integration tests are skipped by default when running:

```bash
go test ./...
```

This is because they are tagged with `//go:build integration` and require the `-tags=integration` flag.

## Test Structure

### Test Files

- `cli_integration_test.go` - Main CLI integration tests

### Test Categories

1. **CLI Method Functionality** - Tests individual CLI methods (GetVersion, Ping, IP, etc.)
2. **Security Validation Integration** - Tests security validation with actual command execution
3. **Execute Command Limits** - Tests context cancellation and command execution limits

### Test Environment

- Uses stub Tailscale binary to avoid requiring actual Tailscale installation
- Tests are hermetic and don't require network connectivity
- Validates command structure and error handling rather than actual Tailscale functionality

## Test Coverage

The integration tests cover:

- End-to-end command execution flow
- Security validation with actual CLI execution
- Error handling and response structure
- Context cancellation behavior
- Authentication key handling (ensuring secrets aren't logged)

## Continuous Integration

In CI environments, you can run integration tests with:

```bash
go test -tags=integration -short ./test/integration/...
```

The `-short` flag will skip longer-running tests within the integration suite.
