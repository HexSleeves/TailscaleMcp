package server

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/pkg/mcp"
)

// Server implements the main MCP server logic
type Server struct {
	config    *config.Config
	mcpServer mcp.Server
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Initializing server",
		"config", cfg.SanitizedCopy(),
		"has_api_credentials", cfg.HasAPICredentials())

	// TODO: Initialize tool registry and MCP server implementation

	return &Server{
		config: cfg,
	}, nil
}

// StartStdio starts the server in stdio mode
func (s *Server) StartStdio(ctx context.Context) error {
	logger.Info("Starting stdio MCP server")

	// TODO: Implement stdio server
	<-ctx.Done()
	logger.Info("Stdio server shutting down")

	return nil
}

// StartHTTP starts the server in HTTP mode
func (s *Server) StartHTTP(ctx context.Context, port int) error {
	logger.Info("Starting HTTP MCP server", "port", port)

	// TODO: Implement HTTP server
	<-ctx.Done()
	logger.Info("HTTP server shutting down")

	return nil
}
