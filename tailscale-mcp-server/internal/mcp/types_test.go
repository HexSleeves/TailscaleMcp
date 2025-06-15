package mcp

import (
	"encoding/json"
	"testing"
)

func TestMessageFactoryFunctions(t *testing.T) {
	// Test NewRequest
	t.Run("NewRequest", func(t *testing.T) {
		id := json.RawMessage(`"test-id"`)
		method := "test/method"
		params := &InitializeRequest{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{},
			ClientInfo:      ClientInfo{Name: "test", Version: "1.0"},
		}

		msg := NewRequest(id, method, params)

		if msg.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC to be '2.0', got '%s'", msg.JSONRPC)
		}
		if string(msg.ID) != `"test-id"` {
			t.Errorf("Expected ID to be '\"test-id\"', got '%s'", string(msg.ID))
		}
		if msg.Method != method {
			t.Errorf("Expected Method to be '%s', got '%s'", method, msg.Method)
		}
		if msg.Params == nil {
			t.Error("Expected Params to be set")
		}
		if msg.Result != nil {
			t.Error("Expected Result to be nil for request")
		}
		if msg.Error != nil {
			t.Error("Expected Error to be nil for request")
		}
	})

	// Test NewResponse
	t.Run("NewResponse", func(t *testing.T) {
		id := json.RawMessage(`123`)
		result := &InitializeResponse{
			ProtocolVersion: "2024-11-05",
			Capabilities:    ServerCapabilities{},
			ServerInfo:      ServerInfo{Name: "test-server", Version: "1.0"},
		}

		msg := NewResponse(id, result)

		if msg.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC to be '2.0', got '%s'", msg.JSONRPC)
		}
		if string(msg.ID) != "123" {
			t.Errorf("Expected ID to be '123', got '%s'", string(msg.ID))
		}
		if msg.Method != "" {
			t.Errorf("Expected Method to be empty for response, got '%s'", msg.Method)
		}
		if msg.Params != nil {
			t.Error("Expected Params to be nil for response")
		}
		if msg.Result == nil {
			t.Error("Expected Result to be set")
		}
		if msg.Error != nil {
			t.Error("Expected Error to be nil for successful response")
		}
	})

	// Test NewErrorMessage
	t.Run("NewErrorMessage", func(t *testing.T) {
		id := json.RawMessage(`null`)
		err := &Error{
			Code:    -32600,
			Message: "Invalid Request",
			Data:    "Additional error data",
		}

		msg := NewErrorMessage(id, err)

		if msg.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC to be '2.0', got '%s'", msg.JSONRPC)
		}
		if string(msg.ID) != "null" {
			t.Errorf("Expected ID to be 'null', got '%s'", string(msg.ID))
		}
		if msg.Method != "" {
			t.Errorf("Expected Method to be empty for error response, got '%s'", msg.Method)
		}
		if msg.Params != nil {
			t.Error("Expected Params to be nil for error response")
		}
		if msg.Result != nil {
			t.Error("Expected Result to be nil for error response")
		}
		if msg.Error == nil {
			t.Error("Expected Error to be set")
		}
		if msg.Error.Code != -32600 {
			t.Errorf("Expected Error.Code to be -32600, got %d", msg.Error.Code)
		}
	})

	// Test NewNotification
	t.Run("NewNotification", func(t *testing.T) {
		method := "notification/method"
		params := &struct {
			Message string `json:"message"`
		}{
			Message: "test notification",
		}

		msg := NewNotification(method, params)

		if msg.JSONRPC != "2.0" {
			t.Errorf("Expected JSONRPC to be '2.0', got '%s'", msg.JSONRPC)
		}
		if msg.ID != nil {
			t.Errorf("Expected ID to be nil for notification, got '%s'", string(msg.ID))
		}
		if msg.Method != method {
			t.Errorf("Expected Method to be '%s', got '%s'", method, msg.Method)
		}
		if msg.Params == nil {
			t.Error("Expected Params to be set")
		}
		if msg.Result != nil {
			t.Error("Expected Result to be nil for notification")
		}
		if msg.Error != nil {
			t.Error("Expected Error to be nil for notification")
		}
	})
}

