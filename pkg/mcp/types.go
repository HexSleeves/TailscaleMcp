package mcp

import (
	"context"
	"encoding/json"
)

// Protocol constants
const (
	ProtocolVersion = "2024-11-05"

	// Request types
	RequestTypeInitialize = "initialize"
	RequestTypeListTools  = "tools/list"
	RequestTypeCallTool   = "tools/call"
	RequestTypeShutdown   = "shutdown"

	// Response types
	ResponseTypeInitialized = "initialized"
	ResponseTypeResult      = "result"
	ResponseTypeError       = "error"
)

// Base message types
type Message struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Initialize request/response
type InitializeRequest struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

type ToolsCapability struct {
	ListChanged *bool `json:"listChanged,omitempty"`
}

// Tools
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ListToolsRequest struct{}

type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

type CallToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type CallToolResponse struct {
	Content []ContentBlock `json:"content"`
	IsError *bool          `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Server interface
type Server interface {
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error)
	CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error)
	Shutdown(ctx context.Context) error
}

// Helper functions
func NewTextContent(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}

func NewErrorResponse(text string) *CallToolResponse {
	isError := true
	return &CallToolResponse{
		Content: []ContentBlock{NewTextContent(text)},
		IsError: &isError,
	}
}

func NewSuccessResponse(text string) *CallToolResponse {
	return &CallToolResponse{
		Content: []ContentBlock{NewTextContent(text)},
	}
}
