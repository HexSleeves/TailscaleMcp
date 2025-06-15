package tools

import (
	"context"
	"fmt"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// ACLTool provides Access Control List management functionality
type ACLTool struct {
	cli *tailscale.TailscaleCLI
	api *tailscale.APIClient
}

// NewACLTool creates a new ACL tool
func NewACLTool(cli *tailscale.TailscaleCLI, api *tailscale.APIClient) *ACLTool {
	return &ACLTool{
		cli: cli,
		api: api,
	}
}

// Name returns the tool name
func (a *ACLTool) Name() string {
	return "acl"
}

// Description returns the tool description
func (a *ACLTool) Description() string {
	return "Access Control List management including viewing, updating, and validating ACL policies"
}

// InputSchema returns the JSON schema for tool input
func (a *ACLTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "ACL action to perform",
				"enum":        []string{"get", "set", "validate", "test"},
			},
			"policy": map[string]interface{}{
				"type":        "string",
				"description": "ACL policy JSON for set operations",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Source IP or user for ACL testing",
			},
			"destination": map[string]interface{}{
				"type":        "string",
				"description": "Destination IP or service for ACL testing",
			},
			"port": map[string]interface{}{
				"type":        "integer",
				"description": "Port number for ACL testing",
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the ACL tool
func (a *ACLTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	action, ok := input["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required and must be a string")
	}

	switch action {
	case "get":
		return a.getACL(ctx)
	case "set":
		policy, _ := input["policy"].(string)
		return a.setACL(ctx, policy)
	case "validate":
		policy, _ := input["policy"].(string)
		return a.validateACL(ctx, policy)
	case "test":
		source, _ := input["source"].(string)
		destination, _ := input["destination"].(string)
		port, _ := input["port"].(float64)
		return a.testACL(ctx, source, destination, int(port))
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (a *ACLTool) getACL(ctx context.Context) (interface{}, error) {
	if a.api != nil {
		acl := a.api.GetACL(ctx) // ← single result
		return acl, nil
	}

	return nil, fmt.Errorf("ACL retrieval not available – requires API access")
}

func (a *ACLTool) setACL(ctx context.Context, policy string) (interface{}, error) {
	if policy == "" {
		return nil, fmt.Errorf("policy is required for set action")
	}

	// SetACL method not implemented in API client yet
	return nil, fmt.Errorf("ACL modification not available - SetACL method not implemented")
}

func (a *ACLTool) validateACL(ctx context.Context, policy string) (interface{}, error) {
	if policy == "" {
		return nil, fmt.Errorf("policy is required for validate action")
	}

	// ValidateACL method not implemented in API client yet
	return nil, fmt.Errorf("ACL validation not available - ValidateACL method not implemented")
}

func (a *ACLTool) testACL(ctx context.Context, source, destination string, port int) (interface{}, error) {
	if source == "" || destination == "" {
		return nil, fmt.Errorf("source and destination are required for test action")
	}

	// TestACL method not implemented in API client yet
	return nil, fmt.Errorf("ACL testing not available - TestACL method not implemented")
}
