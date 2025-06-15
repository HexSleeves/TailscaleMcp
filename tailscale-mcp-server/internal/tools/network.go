package tools

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// NetworkTool provides network management functionality
type NetworkTool struct {
	cli *tailscale.TailscaleCLI
	api *tailscale.APIClient
}

// NewNetworkTool creates a new network tool
func NewNetworkTool(cli *tailscale.TailscaleCLI, api *tailscale.APIClient) *NetworkTool {
	return &NetworkTool{
		cli: cli,
		api: api,
	}
}

// Name returns the tool name
func (n *NetworkTool) Name() string {
	return "network"
}

// Description returns the tool description
func (n *NetworkTool) Description() string {
	return "Network operations including ping, connectivity tests, and route management"
}

// InputSchema returns the JSON schema for tool input
func (n *NetworkTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Network action to perform",
				"enum":        []string{"ping", "routes", "connectivity", "ip"},
			},
			"target": map[string]interface{}{
				"type":        "string",
				"description": "Target host or IP for network operations",
			},
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of ping packets to send",
				"default":     4,
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the network tool
func (n *NetworkTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action, ok := input["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required and must be a string")
	}

	switch action {
	case "ping":
		target, _ := input["target"].(string)
		count, _ := input["count"].(float64)
		if count == 0 {
			count = 4
		}
		return n.ping(ctx, target, int(count))
	case "routes":
		return n.getRoutes(ctx)
	case "connectivity":
		return n.testConnectivity(ctx)
	case "ip":
		return n.getIP(ctx)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (n *NetworkTool) ping(ctx context.Context, target string, count int) (interface{}, error) {
	if target == "" {
		return nil, fmt.Errorf("target is required for ping action")
	}

	result, err := n.cli.Ping(target, count)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return result, nil
}

func (n *NetworkTool) getRoutes(ctx context.Context) (interface{}, error) {
	// Implementation for getting routes
	result, err := n.cli.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	return map[string]interface{}{
		"routes":  result,
		"message": "Route information retrieved",
	}, nil
}

func (n *NetworkTool) testConnectivity(ctx context.Context) (interface{}, error) {
	// Test basic connectivity using netcheck
	result, err := n.cli.Netcheck()
	if err != nil {
		return nil, fmt.Errorf("connectivity test failed: %w", err)
	}

	return map[string]interface{}{
		"connected": true,
		"result":    result,
		"message":   "Connectivity test successful",
	}, nil
}

func (n *NetworkTool) getIP(ctx context.Context) (interface{}, error) {
	result, err := n.cli.IP()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP: %w", err)
	}

	return result, nil
}
