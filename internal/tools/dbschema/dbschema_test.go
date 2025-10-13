package dbschema

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
)

type mockDBConn struct {
	schemas []string
	tables  map[string][]string
	columns map[string]map[string][]map[string]any
	err     error
}

func (m *mockDBConn) Schemas(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.schemas, nil
}

func (m *mockDBConn) Tables(ctx context.Context, schema string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tables[schema], nil
}

func (m *mockDBConn) Columns(ctx context.Context, schema, table string) ([]map[string]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.columns[schema][table], nil
}

func (m *mockDBConn) QueryJSON(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	return nil, nil // not used in this test
}

type mockToolAdder struct {
	tool    mcp.Tool
	handler internal_mcp.ToolHandler
}

func (m *mockToolAdder) AddTool(tool mcp.Tool, handler internal_mcp.ToolHandler) error {
	m.tool = tool
	m.handler = handler
	return nil
}

func (m *mockToolAdder) AddResource(resource mcp.Resource, handler internal_mcp.ResourceHandler) error {
	// not needed for this test
	return nil
}

func TestRegister(t *testing.T) {
	db := &mockDBConn{
		schemas: []string{"public", "private"},
		tables: map[string][]string{
			"public":  {"users", "products"},
			"private": {"secrets"},
		},
		columns: map[string]map[string][]map[string]any{
			"public": {
				"users": {
					{"name": "id", "type": "integer", "nullable": false},
					{"name": "name", "type": "text", "nullable": true},
				},
				"products": {
					{"name": "id", "type": "integer", "nullable": false},
					{"name": "price", "type": "numeric", "nullable": false},
				},
			},
			"private": {
				"secrets": {
					{"name": "key", "type": "text", "nullable": false},
					{"name": "value", "type": "text", "nullable": false},
				},
			},
		},
	}

	t.Run("allowlist disabled", func(t *testing.T) {
		toolAdder := &mockToolAdder{}
		err := Register(toolAdder, db, nil)
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		req := mcp.CallToolRequest{}
		result, err := toolAdder.handler(context.Background(), req)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}

		resMap, ok := result.StructuredContent.(map[string]any)
		if !ok {
			t.Fatalf("result.StructuredContent is not a map[string]any")
		}
		tables, ok := resMap["tables"].([]map[string]interface{})
		if !ok {
			t.Fatalf("tables is not a []map[string]interface{}, but %T", resMap["tables"])
		}

		if len(tables) != 3 {
			t.Errorf("len(tables) = %d, want 3", len(tables))
		}
	})

	t.Run("allowlist enabled", func(t *testing.T) {
		toolAdder := &mockToolAdder{}
		err := Register(toolAdder, db, []string{"public"})
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		req := mcp.CallToolRequest{}
		result, err := toolAdder.handler(context.Background(), req)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}

		resMap, ok := result.StructuredContent.(map[string]any)
		if !ok {
			t.Fatalf("result.StructuredContent is not a map[string]any")
		}
		tables, ok := resMap["tables"].([]map[string]interface{})
		if !ok {
			t.Fatalf("tables is not a []map[string]interface{}, but %T", resMap["tables"])
		}

		if len(tables) != 2 {
			t.Errorf("len(tables) = %d, want 2", len(tables))
		}
		for _, tableMap := range tables {
			if tableMap["schema"] != "public" {
				t.Errorf("schema = %s, want public", tableMap["schema"])
			}
		}
	})
}
