# CLAUDE.md — scg-logger

This file is the **single source of truth** for how Claude must work in this repository.

## What This Repo Is

## Repo Scope

`scg-logger` defines logging interfaces/adapters used across services.

## Hard Boundaries

- Keep the interface minimal.
- No business logic here.
- Provide structured logging primitives; callers own fields.

## Development Commands

```bash
go test ./...
golangci-lint run ./...
```

## Non-Negotiables

- **SOLID / KISS / DRY.** No cleverness. No “just in case” abstractions.
- **No random dependencies.** If you introduce a new third-party package, you must justify it and keep it minimal.
- **No cross-layer leaks.** Domain rules do not depend on transport, DB, or external clients.
- **No silent behavior changes.** Existing API behaviors and CI expectations must remain stable unless explicitly stated.
- **Prefer deterministic, reproducible workflows.** Scripts and CI must be idempotent.

## How to Work (Claude Operating Mode)

1. Read repo docs and existing conventions first.
2. Make the smallest change that solves the problem.
3. Update tests/docs when a behavior or contract changes.
4. If you’re unsure, search locally (ripgrep) instead of guessing.
5. Keep commits scoped; avoid drive-by refactors.

## Output Format

- When providing commands, list them as copy/paste shell blocks.
- When providing file changes, show **full file content** for new files and **precise diffs** for edits.


## Architecture Context (Project-wide)

SupplyChainGuard enforces **Control Plane vs Data Plane** separation:

- Services must prefer shared SCG libraries over random third-party deps.
- Contracts are sacred: shared DTOs/schemas must be backward compatible.
- Redis (if used) is an enforcement cache, not a source of truth.

Project references:
- SAD/ADR documents in the project describe the rules; do not contradict them.


## Token/Cost Control (Mandatory)

Use subagents and compaction intentionally:
- For **search/triage**: use `repo-radar` subagent (keeps logs out of main context).
- For **tests/lint**: use `go-test-runner` subagent.
- For **contract safety**: use `contract-guard` subagent.
- When a conversation grows, use `/compact` to summarize and reset context.

Reason: large logs and long back-and-forth are what burn tokens. Subagents keep heavy work isolated, and `/compact` preserves decisions without carrying every message.

