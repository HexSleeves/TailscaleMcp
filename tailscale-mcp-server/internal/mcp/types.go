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

	// JSON-RPC 2.0 Error Codes
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603

	// MCP-specific error codes
	ErrorCodeUnsupportedProtocol = -32000
	ErrorCodeToolNotFound        = -32001
	ErrorCodeToolExecutionError  = -32002
)

// Base message types
type Message[T any, R any] struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // number, string or null
	Method  string          `json:"method,omitempty"`
	Params  *T              `json:"params,omitempty"`
	Result  *R              `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error implements the Go error interface
func (e *Error) Error() string {
	return e.Message
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
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourceCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged *bool `json:"listChanged,omitempty"`
}

type ResourceCapability struct {
	// No specific capabilities defined yet
}

// Tools
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
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

// Resource represents a data source provided by the server.
type Resource struct {
	URI         string `json:"uri"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Content     any    `json:"content,omitempty"`
}

// Shutdown
type ShutdownRequest struct{}

// Server interface
type Server interface {
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error)
	CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error)
	Shutdown(ctx context.Context, req *ShutdownRequest) error
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

// Message factory functions for type-safe construction

// NewRequest creates a new request message with proper JSONRPC version and method.
// Only the Params field should be set for requests.
func NewRequest[T any](id json.RawMessage, method string, params *T) *Message[T, any] {
	return &Message[T, any]{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewResponse creates a new response message with proper JSONRPC version.
// Only the Result field should be set for successful responses.
func NewResponse[R any](id json.RawMessage, result *R) *Message[any, R] {
	return &Message[any, R]{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorMessage creates a new error response message with proper JSONRPC version.
// Only the Error field should be set for error responses.
func NewErrorMessage(id json.RawMessage, err *Error) *Message[any, any] {
	return &Message[any, any]{
		JSONRPC: "2.0",
		ID:      id,
		Error:   err,
	}
}

// NewNotification creates a new notification message (request without ID).
// Notifications don't expect a response.
func NewNotification[T any](method string, params *T) *Message[T, any] {
	return &Message[T, any]{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// Error constructors for common MCP errors

// NewParseError creates a parse error for malformed JSON
func NewParseError(data any) *Error {
	return &Error{
		Code:    ErrorCodeParseError,
		Message: "Parse error",
		Data:    data,
	}
}

// NewInvalidRequestError creates an invalid request error
func NewInvalidRequestError(data any) *Error {
	return &Error{
		Code:    ErrorCodeInvalidRequest,
		Message: "Invalid Request",
		Data:    data,
	}
}

// NewMethodNotFoundError creates a method not found error
func NewMethodNotFoundError(method string) *Error {
	return &Error{
		Code:    ErrorCodeMethodNotFound,
		Message: "Method not found",
		Data:    map[string]string{"method": method},
	}
}

// NewInvalidParamsError creates an invalid parameters error
func NewInvalidParamsError(data any) *Error {
	return &Error{
		Code:    ErrorCodeInvalidParams,
		Message: "Invalid params",
		Data:    data,
	}
}

// NewInternalError creates an internal error
func NewInternalError(data any) *Error {
	return &Error{
		Code:    ErrorCodeInternalError,
		Message: "Internal error",
		Data:    data,
	}
}

// NewUnsupportedProtocolError creates an unsupported protocol version error
func NewUnsupportedProtocolError(clientVersion, serverVersion string) *Error {
	return &Error{
		Code:    ErrorCodeUnsupportedProtocol,
		Message: "Unsupported protocol version",
		Data: map[string]string{
			"clientVersion": clientVersion,
			"serverVersion": serverVersion,
		},
	}
}

// NewToolNotFoundError creates a tool not found error
func NewToolNotFoundError(toolName string) *Error {
	return &Error{
		Code:    ErrorCodeToolNotFound,
		Message: "Tool not found",
		Data:    map[string]string{"tool": toolName},
	}
}

// NewToolExecutionError creates a tool execution error
func NewToolExecutionError(toolName string, err error) *Error {
	return &Error{
		Code:    ErrorCodeToolExecutionError,
		Message: "Tool execution failed",
		Data: map[string]string{
			"tool":  toolName,
			"error": err.Error(),
		},
	}
}

// Protocol version compatibility checking

// IsCompatibleProtocolVersion checks if the client protocol version is compatible
// with the server's supported version
func IsCompatibleProtocolVersion(clientVersion string) bool {
	// For now, we only support the exact version
	// In the future, this could be enhanced to support version ranges
	return clientVersion == ProtocolVersion
}

// ValidateInitializeRequest validates an initialize request and returns an error if invalid
func ValidateInitializeRequest(req *InitializeRequest) *Error {
	if req == nil {
		return NewInvalidRequestError("Initialize request cannot be nil")
	}

	if req.ProtocolVersion == "" {
		return NewInvalidParamsError("Protocol version is required")
	}

	if !IsCompatibleProtocolVersion(req.ProtocolVersion) {
		return NewUnsupportedProtocolError(req.ProtocolVersion, ProtocolVersion)
	}

	if req.ClientInfo.Name == "" {
		return NewInvalidParamsError("Client name is required")
	}

	if req.ClientInfo.Version == "" {
		return NewInvalidParamsError("Client version is required")
	}

	return nil
}
