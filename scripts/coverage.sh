#!/usr/bin/env bash
# Test Coverage Enforcement for scg-boost library
# Runs tests with coverage and validates against threshold

set -euo pipefail

# ===== Configuration =====
readonly COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-70.0}"
readonly COVERAGE_FILE="${COVERAGE_FILE:-coverage.out}"
readonly COVERAGE_HTML="${COVERAGE_HTML:-coverage.html}"

# Package list for coverage (scg-boost specific)
readonly COVERPKG="./boost/...,./internal/..."

# ===== Colors =====
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===== Helper functions =====
info() {
  echo -e "${GREEN}INFO:${NC} $*" >&2
}

warn() {
  echo -e "${YELLOW}WARN:${NC} $*" >&2
}

error() {
  echo -e "${RED}ERROR:${NC} $*" >&2
}

section() {
  echo -e "\n${BLUE}===${NC} $* ${BLUE}===${NC}" >&2
}

# ===== Main functions =====

show_help() {
  cat <<'EOF'
coverage.sh - Test coverage enforcement for scg-boost

Usage:
  ./scripts/coverage.sh [OPTIONS]

Options:
  -t, --threshold PERCENT    Set coverage threshold (default: 70.0)
  -s, --show-best-practices  Display coverage best practices
  -h, --help                 Show this help message

Environment Variables:
  COVERAGE_THRESHOLD         Coverage percentage threshold (default: 70.0)
  COVERAGE_FILE              Coverage output file (default: coverage.out)
  COVERAGE_HTML              HTML coverage file (default: coverage.html)

Examples:
  # Run coverage check with default threshold
  ./scripts/coverage.sh

  # Custom threshold
  ./scripts/coverage.sh --threshold 80

  # Show best practices
  ./scripts/coverage.sh --show-best-practices
EOF
}

show_best_practices() {
  section "Go Test Coverage Best Practices"
  cat <<'EOF'
  📚 Industry Standards:
     • Production Services: 80% minimum
     • Critical Services: 85-90%
     • Acceptable Baseline: 70%
     • Exemptions: cmd/, bootstrap/, generated code

  💡 Coverage Guidelines:
     • Core Logic: Aim for 90%+ (boost/ package)
     • Internal: Aim for 80%+ (internal/ packages)
     • Integration: Aim for 70%+ (transport, handlers)

  🎯 Current Threshold: 70% (Industry baseline)
EOF
  echo ""
}

run_tests_with_coverage() {
  section "Running tests with coverage"

  info "Executing: go test -count=1 -coverpkg=${COVERPKG} -coverprofile=${COVERAGE_FILE} ./..."
  info "Coverage packages: ${COVERPKG}"
  echo ""

  if ! go test -count=1 -race -coverpkg="${COVERPKG}" -coverprofile="${COVERAGE_FILE}" -covermode=atomic -timeout=5m ./...; then
    error "Tests failed"
    return 1
  fi

  if [[ ! -f "${COVERAGE_FILE}" ]]; then
    error "Coverage file ${COVERAGE_FILE} not generated"
    return 1
  fi

  info "Tests completed successfully ✓"
}

generate_reports() {
  section "Generating coverage reports"

  # HTML report
  info "Generating HTML: ${COVERAGE_HTML}"
  go tool cover -html="${COVERAGE_FILE}" -o "${COVERAGE_HTML}"

  # Detailed report
  info "Generating detailed function coverage"
  go tool cover -func="${COVERAGE_FILE}" > coverage-detailed.txt

  info "Reports generated ✓"
}

analyze_coverage() {
  local threshold="${COVERAGE_THRESHOLD}"

  section "Analyzing coverage"

  # Get total coverage
  local total_coverage
  total_coverage=$(go tool cover -func="${COVERAGE_FILE}" | grep "^total:" | awk '{print $3}' | sed 's/%//')

  if [[ -z "$total_coverage" ]]; then
    error "Could not extract coverage percentage"
    return 1
  fi

  # Show top uncovered functions
  echo ""
  info "📦 Coverage by Package (bottom 10 functions):"
  go tool cover -func="${COVERAGE_FILE}" | grep -v "^total:" | sort -t'	' -k3 -n | head -10

  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  printf "📊 Total Coverage: ${BLUE}%.1f%%${NC}\n" "$total_coverage"
  printf "🎯 Threshold: ${YELLOW}%.1f%%${NC}\n" "$threshold"

  # Calculate gap/surplus
  local gap
  gap=$(awk "BEGIN {printf \"%.1f\", $total_coverage - $threshold}")

  if awk "BEGIN {exit !($total_coverage >= $threshold)}"; then
    printf "✅ Surplus: ${GREEN}+%.1f%%${NC} above threshold\n" "$gap"
  else
    gap=$(awk "BEGIN {printf \"%.1f\", $threshold - $total_coverage}")
    printf "❌ Deficit: ${RED}-%.1f%%${NC} below threshold\n" "$gap"
  fi
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  # Compare with threshold
  if awk "BEGIN {exit !($total_coverage >= $threshold)}"; then
    info "✅ Coverage ($total_coverage%) meets threshold ($threshold%)"
    return 0
  else
    error "❌ Coverage ($total_coverage%) below threshold ($threshold%)"
    error "   Need ${gap}% more coverage"
    return 1
  fi
}

# ===== Main =====
main() {
  local custom_threshold="$COVERAGE_THRESHOLD"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -t|--threshold)
        custom_threshold="$2"
        shift 2
        ;;
      -s|--show-best-practices)
        show_best_practices
        exit 0
        ;;
      -h|--help)
        show_help
        exit 0
        ;;
      *)
        error "Unknown option: $1"
        show_help
        exit 1
        ;;
    esac
  done

  # Override threshold if custom
  COVERAGE_THRESHOLD="$custom_threshold"

  section "Test Coverage Analysis"
  info "Threshold: ${COVERAGE_THRESHOLD}%"
  echo ""

  # Run tests with coverage
  if ! run_tests_with_coverage; then
    exit 1
  fi

  # Generate reports
  generate_reports

  # Analyze and validate
  if analyze_coverage; then
    section "✅ Coverage Check PASSED"
    exit 0
  else
    section "❌ Coverage Check FAILED"
    echo ""
    error "Please add tests to improve coverage"
    echo ""
    show_best_practices
    exit 1
  fi
}

main "$@"
