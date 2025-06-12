# Cobra CLI Migration

## Overview

Migrate from Go's built-in `flag` package to `github.com/spf13/cobra` for improved CLI experience and future extensibility.

## Current State (Phase 1)

Using basic `flag` package with manual help/version handling:

```go
var (
    serverMode  = flag.String("mode", "stdio", "Server mode: stdio or http")
    httpPort    = flag.Int("port", 8080, "HTTP server port")
    showHelp    = flag.Bool("help", false, "Show help message")
    showVersion = flag.Bool("version", false, "Show version information")
)
```

## Target State (Cobra)

Professional CLI with subcommands and enhanced UX:

```go
var rootCmd = &cobra.Command{
    Use:   "tailscale-mcp-server",
    Short: "Tailscale MCP Server",
    Long: `A Model Context Protocol server that provides seamless integration 
with Tailscale's CLI commands and REST API, enabling automated network 
management and monitoring through a standardized interface.`,
}

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the MCP server",
    Long:  "Start the Tailscale MCP server in stdio or HTTP mode",
    Run:   runServer,
}
```

## Migration Benefits

### 1. **Professional UX**
- Automatic help formatting with proper alignment
- Color support for better readability
- Consistent command structure following Unix conventions

### 2. **Subcommand Support**
- `tailscale-mcp-server serve` - Start server
- `tailscale-mcp-server version` - Show version info
- `tailscale-mcp-server config validate` - Validate configuration
- Future extensibility for debugging/testing commands

### 3. **Enhanced Features**
- Shell completion (bash, zsh, fish, PowerShell)
- Built-in flag validation
- Required vs optional parameters
- Flag aliases (e.g., `-m` and `--mode`)

### 4. **Industry Standard**
- Used by major tools: kubectl, helm, docker, gh, etc.
- Familiar UX for developers
- Extensive documentation and community support

## Implementation Plan

### 1. Add Cobra Dependency
```bash
go get github.com/spf13/cobra@latest
```

### 2. Create Command Structure
```
cmd/
├── root.go          # Root command definition
├── serve.go         # Serve command (main functionality)
├── version.go       # Version command
└── completion.go    # Shell completion
```

### 3. Migrate Flags
| Current Flag | Cobra Equivalent |
|-------------|------------------|
| `-mode` | `--mode` with validation |
| `-port` | `--port` with range validation |
| `-help` | Automatic with `-h`/`--help` |
| `-version` | Subcommand + root `--version` |

### 4. Command Structure
```bash
# Current
tailscale-mcp-server -mode=stdio
tailscale-mcp-server -mode=http -port=8080

# New (backward compatible)
tailscale-mcp-server serve --mode=stdio
tailscale-mcp-server serve --mode=http --port=8080

# Also support direct mode (backward compatibility)
tailscale-mcp-server --mode=stdio  # Auto-serve
```

## Implementation Details

### Root Command (cmd/root.go)
```go
var (
    cfgFile string
    verbose bool
)

var rootCmd = &cobra.Command{
    Use:   "tailscale-mcp-server",
    Short: "Tailscale MCP Server",
    Long: `A Model Context Protocol server that provides seamless integration 
with Tailscale's CLI commands and REST API.`,
    Version: getVersion(),
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        logger.Fatal("Command execution failed", "error", err)
    }
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
```

### Serve Command (cmd/serve.go)
```go
var (
    serverMode string
    httpPort   int
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the MCP server",
    Long:  "Start the Tailscale MCP server in stdio or HTTP mode",
    Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
    // Current server logic
}

func init() {
    rootCmd.AddCommand(serveCmd)
    
    serveCmd.Flags().StringVarP(&serverMode, "mode", "m", "stdio", 
        "Server mode (stdio|http)")
    serveCmd.Flags().IntVarP(&httpPort, "port", "p", 8080, 
        "HTTP server port (only used in http mode)")
        
    // Validation
    serveCmd.MarkFlagRequired("mode")
}
```

### Version Command (cmd/version.go)
```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show version information",
    Long:  "Display version, build info, and Go runtime details",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Tailscale MCP Server %s\n", getVersion())
        fmt.Printf("Built with %s\n", runtime.Version())
        fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

## Migration Steps

1. **Add Cobra dependency**
2. **Create command files** in `cmd/` directory
3. **Migrate flag handling** to Cobra flags
4. **Update main.go** to use `cmd.Execute()`
5. **Add shell completion** support
6. **Update documentation** and help text
7. **Test backward compatibility**

## Backward Compatibility

Maintain compatibility with existing usage:
- Support old flag syntax as aliases
- Preserve environment variable behavior
- Keep same exit codes and error messages

## Testing

- Test all command combinations
- Verify help output formatting
- Test shell completion
- Ensure environment variable precedence
- Validate error handling and messages

## Documentation Updates

- Update README.md with new command examples
- Add shell completion installation instructions
- Document subcommand usage
- Update Docker/deployment examples

This migration enhances the professional appearance and usability of the CLI while maintaining full backward compatibility.