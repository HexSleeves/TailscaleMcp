# Tailscale MCP Server

A modern [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that provides seamless integration with Tailscale's CLI commands and REST API, enabling automated network management and monitoring through a standardized interface.

## Features

- **MCP Protocol Support**: Full implementation of MCP 2024-11-05 specification
- **Dual Interface**: Both CLI and REST API integration with Tailscale
- **Comprehensive Tools**: Device management, network operations, ACL management, and administrative functions
- **Multiple Server Modes**: stdio (for MCP clients) and HTTP (for testing/development)
- **Robust Architecture**: Built with Go best practices, comprehensive error handling, and structured logging
- **Docker Support**: Multi-stage Docker builds with security best practices
- **Cross-Platform**: Supports Linux, macOS, and Windows

## Quick Start

### Installation

#### From Release

```bash
# Download latest release
wget https://github.com/hexsleeves/tailscale-mcp-server/releases/latest/download/tailscale-mcp-server-linux-amd64.tar.gz
tar -xzf tailscale-mcp-server-linux-amd64.tar.gz
chmod +x tailscale-mcp-server
```

#### From Source

```bash
git clone https://github.com/hexsleeves/tailscale-mcp-server.git
cd tailscale-mcp-server
make build
```

#### Using Docker

```bash
docker run -e TAILSCALE_API_KEY=your_key -e TAILSCALE_TAILNET=your_tailnet \
  hexsleeves/tailscale-mcp-server:latest serve
```

### Configuration

Set the required environment variables:

```bash
export TAILSCALE_API_KEY="your_api_key_here"
export TAILSCALE_TAILNET="your_tailnet_name"
```

Optional configuration:

```bash
export TAILSCALE_API_BASE_URL="https://api.tailscale.com"  # Custom API URL
export LOG_LEVEL=1                                         # 0=debug, 1=info, 2=warn, 3=error
export MCP_SERVER_LOG_FILE="/var/log/tailscale-mcp.log"   # Log file path
```

### Usage

#### MCP Client Integration (stdio mode)

```bash
tailscale-mcp-server serve
```

#### HTTP Server Mode (for testing)

```bash
tailscale-mcp-server serve --mode=http --port=8080
```

#### Version Information

```bash
tailscale-mcp-server version
```

## Available Tools

### Device Management
- List all devices in your tailnet
- Get device status and information
- Enable/disable devices

### Network Operations
- Ping devices on your tailnet
- Test connectivity
- Get IP information
- View routing information

### Administrative Functions
- Get Tailscale status
- Login/logout operations
- Bring Tailscale up/down
- Version information

### ACL Management
- View current ACL policies
- Update ACL configurations
- Validate ACL syntax
- Test ACL rules

## MCP Client Configuration

### Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "tailscale": {
      "command": "tailscale-mcp-server",
      "args": ["serve"],
      "env": {
        "TAILSCALE_API_KEY": "your_api_key_here",
        "TAILSCALE_TAILNET": "your_tailnet_name"
      }
    }
  }
}
```

### Continue.dev

```json
{
  "mcpServers": [
    {
      "name": "tailscale",
      "command": "tailscale-mcp-server",
      "args": ["serve"],
      "env": {
        "TAILSCALE_API_KEY": "your_api_key_here",
        "TAILSCALE_TAILNET": "your_tailnet_name"
      }
    }
  ]
}
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (for linting)

### Development Commands

```bash
make build              # Build the binary
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make lint               # Run linter
make fmt                # Format code
make clean              # Clean build artifacts
```

### Project Structure

```
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── cli/               # CLI commands
│   ├── config/            # Configuration management
│   ├── logger/            # Logging utilities
│   ├── mcp/               # MCP protocol implementation
│   ├── server/            # Server implementations
│   ├── tailscale/         # Tailscale integration
│   └── tools/             # MCP tools
├── pkg/                   # Public packages
│   └── schema/            # Data schemas
├── api/                   # API specifications
├── docs/                  # Documentation
├── examples/              # Usage examples
├── test/                  # Test files
└── deployments/           # Deployment configurations
```

## Documentation

- [API Reference](api/mcp-spec.md)
- [Configuration Guide](docs/README.md)
- [Examples](examples/README.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/hexsleeves/tailscale-mcp-server/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hexsleeves/tailscale-mcp-server/discussions)
- **Documentation**: [docs/](docs/)

## Acknowledgments

- [Tailscale](https://tailscale.com) for their excellent VPN service and APIs
- [Model Context Protocol](https://modelcontextprotocol.io) for the standardized interface
- The Go community for excellent tooling and libraries
