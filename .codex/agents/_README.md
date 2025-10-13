# Subagents (Project-level)

These subagents live in `.codex/agents/` and are checked into git.

Available agents:
- `repo-radar` - Search repo and return minimal findings
- `code-reviewer` - Review code for security/best practices
- `go-test-runner` - Run tests and summarize failures
- `contract-guard` - Guard auth/token contracts
- `debug-assistant` - Debug production issues
- `schema-explorer` - Explore database schema

Purpose:
- Keep the main chat clean (token control)
- Delegate high-volume actions (search/test/diagnosis)
- Return only summarized results

Usage in Codex:
- Ask: "Use the <agent-name> agent to ..."

Notes:
- Built-in tools already exist; our custom agents are SCG-specific.
