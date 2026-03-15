also# GEMINI.md — scg-boost

This file is the **single source of truth** for how Gemini must work in this repository.

If you violate these rules, you create chaos. Don't.

---

## What This Repo Is

**scg-boost** is a MCP server and context bootstrapper for SupplyChainGuard repositories.

**Purpose:** Give AI assistants (Gemini, Cursor, Junie) a **low-token**, **high-signal** view of Go services:

- Project resources (summary, tree, GEMINI.md)
- Safe tools (config, DB queries, logs, health, metrics)
- Bootstrap templates for `.claude/` directories

**Stack:** Go 1.22+ | MCP over stdio | mark3labs/mcp-go

---

## Non-Negotiables

- **SOLID / KISS / DRY.** No cleverness. No "just in case" abstractions.
- **No random dependencies.** Justify any new third-party package; keep it minimal.
- **No cross-layer leaks.** Public API (`boost/`) is stable; internal (`internal/`) can change.
- **No silent behavior changes.** Existing tool/resource contracts must remain stable.
- **Prefer deterministic, reproducible workflows.** CLI commands and CI must be idempotent.
- **Read-only by default.** All DB tools enforce read-only; no shell execution.

---

## How to Work (Gemini Operating Mode)

1. Read repo docs and existing conventions first.
2. Make the smallest change that solves the problem.
3. Update tests/docs when a behavior or contract changes.
4. If you're unsure, search locally (ripgrep) instead of guessing.
5. Keep commits scoped; avoid drive-by refactors.
6. Run `/test` before considering any change complete.

---

## Output Format

- When providing commands, list them as copy/paste shell blocks.
- When providing file changes, show **full file content** for new files and **precise diffs** for edits.

---

## Architecture

```
scg-boost/
├── boost/                    # PUBLIC API (stable surface)
│   ├── boost.go              # Server interface, New() factory
│   └── options.go            # Functional options (WithDB, WithConfig, etc.)
│
├── types/                    # Cross-layer type contracts
│   └── types.go              # Interfaces: Logger, DBConn, Config, Authorizer
│
├── internal/                 # INTERNAL (unstable, implementation details)
│   ├── mcp/                  # MCP protocol layer
│   ├── tools/                # Tool implementations (appinfo, dbquery, etc.)
│   ├── runtime/              # Utilities (redact, exec sandbox)
│   ├── bootstrap/            # .claude/ installer
│   ├── project/              # Project detection & summary
│   └── security/             # Scope management
│
├── resources/                # Embedded resources
│   ├── guidelines/           # SCG_CODING_GUIDELINES.md
│   ├── schemas/              # JSON schemas
│   └── bootstrap_templates/  # Per-service .claude/ templates
│
├── cmd/scg-boost/            # CLI entrypoint
└── examples/                 # Usage examples
```

### Layer Rules

| Layer        | Stability  | Can Import                 |
|--------------|------------|----------------------------|
| `boost/`     | **Stable** | types, internal            |
| `types/`     | **Stable** | stdlib only                |
| `internal/`  | Unstable   | types, stdlib, third-party |
| `resources/` | Stable     | (embedded assets)          |

---

## Available Tools (MCP)

| Tool                 | Description                           | Requires           |
|----------------------|---------------------------------------|--------------------|
| `appinfo.get`        | App name, version, Go runtime, uptime | Always             |
| `config.get`         | Get config value (redacted)           | `SafeConfig`       |
| `config.list`        | List config keys                      | `SafeConfig`       |
| `dbquery.run`        | Execute read-only SQL                 | `DBConn`           |
| `dbschema.list`      | List schemas/tables/columns           | `DBConn`           |
| `logs.lastError`     | Get last error log entry              | `LogStore`         |
| `health.status`      | Liveness/readiness checks             | `HealthProbe`      |
| `events.outbox.peek` | Peek event outbox                     | `OutboxReader`     |
| `trace.lookup`       | Lookup recent traces                  | `TraceReader`      |
| `service.topology`   | Service topology snapshot             | `TopologyProvider` |
| `routes.list`        | List HTTP/gRPC routes                 | `RouteProvider`    |
| `migrations.status`  | Migration status                      | `MigrationReader`  |
| `cache.stats`        | Cache statistics                      | `CacheInspector`   |
| `docs.search`        | Search documentation                  | `DocsSearcher`     |
| `metrics.summary`    | Metrics summary                       | `MetricsReader`    |
| `env.check`          | Environment validation                | `EnvChecker`       |

