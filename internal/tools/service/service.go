package service

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the service.topology tool.
func Register(s internal_mcp.ToolAdder, tp types.TopologyProvider) error {
	if tp == nil {
		return nil // Tool not registered if no topology provider
	}

	tool := mcp.NewTool(
		"service.topology",
		mcp.WithDescription("Get a snapshot of the service topology."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		topology, err := tp.Snapshot(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to get topology", nil), nil
		}
		return internal_mcp.NewToolResultJSON(topology)
	}

	return s.AddTool(tool, handler)
}
