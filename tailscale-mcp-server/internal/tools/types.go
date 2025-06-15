package tools

import (
	"context"
	"encoding/json"
)

// Tool defines the interface for all tools in the system.
type Tool interface {
	Name() string
	Description() string
	InputSchema() any
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}
