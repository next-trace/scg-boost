package runtime

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ReadOnlyDB struct{ DB *sqlx.DB }

func (r *ReadOnlyDB) QueryJSON(ctx context.Context, query string, params map[string]any) (_ []map[string]any, retErr error) {
	// Use sqlx.Named to bind params, then Rebind for PostgreSQL
	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return nil, fmt.Errorf("bind params: %w", err)
	}
	reboundQuery := sqlx.Rebind(sqlx.DOLLAR, namedQuery)

	rows, err := r.DB.QueryxContext(ctx, reboundQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	var out []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// Schemas returns a list of all non-system schemas.
func (r *ReadOnlyDB) Schemas(ctx context.Context) (_ []string, retErr error) {
	q := `SELECT schema_name FROM information_schema.schemata 
          WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
          ORDER BY schema_name;`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query schemas: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schemas: %w", err)
	}
	return schemas, nil
}

// Tables returns a list of tables in a given schema.
func (r *ReadOnlyDB) Tables(ctx context.Context, schema string) (_ []string, retErr error) {
	q := `SELECT table_name FROM information_schema.tables 
          WHERE table_schema = $1 
          ORDER BY table_name;`
	rows, err := r.DB.QueryContext(ctx, q, schema)
	if err != nil {
		return nil, fmt.Errorf("query tables for schema %q: %w", schema, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}
	return tables, nil
}

// Columns returns metadata for columns in a given table.
func (r *ReadOnlyDB) Columns(ctx context.Context, schema, table string) (_ []map[string]any, retErr error) {
	q := `SELECT column_name, data_type, is_nullable = 'YES' AS nullable
          FROM information_schema.columns
          WHERE table_schema = $1 AND table_name = $2
          ORDER BY ordinal_position;`
	rows, err := r.DB.QueryContext(ctx, q, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query columns for %q.%q: %w", schema, table, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()

	var cols []map[string]any
	for rows.Next() {
		var colName, colType string
		var isNullable bool
		if err := rows.Scan(&colName, &colType, &isNullable); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		cols = append(cols, map[string]any{
			"name":     colName,
			"type":     colType,
			"nullable": isNullable,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate columns: %w", err)
	}
	return cols, nil
}
