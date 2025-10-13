package dbquery

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
)

func TestIsReadOnly(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{"simple select", "SELECT * FROM users", true},
		{"simple with", "WITH u AS (SELECT * FROM users) SELECT * FROM u", true},
		{"leading space", "  SELECT * FROM users", true},
		{"leading comment", "-- comment\nSELECT * FROM users", true},
		{"block comment", "/* comment */ SELECT * FROM users", true},
		{"trailing semicolon", "SELECT * FROM users;", true},
		{"simple insert", "INSERT INTO users VALUES (1, 'test')", false},
		{"simple update", "UPDATE users SET name = 'test' WHERE id = 1", false},
		{"simple delete", "DELETE FROM users WHERE id = 1", false},
		{"simple alter", "ALTER TABLE users ADD COLUMN new_col INT", false},
		{"simple drop", "DROP TABLE users", false},
		{"simple truncate", "TRUNCATE TABLE users", false},
		{"select with semicolon and insert", "SELECT * FROM users; INSERT INTO logs VALUES (1, 'test')", false},
		{"insert with semicolon and select", "INSERT INTO users VALUES (1, 'test'); SELECT * FROM users", false},
		{"multiple statements with write", "SELECT * FROM users; UPDATE users SET name = 'test' WHERE id = 1; SELECT * FROM logs", false},
		{"copy command", "COPY (SELECT 1) TO STDOUT", false},
		{"call procedure", "CALL my_proc()", false},
		{"do block", "DO $$ BEGIN END $$", false},
		{"merge statement", "MERGE INTO t USING s ON t.id = s.id WHEN MATCHED THEN UPDATE SET val = s.val", false},
		{"create table", "CREATE TABLE test (id INT)", false},
		{"grant privileges", "GRANT SELECT ON users TO user1", false},
		{"revoke privileges", "REVOKE SELECT ON users FROM user1", false},
		{"vacuum table", "VACUUM users", false},
		{"analyze table", "ANALYZE users", false},
		{"reindex table", "REINDEX TABLE users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReadOnly(tt.query); got != tt.want {
				t.Errorf("IsReadOnly() = %v, want %v", got, tt.want)
			}
		})
	}
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

func TestRegister_RowLimit(t *testing.T) {
	// 1. Create a mockDBConn with more rows than the limit

	rows := make([]map[string]any, 300)
	for i := 0; i < 300; i++ {

		rows[i] = map[string]any{"id": i}
	}
	db := &mockDBConn{rows: rows}

	// 2. Create a mock ToolAdder
	toolAdder := &mockToolAdder{}

	// 3. Call Register
	maxRows := 100
	timeout := 3 * time.Second
	err := Register(toolAdder, db, maxRows, timeout)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// 4. Call the handler
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "dbquery.run",
			Arguments: map[string]any{
				"query": "SELECT * FROM users",
			},
		},
	}
	result, err := toolAdder.handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}

	// 5. Check the number of rows
	resMap, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("result.StructuredContent is not a map[string]any")
	}
	resRows, ok := resMap["rows"].([]map[string]interface{})
	if !ok {
		t.Fatalf("rows is not a []map[string]interface{}, but %T", resMap["rows"])
	}
	if len(resRows) != maxRows {
		t.Errorf("len(rows) = %d, want %d", len(resRows), maxRows)
	}
}
