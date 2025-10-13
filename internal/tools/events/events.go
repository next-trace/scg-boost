package events

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the events.outbox.peek tool.
func Register(s internal_mcp.ToolAdder, or types.OutboxReader) error {
	if or == nil {
		return nil // Tool not registered if no outbox reader
	}

	tool := mcp.NewTool(
		"events.outbox.peek",
		mcp.WithDescription("Peek at the most recent events in the outbox."),
		mcp.WithInputSchema[map[string]any](),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int(mcp.ParseFloat64(request, "limit", 10))
		events, err := or.Peek(ctx, limit)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to peek outbox", map[string]any{"error": err.Error()}), nil
		}
		return internal_mcp.NewToolResultJSON(map[string]any{"events": events})
	}

	return s.AddTool(tool, handler)
}
