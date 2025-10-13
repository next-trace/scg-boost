package migrations

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the migrations.status tool.
func Register(s internal_mcp.ToolAdder, mr types.MigrationReader) error {
	if mr == nil {
		return nil // Tool not registered if no migration reader
	}

	tool := mcp.NewTool(
		"migrations.status",
		mcp.WithDescription("Get the status of all database migrations."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		statuses, err := mr.Status(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to get migration status", map[string]any{"error": err.Error()}), nil
		}

		pending := 0
		applied := 0
		for _, st := range statuses {
			if st.Applied {
				applied++
			} else {
				pending++
			}
		}

		result := map[string]any{
			"total":      len(statuses),
			"applied":    applied,
			"pending":    pending,
			"migrations": statuses,
		}

		return internal_mcp.NewToolResultJSON(result)
	}

	return s.AddTool(tool, handler)
}
