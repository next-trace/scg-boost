# Security Policy

## Supported Versions

We actively support the following versions of SCG-Boost:

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

**Do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in SCG-Boost, please report it to:

**security@supplychainguard.dev**

Please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

### Response Timeline

- **Initial response:** Within 48 hours
- **Status update:** Within 7 days
- **Fix timeline:** Depends on severity (critical issues prioritized)

## Security Considerations

SCG-Boost is designed with security in mind:

### Read-Only Database Access

All database tools (`dbquery.run`, `dbschema.list`) enforce read-only operations:
- No `INSERT`, `UPDATE`, `DELETE`, or `DROP` statements allowed
- Query validation before execution
- Transaction isolation to prevent modifications

### Configuration Redaction

Sensitive configuration values are automatically redacted:
- API keys, tokens, passwords masked
- Only safe values exposed through `config.get`
- `config.list` shows keys without values

### No Shell Execution

- No arbitrary shell command execution
- All file operations are sandboxed
- MCP server runs with minimal privileges

### MCP Protocol Safety

- Validates all incoming requests
- Enforces schema compliance
- Rejects malformed payloads

### Dependency Management

- Minimal third-party dependencies
- Regular security audits via `go mod verify`
- Dependabot alerts enabled

## Best Practices for Users

When embedding SCG-Boost in your service:

1. **Run with least privilege** - Don't run as root
2. **Isolate the MCP server** - Use separate process/container if possible
3. **Monitor logs** - Watch for unusual query patterns
4. **Validate inputs** - Even though tools validate, add your own checks
5. **Keep updated** - Apply security patches promptly

## Security Updates

Security fixes are released as:
- **Patch versions** for backward-compatible fixes (e.g., 0.2.1 → 0.2.2)
- **Minor versions** if API changes required (e.g., 0.2.x → 0.3.0)

Subscribe to releases on GitHub to stay informed.
