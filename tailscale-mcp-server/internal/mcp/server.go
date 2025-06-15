package mcp

import (
	"context"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tools"
)

// MCPServer implements the MCP server interface, handling tool registration and execution.
type MCPServer struct {
	registry *tools.ToolRegistry
	name     string
	version  string
}

// NewMCPServer creates a new MCP server instance.
func NewMCPServer(registry *tools.ToolRegistry, name, version string) Server {
	return &MCPServer{
		registry: registry,
		name:     name,
		version:  version,
	}
}

// Initialize handles MCP initialization.
func (s *MCPServer) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	if err := ValidateInitializeRequest(req); err != nil {
		return nil, err
	}

	logger.Info("MCP server initialized", "client_info", req.ClientInfo)

	return &InitializeResponse{
		ProtocolVersion: ProtocolVersion,
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
	}, nil
}

// ListTools returns available tools from the tool registry.
func (s *MCPServer) ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error) {
	registeredTools := s.registry.GetTools()
	toolList := make([]Tool, 0, len(registeredTools))

	for _, t := range registeredTools {
		toolList = append(toolList, Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}

	return &ListToolsResponse{
		Tools: toolList,
	}, nil
}

// CallTool executes a tool from the tool registry.
func (s *MCPServer) CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	tool, ok := s.registry.GetTool(req.Name)
	if !ok {
		return nil, NewToolNotFoundError(req.Name)
	}

	// Create a new tool context
	toolCtx := tools.NewContext(ctx, s.registry)

	// Execute the tool
	result, err := tool.Execute(toolCtx, req.Arguments)
	if err != nil {
		return nil, NewToolExecutionError(req.Name, err)
	}

	return &CallToolResponse{
		Content: []ContentBlock{
			NewTextContent(result),
		},
	}, nil
}

// Shutdown handles server shutdown.
func (s *MCPServer) Shutdown(ctx context.Context, req *ShutdownRequest) error {
	logger.Info("MCP server shutting down")
	return nil
}
