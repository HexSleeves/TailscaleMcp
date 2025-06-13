package server

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
	"github.com/hexsleeves/tailscale-mcp-server/pkg/mcp"
)

// ServerOption configures the TailscaleMCPServer (functional options pattern)
type ServerOption func(*TailscaleMCPServer) error

// TailscaleMCPServer implements the main MCP server logic
// Follows Go idioms: composition over inheritance, interfaces for flexibility
type TailscaleMCPServer struct {
	config    *config.Config
	api       *tailscale.APIClient
	mcpServer mcp.Server
}

// WithCustomMCPServer allows injecting a custom MCP server implementation
func WithCustomMCPServer(server mcp.Server) ServerOption {
	return func(s *TailscaleMCPServer) error {
		s.mcpServer = server
		return nil
	}
}

// New creates a new server instance using Go best practices
func New(cfg *config.Config, opts ...ServerOption) (*TailscaleMCPServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Initializing server",
		"config", cfg.SanitizedCopy(),
		"has_api_credentials", cfg.HasAPICredentials())

	// Initialize Tailscale API client
	api := tailscale.NewAPIClient(cfg)

	// Create server with default MCP implementation
	server := &TailscaleMCPServer{
		config: cfg,
		api:    api,
		mcpServer: &mcp.BasicMCPServer{
			Api: api,
		},
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(server); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}

	return server, nil
}

// StartStdio starts the server in stdio mode (Go way: create and start in one call)
func (s *TailscaleMCPServer) StartStdio(ctx context.Context) error {
	logger.Info("Starting stdio MCP server")

	server := NewStdioServer(s.mcpServer)
	return server.Start(ctx)
}

// StartHTTP starts the server in HTTP mode (Go way: create and start in one call)
func (s *TailscaleMCPServer) StartHTTP(ctx context.Context, port int) error {
	logger.Info("Starting HTTP MCP server", "port", port)

	server := NewHTTPServer(s.mcpServer, port)
	return server.Start(ctx)
}

// Shutdown gracefully shuts down the server
func (s *TailscaleMCPServer) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down TailscaleMCPServer")

	if err := s.mcpServer.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down MCP server", "error", err)
		return err
	}

	return nil
}
