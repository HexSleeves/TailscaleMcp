package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/mcp"
)

// HTTPServer implements MCP protocol over HTTP
type HTTPServer struct {
	server     mcp.Server
	httpServer *http.Server
	router     *mux.Router
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(server mcp.Server, port int) *HTTPServer {
	router := mux.NewRouter()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s := &HTTPServer{
		server:     server,
		httpServer: httpServer,
		router:     router,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures HTTP routes and middleware
func (s *HTTPServer) setupRoutes() {
	// Add CORS middleware
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.loggingMiddleware)

	// MCP endpoints
	s.router.HandleFunc("/mcp", s.handleMCPRequest).Methods("POST", "OPTIONS")

	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Server info endpoint
	s.router.HandleFunc("/info", s.handleInfo).Methods("GET")
}

// Start begins the HTTP server
func (s *HTTPServer) Start(ctx context.Context) error {
	logger.Info("Starting HTTP MCP server", "addr", s.httpServer.Addr)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		logger.Info("HTTP server shutting down")

		// Graceful shutdown with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during HTTP server shutdown", "error", err)
			return err
		}

		return ctx.Err()
	case err := <-errChan:
		logger.Error("HTTP server error", "error", err)
		return err
	}
}

// handleMCPRequest processes MCP JSON-RPC requests
func (s *HTTPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		// CORS preflight handled by middleware
		return
	}

	// Parse request body
	var rawMsg map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawMsg); err != nil {
		s.sendError(w, nil, mcp.NewParseError(err.Error()))
		return
	}

	// Extract ID and method
	var id json.RawMessage
	if idRaw, exists := rawMsg["id"]; exists {
		id = idRaw
	}

	var method string
	if methodRaw, exists := rawMsg["method"]; exists {
		if err := json.Unmarshal(methodRaw, &method); err != nil {
			s.sendError(w, id, mcp.NewInvalidRequestError("invalid method"))
			return
		}
	}

	// Route the request based on method
	switch method {
	case mcp.RequestTypeInitialize:
		s.handleInitialize(w, r, id, rawMsg)
	case mcp.RequestTypeListTools:
		s.handleListTools(w, r, id, rawMsg)
	case mcp.RequestTypeCallTool:
		s.handleCallTool(w, r, id, rawMsg)
	case mcp.RequestTypeShutdown:
		s.handleShutdown(w, r, id, rawMsg)
	default:
		s.sendError(w, id, mcp.NewMethodNotFoundError(method))
	}
}

// handleInitialize processes initialize requests
func (s *HTTPServer) handleInitialize(w http.ResponseWriter, r *http.Request, id json.RawMessage, rawMsg map[string]json.RawMessage) {
	var msg mcp.Message[mcp.InitializeRequest, any]
	if err := s.parseMessage(rawMsg, &msg); err != nil {
		s.sendError(w, id, mcp.NewInvalidParamsError(err.Error()))
		return
	}

	if msg.Params == nil {
		s.sendError(w, id, mcp.NewInvalidParamsError("missing params"))
		return
	}

	response, err := s.server.Initialize(r.Context(), msg.Params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			s.sendError(w, id, mcpErr)
		} else {
			s.sendError(w, id, mcp.NewInternalError(err.Error()))
		}
		return
	}

	s.sendResponse(w, id, response)
}

// handleListTools processes list tools requests
func (s *HTTPServer) handleListTools(w http.ResponseWriter, r *http.Request, id json.RawMessage, rawMsg map[string]json.RawMessage) {
	var msg mcp.Message[mcp.ListToolsRequest, any]
	if err := s.parseMessage(rawMsg, &msg); err != nil {
		s.sendError(w, id, mcp.NewInvalidParamsError(err.Error()))
		return
	}

	params := &mcp.ListToolsRequest{}
	if msg.Params != nil {
		params = msg.Params
	}

	response, err := s.server.ListTools(r.Context(), params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			s.sendError(w, id, mcpErr)
		} else {
			s.sendError(w, id, mcp.NewInternalError(err.Error()))
		}
		return
	}

	s.sendResponse(w, id, response)
}

// handleCallTool processes call tool requests
func (s *HTTPServer) handleCallTool(w http.ResponseWriter, r *http.Request, id json.RawMessage, rawMsg map[string]json.RawMessage) {
	var msg mcp.Message[mcp.CallToolRequest, any]
	if err := s.parseMessage(rawMsg, &msg); err != nil {
		s.sendError(w, id, mcp.NewInvalidParamsError(err.Error()))
		return
	}

	if msg.Params == nil {
		s.sendError(w, id, mcp.NewInvalidParamsError("missing params"))
		return
	}

	response, err := s.server.CallTool(r.Context(), msg.Params)
	if err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			s.sendError(w, id, mcpErr)
		} else {
			s.sendError(w, id, mcp.NewToolExecutionError(msg.Params.Name, err))
		}
		return
	}

	s.sendResponse(w, id, response)
}

// handleShutdown processes shutdown requests
func (s *HTTPServer) handleShutdown(w http.ResponseWriter, r *http.Request, id json.RawMessage, rawMsg map[string]json.RawMessage) {
	var msg mcp.Message[mcp.ShutdownRequest, any]
	if err := s.parseMessage(rawMsg, &msg); err != nil {
		s.sendError(w, id, mcp.NewInvalidParamsError(err.Error()))
		return
	}

	if err := s.server.Shutdown(r.Context(), msg.Params); err != nil {
		var mcpErr *mcp.Error
		if errors.As(err, &mcpErr) {
			s.sendError(w, id, mcpErr)
		} else {
			s.sendError(w, id, mcp.NewInternalError(err.Error()))
		}
		return
	}

	s.sendResponse(w, id, map[string]interface{}{})
}

// handleHealth provides a health check endpoint
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"server": "tailscale-mcp-server",
	}); err != nil {
		logger.Error("Failed to encode health response", "error", err)
	}
}

// handleInfo provides server information
func (s *HTTPServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"name":            "tailscale-mcp-server",
		"version":         "dev", // TODO: Get from build info
		"protocolVersion": mcp.ProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
	}); err != nil {
		logger.Error("Failed to encode info response", "error", err)
	}
}

// parseMessage parses a raw message into a typed message
func (s *HTTPServer) parseMessage(rawMsg map[string]json.RawMessage, target interface{}) error {
	data, err := json.Marshal(rawMsg)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// sendResponse sends a successful response
func (s *HTTPServer) sendResponse(w http.ResponseWriter, id json.RawMessage, result interface{}) {
	response := mcp.NewResponse(id, &result)
	s.writeJSON(w, http.StatusOK, response)
}

// sendError sends an error response
func (s *HTTPServer) sendError(w http.ResponseWriter, id json.RawMessage, err *mcp.Error) {
	response := mcp.NewErrorMessage(id, err)

	// Map MCP error codes to HTTP status codes
	statusCode := http.StatusInternalServerError
	switch err.Code {
	case mcp.ErrorCodeParseError, mcp.ErrorCodeInvalidRequest, mcp.ErrorCodeInvalidParams:
		statusCode = http.StatusBadRequest
	case mcp.ErrorCodeMethodNotFound:
		statusCode = http.StatusNotFound
	case mcp.ErrorCodeToolNotFound:
		statusCode = http.StatusNotFound
	}

	s.writeJSON(w, statusCode, response)
}

// writeJSON writes a JSON response
func (s *HTTPServer) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}

// corsMiddleware adds CORS headers
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapper.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
