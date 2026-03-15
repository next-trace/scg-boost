# SCG Skills (Claude Code Slash Commands)

These are project-local slash commands (skills).
They live in `.claude/commands/` and show up in Claude Code autocomplete.

Commands:
- `/triage <problem>`
- `/test [scope]`
- `/contract-check <change-summary>`
- `/review <diff-summary>`
- `/terraform-fix <error-log>` (infra only)

Guidelines:
- Use these to delegate to subagents and keep main context small.
- Prefer minimal outputs; avoid pasting huge logs into the main chat.
