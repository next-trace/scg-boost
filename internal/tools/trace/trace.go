package trace

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the trace.lookup tool.
func Register(s internal_mcp.ToolAdder, tr types.TraceReader) error {
	if tr == nil {
		return nil // Tool not registered if no trace reader
	}

	tool := mcp.NewTool(
		"trace.lookup",
		mcp.WithDescription("Lookup the most recent traces."),
		mcp.WithInputSchema[map[string]any](),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lastN := int(mcp.ParseFloat64(request, "lastN", 10))
		traces, err := tr.Lookup(ctx, lastN)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to lookup traces", map[string]any{"error": err.Error()}), nil
		}
		return internal_mcp.NewToolResultJSON(map[string]any{"traces": traces})
	}

	return s.AddTool(tool, handler)
}
