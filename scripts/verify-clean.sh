#!/usr/bin/env bash
set -euo pipefail

root="$(pwd)"

artifacts=$(find "$root" -maxdepth 2 \
  \( -path "$root/.git" -o -path "$root/.git/*" \) -prune -o \
  \( -path "$root/scripts/coverage.sh" \) -prune -o \
  \( -name "*.out" -o -name "coverage.*" -o -name "dist" \) -print)

if [[ -n "$artifacts" ]]; then
  echo "Artifact(s) found:" >&2
  echo "$artifacts" >&2
  exit 1
fi

exit 0
