# /repo-search <repo> <query> [filePattern]

Use MCP tool `scg_repo_search_text`.

Inputs:
- `repo`: manifest slug
- `query`: text to search
- optional `filePattern` (for example: `.go`, `README`, `Dockerfile`)

Return:
- top hits (file + line)
- grouped findings by concern
- concrete follow-up checks