---

## CLI Commands

```bash
# Bootstrap .claude directory
scg-boost install [--root .] [--repo <name>] [--force]

# Generate MCP client config
scg-boost config --client claude|cursor|junie [--root .] [--name <server>]

# Scan project and generate summary
scg-boost scan [--root .]

# Run MCP server over stdio
scg-boost mcp [--root .] [--name <app>] [--version <v>]

# List available tools
scg-boost tools [--json]

# Show version
scg-boost version

# Validate project setup
scg-boost validate [--root .]
```

---

## Local Dev

```bash
# Quick checks
./scg ci

# Build CLI
go build -o scg-boost ./cmd/scg-boost

# Test install
./scg-boost install --root /tmp/test-repo --force
./scg-boost validate --root /tmp/test-repo
```

---

## Adding New Tools

1. Create interface in `types/types.go`
2. Add option in `boost/options.go`
3. Create tool package in `internal/tools/<name>/`
4. Register in `boost/boost.go` (conditional on option)
5. Add to tools list in CLI

---

## Token/Cost Control (Mandatory)

Use subagents and compaction intentionally:

| Task               | Use This Agent    |
|--------------------|-------------------|
| Search/triage      | `repo-radar`      |
| Tests/lint         | `go-test-runner`  |
| Contract safety    | `contract-guard`  |
| Code review        | `code-reviewer`   |
| Debug issues       | `debug-assistant` |
| Schema exploration | `schema-explorer` |

**Commands available:**

- `/test [scope]` — Run tests via subagent
- `/triage <problem>` — Find relevant files
- `/contract-check <summary>` — Check for breaking changes
- `/review <diff-summary>` — Code review
- `/debug <issue>` — Debug production issues
- `/docs <query>` — Search documentation
- `/migrate [action]` — Migration guidance

When a conversation grows, use `/compact` to summarize and reset context.

**Reason:** Large logs and long back-and-forth burn tokens. Subagents keep heavy work isolated.

---

## Security Rules

- **Read-only by default.** DB tools prevent writes.
- **Redaction.** Config values are redacted before exposure.
- **Fail closed.** Default authorizer denies all.
- **Never log secrets/tokens.**

---

## Go Best Practices (Enforced)

### Functional Options Pattern

```go
srv, err := boost.New(
boost.WithName("my-app"),
boost.WithVersion("1.0.0"),
boost.WithDB(myDB),
boost.WithConfig(myCfg),
)
```

### Error Handling

```go
// ✅ CORRECT: Wrap with context
if err != nil {
return fmt.Errorf("register tool %s: %w", name, err)
}

// ❌ WRONG: Naked returns
if err != nil {
return err
}
```

### Interface Design

```go
// ✅ CORRECT: Small, focused interfaces
type DBConn interface {
QueryJSON(ctx context.Context, query string, params map[string]any) ([]map[string]any, error)
}

// ❌ WRONG: God interfaces with 20+ methods
```

---

## Quick Reference

| What                 | Where                            |
|----------------------|----------------------------------|
| Public API           | `boost/`                         |
| Type contracts       | `types/`                         |
| Tool implementations | `internal/tools/`                |
| MCP protocol         | `internal/mcp/`                  |
| Bootstrap templates  | `resources/bootstrap_templates/` |
| CLI                  | `cmd/scg-boost/`                 |
| Examples             | `examples/`                      |
