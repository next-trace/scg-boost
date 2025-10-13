package logs

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the logs.lastError tool.
func Register(s internal_mcp.ToolAdder, lr types.LogReader) error {
	if lr == nil {
		return nil // Tool not registered if no log reader
	}

	tool := mcp.NewTool(
		"logs.lastError",
		mcp.WithDescription("Get the last recorded error message."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log, err := lr.LastError(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to get last error", nil), nil
		}
		if log == nil {
			return internal_mcp.NewToolResultJSON(map[string]any{"message": "no errors recorded"})
		}
		return internal_mcp.NewToolResultJSON(map[string]any{
			"ts":  log.Timestamp.Format(time.RFC3339),
			"msg": log.Message,
			"lvl": log.Level,
		})
	}

	return s.AddTool(tool, handler)
}
