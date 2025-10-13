package boost

import (
	"time"

	"github.com/next-trace/scg-boost/types"
)

// Options configures the SCG-Boost server behavior.
// Zero value is valid; use functional options to override.
type Options struct {
	Name             string
	Version          string
	Logger           types.Logger
	DB               types.DBConn
	Config           types.SafeConfig
	Authorizer       types.Authorizer
	LogStore         types.LogStore
	HealthProbe      types.HealthProbe
	OutboxReader     types.OutboxReader
	TraceReader      types.TraceReader
	TopologyProvider types.TopologyProvider
	AllowSchemas     []string
	MaxRows          int
	DBQueryTimeout   time.Duration

	// ProjectRoot, when set, enables project-scoped MCP resources (summary, tree,...).
	ProjectRoot            string
	ProjectSummaryMarkdown string

	// New providers (Laravel Boost inspired)
	RouteProvider   types.RouteProvider
	MigrationReader types.MigrationReader
	CacheInspector  types.CacheInspector
	DocsSearcher    types.DocsSearcher
	MetricsReader   types.MetricsReader
	EnvChecker      types.EnvChecker
}

// Option applies configuration to Options.
type Option func(*Options)

// WithName sets the application name.
func WithName(name string) Option { return func(o *Options) { o.Name = name } }

// WithVersion sets the application version.
func WithVersion(version string) Option { return func(o *Options) { o.Version = version } }

// WithLogger supplies an optional logger implementation.
func WithLogger(l types.Logger) Option { return func(o *Options) { o.Logger = l } }

// WithDB supplies an optional read-only DB adapter.
func WithDB(db types.DBConn) Option { return func(o *Options) { o.DB = db } }

// WithConfig supplies an optional safe configuration provider.
func WithConfig(cfg types.SafeConfig) Option { return func(o *Options) { o.Config = cfg } }

// WithAllowSchemas restricts DB tools to these schemas.
func WithAllowSchemas(schemas []string) Option {
	return func(o *Options) { o.AllowSchemas = append([]string{}, schemas...) }
}

// WithMaxRows caps result rows for DB queries.
func WithMaxRows(n int) Option { return func(o *Options) { o.MaxRows = n } }

// WithDBQueryTimeout sets the timeout for DB queries.
func WithDBQueryTimeout(d time.Duration) Option { return func(o *Options) { o.DBQueryTimeout = d } }

// WithAuthorizer supplies an optional authorizer for tool access control.
func WithAuthorizer(a types.Authorizer) Option { return func(o *Options) { o.Authorizer = a } }

// WithLogStore supplies an optional log store for retrieving last errors.
func WithLogStore(ls types.LogStore) Option { return func(o *Options) { o.LogStore = ls } }

// WithHealthProbe supplies an optional health probe for liveness and readiness checks.
func WithHealthProbe(h types.HealthProbe) Option { return func(o *Options) { o.HealthProbe = h } }

// WithOutboxReader supplies an optional outbox reader for event peeking.
func WithOutboxReader(or types.OutboxReader) Option { return func(o *Options) { o.OutboxReader = or } }

// WithTraceReader supplies an optional trace reader for looking up recent traces.
func WithTraceReader(tr types.TraceReader) Option { return func(o *Options) { o.TraceReader = tr } }

// WithTopologyProvider supplies an optional topology provider for service topology snapshots.
func WithTopologyProvider(tp types.TopologyProvider) Option {
	return func(o *Options) { o.TopologyProvider = tp }
}

// WithProjectRoot enables project-scoped MCP resources (summary, tree, CLAUDE.md).
// This is used by the scg-boost CLI when running as a per-project Boost server.
func WithProjectRoot(root string) Option {
	return func(o *Options) { o.ProjectRoot = root }
}

// WithProjectSummaryMarkdown overrides the project summary resource content.
func WithProjectSummaryMarkdown(md string) Option {
	return func(o *Options) { o.ProjectSummaryMarkdown = md }
}

// WithProjectResources enables project-scoped resources (summary/tree/claude) for the given root.
func WithProjectResources(root string, summaryMarkdown string) Option {
	return func(o *Options) {
		o.ProjectRoot = root
		o.ProjectSummaryMarkdown = summaryMarkdown
	}
}

// WithRouteProvider supplies an optional route provider for route inspection.
func WithRouteProvider(rp types.RouteProvider) Option {
	return func(o *Options) { o.RouteProvider = rp }
}

// WithMigrationReader supplies an optional migration reader for status checks.
func WithMigrationReader(mr types.MigrationReader) Option {
	return func(o *Options) { o.MigrationReader = mr }
}

// WithCacheInspector supplies an optional cache inspector for statistics.
func WithCacheInspector(ci types.CacheInspector) Option {
	return func(o *Options) { o.CacheInspector = ci }
}

// WithDocsSearcher supplies an optional docs searcher for documentation search.
func WithDocsSearcher(ds types.DocsSearcher) Option {
	return func(o *Options) { o.DocsSearcher = ds }
}

// WithMetricsReader supplies an optional metrics reader for summaries.
func WithMetricsReader(mr types.MetricsReader) Option {
	return func(o *Options) { o.MetricsReader = mr }
}

// WithEnvChecker supplies an optional environment checker.
func WithEnvChecker(ec types.EnvChecker) Option {
	return func(o *Options) { o.EnvChecker = ec }
}
