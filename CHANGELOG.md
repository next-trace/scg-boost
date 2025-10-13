# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Skills system with metadata-driven discovery
  - `skills:list` command to browse available skills (table and JSON formats)
  - `skills:install` command to install specific skills
  - `skills:sync` command to update installed skills
  - Auto-detection of repo type with skill suggestions
  - 16 pre-packaged skills for SCG repositories
- Production-ready infrastructure
  - Root `.gitignore` for binaries, IDE files, and test artifacts
  - `SECURITY.md` with vulnerability reporting and security guarantees
  - `RELEASING.md` with complete release process documentation
  - `docs/DEVELOPMENT.md` with Go toolchain setup guide
- CI improvements
  - Deterministic Go 1.25.7 toolchain enforcement
  - Go version verification step in build job
  - `GOTOOLCHAIN=local` in all CI jobs to prevent automatic upgrades
- MCP integration tests
  - Basic server lifecycle validation
  - Tool registration verification

### Changed
- Enhanced `install` command with auto-detection and skill suggestions
- Refactored `bootstrap.Install()` to support `InstallSkill()` function
- Updated `resources.go` to embed `skill.json` metadata files
- Fixed `.gitignore` to allow `resources/bootstrap_templates/scg-boost/`

### Internal
- Created `internal/skills` package with:
  - `Metadata` type for skill properties
  - `Registry` type for skill discovery
  - `DetectRepoType()` function for auto-detection
- Skills stored in `.claude/skill.json` for tracking

## [0.1.0] - 2024-02-04

### Added
- Initial MCP server implementation over stdio
- 16 MCP tools for observability:
  - `appinfo.get` - Application info
  - `config.get` / `config.list` - Configuration access
  - `dbquery.run` / `dbschema.list` - Database queries
  - `logs.lastError` - Log access
  - `health.status` - Health checks
  - `events.outbox.peek` - Event outbox
  - `trace.lookup` - Trace lookup
  - `service.topology` - Service topology
  - `routes.list` - Route listing
  - `migrations.status` - Migration status
  - `cache.stats` - Cache statistics
  - `docs.search` - Documentation search
  - `metrics.summary` - Metrics summary
  - `env.check` - Environment validation
- CLI commands:
  - `install` - Bootstrap `.claude/` directory
  - `config` - Generate MCP client config (Claude/Cursor/Junie)
  - `scan` - Project detection and summary
  - `mcp` - Run MCP server over stdio
  - `tools` - List available tools
  - `version` - Show version
  - `validate` - Validate project setup
- Bootstrap templates for 17 SCG repositories
- Functional options pattern for server configuration
- Read-only database enforcement
- Configuration value redaction
- Embedded resources (guidelines, schemas, templates)
- CI/CD with GitHub Actions:
  - Build, test, quality, security jobs
  - 70% test coverage enforcement
  - golangci-lint, govulncheck, gosec
- Comprehensive documentation

### Security
- Read-only database operations only
- Automatic redaction of sensitive config values
- No shell command execution
- MCP request validation

[Unreleased]: https://github.com/next-trace/scg-boost/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/next-trace/scg-boost/releases/tag/v0.1.0
