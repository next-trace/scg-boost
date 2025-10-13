# scg-boost - API Surface

## CLI (cmd/scg-boost)
- `scg-boost install` (bootstrap `.claude/`)
- `scg-boost config --client <claude|cursor|junie>`
- `scg-boost scan`
- `scg-boost mcp`
- `scg-boost tools`
- `scg-boost version`
- `scg-boost validate`

## Go public API (boost/)
- `boost.New(...)` with functional options (e.g., name, version, DB, config)
- `boost/` and `types/` are the stable public surface

## MCP tools (registered conditionally)
- `appinfo.get`, `config.get`, `config.list`
- `dbquery.run`, `dbschema.list`
- `logs.lastError`, `health.status`
- `events.outbox.peek`, `trace.lookup`, `service.topology`
- `routes.list`, `migrations.status`, `cache.stats`
- `docs.search`, `metrics.summary`, `env.check`
