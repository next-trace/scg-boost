#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
PLAN_DIR="$ROOT_DIR/.plan"
TS="$(date +%Y%m%d-%H%M%S)"
REPORT="$PLAN_DIR/pre-push-review-${TS}.md"

mkdir -p "$PLAN_DIR"

branch="$(git rev-parse --abbrev-ref HEAD)"
base_ref="origin/main"
if ! git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  base_ref="main"
fi

has_issue=0

write_check() {
  local name="$1"
  local status="$2"
  local details="$3"
  printf -- "- [%s] %s\n" "$status" "$name" >>"$REPORT"
  if [[ -n "$details" ]]; then
    printf -- "  - %s\n" "$details" >>"$REPORT"
  fi
}

run_check() {
  local name="$1"
  shift
  if "$@" >/tmp/prepush-check.out 2>&1; then
    write_check "$name" "PASS" ""
  else
    has_issue=1
    write_check "$name" "FAIL" "$(head -n 5 /tmp/prepush-check.out | tr '\n' ' ' | sed 's/  */ /g')"
  fi
}

{
  echo "# Pre-Push Review Report"
  echo
  echo "- Timestamp: $TS"
  echo "- Branch: $branch"
  echo "- Base: $base_ref"
  echo
  echo "## Automated Checks"
} >"$REPORT"

run_check "Go fmt" bash -lc "cd '$ROOT_DIR' && test -z \"\$(gofmt -l .)\""
run_check "Go vet" bash -lc "cd '$ROOT_DIR' && go vet ./..."
run_check "Go test" bash -lc "cd '$ROOT_DIR' && go test ./..."

if command -v golangci-lint >/dev/null 2>&1; then
  run_check "golangci-lint" bash -lc "cd '$ROOT_DIR' && golangci-lint run ./..."
else
  write_check "golangci-lint" "SKIP" "golangci-lint not installed"
fi

if rg -n "TODO|FIXME" "$ROOT_DIR" --glob '!**/.git/**' --glob '!**/vendor/**' >/tmp/prepush-todos.out 2>&1; then
  has_issue=1
  write_check "No unresolved TODO/FIXME" "FAIL" "Found TODO/FIXME entries"
else
  write_check "No unresolved TODO/FIXME" "PASS" ""
fi

if [[ "$base_ref" != "main" ]]; then
  if git diff --name-only "$base_ref...HEAD" >/tmp/prepush-diff.out 2>&1; then
    changed_count="$(wc -l < /tmp/prepush-diff.out | tr -d ' ')"
    write_check "Diff against base" "PASS" "Changed files: $changed_count"
  else
    has_issue=1
    write_check "Diff against base" "FAIL" "Unable to compute diff against $base_ref"
  fi
fi

echo >>"$REPORT"
echo "## Recommendation" >>"$REPORT"
if [[ "$has_issue" -eq 0 ]]; then
  echo "No blocking issues detected. Push is safe." >>"$REPORT"
  echo "Pre-push checks passed. Report: $REPORT"
  exit 0
fi

echo "Issues detected. Review report before pushing: $REPORT" >>"$REPORT"
echo "Pre-push issues detected. Report: $REPORT"

echo -n "Continue push anyway? [y/N]: "
read -r answer
case "$answer" in
  y|Y|yes|YES)
    echo "Push allowed by user confirmation."
    exit 0
    ;;
  *)
    echo "Push blocked. Resolve issues first."
    exit 1
    ;;
esac
