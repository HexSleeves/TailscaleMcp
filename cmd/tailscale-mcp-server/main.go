package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/server"
)

func main() {
	var (
		serverMode  = flag.String("mode", "stdio", "Server mode: stdio or http")
		httpPort    = flag.Int("port", 8080, "HTTP server port (only used in http mode)")
		showHelp    = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVersion {
		printVersion()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := logger.Initialize(cfg.LogLevel, cfg.LogFile); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", "error", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start server
	logger.Info("Starting Tailscale MCP Server",
		"mode", *serverMode,
		"version", getVersion())

	var serverErr error
	switch *serverMode {
	case "stdio":
		serverErr = srv.StartStdio(ctx)
	case "http":
		serverErr = srv.StartHTTP(ctx, *httpPort)
	default:
		logger.Fatal("Invalid server mode", "mode", *serverMode)
	}

	if serverErr != nil {
		logger.Fatal("Server error", "error", serverErr)
	}

	logger.Info("Server stopped")
}

func printHelp() {
	fmt.Println("Tailscale MCP Server")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("    tailscale-mcp-server [OPTIONS]")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("    -mode string     Server mode: stdio or http (default: stdio)")
	fmt.Println("    -port int        HTTP server port, only used in http mode (default: 8080)")
	fmt.Println("    -help            Show this help message")
	fmt.Println("    -version         Show version information")
	fmt.Println("")
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("    TAILSCALE_API_KEY        Tailscale API key (required for API operations)")
	fmt.Println("    TAILSCALE_TAILNET        Tailnet name (required for API operations)")
	fmt.Println("    TAILSCALE_API_BASE_URL   Custom API base URL (optional)")
	fmt.Println("    LOG_LEVEL                Logging level: 0=debug, 1=info, 2=warn, 3=error (default: 1)")
	fmt.Println("    MCP_SERVER_LOG_FILE      Log file path (optional)")
}

func printVersion() {
	fmt.Printf("Tailscale MCP Server %s\n", getVersion())
	fmt.Println("Built with Go")
}

func getVersion() string {
	// This will be replaced during build with actual version
	return "dev"
}
