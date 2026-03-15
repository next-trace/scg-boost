# Subagents (Project-level)

These subagents live in `.claude/agents/` and are checked into git.

Purpose:
- Keep the main chat clean (token control)
- Delegate high-volume actions (search/test/terraform diagnosis)
- Return only summarized results

Usage in Claude Code:
- `/agents` to view them
- Ask Claude: "Use the <agent-name> agent to ..."

Notes:
- Built-in `Explore` and `Plan` agents already exist; our custom agents are SCG-specific.
