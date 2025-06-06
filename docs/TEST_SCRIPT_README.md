# Tailscale MCP Server Test Script

A comprehensive test script that allows you to interact with the Tailscale MCP server locally without requiring an AI assistant. This is perfect for development, testing, and debugging.

## 🚀 Quick Start

### Prerequisites

1. **Build the server first:**

   ```bash
   npm run build
   ```

2. **Set up environment variables (optional but recommended):**

   ```bash
   export TAILSCALE_API_KEY="your-api-key-here"
   export TAILSCALE_TAILNET="your-tailnet-name"
   ```

### Running the Test Script

```bash
# Make executable (first time only)
chmod +x test-mcp-server.js

# Run the script
node test-mcp-server.js
```

## 📋 Features

### Interactive CLI Menu

- **User-friendly interface** with color-coded output
- **Environment validation** - checks for API keys and configuration
- **Error handling** with clear error messages
- **Graceful shutdown** with proper cleanup

### Pre-built Tool Tests

1. **List all available tools** - See what tools are available
2. **Get Tailscale version** - Test basic CLI functionality
3. **Get network status** - View current network state (JSON or summary format)
4. **List devices** - Show all devices in your tailnet
5. **Ping a peer** - Test connectivity to other devices
6. **Get tailnet info** - View detailed network information
7. **Custom tool call** - Call any tool with custom parameters
8. **Show examples** - View example tool calls with parameters

### Advanced Features

- **Real-time server communication** via JSON-RPC over stdio
- **Automatic server startup** and management
- **Response formatting** with syntax highlighting
- **Parameter validation** and input assistance
- **Timeout handling** for long-running operations

## 🛠️ Usage Examples

### Basic Operations

#### 1. Get Tailscale Version

```text
Select option: 2
```

Output:

```text
✅ Tool Result:
Tailscale version information:

1.84.1
  tailscale commit: 1b829929a79c00c264bed731145e0c817e4ca452
  other commit: 1343260d13748e82f6ca2c3d92e870630faa8730
  go version: go1.24.2
```

#### 2. Network Status (Summary Format)

```text
Select option: 3
Format (json/summary) [summary]: summary
```

#### 3. List All Devices

```text
Select option: 4
```

### Advanced Operations

#### 4. Custom Tool Call Example

```text
Select option: 7
Enter tool name: device_action
Enter arguments as JSON: {"deviceId": "your-device-id", "action": "authorize"}
```

#### 5. Ping a Peer

```text
Select option: 5
Enter target hostname or IP: peer-hostname
Number of pings [4]: 3
```

## 🔧 Environment Configuration

### Required for API Operations

```bash
# Your Tailscale API key (get from https://login.tailscale.com/admin/settings/keys)
export TAILSCALE_API_KEY="tskey-api-xxxxx"

# Your tailnet name (optional, defaults to current user's tailnet)
export TAILSCALE_TAILNET="your-org.tailscale.net"
```

### CLI-Only Operations

Some tools work without API keys:

- `get_version`
- `get_network_status`
- `connect_network`
- `disconnect_network`
- `ping_peer`

### API-Required Operations

These tools require `TAILSCALE_API_KEY`:

- `list_devices`
- `device_action`
- `manage_routes`
- `get_tailnet_info`
- Most admin and management tools

## 📚 Example Tool Calls

### Device Management

```json
{
  "name": "device_action",
  "arguments": {
    "deviceId": "device-id-here",
    "action": "authorize"
  }
}
```

### Route Management

```json
{
  "name": "manage_routes",
  "arguments": {
    "deviceId": "device-id-here",
    "routes": ["192.168.1.0/24", "10.0.0.0/8"],
    "action": "enable"
  }
}
```

### Network Connection

```json
{
  "name": "connect_network",
  "arguments": {
    "acceptRoutes": true,
    "acceptDNS": true,
    "hostname": "my-custom-hostname"
  }
}
```

### ACL Management

```json
{
  "name": "manage_acl",
  "arguments": {
    "operation": "get"
  }
}
```

## 🐛 Troubleshooting

### Common Issues

#### 1. "Server build not found"

```bash
# Solution: Build the server first
npm run build
```

#### 2. "Server initialization timeout"

```bash
# Check if Tailscale is installed and accessible
tailscale version

# Check server logs for errors
npm run dev
```

#### 3. "API operations fail"

```bash
# Set your API key
export TAILSCALE_API_KEY="your-key-here"

# Verify the key works
curl -H "Authorization: Bearer $TAILSCALE_API_KEY" \
     https://api.tailscale.com/api/v2/tailnet/-/devices
```

#### 4. "Tool not found" errors

```bash
# List available tools first
Select option: 1

# Check tool name spelling
# Tool names are case-sensitive
```

### Debug Mode

For detailed debugging, you can modify the script to show server logs:

```javascript
// In the startServer() method, add:
this.serverProcess.stdout.on("data", (data) => {
  console.log("Server:", data.toString());
});

this.serverProcess.stderr.on("data", (data) => {
  console.log("Server Error:", data.toString());
});
```

## 🔍 Development Tips

### Testing New Tools

1. Add your tool to the appropriate module in `src/tools/`
2. Rebuild: `npm run build`
3. Use option 7 (Custom tool call) to test
4. Check the tool list (option 1) to verify registration

### Debugging Tool Issues

1. Use the custom tool call feature to test specific parameters
2. Check the formatted response for error details
3. Verify your tool's input schema matches the arguments
4. Use the examples (option 8) as reference

### Performance Testing

- The script includes 30-second timeouts for tool calls
- Monitor server startup time (should be < 10 seconds)
- Test with various parameter combinations

## 📖 Script Architecture

### Key Components

1. **TailscaleMCPTester Class**

   - Manages server lifecycle
   - Handles MCP communication
   - Provides CLI interface

2. **Server Communication**

   - JSON-RPC 2.0 protocol over stdio
   - Request/response correlation
   - Timeout and error handling

3. **User Interface**
   - Color-coded output
   - Interactive menus
   - Input validation

### Communication Flow

```text
Test Script → JSON-RPC Request → MCP Server → Tailscale CLI/API → Response → Formatted Output
```

## 🚀 Advanced Usage

### Batch Testing

You can modify the script to run automated tests:

```javascript
// Add to the TailscaleMCPTester class
async runBatchTests() {
  const tests = [
    { name: 'get_version', args: {} },
    { name: 'get_network_status', args: { format: 'summary' } },
    // Add more tests...
  ];

  for (const test of tests) {
    console.log(`Testing ${test.name}...`);
    const result = await this.callTool(test.name, test.args);
    this.formatToolResponse(result);
  }
}
```

### Integration with CI/CD

The script can be adapted for automated testing in CI/CD pipelines:

```bash
# Example CI script
export TAILSCALE_API_KEY="$CI_TAILSCALE_API_KEY"
node test-mcp-server.js --batch --test-suite=basic
```

This test script provides a comprehensive way to interact with and test your Tailscale MCP server locally, making development and debugging much more efficient!
