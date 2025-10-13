# PLAN — SCG Claude Structure (Commands + Skills + Subagents)

Goal:
- Reduce token usage and keep Claude outputs deterministic and safe across many repos.

## 1) Commands/Skills (Slash Commands)

We use Claude Code custom slash commands stored in `.claude/commands/` (per repo).
These commands are tiny prompt wrappers that always delegate heavy work to subagents.

Standard commands:
- triage: find minimal relevant files (repo-radar)
- test: run tests/lint and summarize failures (go-test-runner)
- contract-check: detect breaking changes (contract-guard)
- review: ruthless review with evidence (code-reviewer)
- terraform-fix: infra-only terraform diagnosis (terraform-doctor)

## 2) Subagents

Subagents live in `.claude/agents/` with YAML frontmatter.

Design principles:
- Separate context to avoid polluting the main thread.
- Minimal tool permissions (default: read-only; only test/terraform agents get Bash).
- Models:
  - Haiku for fast triage/test summaries (cheap, short)
  - Sonnet for reviews/contract safety (more reasoning)

## 3) Delegation Rules

Main Claude must:
- delegate search/test/terraform diagnosis to subagents
- keep main conversation focused on decisions + final patch

Use `/compact` when conversations get long to keep only decisions and constraints.

## 4) Repo Contracts

Every repo has `.claude/CLAUDE.md` as the constitution:
- boundaries (control plane vs data plane)
- no random deps
- stability of API/contracts
- idempotent infra

## Repositories Covered (Current)

- identity-service
- scg-infra
- gateway-service
- scg-test-kit
- scg-validator
- scg-database
- scg-service-api
- scg-config
- scg-contracts
- scg-notification-contracts
- scg-boost
- scg-service-bus
- scg-support
- scg-error
- scg-logger

