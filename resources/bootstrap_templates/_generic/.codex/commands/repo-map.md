# /repo-map <repo> <pattern>

Use MCP tool `scg_repo_find_files`.

Inputs:
- `repo`: manifest slug
- `pattern`: case-insensitive path fragment (for example: `migration`, `handler`, `.proto`)

Return:
- matched file list
- likely entrypoints
- suggested next inspection files
