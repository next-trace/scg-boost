package dbschema

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/types"
)

// Register registers the dbschema.list tool.
func Register(s internal_mcp.ToolAdder, db types.DBConn, allowSchemas []string) error {
	if db == nil {
		return fmt.Errorf("dbschema: nil db")
	}

	tool := mcp.NewTool(
		"dbschema.list",
		mcp.WithDescription("List tables and columns for allowed schemas."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		schemasToList := allowSchemas
		if len(schemasToList) == 0 {
			var err error
			schemasToList, err = db.Schemas(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list schemas: %v", err)), nil
			}
		}

		var tables []map[string]any
		for _, schema := range schemasToList {
			tableNames, err := db.Tables(ctx, schema)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list tables for schema %s: %v", schema, err)), nil
			}

			for _, tableName := range tableNames {
				cols, err := db.Columns(ctx, schema, tableName)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to list columns for table %s.%s: %v", schema, tableName, err)), nil
				}
				tables = append(tables, map[string]any{
					"schema":  schema,
					"table":   tableName,
					"columns": cols,
				})
			}
		}

		return mcp.NewToolResultJSON(map[string]any{"tables": tables})
	}

	if err := s.AddTool(tool, handler); err != nil {
		return fmt.Errorf("register dbschema.list: %w", err)
	}
	return nil
}
