package dbquery

import (
	"context"
	"fmt"
)

type mockDBConn struct {
	rows []map[string]any
	err  error
}

func (m *mockDBConn) QueryJSON(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rows, nil
}

func (m *mockDBConn) Schemas(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockDBConn) Tables(ctx context.Context, schema string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockDBConn) Columns(ctx context.Context, schema, table string) ([]map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}
