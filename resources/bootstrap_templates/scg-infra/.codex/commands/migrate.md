# /migrate [action]

Get migration status and guidance.

Actions:
- (no action) - show migration status
- `plan` - plan a new migration
- `check` - validate migration safety

Uses `migrations.status` MCP tool if available.

Return:
- Current migration status
- Pending migrations
- Safety recommendations
