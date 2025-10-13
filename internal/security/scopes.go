package security

import (
	"context"

	"github.com/next-trace/scg-boost/types"
)

// Scopes required for each tool.
const (
	ScopeAppInfoGet         = "appinfo.get"
	ScopeConfigGet          = "config.get"
	ScopeConfigList         = "config.list"
	ScopeDBSchemaList       = "dbschema.list"
	ScopeDBQueryRun         = "dbquery.run"
	ScopeLogsLastError      = "logs.lastError"
	ScopeHealthStatus       = "health.status"
	ScopeEventsOutboxPeek   = "events.outbox.peek"
	ScopeTraceLookup        = "trace.lookup"
	ScopeServiceTopology    = "service.topology"
	ScopeResourceGuidelines = "resource.guidelines" // For accessing static resources
	ScopeDBRead             = "db.read"             // For database read operations
)

// Additional scopes for new tools.
const (
	ScopeRoutesList       = "routes.list"
	ScopeMigrationsStatus = "migrations.status"
	ScopeCacheStats       = "cache.stats"
	ScopeDocsSearch       = "docs.search"
	ScopeMetricsSummary   = "metrics.summary"
	ScopeEnvCheck         = "env.check"
)

// ToolScopes maps tool names to their required scopes.
var ToolScopes = map[string][]string{
	"appinfo.get":         {ScopeAppInfoGet},
	"config.get":          {ScopeConfigGet},
	"config.list":         {ScopeConfigList},
	"dbschema.list":       {ScopeDBSchemaList, ScopeDBRead},
	"dbquery.run":         {ScopeDBQueryRun, ScopeDBRead},
	"logs.lastError":      {ScopeLogsLastError},
	"health.status":       {ScopeHealthStatus},
	"events.outbox.peek":  {ScopeEventsOutboxPeek},
	"trace.lookup":        {ScopeTraceLookup},
	"service.topology":    {ScopeServiceTopology},
	"routes.list":         {ScopeRoutesList},
	"migrations.status":   {ScopeMigrationsStatus},
	"cache.stats":         {ScopeCacheStats},
	"docs.search":         {ScopeDocsSearch},
	"metrics.summary":     {ScopeMetricsSummary},
	"env.check":           {ScopeEnvCheck},
	"resource.guidelines": {ScopeResourceGuidelines},
}

// AllowAllAuthorizer is a development-only authorizer that grants all scopes.
type AllowAllAuthorizer struct{}

// HasScope always returns true for AllowAllAuthorizer.
func (a *AllowAllAuthorizer) HasScope(ctx context.Context, tool string) bool {
	return true
}

// NewAllowAllAuthorizer creates a new AllowAllAuthorizer.
func NewAllowAllAuthorizer() types.Authorizer {
	return &AllowAllAuthorizer{}
}

// GetToolScopes returns the required scopes for a tool.
// Returns nil if tool has no scope requirements.
func GetToolScopes(toolName string) []string {
	return ToolScopes[toolName]
}
