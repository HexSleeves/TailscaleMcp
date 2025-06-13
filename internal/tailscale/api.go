package tailscale

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hexsleeves/tailscale-mcp-server/internal/config"
	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
)

// APIClient represents a Tailscale API client
type APIClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	tailnet    string
}

// APIResponse represents a standardized API response
type APIResponse[T any] struct {
	Success    bool   `json:"success"`
	Data       T      `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
}

// APIError represents a Tailscale API error
type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Tailscale API error (status %d): %s", e.StatusCode, e.Message)
}

// NewAPIClient creates a new Tailscale API client
func NewAPIClient(cfg *config.Config) *APIClient {
	if cfg.TailscaleAPIKey == "" {
		logger.Warn("No Tailscale API key provided. API operations will fail until TAILSCALE_API_KEY is set.")
	}

	tailnet := cfg.TailscaleTailnet
	if tailnet == "" {
		tailnet = "-" // Default to current user's tailnet
	}

	return &APIClient{

		baseURL: strings.TrimSuffix(cfg.TailscaleAPIBaseURL, "/") + "/api/v2",
		apiKey:  cfg.TailscaleAPIKey,
		tailnet: tailnet,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest performs an HTTP request with proper authentication and logging
func (c *APIClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Log request
	logger.Debug(fmt.Sprintf("API Request: %s %s", method, url))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("API Request Error", "url", url, "method", method, "error", err.Error())
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log response
	logger.Debug(fmt.Sprintf("API Response: %d %s", resp.StatusCode, url))

	return resp, nil
}

// handleResponse processes an HTTP response and returns the raw response data
func (c *APIClient) handleResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err.Error())
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to parse error response
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				return nil, &APIError{Message: msg, StatusCode: resp.StatusCode}
			}
			if errMsg, ok := errorResp["error"].(string); ok {
				return nil, &APIError{Message: errMsg, StatusCode: resp.StatusCode}
			}
		}

		return nil, &APIError{
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
		}
	}

	return body, nil
}

// createSuccessResponse creates a successful API response
func createSuccessResponse[T any](data T, statusCode int) APIResponse[T] {
	return APIResponse[T]{
		Success:    true,
		Data:       data,
		StatusCode: statusCode,
	}
}

// createErrorResponse creates an error API response
func createErrorResponse[T any](err error) APIResponse[T] {
	if apiErr, ok := err.(*APIError); ok {
		return APIResponse[T]{
			Success:    false,
			Error:      apiErr.Message,
			StatusCode: apiErr.StatusCode,
		}
	}

	return APIResponse[T]{
		Success: false,
		Error:   err.Error(),
	}
}

// Device Management Methods

// ListDevices retrieves all devices in the tailnet
func (c *APIClient) ListDevices(ctx context.Context) APIResponse[DeviceListResponse] {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/tailnet/%s/devices", c.tailnet), nil)
	if err != nil {
		return createErrorResponse[DeviceListResponse](err)
	}

	body, err := c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[DeviceListResponse](err)
	}

	var deviceList DeviceListResponse
	if err := json.Unmarshal(body, &deviceList); err != nil {
		return createErrorResponse[DeviceListResponse](fmt.Errorf("failed to parse device list: %w", err))
	}

	return createSuccessResponse(deviceList, resp.StatusCode)
}

// GetDevice retrieves a specific device by ID
func (c *APIClient) GetDevice(ctx context.Context, deviceID string) APIResponse[Device] {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/device/%s", deviceID), nil)
	if err != nil {
		return createErrorResponse[Device](err)
	}

	body, err := c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[Device](err)
	}

	var device Device
	if err := json.Unmarshal(body, &device); err != nil {
		return createErrorResponse[Device](fmt.Errorf("failed to parse device: %w", err))
	}

	return createSuccessResponse(device, resp.StatusCode)
}

// DeleteDevice removes a device from the tailnet
func (c *APIClient) DeleteDevice(ctx context.Context, deviceID string) APIResponse[interface{}] {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/device/%s", deviceID), nil)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}

// SetDeviceAuthorization authorizes or deauthorizes a device
func (c *APIClient) SetDeviceAuthorization(ctx context.Context, deviceID string, authorized bool) APIResponse[interface{}] {
	auth := DeviceAuthorization{Authorized: authorized}
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/device/%s/authorized", deviceID), auth)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}

// SetDeviceKeyExpiry sets whether a device's key expires
func (c *APIClient) SetDeviceKeyExpiry(ctx context.Context, deviceID string, keyExpiryDisabled bool) APIResponse[interface{}] {
	keyConfig := DeviceKey{KeyExpiryDisabled: keyExpiryDisabled}
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/device/%s/key", deviceID), keyConfig)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}

// SetDeviceTags sets tags for a device
func (c *APIClient) SetDeviceTags(ctx context.Context, deviceID string, tags []string) APIResponse[interface{}] {
	deviceTags := DeviceTags{Tags: tags}
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/device/%s/tags", deviceID), deviceTags)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}

// SetDeviceRoutes sets the enabled routes for a device
func (c *APIClient) SetDeviceRoutes(ctx context.Context, deviceID string, routes []string) APIResponse[interface{}] {
	deviceRoutes := DeviceRoutes{EnabledRoutes: routes}
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/device/%s/routes", deviceID), deviceRoutes)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}

// Tailnet Management Methods

// GetTailnetInfo retrieves information about the tailnet
func (c *APIClient) GetTailnetInfo(ctx context.Context) APIResponse[TailnetInfo] {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/tailnet/%s", c.tailnet), nil)
	if err != nil {
		return createErrorResponse[TailnetInfo](err)
	}

	body, err := c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[TailnetInfo](err)
	}

	var tailnetInfo TailnetInfo
	if err := json.Unmarshal(body, &tailnetInfo); err != nil {
		return createErrorResponse[TailnetInfo](fmt.Errorf("failed to parse tailnet info: %w", err))
	}

	return createSuccessResponse(tailnetInfo, resp.StatusCode)
}

// TestConnection tests the API connectivity
func (c *APIClient) TestConnection(ctx context.Context) APIResponse[map[string]string] {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/tailnet/%s", c.tailnet), nil)
	if err != nil {
		return createErrorResponse[map[string]string](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[map[string]string](err)
	}

	result := map[string]string{"status": "connected"}
	return createSuccessResponse(result, resp.StatusCode)
}

// Authentication Key Methods

// ListAuthKeys retrieves all authentication keys for the tailnet
func (c *APIClient) ListAuthKeys(ctx context.Context) APIResponse[AuthKeyListResponse] {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/tailnet/%s/keys", c.tailnet), nil)
	if err != nil {
		return createErrorResponse[AuthKeyListResponse](err)
	}

	body, err := c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[AuthKeyListResponse](err)
	}

	var authKeys AuthKeyListResponse
	if err := json.Unmarshal(body, &authKeys); err != nil {
		return createErrorResponse[AuthKeyListResponse](fmt.Errorf("failed to parse auth keys: %w", err))
	}

	return createSuccessResponse(authKeys, resp.StatusCode)
}

// CreateAuthKey creates a new authentication key
func (c *APIClient) CreateAuthKey(ctx context.Context, keyRequest AuthKeyRequest) APIResponse[AuthKey] {
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/tailnet/%s/keys", c.tailnet), keyRequest)
	if err != nil {
		return createErrorResponse[AuthKey](err)
	}

	body, err := c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[AuthKey](err)
	}

	var authKey AuthKey
	if err := json.Unmarshal(body, &authKey); err != nil {
		return createErrorResponse[AuthKey](fmt.Errorf("failed to parse auth key: %w", err))
	}

	return createSuccessResponse(authKey, resp.StatusCode)
}

// DeleteAuthKey deletes an authentication key
func (c *APIClient) DeleteAuthKey(ctx context.Context, keyID string) APIResponse[interface{}] {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/tailnet/%s/keys/%s", c.tailnet, keyID), nil)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	_, err = c.handleResponse(resp)
	if err != nil {
		return createErrorResponse[interface{}](err)
	}

	return createSuccessResponse[interface{}](nil, resp.StatusCode)
}
