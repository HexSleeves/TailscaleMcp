package tools

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// AdminTool provides administrative functionality
type AdminTool struct {
	cli *tailscale.TailscaleCLI
	api *tailscale.APIClient
}

// NewAdminTool creates a new admin tool
func NewAdminTool(cli *tailscale.TailscaleCLI, api *tailscale.APIClient) *AdminTool {
	return &AdminTool{
		cli: cli,
		api: api,
	}
}

// Name returns the tool name
func (a *AdminTool) Name() string {
	return "admin"
}

// Description returns the tool description
func (a *AdminTool) Description() string {
	return "Administrative operations including user management, settings, and system configuration"
}

// InputSchema returns the JSON schema for tool input
func (a *AdminTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Administrative action to perform",
				"enum":        []string{"status", "logout", "login", "up", "down", "version"},
			},
			"auth_key": map[string]interface{}{
				"type":        "string",
				"description": "Authentication key for login operations",
			},
			"hostname": map[string]interface{}{
				"type":        "string",
				"description": "Hostname for the device",
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the admin tool
func (a *AdminTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action, ok := input["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required and must be a string")
	}

	switch action {
	case "status":
		return a.getStatus(ctx)
	case "logout":
		return a.logout(ctx)
	case "login":
		authKey, _ := input["auth_key"].(string)
		return a.login(ctx, authKey)
	case "up":
		hostname, _ := input["hostname"].(string)
		return a.up(ctx, hostname)
	case "down":
		return a.down(ctx)
	case "version":
		return a.getVersion(ctx)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (a *AdminTool) getStatus(ctx context.Context) (interface{}, error) {
	result, err := a.cli.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return result, nil
}

func (a *AdminTool) logout(ctx context.Context) (interface{}, error) {
	err := a.cli.Logout()
	if err != nil {
		return nil, fmt.Errorf("logout failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Successfully logged out",
	}, nil
}

func (a *AdminTool) login(ctx context.Context, authKey string) (interface{}, error) {
	if authKey == "" {
		return nil, fmt.Errorf("auth_key is required for login action")
	}

	// Use Up command with auth key for login
	options := &tailscale.UpOptions{
		AuthKey: authKey,
	}
	err := a.cli.Up(options)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Successfully logged in",
	}, nil
}

func (a *AdminTool) up(ctx context.Context, hostname string) (interface{}, error) {
	options := &tailscale.UpOptions{}
	if hostname != "" {
		options.Hostname = hostname
	}

	err := a.cli.Up(options)
	if err != nil {
		return nil, fmt.Errorf("up command failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Tailscale is now up",
	}, nil
}

func (a *AdminTool) down(ctx context.Context) (interface{}, error) {
	err := a.cli.Down()
	if err != nil {
		return nil, fmt.Errorf("down command failed: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Tailscale is now down",
	}, nil
}

func (a *AdminTool) getVersion(ctx context.Context) (interface{}, error) {
	result, err := a.cli.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return map[string]interface{}{
		"version": result,
	}, nil
}
