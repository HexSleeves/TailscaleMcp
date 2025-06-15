package device

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tools"
)

// ListDevicesInput defines the input schema for the list_devices tool
type ListDevicesInput struct {
	IncludeRoutes bool `json:"includeRoutes" description:"Include route information for each device"`
}

// ListDevicesTool is a tool for listing devices in a Tailscale tailnet.
type ListDevicesTool struct {
	Client *tailscale.APIClient
}

// NewListDevicesTool creates a new ListDevicesTool
func NewListDevicesTool(client *tailscale.APIClient) *ListDevicesTool {
	return &ListDevicesTool{
		Client: client,
	}
}

// Name returns the name of the tool
func (t *ListDevicesTool) Name() string {
	return "list_devices"
}

// Description returns a description of the tool
func (t *ListDevicesTool) Description() string {
	return "Lists all devices in the tailnet, with an option to include route information."
}

// InputSchema returns the input schema for the tool
func (t *ListDevicesTool) InputSchema() any {
	return ListDevicesInput{}
}

// Execute runs the tool
func (t *ListDevicesTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input ListDevicesInput
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("failed to unmarshal input: %w", err)
	}

	resp := t.Client.ListDevices(ctx)
	if !resp.Success {
		return "", fmt.Errorf("failed to list devices: %s", resp.Error)
	}

	var output strings.Builder
	for _, device := range resp.Data.Devices {
		authStatus := "authorized"
		if !device.Authorized {
			authStatus = "unauthorized"
		}

		output.WriteString(fmt.Sprintf("Device: %s (%s) - %s\n", device.Name, device.ID, authStatus))
		output.WriteString(fmt.Sprintf("  OS: %s, Version: %s\n", device.OS, device.ClientVersion))
		output.WriteString(fmt.Sprintf("  Addresses: %s\n", strings.Join(device.Addresses, ", ")))
		output.WriteString(fmt.Sprintf("  Last Seen: %s\n", device.LastSeen.Format("2006-01-02 15:04:05")))

		if input.IncludeRoutes {
			output.WriteString(fmt.Sprintf("  Enabled Routes: %s\n", strings.Join(device.EnabledRoutes, ", ")))
			output.WriteString(fmt.Sprintf("  Advertised Routes: %s\n", strings.Join(device.AdvertisedRoutes, ", ")))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

var _ tools.Tool = (*ListDevicesTool)(nil)
