// tailscale-mcp-server/internal/server/server.go
package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/mcp"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tools"
)

// ServerOption configures the TailscaleMCPServer (functional options pattern)
type ServerOption func(*TailscaleMCPServer) error

// TailscaleMCPServer implements the main MCP server logic
// Follows Go idioms: composition over inheritance, interfaces for flexibility
type TailscaleMCPServer struct {
	mu        sync.RWMutex
	config    *config.Config
	api       *tailscale.APIClient
	cli       *tailscale.TailscaleCLI
	registry  *tools.ToolRegistry
	mcpServer mcp.Server
	running   bool
}

// WithCustomMCPServer allows injecting a custom MCP server implementation
func WithCustomMCPServer(server mcp.Server) ServerOption {
	return func(s *TailscaleMCPServer) error {
		if server == nil {
			return fmt.Errorf("custom MCP server cannot be nil")
		}

		s.mcpServer = server
		return nil
	}
}

// WithCustomRegistry allows injecting a custom tool registry
func WithCustomRegistry(registry *tools.ToolRegistry) ServerOption {
	return func(s *TailscaleMCPServer) error {
		if registry == nil {
			return fmt.Errorf("custom registry cannot be nil")
		}

		s.registry = registry
		return nil
	}
}

// New creates a new server instance using Go best practices
func New(cfg *config.Config, opts ...ServerOption) (*TailscaleMCPServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration must not be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Initializing server",
		"config", cfg.SanitizedCopy(),
		"has_api_credentials", cfg.HasAPICredentials())

	// Initialize Tailscale clients
	api := tailscale.NewAPIClient(cfg)
	cli, err := tailscale.NewTailscaleCLI()
	if err != nil {
		return nil, fmt.Errorf("failed to create tailscale cli: %w", err)
	}

	// Create tool registry
	registry := tools.NewToolRegistry(api, cli)

	// Create server with default MCP implementation
	server := &TailscaleMCPServer{
		config:   cfg,
		api:      api,
		cli:      cli,
		registry: registry,
		mcpServer: mcp.NewMCPServer(
			registry,
			"tailscale-mcp-server",
			"0.1.0", // TODO: Get version from config or build flags
		),
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(server); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}

	return server, nil
}

// StartStdio starts the server in stdio mode
func (s *TailscaleMCPServer) StartStdio(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	logger.Info("Starting stdio MCP server")

	server := NewStdioServer(s.mcpServer)
	return server.Start(ctx)
}

// StartHTTP starts the server in HTTP mode
func (s *TailscaleMCPServer) StartHTTP(ctx context.Context, port int) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	logger.Info("Starting HTTP MCP server", "port", port)

	server := NewHTTPServer(s.mcpServer, port)
	return server.Start(ctx)
}

// Shutdown gracefully shuts down the server
func (s *TailscaleMCPServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil // Already stopped
	}

	logger.Info("Shutting down TailscaleMCPServer")

	// Shutdown MCP server
	if err := s.mcpServer.Shutdown(ctx, &mcp.ShutdownRequest{}); err != nil {
		logger.Error("Error shutting down MCP server", "error", err)
		return fmt.Errorf("failed to shutdown MCP server: %w", err)
	}

	// Close tool registry
	if err := s.registry.Close(); err != nil {
		logger.Error("Error closing tool registry", "error", err)
		return fmt.Errorf("failed to close tool registry: %w", err)
	}

	s.running = false
	return nil
}

// IsRunning returns true if the server is currently running
func (s *TailscaleMCPServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Config returns a copy of the server configuration
func (s *TailscaleMCPServer) Config() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *s.config
	return &configCopy
}

// ToolCount returns the number of registered tools
func (s *TailscaleMCPServer) ToolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registry.Count()
}
