package mcp

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// BasicMCPServer provides a basic implementation of the MCP server interface
// TODO: This should be replaced with a proper tool registry implementation
type BasicMCPServer struct {
	Api *tailscale.APIClient
}

// Initialize handles MCP initialization
func (s *BasicMCPServer) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	logger.Info("MCP server initialized", "client_info", req.ClientInfo)

	return &InitializeResponse{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "tailscale-mcp-server",
			Version: "0.1.0",
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
	}, nil
}

// ListTools returns available tools
func (s *BasicMCPServer) ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error) {
	// TODO: Return actual tools from tool registry
	return &ListToolsResponse{
		Tools: []Tool{
			{
				Name:        "tailscale_status",
				Description: "Get Tailscale network status",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
	}, nil
}

// CallTool executes a tool
func (s *BasicMCPServer) CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	// TODO: Implement actual tool execution
	switch req.Name {
	case "tailscale_status":
		return &CallToolResponse{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "Tailscale status: Connected",
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}

// Shutdown handles server shutdown
func (s *BasicMCPServer) Shutdown(ctx context.Context) error {
	logger.Info("MCP server shutting down")
	return nil
}
