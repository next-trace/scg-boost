# Release Process

This document outlines the release process for SCG-Boost.

## Pre-Release Checklist

Before creating a release, ensure:

- [ ] All CI checks passing on main branch
- [ ] Test coverage >= 70% (enforced by CI)
- [ ] No known critical bugs
- [ ] Dependencies up to date (`go mod tidy`)
- [ ] Version updated in `boost/boost.go` (`Version` constant)
- [ ] CHANGELOG.md updated with release notes
- [ ] Documentation reflects current state
- [ ] Skills metadata current (all skill.json files valid)

## Versioning

SCG-Boost follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) - Breaking changes to public API or CLI
- **MINOR** (0.X.0) - New features, backward compatible
- **PATCH** (0.0.X) - Bug fixes, backward compatible

### What Constitutes a Breaking Change?

- Removing or renaming CLI commands
- Changing CLI flag names or behavior
- Removing or changing MCP tool signatures
- Breaking changes to `boost/` public API
- Removing or renaming skills

### Backward Compatibility Promise

We maintain backward compatibility for:

- CLI commands and flags (deprecation warnings before removal)
- MCP tool contracts (names, input/output schemas)
- Public API in `boost/` package
- Skill names and structure

We do NOT guarantee compatibility for:

- Internal packages (`internal/*`)
- Resource file formats (may evolve)
- Development tools and scripts

## Release Steps

### 1. Prepare Release Branch

```sh
# Ensure main is up to date
git checkout main
git pull origin main

# Create release branch (optional, for final checks)
git checkout -b release/v0.X.0
```

### 2. Update Version

Edit `boost/boost.go`:

```go
const Version = "0.X.0"
```

### 3. Update CHANGELOG

Add release section to `CHANGELOG.md`:

```markdown
## [0.X.0] - YYYY-MM-DD

### Added
- New skill: scg-example
- CLI command: skills:export

### Changed
- Improved auto-detection logic

### Fixed
- Bug in skills:sync with missing skill.json

### Deprecated
- Old install --legacy flag (use --force instead)
```

### 4. Commit Version Bump

```sh
git add boost/boost.go CHANGELOG.md
git commit -m "chore: bump version to 0.X.0"
git push origin release/v0.X.0  # or main
```

### 5. Create Git Tag

```sh
# Create annotated tag
git tag -a v0.X.0 -m "Release v0.X.0"

# Push tag to trigger release workflow
git push origin v0.X.0
```

### 6. Create GitHub Release

Go to [GitHub Releases](https://github.com/next-trace/scg-boost/releases) and:

1. Click "Draft a new release"
2. Select tag `v0.X.0`
3. Title: `v0.X.0`
4. Copy relevant CHANGELOG section to description
5. Check "Set as latest release" (if applicable)
6. Click "Publish release"

GitHub Actions will automatically:

- Build binaries for multiple platforms
- Attach binaries to release
- Publish to package registries (if configured)

### 7. Verify Release

```sh
# Install from released version
go install github.com/next-trace/scg-boost/cmd/scg-boost@v0.X.0

# Verify version
scg-boost version
# Output: scg-boost version 0.X.0

# Test basic functionality
scg-boost skills:list
scg-boost tools
```

## Post-Release

### Announce

- Post release notes in internal #scg-tools channel
- Update documentation site (if applicable)
- Notify dependent projects

### Monitor

- Watch for bug reports
- Monitor CI for flaky tests
- Check download metrics

### Hotfix Process

If a critical bug is found post-release:

1. Create hotfix branch from release tag:
   ```sh
   git checkout -b hotfix/v0.X.1 v0.X.0
   ```

2. Fix bug and add tests

3. Update version to 0.X.1

4. Follow normal release process

## Rollback Procedure

If a release has critical issues:

### Option 1: Yank Release (Preferred)

1. Delete Git tag locally and remotely:
   ```sh
   git tag -d v0.X.0
   git push origin :refs/tags/v0.X.0
   ```

2. Mark GitHub release as "Pre-release" or delete it

3. Communicate to users to avoid version

### Option 2: Publish Hotfix

1. Follow hotfix process above
2. Release 0.X.1 immediately
3. Document issues in CHANGELOG

## Release Cadence

- **Patch releases**: As needed for bug fixes
- **Minor releases**: Monthly or when features accumulate
- **Major releases**: Only when breaking changes necessary

## Automation

Future improvements:

- [ ] Automated changelog generation from commit messages
- [ ] Automated version bumping via CI
- [ ] Automated binary uploads to releases
- [ ] Automated Docker image publishing

## Release History

| Version | Date       | Type  | Highlights                  |
|---------|------------|-------|-----------------------------|
| 0.2.0   | TBD        | Minor | Skills system, production-ready |
| 0.1.0   | 2024-02-04 | Minor | Initial MCP server, 16 tools |

## Support Policy

- **Latest version**: Full support, active development
- **Previous minor**: Security fixes for 6 months
- **Older versions**: No support

Users are encouraged to upgrade to the latest version.
