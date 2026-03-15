# CLAUDE.md — scg-infra

This file is the **single source of truth** for how Claude must work in this repository.

If you violate these rules, you create chaos. Don't.

## What This Repo Is

Infrastructure-as-code repo (Terraform + ./scg helper). Owns AWS foundations and runtime stacks for dev/staging/prod.

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

This repository owns:
- Terraform stacks/modules for AWS (dev/staging/prod)
- State backend (S3) + locking (DynamoDB)
- Environment orchestration via `./scg` helper
- **IAM/OIDC roles for GitHub Actions**
- Network foundations and runtime infrastructure

This repository does **not** own:
- Application code changes (services/libraries)
- Manual click‑ops in AWS console (except emergency break-glass)

## Hard Boundaries

- Never hardcode secrets. Everything goes via SSM/Secrets Manager with clear ownership.
- Never “fix” things by adding random `depends_on` everywhere. Root-cause it.
- No environment drift: dev/staging/prod must be structurally identical (sizes may differ).

## Stack Rules

- Stacks must be **idempotent**.
- Stacks must support **import** for existing resources rather than “delete and recreate”.
- Names must be deterministic and include env/region.

### Apply Order (typical)

```bash
# 1) Foundation (network + iam)
./scg plan  --stack foundation-network --env dev --region eu-central-1
./scg apply --stack foundation-network --env dev --region eu-central-1 --yes

./scg plan  --stack foundation-iam --env dev --region eu-central-1
./scg apply --stack foundation-iam --env dev --region eu-central-1 --yes

# 2) Shared services
./scg plan  --stack postgres --env dev --region eu-central-1
./scg apply --stack postgres --env dev --region eu-central-1 --yes

./scg plan  --stack kafka --env dev --region eu-central-1
./scg apply --stack kafka --env dev --region eu-central-1 --yes

# 3) Runtime
./scg plan  --stack runtime-ecs --env dev --region eu-central-1
./scg apply --stack runtime-ecs --env dev --region eu-central-1 --yes
```

## CI/CD Rules

- GitHub Actions must assume environment‑scoped OIDC roles (dev/staging/prod).
- No CI job is allowed to create IAM roles for runtime services unless that’s explicitly part of foundation-iam.

## “Destroy” Rules (Safety)

- No blanket “nuke everything” defaults.
- Any destroy command must require `--confirm-env <env>` and refuse production unless `--force-production` is explicitly passed.

## Terraform Style

- Avoid dynamic magic. Prefer explicit modules and outputs.
- Guard rails must be implemented via `terraform_data` + `precondition` (or equivalent), with clean heredocs (no broken multi-line strings).

## Token/Cost Control (Mandatory)

Use subagents and compaction intentionally:
- For **search/triage**: use `repo-radar` subagent (keeps logs out of main context).
- For **tests/lint**: use `go-test-runner` subagent.
- For **contract safety**: use `contract-guard` subagent.
- For **terraform failures** (infra only): use `terraform-doctor` subagent.
- When a conversation grows, use `/compact` to summarize and reset context.

Reason: large logs and long back-and-forth are what burn tokens. Subagents keep heavy work isolated, and `/compact` preserves decisions without carrying every message.
