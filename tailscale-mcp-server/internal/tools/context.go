package tools

import (
	"context"

	"github.com/hexsleeves/tailscale-mcp-server/internal/tailscale"
)

type contextKey string

const (
	registryKey  contextKey = "toolRegistry"
	apiClientKey contextKey = "apiClient"
	cliClientKey contextKey = "cliClient"
)

// NewContext creates a new context with the tool registry and clients.
func NewContext(ctx context.Context, r *ToolRegistry) context.Context {
	ctx = context.WithValue(ctx, registryKey, r)
	ctx = context.WithValue(ctx, apiClientKey, r.api)
	ctx = context.WithValue(ctx, cliClientKey, r.cli)
	return ctx
}

// RegistryFromContext retrieves the tool registry from the context.
func RegistryFromContext(ctx context.Context) (*ToolRegistry, bool) {
	r, ok := ctx.Value(registryKey).(*ToolRegistry)
	return r, ok
}

// APIClientFromContext retrieves the Tailscale API client from the context.
func APIClientFromContext(ctx context.Context) (*tailscale.APIClient, bool) {
	api, ok := ctx.Value(apiClientKey).(*tailscale.APIClient)
	return api, ok
}

// CLIClientFromContext retrieves the Tailscale CLI client from the context.
func CLIClientFromContext(ctx context.Context) (*tailscale.TailscaleCLI, bool) {
	cli, ok := ctx.Value(cliClientKey).(*tailscale.TailscaleCLI)
	return cli, ok
}
