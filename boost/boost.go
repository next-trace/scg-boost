package boost

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	internal_mcp "github.com/next-trace/scg-boost/internal/mcp"
	"github.com/next-trace/scg-boost/internal/tools/appinfo"
	"github.com/next-trace/scg-boost/internal/tools/cache"
	"github.com/next-trace/scg-boost/internal/tools/config"
	"github.com/next-trace/scg-boost/internal/tools/dbquery"
	"github.com/next-trace/scg-boost/internal/tools/dbschema"
	"github.com/next-trace/scg-boost/internal/tools/docs"
	"github.com/next-trace/scg-boost/internal/tools/env"
	"github.com/next-trace/scg-boost/internal/tools/events"
	"github.com/next-trace/scg-boost/internal/tools/health"
	"github.com/next-trace/scg-boost/internal/tools/logs"
	"github.com/next-trace/scg-boost/internal/tools/metrics"
	"github.com/next-trace/scg-boost/internal/tools/migrations"
	"github.com/next-trace/scg-boost/internal/tools/routes"
	"github.com/next-trace/scg-boost/internal/tools/service"
	"github.com/next-trace/scg-boost/internal/tools/trace"
	"github.com/next-trace/scg-boost/resources"
	"github.com/next-trace/scg-boost/types"
)

// Version is the current version of scg-boost.
const Version = "0.2.0"

// New creates a new Server with the provided options. It returns a server
// instance or an error if the configuration is invalid.
func New(opts ...Option) (Server, error) {
	o := Options{
		Name:           "scg-boost",
		Version:        "0.1.0",
		MaxRows:        500,
		DBQueryTimeout: 3 * time.Second,
	}
	for _, fn := range opts {
		if fn != nil {
			fn(&o)
		}
	}

	if o.Logger == nil {
		o.Logger = &nopLogger{}
	}
	if o.Authorizer == nil {
		// Default to deny all if no authorizer is provided.
		o.Authorizer = &denyAllAuthorizer{}
	}

	mcpServer := internal_mcp.NewStdioServer(o.Name, o.Version)
	authorizedServer := internal_mcp.NewAuthorizedServer(mcpServer, o.Authorizer, o.Logger)

	s := &server{
		o:   o,
		mcp: authorizedServer,
	}

	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("register tools: %w", err)
	}

	return s, nil
}

// Server is the main SCG-Boost server interface.
type Server interface {
	// Start runs the MCP server in a background goroutine over stdio. It returns
	// a stop function and an error if startup fails. The server will run until
	// the provided context is canceled.
	Start(ctx context.Context) (stop func() error, err error)
}

type server struct {
	o   Options
	mcp *internal_mcp.AuthorizedServer
}

// Start implements the Server interface.
func (s *server) Start(ctx context.Context) (func() error, error) {
	go func() {
		if err := s.mcp.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.o.Logger.Error("mcp server run failed", map[string]any{"error": err.Error()})
		}
	}()

	stop := func() error {
		// The server is stopped by canceling the context passed to Start.
		return nil
	}

	return stop, nil
}

func (s *server) registerTools() error {
	// App Info
	appInfoData := map[string]any{
		"name":    s.o.Name,
		"version": s.o.Version,
		"go":      runtime.Version(),
	}
	s.registerTool("appinfo.get", appinfo.Register(s.mcp, appInfoData))

	// Config
	if s.o.Config != nil {
		s.registerTool("config.get", config.Register(s.mcp, s.o.Config))
	}

	// DB
	if s.o.DB != nil {
		s.registerTool("dbschema.list", dbschema.Register(s.mcp, s.o.DB, s.o.AllowSchemas))
		s.registerTool("dbquery.run", dbquery.Register(s.mcp, s.o.DB, s.o.MaxRows, s.o.DBQueryTimeout))
	}

	// Logs
	if s.o.LogStore != nil {
		// Adapt LogStore to LogReader
		logReader := &logStoreAdapter{store: s.o.LogStore}
		s.registerTool("logs.lastError", logs.Register(s.mcp, logReader))
	}

	// Health
	if s.o.HealthProbe != nil {
		s.registerTool("health.status", health.Register(s.mcp, s.o.HealthProbe))
	}

	// Events
	if s.o.OutboxReader != nil {
		s.registerTool("events.outbox.peek", events.Register(s.mcp, s.o.OutboxReader))
	}

	// Trace
	if s.o.TraceReader != nil {
		s.registerTool("trace.lookup", trace.Register(s.mcp, s.o.TraceReader))
	}

	// Service Topology
	if s.o.TopologyProvider != nil {
		s.registerTool("service.topology", service.Register(s.mcp, s.o.TopologyProvider))
	}

	// Routes
	if s.o.RouteProvider != nil {
		s.registerTool("routes.list", routes.Register(s.mcp, s.o.RouteProvider))
	}

	// Migrations
	if s.o.MigrationReader != nil {
		s.registerTool("migrations.status", migrations.Register(s.mcp, s.o.MigrationReader))
	}

	// Cache
	if s.o.CacheInspector != nil {
		s.registerTool("cache.stats", cache.Register(s.mcp, s.o.CacheInspector))
	}

	// Docs
	if s.o.DocsSearcher != nil {
		s.registerTool("docs.search", docs.Register(s.mcp, s.o.DocsSearcher))
	}

	// Metrics
	if s.o.MetricsReader != nil {
		s.registerTool("metrics.summary", metrics.Register(s.mcp, s.o.MetricsReader))
	}

	// Env
	if s.o.EnvChecker != nil {
		s.registerTool("env.check", env.Register(s.mcp, s.o.EnvChecker))
	}

	// Guidelines
	guidelinesContent, err := resources.Guidelines()
	if err != nil {
		return fmt.Errorf("read guidelines: %w", err)
	}
	if err := internal_mcp.RegisterBaseResources(s.mcp, string(guidelinesContent)); err != nil {
		return fmt.Errorf("register base resources: %w", err)
	}

	// Project-scoped resources (optional)
	if s.o.ProjectRoot != "" {
		summary := s.o.ProjectSummaryMarkdown
		if summary == "" {
			summary = "# SCG Project Summary\n\n(no summary provided)\n"
		}
		if err := internal_mcp.RegisterProjectResources(s.mcp, internal_mcp.ProjectResourceOptions{Root: s.o.ProjectRoot}, summary); err != nil {
			s.o.Logger.Error("failed to register project resources", map[string]any{"error": err.Error()})
		}
	}

	return nil
}

func (s *server) registerTool(name string, toolErr error) {
	if toolErr != nil {
		s.o.Logger.Error(fmt.Sprintf("failed to register tool %s", name), map[string]any{"error": toolErr.Error()})
	}
}

type nopLogger struct{}

func (l *nopLogger) Debug(string, map[string]any) {}
func (l *nopLogger) Error(string, map[string]any) {}

type denyAllAuthorizer struct{}

func (a *denyAllAuthorizer) HasScope(ctx context.Context, tool string) bool {
	return false
}

// logStoreAdapter adapts LogStore to LogReader interface.
type logStoreAdapter struct {
	store types.LogStore
}

func (a *logStoreAdapter) LastError(ctx context.Context) (*types.LogEntry, error) {
	ts, msg, fields, err := a.store.LastError(ctx)
	if err != nil {
		return nil, err
	}
	if ts == "" && msg == "" {
		return nil, nil
	}
	// Parse timestamp, assuming it's RFC3339
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, err
	}
	level := "error"
	if lvl, ok := fields["level"].(string); ok {
		level = lvl
	}
	return &types.LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   msg,
	}, nil
}
