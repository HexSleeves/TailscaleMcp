package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/server"
	"github.com/hexsleeves/tailscale-mcp-server/version"
)

var (
	serverMode    string
	httpPort      int
	cachedVersion string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the Tailscale MCP server in stdio or HTTP mode.

The server provides Model Context Protocol integration with Tailscale,
allowing automated network management through standardized interfaces.

Modes:
  stdio  - Standard input/output communication (default, for MCP clients)
  http   - HTTP server mode (for testing and development)

Examples:
  # Start in stdio mode (default)
  tailscale-mcp-server serve

  # Start in HTTP mode on custom port
  tailscale-mcp-server serve --mode=http --port=9000

  # With verbose logging
  tailscale-mcp-server serve --verbose`,
	Run: runServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Command-specific flags
	serveCmd.Flags().StringVarP(&serverMode, "mode", "m", "stdio", "Server mode (stdio|http)")
	serveCmd.Flags().IntVarP(&httpPort, "port", "p", 8080, "HTTP server port (only used in http mode)")

	// Flag validation
	serveCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Validate server mode
		if serverMode != "stdio" && serverMode != "http" {
			return fmt.Errorf("invalid server mode: must be 'stdio' or 'http'")
		}

		// Validate port range
		if serverMode == "http" {
			if httpPort < 1 || httpPort > 65535 {
				return fmt.Errorf("invalid port: must be between 1 and 65535")
			}
		}

		return nil
	}

	// Cache the version string once during package initialization
	cachedVersion = version.Short()
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize logger with verbose flag consideration
	logLevel := cfg.LogLevel
	if verbose {
		logLevel = 0 // Debug level when verbose
	}

	if err := logger.Initialize(logLevel, cfg.LogFile); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	// Create server
	tailscaleMCPServer, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", "error", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start server
	logger.Info("Starting Tailscale MCP Server",
		"mode", serverMode,
		"version", cachedVersion,
		"verbose", verbose)

	var serverErr error
	switch serverMode {
	case "stdio":
		serverErr = tailscaleMCPServer.StartStdio(ctx)
	case "http":
		serverErr = tailscaleMCPServer.StartHTTP(ctx, httpPort)
	default:
		logger.Fatal("Invalid server mode", "mode", serverMode, "valid_modes", []string{"stdio", "http"})
	}

	if serverErr != nil {
		logger.Fatal("Server error", "error", serverErr)
	}

	logger.Info("Server stopped")
}
