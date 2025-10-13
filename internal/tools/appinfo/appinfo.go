package appinfo

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
)

// Register registers the appinfo.get tool.
func Register(s internal_mcp.ToolAdder, appInfoData map[string]any) error {
	tool := mcp.NewTool(
		"appinfo.get",
		mcp.WithDescription("Get application information, including name, version, Go runtime, OS, Arch, and process details."),
	)

	// Enrich appInfoData with runtime and process information
	appInfoData["os"] = runtime.GOOS
	appInfoData["arch"] = runtime.GOARCH
	appInfoData["pid"] = os.Getpid()
	appInfoData["ppid"] = os.Getppid()
	startTime := time.Now() // Assuming the process starts when the server is created

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Calculate uptime dynamically
		appInfoData["uptime"] = time.Since(startTime).String()
		return internal_mcp.NewToolResultJSON(appInfoData)
	}

	return s.AddTool(tool, handler)
}