func TestMessageSerialization(t *testing.T) {
	// Test that factory-created messages serialize correctly
	t.Run("Request serialization", func(t *testing.T) {
		id := json.RawMessage(`"req-1"`)
		method := RequestTypeInitialize
		params := &InitializeRequest{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    map[string]any{"test": true},
			ClientInfo:      ClientInfo{Name: "test-client", Version: "1.0.0"},
		}

		msg := NewRequest(id, method, params)

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		// Verify the JSON contains expected fields
		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if parsed["jsonrpc"] != "2.0" {
			t.Errorf("Expected jsonrpc to be '2.0', got %v", parsed["jsonrpc"])
		}
		if parsed["id"] != "req-1" {
			t.Errorf("Expected id to be 'req-1', got %v", parsed["id"])
		}
		if parsed["method"] != method {
			t.Errorf("Expected method to be '%s', got %v", method, parsed["method"])
		}
		if parsed["params"] == nil {
			t.Error("Expected params to be present")
		}
	})

	t.Run("Response serialization", func(t *testing.T) {
		id := json.RawMessage(`42`)
		result := &InitializeResponse{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    ServerCapabilities{},
			ServerInfo:      ServerInfo{Name: "test-server", Version: "1.0.0"},
		}

		msg := NewResponse(id, result)

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal response: %v", err)
		}

		// Verify the JSON contains expected fields
		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if parsed["jsonrpc"] != "2.0" {
			t.Errorf("Expected jsonrpc to be '2.0', got %v", parsed["jsonrpc"])
		}

		idValue, ok := parsed["id"].(float64)
		if !ok || idValue != 42 {
			t.Errorf("Expected id to be 42 (float64), got %v (%T)", parsed["id"], parsed["id"])
		}

		if parsed["result"] == nil {
			t.Error("Expected result to be present")
		}
		// Method should not be present in response
		if _, exists := parsed["method"]; exists {
			t.Error("Expected method to be omitted in response")
		}
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewParseError", func(t *testing.T) {
		err := NewParseError("invalid json")
		if err.Code != ErrorCodeParseError {
			t.Errorf("Expected code %d, got %d", ErrorCodeParseError, err.Code)
		}
		if err.Message != "Parse error" {
			t.Errorf("Expected message 'Parse error', got '%s'", err.Message)
		}
		if err.Data != "invalid json" {
			t.Errorf("Expected data 'invalid json', got %v", err.Data)
		}
	})

	t.Run("NewMethodNotFoundError", func(t *testing.T) {
		err := NewMethodNotFoundError("unknown/method")
		if err.Code != ErrorCodeMethodNotFound {
			t.Errorf("Expected code %d, got %d", ErrorCodeMethodNotFound, err.Code)
		}
		if err.Message != "Method not found" {
			t.Errorf("Expected message 'Method not found', got '%s'", err.Message)
		}
		data, ok := err.Data.(map[string]string)
		if !ok {
			t.Errorf("Expected data to be map[string]string, got %T", err.Data)
		} else if data["method"] != "unknown/method" {
			t.Errorf("Expected method 'unknown/method', got '%s'", data["method"])
		}
	})

	t.Run("NewUnsupportedProtocolError", func(t *testing.T) {
		err := NewUnsupportedProtocolError("1.0.0", "2.0.0")
		if err.Code != ErrorCodeUnsupportedProtocol {
			t.Errorf("Expected code %d, got %d", ErrorCodeUnsupportedProtocol, err.Code)
		}
		data, ok := err.Data.(map[string]string)
		if !ok {
			t.Errorf("Expected data to be map[string]string, got %T", err.Data)
		} else {
			if data["clientVersion"] != "1.0.0" {
				t.Errorf("Expected clientVersion '1.0.0', got '%s'", data["clientVersion"])
			}
			if data["serverVersion"] != "2.0.0" {
				t.Errorf("Expected serverVersion '2.0.0', got '%s'", data["serverVersion"])
			}
		}
	})

	t.Run("Error interface implementation", func(t *testing.T) {
		err := NewInternalError("test error")
		var goErr error = err
		if goErr.Error() != "Internal error" {
			t.Errorf("Expected error message 'Internal error', got '%s'", goErr.Error())
		}
	})
}

