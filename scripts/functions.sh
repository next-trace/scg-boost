#!/usr/bin/env bash
# SCG Functions - Core automation for scg-boost library
# Adapted from identity-service for library repos (no docker/db/terraform)

set -euo pipefail

# ===== Colors =====
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===== Helper functions =====
info() {
  echo -e "${GREEN}INFO:${NC} $*" >&2
}

warn() {
  echo -e "${YELLOW}WARN:${NC} $*" >&2
}

fail() {
  echo -e "${RED}ERROR:${NC} $*" >&2
  exit 1
}

section() {
  echo -e "\n${BLUE}===${NC} $* ${BLUE}===${NC}" >&2
}

# ===== Version checks =====

fn_require_go() {
  section "Checking Go version"

  if ! command -v go &>/dev/null; then
    fail "Go is not installed"
  fi

  local current
  current=$(go version | awk '{print $3}' | sed 's/go//')

  # Compare major.minor.patch
  local required_major required_minor required_patch
  local current_major current_minor current_patch
  required_major=$(echo "$GO_REQUIRED" | cut -d. -f1)
  required_minor=$(echo "$GO_REQUIRED" | cut -d. -f2)
  required_patch=$(echo "$GO_REQUIRED" | cut -d. -f3)
  current_major=$(echo "$current" | cut -d. -f1)
  current_minor=$(echo "$current" | cut -d. -f2)
  current_patch=$(echo "$current" | cut -d. -f3)

  required_patch=${required_patch:-0}
  current_patch=${current_patch:-0}

  if [[ "$current_major" -lt "$required_major" ]] || \
     [[ "$current_major" -eq "$required_major" && "$current_minor" -lt "$required_minor" ]] || \
     [[ "$current_major" -eq "$required_major" && "$current_minor" -eq "$required_minor" && "$current_patch" -lt "$required_patch" ]]; then
    fail "Go $GO_REQUIRED required, found $current"
  fi

  info "Go version: $current ✓"
}

fn_require_tools() {
  section "Checking required tools"

  # golangci-lint
  if ! command -v golangci-lint &>/dev/null; then
    warn "golangci-lint not found, installing..."
    go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"
  fi

  local lint_version
  lint_version=$(golangci-lint --version 2>/dev/null | head -1 || echo "unknown")
  info "golangci-lint: $lint_version ✓"

  # goimports (optional but recommended)
  if ! command -v goimports &>/dev/null; then
    warn "goimports not found, installing..."
    go install golang.org/x/tools/cmd/goimports@latest
  fi
  info "goimports: installed ✓"

  # govulncheck
  if ! command -v govulncheck &>/dev/null; then
    warn "govulncheck not found, installing..."
    go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
  fi
  info "govulncheck: installed ✓"

  # gosec
  if ! command -v gosec &>/dev/null; then
    warn "gosec not found, installing..."
    go install "github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}"
  fi
  info "gosec: installed ✓"
}

# ===== Go commands =====

fn_deps() {
  section "Managing dependencies"

  info "Running go mod tidy..."
  go mod tidy

  info "Running go mod download..."
  go mod download

  info "Dependencies updated ✓"
}

fn_fmt() {
  section "Formatting code"

  if command -v goimports &>/dev/null; then
    info "Running goimports..."
    goimports -w .
  else
    info "Running gofmt..."
    gofmt -w .
  fi

  info "Code formatted ✓"
}

fn_lint() {
  section "Running linter"

  if ! command -v golangci-lint &>/dev/null; then
    fn_require_tools
  fi

  info "Running golangci-lint..."
  golangci-lint run ./...

  info "Lint passed ✓"
}

fn_vet() {
  section "Running go vet"

  go vet ./...

  info "Vet passed ✓"
}

fn_build() {
  section "Building packages"

  info "Running go build..."
  go build ./...

  info "Build successful ✓"
}

fn_vulncheck() {
  section "Running govulncheck"

  if ! command -v govulncheck &>/dev/null; then
    fn_require_tools
  fi

  govulncheck ./...

  info "govulncheck passed ✓"
}

fn_gosec() {
  section "Running gosec"

  if ! command -v gosec &>/dev/null; then
    fn_require_tools
  fi

  gosec -exclude-generated ./...

  info "gosec passed ✓"
}

# ===== Testing =====

fn_test() {
  section "Running all tests"

  info "Running tests with race detector..."
  go test -race -v ./...

  info "All tests passed ✓"
}

fn_test_unit() {
  section "Running unit tests"

  info "Running unit tests (short mode)..."
  go test -race -short ./...

  info "Unit tests passed ✓"
}

fn_coverage() {
  section "Running coverage"

  if [[ -f "scripts/coverage.sh" ]]; then
    bash scripts/coverage.sh "$@"
  else
    # Fallback inline coverage
    local coverage_file="coverage.out"
    local coverpkg="./boost/...,./internal/..."

    info "Running tests with coverage..."
    go test -count=1 -coverpkg="${coverpkg}" -coverprofile="${coverage_file}" -covermode=atomic ./...

    local total
    total=$(go tool cover -func="${coverage_file}" | grep "^total:" | awk '{print $3}' | sed 's/%//')

    info "Total coverage: ${total}%"
    info "Threshold: ${COVERAGE_THRESHOLD}%"

    if awk "BEGIN {exit !($total >= $COVERAGE_THRESHOLD)}"; then
      info "Coverage passed ✓"
    else
      fail "Coverage ${total}% below threshold ${COVERAGE_THRESHOLD}%"
    fi
  fi
}

# ===== CI/CD =====

fn_ci() {
  section "Running CI Pipeline"

  local start_time
  start_time=$(date +%s)

  info "Step 1/8: Dependencies"
  fn_deps

  info "Step 2/8: Format"
  fn_fmt

  info "Step 3/8: Vet"
  fn_vet

  info "Step 4/8: Lint"
  fn_lint

  info "Step 5/8: Vulncheck"
  fn_vulncheck

  info "Step 6/8: Build"
  fn_build

  info "Step 7/8: Test"
  fn_test

  info "Step 8/8: Security"
  fn_gosec

  local end_time duration
  end_time=$(date +%s)
  duration=$((end_time - start_time))

  section "CI Pipeline Complete"
  info "Duration: ${duration}s"
  info "All checks passed ✓"
}

fn_doctor() {
  section "Doctor - Checking prerequisites"

  fn_require_go
  fn_require_tools

  # Check go.mod exists
  if [[ ! -f "go.mod" ]]; then
    fail "go.mod not found - not in a Go module"
  fi
  info "go.mod found ✓"

  # Check for .golangci.yml
  if [[ ! -f ".golangci.yml" ]]; then
    warn ".golangci.yml not found"
  else
    info ".golangci.yml found ✓"
  fi

  section "Doctor Complete"
  info "All prerequisites satisfied ✓"
}

fn_clean() {
  section "Cleaning build artifacts"

  # Remove coverage files
  rm -f coverage.out coverage.html coverage-detailed.txt

  # Remove test cache
  go clean -testcache

  # Remove build cache (optional, can be slow)
  # go clean -cache

  info "Clean complete ✓"
}

fn_guards() {
  section "Running guards"

  if [[ -f "scripts/guards.sh" ]]; then
    bash scripts/guards.sh "$@"
  else
    warn "scripts/guards.sh not found"
  fi
}
