package metrics

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the metrics.summary tool.
func Register(s internal_mcp.ToolAdder, mr types.MetricsReader) error {
	if mr == nil {
		return nil // Tool not registered if no metrics reader
	}

	tool := mcp.NewTool(
		"metrics.summary",
		mcp.WithDescription("Get a summary of application metrics (Prometheus/OpenMetrics)."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		summary, err := mr.Summary(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to get metrics summary", map[string]any{"error": err.Error()}), nil
		}

		return internal_mcp.NewToolResultJSON(summary)
	}

	return s.AddTool(tool, handler)
}
