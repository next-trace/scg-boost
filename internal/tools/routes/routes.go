package routes

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the routes.list tool.
func Register(s internal_mcp.ToolAdder, rp types.RouteProvider) error {
	if rp == nil {
		return nil // Tool not registered if no route provider
	}

	tool := mcp.NewTool(
		"routes.list",
		mcp.WithDescription("List all registered HTTP/gRPC routes in the application."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		routes, err := rp.List(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to list routes", map[string]any{"error": err.Error()}), nil
		}

		result := map[string]any{
			"count":  len(routes),
			"routes": routes,
		}

		return internal_mcp.NewToolResultJSON(result)
	}

	return s.AddTool(tool, handler)
}
