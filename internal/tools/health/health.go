package health

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the health.status tool.
func Register(s internal_mcp.ToolAdder, hp types.HealthProbe) error {
	if hp == nil {
		return nil // Tool not registered if no health probe
	}

	tool := mcp.NewTool(
		"health.status",
		mcp.WithDescription("Get the health status of the service, including liveness and readiness."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		livenessErr := hp.Liveness(ctx)
		readinessErr := hp.Readiness(ctx)

		status := map[string]any{
			"liveness":  "ok",
			"readiness": "ok",
		}
		if livenessErr != nil {
			status["liveness"] = "fail"
		}
		if readinessErr != nil {
			status["readiness"] = "fail"
		}

		return internal_mcp.NewToolResultJSON(status)
	}

	return s.AddTool(tool, handler)
}
