# CLAUDE.md — identity-service

This file is the **single source of truth** for how Claude must work in this repository.

If you violate these rules, you create chaos. Don't.

## What This Repo Is

AuthN/AuthZ service. Issues/validates tokens, user/tenant auth flows. Read-only enforcement via Redis.

## Non‑Negotiables

- **SOLID / KISS / DRY.** No cleverness. No “just in case” abstractions.
- **No random dependencies.** If you introduce a new third‑party package, you must justify it and keep it minimal.
- **No cross‑layer leaks.** Domain rules do not depend on transport, DB, or external clients.
- **No silent behavior changes.** Existing API behaviors and CI expectations must remain stable unless explicitly stated.
- **Prefer deterministic, reproducible workflows.** Scripts and CI must be idempotent.

## How to Work (Claude Operating Mode)

1. Read repo docs and existing conventions first.
2. Make the smallest change that solves the problem.
3. Update tests/docs when a behavior or contract changes.
4. If you’re unsure, search locally (ripgrep) instead of guessing.
5. Keep commits scoped; avoid drive‑by refactors.

## Output Format

- When providing commands, list them as copy/paste shell blocks.
- When providing file changes, show **full file content** for new files and **precise diffs** for edits.


## Architecture Context (Project‑wide)

SupplyChainGuard enforces **Control Plane vs Data Plane** separation:

- **Tenant‑Service = Control Plane owner (policies source of truth; sole Redis writer).**
- **Gateway/Identity = Data Plane only (read‑only enforcement).**
- **Redis is an Enforcement Cache, not a source of truth.**
- Redis admin commands are forbidden; use allow‑listed commands only.

Project references:
- SAD v2.5 (Control/Data plane + Redis enforcement) and ADRs 0001/0002/0003 define the rules.


## Repo-Specific Rules

## Repo Scope

Identity is an internal service for authentication/authorization.
It is **Data Plane only**.

Responsibilities:
- Token issuance/validation
- User and client authentication flows
- Authorization decisions using **read‑only** enforcement cache

Non‑responsibilities:
- Owning tenant policies
- Writing to Redis enforcement cache

## Hard Boundaries

- Never write to Redis enforcement cache.
- Use `scg-validator` for request validation; do not re‑invent validators.
- Keep domain logic independent of transport (HTTP/gRPC).

## Local Dev

```bash
go test ./...
go test -race ./...
golangci-lint run ./...
```

## Security Rules

- Fail closed for auth decisions.
- Never log secrets/tokens.
- Minimal token claims; no over-sharing.

## Token/Cost Control (Mandatory)

Use subagents and compaction intentionally:
- For **search/triage**: use `repo-radar` subagent (keeps logs out of main context).
- For **tests/lint**: use `go-test-runner` subagent.
- For **contract safety**: use `contract-guard` subagent.
- For **terraform failures** (infra only): use `terraform-doctor` subagent.
- When a conversation grows, use `/compact` to summarize and reset context.

Reason: large logs and long back-and-forth are what burn tokens. Subagents keep heavy work isolated, and `/compact` preserves decisions without carrying every message.
