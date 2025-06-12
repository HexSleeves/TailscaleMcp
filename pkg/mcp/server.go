package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// BasicServer provides a basic implementation of the MCP Server interface
// with proper handshake and capability negotiation
type BasicServer struct {
	serverInfo ServerInfo
	tools      []Tool
}

// NewBasicServer creates a new basic MCP server
func NewBasicServer(name, version string) *BasicServer {
	return &BasicServer{
		serverInfo: ServerInfo{
			Name:    name,
			Version: version,
		},
		tools: make([]Tool, 0),
	}
}

// RegisterTool adds a tool to the server's tool registry
func (s *BasicServer) RegisterTool(tool Tool) {
	s.tools = append(s.tools, tool)
}

// Initialize handles the MCP initialization handshake
func (s *BasicServer) Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error) {
	// Validate the initialize request
	if err := ValidateInitializeRequest(req); err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	// Create response with server capabilities
	response := &InitializeResponse{
		ProtocolVersion: ProtocolVersion,
		ServerInfo:      s.serverInfo,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				// Indicate that our tool list can change
				ListChanged: boolPtr(true),
			},
		},
	}

	return response, nil
}

// ListTools returns the list of available tools
func (s *BasicServer) ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error) {
	return &ListToolsResponse{
		Tools: s.tools,
	}, nil
}

// CallTool executes a tool call - this is a basic implementation that should be overridden
func (s *BasicServer) CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	// Find the requested tool
	var tool *Tool
	for i := range s.tools {
		if s.tools[i].Name == req.Name {
			tool = &s.tools[i]
			break
		}
	}

	if tool == nil {
		return NewErrorResponse(fmt.Sprintf("Tool '%s' not found", req.Name)), nil
	}

	// Basic implementation - should be overridden by specific implementations
	return NewSuccessResponse(fmt.Sprintf("Tool '%s' called successfully (basic implementation)", req.Name)), nil
}

// Shutdown handles server shutdown
func (s *BasicServer) Shutdown(ctx context.Context) error {
	// Basic implementation - can be overridden for cleanup
	return nil
}

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// ToolHandler is a function type for handling tool calls
type ToolHandler func(ctx context.Context, args json.RawMessage) (*CallToolResponse, error)

// AdvancedServer extends BasicServer with custom tool handlers
type AdvancedServer struct {
	*BasicServer
	toolHandlers map[string]ToolHandler
}

// NewAdvancedServer creates a new advanced MCP server with custom tool handlers
func NewAdvancedServer(name, version string) *AdvancedServer {
	return &AdvancedServer{
		BasicServer:  NewBasicServer(name, version),
		toolHandlers: make(map[string]ToolHandler),
	}
}

// RegisterToolWithHandler adds a tool with a custom handler
func (s *AdvancedServer) RegisterToolWithHandler(tool Tool, handler ToolHandler) {
	s.RegisterTool(tool)
	s.toolHandlers[tool.Name] = handler
}

// CallTool executes a tool call using registered handlers
func (s *AdvancedServer) CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	handler, exists := s.toolHandlers[req.Name]
	if !exists {
		return NewErrorResponse(fmt.Sprintf("Tool '%s' not found", req.Name)), nil
	}

	// Execute the tool handler
	response, err := handler(ctx, req.Arguments)
	if err != nil {
		return NewErrorResponse(fmt.Sprintf("Tool execution failed: %v", err)), nil
	}

	return response, nil
}
