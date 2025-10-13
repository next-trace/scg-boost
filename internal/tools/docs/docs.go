package docs

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the docs.search tool.
func Register(s internal_mcp.ToolAdder, ds types.DocsSearcher) error {
	if ds == nil {
		return nil // Tool not registered if no docs searcher
	}

	tool := mcp.NewTool(
		"docs.search",
		mcp.WithDescription("Search project documentation for relevant content."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default: 10)")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := request.GetString("query", "")
		if query == "" {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInvalidInput, "query is required", nil), nil
		}

		limit := 10
		args := request.GetArguments()
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}

		matches, err := ds.Search(ctx, query, limit)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "docs search failed", map[string]any{"error": err.Error()}), nil
		}

		result := map[string]any{
			"query":   query,
			"count":   len(matches),
			"matches": matches,
		}

		return internal_mcp.NewToolResultJSON(result)
	}

	return s.AddTool(tool, handler)
}
