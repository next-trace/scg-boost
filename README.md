# SCG-Boost

SCG-Boost is the SupplyChainGuard equivalent of **Laravel Boost**: a **per-repo** bootstrapper + MCP server that makes Claude / Codex / Gemini (and other MCP-capable clients) behave like a senior SCG engineer instead of hallucinating.

It does two things:

1) **Bootstrap context** in your repo: `.claude/`, `.codex/`, `.gemini/` plus `.mcp.json`.
2) **Serve MCP over stdio** so clients can query: project summary, repo tree, AI guidance files, and SCG guidelines.

## Requirements

- **Go 1.26.1+** (required for stdlib security fixes)
- golangci-lint 2.11.3+
- govulncheck 1.1.4+
- gosec 2.22.2+

## Quick Start

### 1) Install the CLI

```sh
go install github.com/next-trace/scg-boost/cmd/scg-boost@latest
```

### 2) Install or update repo context

From the repo root:

```sh
scg-boost install
# Later, refresh to latest bundled templates + mcp wiring
scg-boost update
```

`install`/`update` writes `.mcp.json` in the repo root for project-local MCP detection.

### 3) Optional: print manual client config

Run:

```sh
scg-boost config --client claude
```

Use this when you need manual wiring. In normal flow, `.mcp.json` is already generated.

## Verify

In your AI assistant, call:

- `appinfo.get`
- `scg://project/summary` (resource)
- `scg://project/tree` (resource)
- `scg://project/claude` (resource)
- `scg://project/codex` (resource)
- `scg://project/gemini` (resource)

If you can't see resources, you wired the MCP server wrong. Fix your client config.

### Skills System (Preview)

SCG-Boost includes a skills system for managing Claude context:

```sh
# List available skills
scg-boost skills:list

# Install specific skill
scg-boost skills:install --skill gateway-service

# Sync installed skills with latest bundled versions
scg-boost skills:sync

# Create local overrides that survive sync
scg-boost skills:override --skill gateway-service --path .claude/commands/custom.md
```

Skills are repo-specific context bundles (CLAUDE.md, agents, commands)
optimized for different project types.

See `docs/OVERRIDES.md` for override rules and section overrides.

### Embed in Your Service

To integrate SCG-Boost directly into your application:
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/next-trace/scg-boost/boost"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Use boost.New with options relevant to your app
	// (e.g., providing a DB connection, a config provider, etc.)
	srv, err := boost.New(
		boost.WithName("my-app"),
		boost.WithVersion("1.2.3"),
		// boost.WithDB(myDbConn),
		// boost.WithConfig(myConfig),
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Start runs the MCP server over stdio.
	if _, err := srv.Start(ctx); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	log.Println("MCP server running. Waiting for shutdown signal.")
	<-ctx.Done()
	log.Println("Shutting down.")
}
```

## Development

### Prerequisites

Ensure you have the required Go version and tools:

```sh
# Check Go version
go version  # Should be 1.26.1 or higher

# Install development tools (automated)
./scg doctor
```

### Running Tests

```sh
# Full CI pipeline (recommended)
./scg ci

# Unit tests only
go test ./...

# With race detector
go test -race ./...

# Integration tests (MCP stdio server)
go test -tags=integration ./...
```

### Building from Source

```sh
go build -o scg-boost ./cmd/scg-boost
./scg-boost version
```

### CI Commands

```sh
./scg doctor    # Check prerequisites
./scg ci        # Run full CI pipeline
./scg lint      # Run linter only
./scg test      # Run tests only
./scg guards    # Run custom guards
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for complete development guide.
