package env

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the env.check tool.
func Register(s internal_mcp.ToolAdder, ec types.EnvChecker) error {
	if ec == nil {
		return nil // Tool not registered if no env checker
	}

	tool := mcp.NewTool(
		"env.check",
		mcp.WithDescription("Validate environment configuration and report any issues."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		issues, err := ec.Check(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to check environment", map[string]any{"error": err.Error()}), nil
		}

		errCount := 0
		warnCount := 0
		for _, issue := range issues {
			switch issue.Severity {
			case "error":
				errCount++
			case "warning":
				warnCount++
			}
		}

		status := "ok"
		if errCount > 0 {
			status = "error"
		} else if warnCount > 0 {
			status = "warning"
		}

		result := map[string]any{
			"status":   status,
			"errors":   errCount,
			"warnings": warnCount,
			"issues":   issues,
		}

		return internal_mcp.NewToolResultJSON(result)
	}

	return s.AddTool(tool, handler)
}
