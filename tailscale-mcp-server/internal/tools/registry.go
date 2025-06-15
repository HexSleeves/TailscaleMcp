package tools

import (
	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

// ToolRegistry holds all registered tools.
type ToolRegistry struct {
	tools map[string]Tool
	api   *tailscale.APIClient
	cli   *tailscale.TailscaleCLI
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry(api *tailscale.APIClient, cli *tailscale.TailscaleCLI) *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
		api:   api,
		cli:   cli,
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// GetTool retrieves a tool from the registry by name.
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetTools returns all registered tools.
func (r *ToolRegistry) GetTools() map[string]Tool {
	return r.tools
}
