package dbquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

type dbQueryRunInput struct {
	Query  string         `json:"query"`
	Params map[string]any `json:"params,omitempty"`
}

// Register registers the dbquery.run tool with read-only enforcement.
func Register(s internal_mcp.ToolAdder, db types.DBConn, maxRows int, timeout time.Duration) error {
	if db == nil {
		return fmt.Errorf("dbquery: nil db")
	}
	if maxRows <= 0 {
		maxRows = 500
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	tool := mcp.NewTool(
		"dbquery.run",
		mcp.WithDescription("Execute a read-only SQL query against the database."),
		mcp.WithInputSchema[dbQueryRunInput](),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawQ := request.GetString("query", "")
		if strings.TrimSpace(rawQ) == "" {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInvalidInput, "missing query", nil), nil
		}
		if !IsReadOnly(rawQ) {
			return internal_mcp.ToolError(internal_mcp.ErrCodeReadOnly, "Only SELECT/CTE queries are allowed", map[string]any{"hint": "read-only enforced"}), nil
		}

		var params map[string]any
		if ps, ok := request.GetArguments()["params"]; ok {
			if p, ok := ps.(map[string]any); ok {
				params = p
			}
		}

		// Enforce MaxRows: append LIMIT if missing
		finalQuery := rawQ
		finalParams := params
		if !strings.Contains(strings.ToUpper(rawQ), "LIMIT") {
			finalQuery = rawQ + " LIMIT :__max_rows"
			if finalParams == nil {
				finalParams = make(map[string]any)
			}
			finalParams["__max_rows"] = maxRows
		}

		cctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		rows, err := db.QueryJSON(cctx, finalQuery, finalParams)
		if err != nil {
			return internal_mcp.ToolError(internal_mcp.ErrCodeInternal, "query failed", map[string]any{"error": err.Error()}), nil
		}
		if len(rows) > maxRows {
			rows = rows[:maxRows]
		}
		return internal_mcp.NewToolResultJSON(map[string]any{"rows": rows, "rowCount": len(rows)})
	}

	if err := s.AddTool(tool, handler); err != nil {
		return fmt.Errorf("register dbquery.run: %w", err)
	}
	return nil
}
