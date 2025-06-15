package tailscale

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
)

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*APIClient, *httptest.Server) {
	server := httptest.NewServer(handler)

	// Initialize logger for tests
	err := logger.Initialize(0, "") // Debug level, no file
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	cfg := &config.Config{
		TailscaleAPIKey:     "test-api-key",
		TailscaleTailnet:    "test-tailnet",
		TailscaleAPIBaseURL: server.URL,
	}

	client := NewAPIClient(cfg)
	return client, server
}

func TestNewAPIClient(t *testing.T) {
	cfg := &config.Config{
		TailscaleAPIKey:     "test-key",
		TailscaleTailnet:    "test-tailnet",
		TailscaleAPIBaseURL: "https://api.tailscale.com",
	}

	client := NewAPIClient(cfg)

	if client.apiKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", client.apiKey)
	}

	if client.tailnet != "test-tailnet" {
		t.Errorf("Expected tailnet 'test-tailnet', got '%s'", client.tailnet)
	}

	expectedBaseURL := "https://api.tailscale.com/api/v2"
	if client.baseURL != expectedBaseURL {
		t.Errorf("Expected base URL '%s', got '%s'", expectedBaseURL, client.baseURL)
	}
}

func TestNewAPIClientDefaults(t *testing.T) {
	cfg := &config.Config{
		TailscaleAPIKey:     "test-key",
		TailscaleAPIBaseURL: "https://api.tailscale.com",
		// No tailnet specified
	}

	client := NewAPIClient(cfg)

	if client.tailnet != "-" {
		t.Errorf("Expected default tailnet '-', got '%s'", client.tailnet)
	}
}

func TestListDevices(t *testing.T) {
	mockDevices := DeviceListResponse{
		Devices: []Device{
			{
				ID:       "device1",
				Name:     "test-device",
				Hostname: "test-hostname",
				OS:       "linux",
				Created:  time.Now(),
				LastSeen: time.Now(),
			},
		},
	}

	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/api/v2/tailnet/test-tailnet/devices"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockDevices); err != nil {
			t.Errorf("Failed to encode mock devices: %v", err)
		}
	})
	defer server.Close()

	ctx := context.Background()
	response := client.ListDevices(ctx)

	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}

	if len(response.Data.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(response.Data.Devices))
	}

	device := response.Data.Devices[0]
	if device.ID != "device1" {
		t.Errorf("Expected device ID 'device1', got '%s'", device.ID)
	}
}

func TestGetDevice(t *testing.T) {
	mockDevice := Device{
		ID:       "device1",
		Name:     "test-device",
		Hostname: "test-hostname",
		OS:       "linux",
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/api/v2/device/device1"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockDevice); err != nil {
			t.Errorf("Failed to encode mock device: %v", err)
		}
	})
	defer server.Close()

	ctx := context.Background()
	response := client.GetDevice(ctx, "device1")

	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}

	if response.Data.ID != "device1" {
		t.Errorf("Expected device ID 'device1', got '%s'", response.Data.ID)
	}
}

func TestGetTailnetInfo(t *testing.T) {
	mockTailnet := TailnetInfo{
		Name:      "test-tailnet",
		AccountID: "account123",
		CreatedAt: time.Now(),
		DNSConfig: DNSConfig{
			MagicDNS: true,
		},
	}

	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/api/v2/tailnet/test-tailnet"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockTailnet); err != nil {
			t.Errorf("Failed to encode mock tailnet: %v", err)
		}
	})
	defer server.Close()

	ctx := context.Background()
	response := client.GetTailnetInfo(ctx)

	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}

	if response.Data.Name != "test-tailnet" {
		t.Errorf("Expected tailnet name 'test-tailnet', got '%s'", response.Data.Name)
	}
}

func TestAPIError(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"message": "Unauthorized",
		}); err != nil {
			t.Errorf("Failed to encode error response: %v", err)
		}
	})
	defer server.Close()

	ctx := context.Background()
	response := client.ListDevices(ctx)

	if response.Success {
		t.Error("Expected error response, got success")
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}

	if response.Error != "Unauthorized" {
		t.Errorf("Expected error message 'Unauthorized', got '%s'", response.Error)
	}
}

func TestTestConnection(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"name": "test-tailnet",
		}); err != nil {
			t.Errorf("Failed to encode test response: %v", err)
		}
	})
	defer server.Close()

	ctx := context.Background()
	response := client.TestConnection(ctx)

	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}

	if response.Data["status"] != "connected" {
		t.Errorf("Expected status 'connected', got '%s'", response.Data["status"])
	}
}
