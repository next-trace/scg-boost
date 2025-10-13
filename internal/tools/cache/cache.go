package cache

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the cache.stats tool.
func Register(s internal_mcp.ToolAdder, ci types.CacheInspector) error {
	if ci == nil {
		return nil // Tool not registered if no cache inspector
	}

	tool := mcp.NewTool(
		"cache.stats",
		mcp.WithDescription("Get cache statistics including hits, misses, keys, and memory usage."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := ci.Stats(ctx)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "failed to get cache stats", map[string]any{"error": err.Error()}), nil
		}

		hitRate := float64(0)
		total := stats.Hits + stats.Misses
		if total > 0 {
			hitRate = float64(stats.Hits) / float64(total) * 100
		}

		result := map[string]any{
			"hits":              stats.Hits,
			"misses":            stats.Misses,
			"hit_rate_percent":  hitRate,
			"keys":              stats.Keys,
			"memory_used_bytes": stats.MemoryUsed,
		}

		return internal_mcp.NewToolResultJSON(result)
	}

	return s.AddTool(tool, handler)
}
