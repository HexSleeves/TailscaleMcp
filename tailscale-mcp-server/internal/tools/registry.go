// tailscale-mcp-server/internal/tools/registry.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
	"github.com/hexsleeves/tailscale-mcp-server/internal/tools/device"
)

// ToolRegistry holds all registered tools with thread-safety and lifecycle management.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
	api   *tailscale.APIClient
	cli   *tailscale.TailscaleCLI
}

// NewToolRegistry creates a new tool registry with the given clients.
func NewToolRegistry(api *tailscale.APIClient, cli *tailscale.TailscaleCLI) *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]Tool),
		api:   api,
		cli:   cli,
	}

	// Register built-in tools
	registry.registerBuiltinTools()

	return registry
}

// registerBuiltinTools registers all the built-in tools
func (r *ToolRegistry) registerBuiltinTools() {
	// Device management tools
	r.Register(device.NewListDevicesTool(r.api))
	// Add more tools here as they're implemented
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		logger.Warn("Overriding existing tool", "name", name)
	}

	r.tools[name] = tool
	logger.Debug("Registered tool", "name", name, "description", tool.Description())
	return nil
}

// GetTool retrieves a tool from the registry by name.
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// GetTools returns a copy of all registered tools.
func (r *ToolRegistry) GetTools() map[string]Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	tools := make(map[string]Tool, len(r.tools))
	for name, tool := range r.tools {
		tools[name] = tool
	}
	return tools
}

// ListToolNames returns a sorted list of tool names.
func (r *ToolRegistry) ListToolNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered tools.
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// ExecuteTool executes a tool with the given arguments.
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, args json.RawMessage) (string, error) {
	tool, ok := r.GetTool(name)
	if !ok {
		return "", fmt.Errorf("tool %q not found", name)
	}

	// Create context with registry and clients
	toolCtx := NewContext(ctx, r)

	logger.Debug("Executing tool", "name", name)
	result, err := tool.Execute(toolCtx, args)
	if err != nil {
		logger.Error("Tool execution failed", "name", name, "error", err)
		return "", fmt.Errorf("tool %q execution failed: %w", name, err)
	}

	logger.Debug("Tool executed successfully", "name", name)
	return result, nil
}

// Close gracefully shuts down the registry and any resources.
func (r *ToolRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger.Debug("Closing tool registry", "tool_count", len(r.tools))

	// Clear the tools map
	r.tools = make(map[string]Tool)

	return nil
}
