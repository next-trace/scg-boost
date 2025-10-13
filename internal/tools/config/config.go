package config

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

type configGetInput struct {
	Key string `json:"key"`
}

type configListInput struct {
	Prefix string `json:"prefix,omitempty"`
}

// Register registers read-only config tools: config.get and config.list.
func Register(s internal_mcp.ToolAdder, c types.SafeConfig) error {
	if c == nil {
		return fmt.Errorf("config: nil safe config")
	}

	// config.get
	getTool := mcp.NewTool(
		"config.get",
		mcp.WithDescription("Get a configuration value by key."),
		mcp.WithInputSchema[configGetInput](),
	)
	getHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key := request.GetString("key", "")
		if key == "" {
			return mcp.NewToolResultError("key is required"), nil
		}
		v, ok := c.Get(key)
		if !ok {
			return mcp.NewToolResultJSON(map[string]any{"error": "not found"})
		}
		return mcp.NewToolResultJSON(map[string]any{"key": key, "value": v})
	}
	if err := s.AddTool(getTool, getHandler); err != nil {
		return fmt.Errorf("register config.get: %w", err)
	}

	// config.list
	listTool := mcp.NewTool(
		"config.list",
		mcp.WithDescription("List configuration keys and values, optionally filtered by a prefix."),
		mcp.WithInputSchema[configListInput](),
	)
	listHandler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prefix := request.GetString("prefix", "")
		items := c.List(prefix)
		return mcp.NewToolResultJSON(items)
	}
	if err := s.AddTool(listTool, listHandler); err != nil {
		return fmt.Errorf("register config.list: %w", err)
	}

	return nil
}