func TestProtocolVersionCompatibility(t *testing.T) {
	t.Run("IsCompatibleProtocolVersion", func(t *testing.T) {
		// Test compatible version
		if !IsCompatibleProtocolVersion(ProtocolVersion) {
			t.Errorf("Expected protocol version '%s' to be compatible", ProtocolVersion)
		}

		// Test incompatible version
		if IsCompatibleProtocolVersion("1.0.0") {
			t.Error("Expected protocol version '1.0.0' to be incompatible")
		}

		// Test empty version
		if IsCompatibleProtocolVersion("") {
			t.Error("Expected empty protocol version to be incompatible")
		}
	})
}

func TestValidateInitializeRequest(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		req := &InitializeRequest{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    map[string]any{},
			ClientInfo:      ClientInfo{Name: "test-client", Version: "1.0.0"},
		}

		err := ValidateInitializeRequest(req)
		if err != nil {
			t.Errorf("Expected valid request to pass validation, got error: %v", err)
		}
	})

	t.Run("Nil request", func(t *testing.T) {
		err := ValidateInitializeRequest(nil)

		if err == nil {
			t.Fatal("expected nil request to fail validation")
		}

		if err.Code != ErrorCodeInvalidRequest {
			t.Fatalf("expected error code %d, got %d",
				ErrorCodeInvalidRequest, err.Code)
		}
	})

	t.Run("Missing protocol version", func(t *testing.T) {
		req := &InitializeRequest{
			Capabilities: map[string]any{},
			ClientInfo:   ClientInfo{Name: "test-client", Version: "1.0.0"},
		}

		err := ValidateInitializeRequest(req)
		if err == nil {
			t.Fatal("expected request with missing protocol version to fail validation")
		}
		if err.Code != ErrorCodeInvalidParams {
			t.Errorf("Expected error code %d, got %d", ErrorCodeInvalidParams, err.Code)
		}
	})

	t.Run("Incompatible protocol version", func(t *testing.T) {
		req := &InitializeRequest{
			ProtocolVersion: "1.0.0",
			Capabilities:    map[string]any{},
			ClientInfo:      ClientInfo{Name: "test-client", Version: "1.0.0"},
		}

		err := ValidateInitializeRequest(req)
		if err == nil {
			t.Fatal("expected request with incompatible protocol version to fail validation")
		}
		if err.Code != ErrorCodeUnsupportedProtocol {
			t.Errorf("Expected error code %d, got %d", ErrorCodeUnsupportedProtocol, err.Code)
		}
	})

	t.Run("Missing client name", func(t *testing.T) {
		req := &InitializeRequest{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    map[string]any{},
			ClientInfo:      ClientInfo{Version: "1.0.0"},
		}

		err := ValidateInitializeRequest(req)
		if err == nil {
			t.Fatal("expected request with missing client name to fail validation")
		}
		if err.Code != ErrorCodeInvalidParams {
			t.Errorf("Expected error code %d, got %d", ErrorCodeInvalidParams, err.Code)
		}
	})

	t.Run("Missing client version", func(t *testing.T) {
		req := &InitializeRequest{
			ProtocolVersion: ProtocolVersion,
			Capabilities:    map[string]any{},
			ClientInfo:      ClientInfo{Name: "test-client"},
		}

		err := ValidateInitializeRequest(req)
		if err == nil {
			t.Fatal("expected request with missing client version to fail validation")
		}
		if err.Code != ErrorCodeInvalidParams {
			t.Errorf("Expected error code %d, got %d", ErrorCodeInvalidParams, err.Code)
		}
	})
}
