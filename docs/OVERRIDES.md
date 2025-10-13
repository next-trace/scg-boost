# Skill Overrides

SCG-Boost supports local overrides that survive `skills:sync`.

## How It Works

- Overrides live under `.scg/overrides/<skill-id>/`.
- On install/sync, overrides are applied after the base skill is installed.
- Only paths listed in `override_paths` are allowed.

Precedence:
1. `.scg/overrides/<skill-id>/` (highest)
2. `.claude/` base skill files

## Creating Overrides

List overrideable paths:
```sh
scg-boost skills:override --skill gateway-service
```

Create an override file:
```sh
scg-boost skills:override --skill gateway-service --path .claude/commands/custom.md
```

Section override (Markdown headings):
```sh
scg-boost skills:override --skill gateway-service --path .claude/CLAUDE.md#repo-specific-rules
```

For section overrides, the override file should contain only the section body
(no heading line). The existing heading is preserved.

## Override Layout

```
.scg/
└── overrides/
    └── gateway-service/
        ├── CLAUDE.md
        └── commands/
            └── custom.md
```
