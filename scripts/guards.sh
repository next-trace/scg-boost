#!/usr/bin/env bash
# Guards - Repository validation for scg-boost library
# Simplified from identity-service (no infrastructure/production checks)

set -euo pipefail

# ===== Configuration =====
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "$REPO_ROOT"

# ===== Colors =====
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# ===== Tracking =====
FAILED=0
WARNINGS=0

# ===== Helper functions =====
fail() {
  echo -e "${RED}✗${NC} $*" >&2
  FAILED=1
}

warn() {
  echo -e "${YELLOW}⚠${NC} $*" >&2
  ((WARNINGS++))
}

pass() {
  echo -e "${GREEN}✓${NC} $*" >&2
}

info() {
  echo -e "${BLUE}ℹ${NC} $*" >&2
}

section() {
  echo "" >&2
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
  echo -e "${BLUE}$*${NC}" >&2
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
}

# ===== Checks =====

check_build_artifacts() {
  section "Checking Build Artifacts"

  # Check for compiled test binaries
  if find . -type f -name '*.test' ! -name '.env.test' 2>/dev/null | grep -q '.'; then
    find . -type f -name '*.test' ! -name '.env.test'
    fail "Compiled test binaries (*.test) must not be committed"
  else
    pass "No test binaries found"
  fi

  # Check for tracked binary artifacts
  if git ls-files --stage 2>/dev/null | awk '{print $4}' | grep -E '\.(exe|dll|so|dylib|bin)$' | grep -q '.'; then
    git ls-files --stage | awk '{print $4}' | grep -E '\.(exe|dll|so|dylib|bin)$'
    fail "Binary artifacts detected in repository"
  else
    pass "No binary artifacts tracked"
  fi
}

check_env_files() {
  section "Checking Environment Files"

  # Check for tracked .env files
  if git ls-files 2>/dev/null | grep -qE '^\.env$|^\.env\.[^d]'; then
    fail ".env files are tracked in git - These may contain secrets!"
  else
    pass "No tracked .env files"
  fi
}

check_go_structure() {
  section "Checking Go Structure"

  # Check go.mod exists
  if [[ ! -f "go.mod" ]]; then
    fail "go.mod not found"
  else
    pass "go.mod exists"
  fi

  # Check for vendor directory (should not be committed for libraries)
  if [[ -d "vendor" ]] && git ls-files vendor/ 2>/dev/null | grep -q '.'; then
    warn "vendor/ directory is tracked - consider removing for library"
  else
    pass "No tracked vendor directory"
  fi

  # Check package structure
  if [[ -d "boost" ]]; then
    pass "boost/ package exists"
  else
    warn "boost/ package not found"
  fi

  if [[ -d "internal" ]]; then
    pass "internal/ package exists"
  else
    warn "internal/ package not found"
  fi
}

check_forbidden_patterns() {
  section "Checking Code Quality Patterns"

  local issues_found=0

  # Check for TODO/FIXME/STUB patterns (excluding this script and test files)
  if command -v rg &>/dev/null; then
    if rg -n 'TODO:|FIXME:|STUB:' --type go --glob '!*_test.go' --glob '!scripts/*' . 2>/dev/null | head -10 | grep -q .; then
      warn "Found TODO/FIXME/STUB patterns in code:"
      rg -n 'TODO:|FIXME:|STUB:' --type go --glob '!*_test.go' --glob '!scripts/*' . 2>/dev/null | head -10
      issues_found=1
    fi
  else
    if grep -rn 'TODO:\|FIXME:\|STUB:' --include='*.go' --exclude='*_test.go' . 2>/dev/null | head -10 | grep -q .; then
      warn "Found TODO/FIXME/STUB patterns in code"
      issues_found=1
    fi
  fi

  if [[ $issues_found -eq 0 ]]; then
    pass "No TODO/FIXME/STUB patterns found"
  fi

  # Check for panic in non-test code
  if command -v rg &>/dev/null; then
    if rg -n 'panic\(' --type go --glob '!*_test.go' . 2>/dev/null | grep -v 'recover' | head -5 | grep -q .; then
      warn "Found panic() calls in non-test code (consider returning errors):"
      rg -n 'panic\(' --type go --glob '!*_test.go' . 2>/dev/null | grep -v 'recover' | head -5
    else
      pass "No bare panic() calls found"
    fi
  fi
}

check_test_files() {
  section "Checking Test Coverage"

  # Count test files
  local test_count
  test_count=$(find . -name '*_test.go' -type f 2>/dev/null | wc -l)

  if [[ $test_count -eq 0 ]]; then
    fail "No test files found"
  else
    pass "Found $test_count test files"
  fi

  # Check key packages have tests
  if [[ -d "boost" ]]; then
    if find boost -name '*_test.go' 2>/dev/null | grep -q .; then
      pass "boost/ has tests"
    else
      warn "boost/ has no tests"
    fi
  fi
}

# ===== Help =====

show_help() {
  cat <<'EOF'
guards.sh - Repository validation for scg-boost library

Usage:
  ./scripts/guards.sh [OPTIONS]

Options:
  --skip-quality    Skip code quality pattern checks
  -h, --help        Show this help message

Checks:
  - Build artifacts (no binaries committed)
  - Environment files (no .env tracked)
  - Go structure (go.mod, packages)
  - Code quality (TODO/FIXME/STUB patterns)
  - Test files (coverage exists)

Examples:
  # Run all checks
  ./scripts/guards.sh

  # Skip quality checks (faster)
  ./scripts/guards.sh --skip-quality
EOF
}

# ===== Main =====

main() {
  local skip_quality=0

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --skip-quality)
        skip_quality=1
        shift
        ;;
      -h|--help)
        show_help
        exit 0
        ;;
      *)
        echo "Unknown option: $1" >&2
        show_help
        exit 1
        ;;
    esac
  done

  # Header
  section "🔒 scg-boost Guards Validation"

  # Run checks
  check_build_artifacts
  check_env_files
  check_go_structure
  check_test_files

  [[ $skip_quality -eq 0 ]] && check_forbidden_patterns

  # Summary
  section "Summary"

  if [[ $FAILED -eq 0 ]]; then
    if [[ $WARNINGS -gt 0 ]]; then
      echo -e "${YELLOW}⚠ Validation completed with $WARNINGS warning(s)${NC}" >&2
      echo -e "${GREEN}✓ All critical checks passed${NC}" >&2
    else
      echo -e "${GREEN}✓ All guards passed successfully!${NC}" >&2
    fi
    exit 0
  else
    echo -e "${RED}✗ Validation failed - Fix violations above${NC}" >&2
    exit 1
  fi
}

main "$@"
