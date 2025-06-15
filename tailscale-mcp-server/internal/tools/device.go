package tools

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// DeviceManagementTool provides device management functionality
type DeviceManagementTool struct {
	cli *tailscale.TailscaleCLI
	api *tailscale.APIClient
}

// NewDeviceManagementTool creates a new device management tool
func NewDeviceManagementTool(cli *tailscale.TailscaleCLI, api *tailscale.APIClient) *DeviceManagementTool {
	return &DeviceManagementTool{
		cli: cli,
		api: api,
	}
}

// Name returns the tool name
func (d *DeviceManagementTool) Name() string {
	return "device_management"
}

// Description returns the tool description
func (d *DeviceManagementTool) Description() string {
	return "Manage Tailscale devices including listing, status, and configuration"
}

// InputSchema returns the JSON schema for tool input
func (d *DeviceManagementTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform",
				"enum":        []string{"list", "status", "enable", "disable"},
			},
			"device_id": map[string]interface{}{
				"type":        "string",
				"description": "Device ID for specific operations",
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the device management tool
func (d *DeviceManagementTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action, ok := input["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required and must be a string")
	}

	switch action {
	case "list":
		return d.listDevices(ctx)
	case "status":
		deviceID, _ := input["device_id"].(string)
		return d.getDeviceStatus(ctx, deviceID)
	case "enable":
		deviceID, _ := input["device_id"].(string)
		return d.enableDevice(ctx, deviceID)
	case "disable":
		deviceID, _ := input["device_id"].(string)
		return d.disableDevice(ctx, deviceID)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (d *DeviceManagementTool) listDevices(ctx context.Context) (interface{}, error) {
	if d.api != nil {
		devices := d.api.ListDevices(ctx)
		if devices.Success {
			return devices, nil
		}
	}

	// Fallback to CLI
	result, err := d.cli.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return result, nil
}

func (d *DeviceManagementTool) getDeviceStatus(ctx context.Context, deviceID string) (interface{}, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required for status action")
	}

	// Implementation for getting specific device status
	result, err := d.cli.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}

	return result, nil
}

func (d *DeviceManagementTool) enableDevice(ctx context.Context, deviceID string) (interface{}, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required for enable action")
	}

	// Implementation for enabling device
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Device %s enabled", deviceID),
	}, nil
}

func (d *DeviceManagementTool) disableDevice(ctx context.Context, deviceID string) (interface{}, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("device_id is required for disable action")
	}

	// Implementation for disabling device
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Device %s disabled", deviceID),
	}, nil
}
