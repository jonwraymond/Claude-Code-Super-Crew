# Claude Code Super Crew Hooks

Hooks allow you to automate tasks when using Claude Code. They run automatically based on various events.

## Available Hooks

### 🤖 git-auto-commit
Automatically commits changes made by Claude Code to git.
- **When**: After Write, Edit, or MultiEdit tools
- **Config**: `SUPERCREW_GIT_AUTO_COMMIT=true/false`

### 🔍 lint-on-save
Runs linters after file modifications to maintain code quality.
- **When**: After Write, Edit, or MultiEdit tools
- **Config**: 
  - `SUPERCREW_LINT_AUTOFIX=true/false` - Enable auto-fixing
  - `SUPERCREW_LINT_QUIET=true/false` - Suppress output

### 🧪 test-on-change
Runs relevant tests when code files are modified.
- **When**: After Write, Edit, or MultiEdit tools
- **Config**:
  - `SUPERCREW_TEST_PATTERN=auto|unit|integration|all`
  - `SUPERCREW_TEST_COVERAGE=true/false`

### 🔒 security-scan
Scans code changes for security vulnerabilities.
- **When**: After Write, Edit, or MultiEdit tools
- **Config**:
  - `SUPERCREW_SECURITY_BLOCK=true/false` - Block dangerous operations
  - `SUPERCREW_SECURITY_LEVEL=low|medium|high`

### 📦 backup-before-change
Creates backups before modifying files.
- **When**: Before Write, Edit, or MultiEdit tools
- **Config**:
  - `SUPERCREW_BACKUP_DIR=.claude/backups`
  - `SUPERCREW_BACKUP_DAYS=7` - Days to keep backups

## Installation

Install hooks using the crew command:

```bash
# Interactive installation
crew hooks

# Enable specific hooks
crew hooks --enable git-auto-commit
crew hooks --enable lint-on-save

# List all available hooks
crew hooks --list
```

## Manual Configuration

Hooks are configured in `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$HOME/.claude/hooks/git-auto-commit.sh",
            "env": {
              "SUPERCREW_GIT_AUTO_COMMIT": "true"
            }
          }
        ]
      }
    ]
  }
}
```

## Creating Custom Hooks

You can create your own hooks by:

1. Creating a script in `.claude/hooks/` (project-level) or any location
2. Making it executable: `chmod +x your-hook.sh`
3. Adding it to `~/.claude/settings.json` (user-level) or `.claude/settings.json` (project-level)

Hook scripts receive tool information via stdin as JSON and can:
- Process the data
- Block operations by exiting with non-zero status
- Add output that will be shown to the user

## Troubleshooting

- **Hooks not running**: Check `~/.claude/settings.json` configuration
- **Permission denied**: Ensure hook scripts are executable (`chmod +x`)
- **Hook failing**: Run the hook manually to debug: `echo '{}' | .claude/hooks/your-hook.sh`