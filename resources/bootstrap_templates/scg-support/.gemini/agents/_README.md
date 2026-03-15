# Subagents (Project-level)

These subagents live in `.claude/agents/` and are checked into git.

Available agents:
- `repo-radar` - Search repo and return minimal findings (Haiku)
- `code-reviewer` - Review code for security/best practices (Sonnet)
- `go-test-runner` - Run tests and summarize failures
- `contract-guard` - Guard auth/token contracts
- `debug-assistant` - Debug production issues (Sonnet)
- `schema-explorer` - Explore database schema (Haiku)

Purpose:
- Keep the main chat clean (token control)
- Delegate high-volume actions (search/test/terraform diagnosis)
- Return only summarized results

Usage in Claude Code:
- `/agents` to view them
- Ask Claude: "Use the <agent-name> agent to ..."

Notes:
- Built-in `Explore` and `Plan` agents already exist; our custom agents are SCG-specific.
