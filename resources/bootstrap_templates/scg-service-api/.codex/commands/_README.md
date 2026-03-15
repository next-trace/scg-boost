# Project Commands (Codex)

These are project-local commands.
They live in `.codex/commands/`.

Commands:
- `/triage <problem>`
- `/test [scope]`
- `/contract-check <change-summary>`
- `/review <diff-summary>`
- `/plan <task>`
- `/debug <issue>`
- `/docs <query>`
- `/migrate [action]`

Guidelines:
- Use these to delegate to subagents and keep main context small.
- Prefer minimal outputs; avoid pasting huge logs into the main chat.
