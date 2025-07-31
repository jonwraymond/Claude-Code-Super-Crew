# CLAUDE.md Location Fix

## Issue
The Claude Code SuperCrew was creating `CLAUDE.md` in `.claude/CLAUDE.md` within projects, but Claude's `/init` command creates it in the project root.

## Fix Applied
Updated the load integration code to create `CLAUDE.md` in the project root:

### Before:
```go
claudeFile := filepath.Join(lch.ProjectRoot, ".claude", "CLAUDE.md")
```

### After:
```go
claudeFile := filepath.Join(lch.ProjectRoot, "CLAUDE.md")
```

## Context

### Global Installation
- Global files go in `~/.claude/` (user home directory)
- This includes global `CLAUDE.md`, `COMMANDS.md`, etc.
- This remains unchanged

### Project-Level Installation
- Project `CLAUDE.md` goes in project root (same as `/init`)
- Project `.claude/` directory contains:
  - `agents/` - project-specific agents
  - `commands/` - project-specific commands
  - `analysis.json` - project analysis results
  - `backups/` - backup files from hooks

## Behavior After Fix

When running `crew claude --install` or `/crew:onboard`:
1. Creates `CLAUDE.md` in project root
2. Creates `.claude/` directory for other project files
3. Compatible with Claude's `/init` command

## Migration for Existing Projects

If you have an existing project with `.claude/CLAUDE.md`:
```bash
# Move CLAUDE.md to project root
mv .claude/CLAUDE.md ./CLAUDE.md
```

Or simply run `/init` in Claude to create it in the correct location.