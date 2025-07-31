# Claude Code SuperCrew Hooks Feature

## Overview

The Claude Code SuperCrew now includes a comprehensive hooks system that integrates with Claude Code's hook mechanism, allowing automated tasks to run on various events.

## Implementation Details

### 1. Global Hook Scripts

Created 5 essential hook scripts in `/SuperCrew/Hooks/`:

- **git-auto-commit.sh**: Automatically commits changes made by Claude Code
- **lint-on-save.sh**: Runs language-specific linters after file modifications
- **test-on-change.sh**: Executes relevant tests when code files change
- **security-scan.sh**: Scans for security vulnerabilities and blocks dangerous operations
- **backup-before-change.sh**: Creates backups before file modifications

### 2. Hook Management System

Created a hook manager (`internal/hooks/manager.go`) that:
- Discovers available hook scripts
- Manages hook enable/disable functionality
- Updates Claude Code's `~/.claude/settings.json` automatically
- Provides configuration options for each hook

### 3. Interactive CLI Command

Added `crew hooks` command with features:
- Interactive menu for hook management
- Enable/disable multiple hooks at once
- Configure hook-specific settings
- List all available hooks with status
- Configure hooks in `~/.claude/settings.json`

### 4. Installation Integration

Hooks are now part of the standard installation:
- Included in "Quick Installation" profile
- Available as a component during custom installation
- Automatically copied to installation directory

## Usage

### Interactive Mode
```bash
crew hooks
```

### Command Line Options
```bash
# List all hooks
crew hooks --list

# Enable specific hooks
crew hooks --enable git-auto-commit
crew hooks --enable lint-on-save

# Disable hooks
crew hooks --disable test-on-change

# Install hooks only
crew hooks --install-only
```

### Configuration

Each hook supports environment variables for configuration:

```bash
# Git auto-commit
SUPERCREW_GIT_AUTO_COMMIT=true/false

# Linting
SUPERCREW_LINT_AUTOFIX=true/false
SUPERCREW_LINT_QUIET=true/false

# Testing
SUPERCREW_TEST_PATTERN=auto|unit|integration|all
SUPERCREW_TEST_COVERAGE=true/false

# Security
SUPERCREW_SECURITY_BLOCK=true/false
SUPERCREW_SECURITY_LEVEL=low|medium|high

# Backup
SUPERCREW_BACKUP_DIR=.claude/backups
SUPERCREW_BACKUP_DAYS=7
```

## Technical Architecture

### Hook Types Supported
- **PreToolUse**: Runs before tools (e.g., backup-before-change)
- **PostToolUse**: Runs after tools (e.g., git-auto-commit, lint-on-save)

### Integration with Claude Code
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

## Benefits

1. **Automation**: Reduces manual tasks like committing changes or running tests
2. **Quality**: Automatic linting and security scanning improve code quality
3. **Safety**: Backup hooks protect against accidental data loss
4. **Customization**: Each hook can be configured to match project needs
5. **Integration**: Works seamlessly with Claude Code's existing hook system

## Future Enhancements

- Support for custom project-specific hooks
- Hook templates for common workflows
- Hook chaining and dependencies
- Performance metrics and logging
- Web-based hook configuration UI