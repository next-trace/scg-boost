# SCG Skills (Claude Code Slash Commands)

These are project-local slash commands (skills).
They live in `.claude/commands/` and show up in Claude Code autocomplete.

Commands:
- `/triage <problem>` - Triage issues using repo-radar
- `/test [scope]` - Run tests via go-test-runner
- `/contract-check <change-summary>` - Check auth contracts
- `/review <diff-summary>` - Review code changes
- `/debug <issue>` - Debug production issues
- `/migrate [action]` - Migration status and guidance
- `/docs <query>` - Search project documentation
- `/terraform-fix <error-log>` (infra only)

Guidelines:
- Use these to delegate to subagents and keep main context small.
- Prefer minimal outputs; avoid pasting huge logs into the main chat.
