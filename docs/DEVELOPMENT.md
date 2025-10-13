# Development Guide

## Prerequisites

### Go Toolchain

SCG-Boost requires **Go 1.25.7 exactly** for deterministic builds.

#### Why Strict Versioning?

- **Deterministic builds** - Same code produces identical binaries across environments
- **Explicit upgrades** - Toolchain changes are intentional, not accidental
- **CI/Local parity** - Developers see the same behavior as CI

#### Installation

```sh
# Download and install Go 1.25.7
go install golang.org/dl/go1.25.7@latest
go1.25.7 download

# Verify installation
go1.25.7 version
# Output: go version go1.25.7 linux/amd64
```

#### Using the Correct Toolchain

**Option 1: Set GOTOOLCHAIN environment variable (Recommended)**

```sh
export GOTOOLCHAIN=go1.25.7
go version  # Should show go1.25.7
```

Add to your shell profile (`~/.bashrc`, `~/.zshrc`):
```sh
export GOTOOLCHAIN=go1.25.7
```

**Option 2: Use the specific binary**

```sh
go1.25.7 build ./...
go1.25.7 test ./...
```

**Option 3: Install as default**

If you only work on Go 1.25.7 projects:
```sh
# Backup existing installation
sudo mv /usr/local/go /usr/local/go-backup

# Link to 1.25.7
sudo ln -s $(go1.25.7 env GOROOT) /usr/local/go
```

#### Verification

```sh
# Check active Go version
go version

# Verify it matches requirement
go version | grep -q "go1.25.7" && echo "✓ Correct version" || echo "✗ Wrong version"

# Check GOTOOLCHAIN setting
echo $GOTOOLCHAIN
# Should output: go1.25.7 or local
```

## Local Development

### Quick Start

```sh
# Clone repository
git clone https://github.com/next-trace/scg-boost.git
cd scg-boost

# Set toolchain (if not in profile)
export GOTOOLCHAIN=go1.25.7

# Verify setup
go version
go mod verify

# Run CI checks locally
./scg ci
```

### Common Commands

```sh
# Build CLI
go build -o scg-boost ./cmd/scg-boost

# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.txt

# Run linter
golangci-lint run

# Check vulnerabilities
govulncheck ./...

# Security scan
gosec -exclude-generated ./...
```

### CI Wrapper Script

The `./scg ci` command runs the same checks as GitHub Actions:

```sh
# Run all CI checks
./scg ci

# Format code
./scg fmt

# Tidy dependencies
./scg tidy
```

## Project Structure

```
scg-boost/
├── boost/                    # Public API (stable)
├── types/                    # Interface contracts
├── internal/                 # Implementation (unstable)
│   ├── mcp/                  # MCP protocol
│   ├── tools/                # Tool implementations
│   ├── runtime/              # Utilities
│   ├── bootstrap/            # .claude/ installer
│   ├── project/              # Project detection
│   └── security/             # Scope management
├── resources/                # Embedded assets
├── cmd/scg-boost/            # CLI entrypoint
├── examples/                 # Usage examples
├── scripts/                  # Build/CI scripts
└── docs/                     # Documentation
```

## Adding New Tools

1. **Define interface** in `types/types.go`
2. **Add option** in `boost/options.go`
3. **Implement tool** in `internal/tools/<name>/`
4. **Register in server** in `boost/boost.go`
5. **Add tests** with 70%+ coverage
6. **Update documentation**

Example:
```go
// 1. types/types.go
type MyFeature interface {
    DoSomething(ctx context.Context) (string, error)
}

// 2. boost/options.go
func WithMyFeature(feat MyFeature) Option {
    return func(s *Server) error {
        s.myFeature = feat
        return nil
    }
}

// 3. internal/tools/myfeature/tool.go
// ... implementation ...

// 4. boost/boost.go
if s.myFeature != nil {
    tool := myfeature.New(s.myFeature)
    mcp.RegisterTool(tool)
}
```

## Testing Guidelines

### Unit Tests

- Place tests in `*_test.go` files
- Use table-driven tests for multiple cases
- Mock external dependencies
- Aim for 70%+ coverage

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Use build tag: `//go:build integration`
- Test full server lifecycle
- Validate MCP protocol compliance

```go
//go:build integration

func TestMCPServer(t *testing.T) {
    // ... integration test ...
}
```

Run with: `go test -tags=integration ./...`

## Troubleshooting

### Wrong Go Version

**Symptom:** CI fails with "Go version mismatch"

**Solution:**
```sh
export GOTOOLCHAIN=go1.25.7
go version  # Verify it shows 1.25.7
```

### Tests Failing Locally but Passing in CI

**Symptom:** Different behavior between local and CI

**Possible causes:**
- Wrong Go version locally
- Environment variable differences
- Cached test results

**Solution:**
```sh
# Clean cache and rerun
go clean -testcache
export GOTOOLCHAIN=go1.25.7
go test ./...
```

### Linter Errors

**Symptom:** `golangci-lint` reports issues

**Solution:**
```sh
# Auto-fix what's possible
golangci-lint run --fix

# View specific issue details
golangci-lint run --verbose
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run `./scg ci` to verify
5. Commit with descriptive messages
6. Submit pull request

## Resources

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [MCP Specification](https://spec.modelcontextprotocol.io)
- [SCG Coding Guidelines](../resources/guidelines/SCG_CODING_GUIDELINES.md)
