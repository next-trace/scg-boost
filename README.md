# SCG-Boost

SCG-Boost is the SupplyChainGuard equivalent of **Laravel Boost**: a **per-repo** bootstrapper + MCP server that makes Claude / Junie / Cursor behave like a senior SCG engineer instead of hallucinating.

It does two things:

1) **Bootstrap context** in your repo: `.claude/` with `CLAUDE.md`, subagents, and commands.
2) **Serve MCP over stdio** so clients can query: project summary, repo tree, CLAUDE.md, and SCG guidelines.

## Quick Start

### 1) Install the CLI

```sh
go install github.com/next-trace/scg-boost/cmd/scg-boost@latest
```

### 2) Generate repo context

From the repo root:

```sh
scg-boost install
```

### 3) Connect your AI client to the MCP server

Run:

```sh
scg-boost config --client claude
```

Copy the JSON it prints into your client config. Then restart the client.

## Verify

In your AI assistant, call:

- `appinfo.get`
- `scg://project/summary` (resource)
- `scg://project/tree` (resource)
- `scg://project/claude` (resource)

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

### Running Tests

```sh
# Unit tests
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

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for complete development guide.
