package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/mcp"
)

// StdioServer implements MCP protocol over stdin/stdout
type StdioServer struct {
	server mcp.Server
	reader *bufio.Scanner
	writer io.Writer
	mu     sync.Mutex
}

// NewStdioServer creates a new stdio server instance
func NewStdioServer(server mcp.Server) *StdioServer {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10 MiB hard-cap

	return &StdioServer{
		server: server,
		reader: sc,
		writer: os.Stdout,
	}
}

// Start begins processing MCP messages from stdin
func (s *StdioServer) Start(ctx context.Context) error {
	logger.Info("Starting stdio MCP server")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stdio server shutting down")
			return ctx.Err()
		default:
			if !s.reader.Scan() {
				if err := s.reader.Err(); err != nil {
					logger.Error("Error reading from stdin", "error", err)
					return fmt.Errorf("stdin read error: %w", err)
				}
				// EOF reached
				logger.Info("Stdin closed, shutting down")
				return nil
			}

			line := s.reader.Text()
			if line == "" {
				continue
			}

			logger.Debug("Received message", "message", line)

			if err := s.handleMessage(ctx, line); err != nil {
				logger.Error("Error handling message", "error", err, "message", line)
				// Continue processing other messages
			}
		}
	}
}

// handleMessage processes a single JSON-RPC message
func (s *StdioServer) handleMessage(ctx context.Context, message string) error {
	// Parse the raw message to determine the method
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal([]byte(message), &rawMsg); err != nil {
		return s.sendError(nil, mcp.NewParseError(err.Error()))
	}

	// Extract ID and method
	var id json.RawMessage
	if idRaw, exists := rawMsg["id"]; exists {
		id = idRaw
	}

	var method string
	if methodRaw, exists := rawMsg["method"]; exists {
		if err := json.Unmarshal(methodRaw, &method); err != nil {
			return s.sendError(id, mcp.NewInvalidRequestError("invalid method"))
		}
	}

	// Route the message based on method
	switch method {
	case mcp.RequestTypeInitialize:
		return s.handleInitialize(ctx, id, message)
	case mcp.RequestTypeListTools:
		return s.handleListTools(ctx, id, message)
	case mcp.RequestTypeCallTool:
		return s.handleCallTool(ctx, id, message)
	case mcp.RequestTypeShutdown:
		return s.handleShutdown(ctx, id, message)
	default:
		return s.sendError(id, mcp.NewMethodNotFoundError(method))
	}
}

// handleInitialize processes initialize requests
func (s *StdioServer) handleInitialize(ctx context.Context, id json.RawMessage, message string) error {
	var msg mcp.Message[mcp.InitializeRequest, any]
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return s.sendError(id, mcp.NewInvalidParamsError(err.Error()))
	}

	if msg.Params == nil {
		return s.sendError(id, mcp.NewInvalidParamsError("missing params"))
	}

	response, err := s.server.Initialize(ctx, msg.Params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			return s.sendError(id, mcpErr)
		}
		return s.sendError(id, mcp.NewInternalError(err.Error()))
	}

	return s.sendResponse(id, response)
}

// handleListTools processes list tools requests
func (s *StdioServer) handleListTools(ctx context.Context, id json.RawMessage, message string) error {
	var msg mcp.Message[mcp.ListToolsRequest, any]
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return s.sendError(id, mcp.NewInvalidParamsError(err.Error()))
	}

	params := &mcp.ListToolsRequest{}
	if msg.Params != nil {
		params = msg.Params
	}

	response, err := s.server.ListTools(ctx, params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			return s.sendError(id, mcpErr)
		}
		return s.sendError(id, mcp.NewInternalError(err.Error()))
	}

	return s.sendResponse(id, response)
}

// handleCallTool processes call tool requests
func (s *StdioServer) handleCallTool(ctx context.Context, id json.RawMessage, message string) error {
	var msg mcp.Message[mcp.CallToolRequest, any]
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return s.sendError(id, mcp.NewInvalidParamsError(err.Error()))
	}

	if msg.Params == nil {
		return s.sendError(id, mcp.NewInvalidParamsError("missing params"))
	}

	response, err := s.server.CallTool(ctx, msg.Params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			return s.sendError(id, mcpErr)
		}
		return s.sendError(id, mcp.NewToolExecutionError(msg.Params.Name, err))
	}

	return s.sendResponse(id, response)
}

// handleShutdown processes shutdown requests
func (s *StdioServer) handleShutdown(ctx context.Context, id json.RawMessage, message string) error {
	var msg mcp.Message[mcp.ShutdownRequest, any]
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return s.sendError(id, mcp.NewInvalidParamsError(err.Error()))
	}

	if err := s.server.Shutdown(ctx, msg.Params); err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			return s.sendError(id, mcpErr)
		}
		return s.sendError(id, mcp.NewInternalError(err.Error()))
	}

	// Send success response
	return s.sendResponse(id, map[string]interface{}{})
}

// sendResponse sends a successful response
func (s *StdioServer) sendResponse(id json.RawMessage, result any) error {
	response := mcp.NewResponse(id, &result)
	return s.writeMessage(response)
}

// sendError sends an error response
func (s *StdioServer) sendError(id json.RawMessage, err *mcp.Error) error {
	response := mcp.NewErrorMessage(id, err)
	return s.writeMessage(response)
}

// writeMessage writes a message to stdout
func (s *StdioServer) writeMessage(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal response", "error", err)
		return fmt.Errorf("marshal error: %w", err)
	}

	logger.Debug("Sending message", "message", string(data))

	if _, err := s.writer.Write(data); err != nil {
		logger.Error("Failed to write to stdout", "error", err)
		return fmt.Errorf("write error: %w", err)
	}

	if _, err := s.writer.Write([]byte("\n")); err != nil {
		logger.Error("Failed to write newline to stdout", "error", err)
		return fmt.Errorf("write newline error: %w", err)
	}

	return nil
}
