# scg-boost - Overview

## Purpose
- MCP server and context bootstrapper for SupplyChainGuard repos.
- Generates and maintains `.claude/` bootstrap context.
- Serves project summary/tree/CLAUDE.md resources over MCP (stdio).

## Stack
- Go 1.22+
- MCP over stdio (mark3labs/mcp-go)

## Architecture (high level)
- `boost/` public API (stable surface)
- `types/` cross-layer interfaces
- `internal/` implementation (tools, runtime, bootstrap, project)
- `resources/` embedded assets and bootstrap templates
- `cmd/scg-boost/` CLI entrypoint

## Non-negotiables (from CLAUDE.md)
- Keep `boost/` stable; `internal/` can change.
- Avoid new dependencies unless clearly justified.
- No silent contract changes to tools/resources.
- Prefer deterministic, reproducible workflows.
