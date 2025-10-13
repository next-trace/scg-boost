package types

import (
	"context"
	"time"
)

// Logger is a minimal structured logger.
type Logger interface {
	Debug(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

// SafeConfig provides access to configuration with redaction.
type SafeConfig interface {
	Get(key string) (string, bool)
	List(prefix string) map[string]string
}

// DBConn is a read-only database connection adapter.
// It allows SCG-Boost to work with any database driver.
type DBConn interface {
	// QueryJSON executes a read-only query and returns results as JSON-like maps.
	// Implementations must prevent write operations.
	QueryJSON(ctx context.Context, query string, params map[string]any) ([]map[string]any, error)

	// Schemas returns a list of all schemas accessible to the current user.
	// Used by the dbschema.list tool.
	Schemas(ctx context.Context) ([]string, error)

	// Tables returns a list of all tables in a given schema.
	// Used by the dbschema.list tool.
	Tables(ctx context.Context, schema string) ([]string, error)

	// Columns returns metadata for all columns in a given table.
	// Used by the dbschema.list tool.
	Columns(ctx context.Context, schema, table string) ([]map[string]any, error)
}

// Authorizer is an interface for checking if a tool can be executed.
type Authorizer interface {
	HasScope(ctx context.Context, tool string) bool
}

// LogEntry represents a log entry.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// LogReader is an interface for retrieving logs.
type LogReader interface {
	LastError(ctx context.Context) (*LogEntry, error)
}

// LogStore is an interface for retrieving the last error log.
type LogStore interface {
	LastError(ctx context.Context) (ts string, msg string, fields map[string]any, err error)
}

// HealthProbe is an interface for checking the liveness and readiness of the service.
type HealthProbe interface {
	Liveness(ctx context.Context) error
	Readiness(ctx context.Context) error
}

// OutboxReader is an interface for peeking into an outbox of events.
type OutboxReader interface {
	Peek(ctx context.Context, limit int) ([]map[string]any, error)
}

// TraceReader is an interface for looking up recent traces.
type TraceReader interface {
	Lookup(ctx context.Context, lastN int) ([]map[string]any, error)
}

// TopologyProvider is an interface for getting a snapshot of the service topology.
type TopologyProvider interface {
	Snapshot(ctx context.Context) (map[string]any, error)
}

// RouteInfo describes a single HTTP/gRPC route.
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	Middlewares []string `json:"middlewares,omitempty"`
}

// RouteProvider exposes registered routes for inspection.
type RouteProvider interface {
	List(ctx context.Context) ([]RouteInfo, error)
}

// MigrationStatus describes the state of a single migration.
type MigrationStatus struct {
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
	Batch   int    `json:"batch,omitempty"`
}

// MigrationReader exposes migration status.
type MigrationReader interface {
	Status(ctx context.Context) ([]MigrationStatus, error)
}

// CacheStats contains cache statistics.
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Keys       int64 `json:"keys"`
	MemoryUsed int64 `json:"memory_used_bytes"`
}

// CacheInspector exposes cache statistics.
type CacheInspector interface {
	Stats(ctx context.Context) (CacheStats, error)
}

// DocMatch represents a documentation search result.
type DocMatch struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}

// DocsSearcher searches project documentation.
type DocsSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]DocMatch, error)
}

// MetricsReader exposes metrics summary.
type MetricsReader interface {
	Summary(ctx context.Context) (map[string]any, error)
}

// EnvIssue represents an environment configuration issue.
type EnvIssue struct {
	Key      string `json:"key"`
	Severity string `json:"severity"` // "error", "warning", "info"
	Message  string `json:"message"`
}

// EnvChecker validates environment configuration.
type EnvChecker interface {
	Check(ctx context.Context) ([]EnvIssue, error)
}
