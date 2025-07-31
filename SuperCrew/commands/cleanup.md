---
allowed-tools: [Read, Grep, Glob, Bash, Edit, MultiEdit]
description: "Clean up code, remove dead code, optimize structure, and clean project artifacts"
---

# /crew:cleanup - Code and Project Cleanup

## Purpose
Systematically clean up code, remove dead code, optimize imports, improve project structure, and clean development artifacts.

## Usage
```
/crew:cleanup [target] [--type code|imports|files|artifacts|all] [--project] [--safe|--aggressive]
```

## Arguments
- `target` - Files, directories, or entire project to clean (default: current directory)
- `--type` - Cleanup type:
  - `code` - Dead code, unused functions, commented code
  - `imports` - Unused imports, duplicate imports
  - `files` - Empty files, redundant files
  - `artifacts` - Build outputs, logs, temp files
  - `all` - All types (default)
- `--project` - Clean entire project including artifacts (safety backup first)
- `--safe` - Conservative cleanup (default)
- `--aggressive` - More thorough cleanup with higher risk
- `--dry-run` - Preview changes without applying them
- `--no-backup` - Skip backup creation (not recommended with --project)
- `--exclude <pattern>` - Exclude files matching pattern

## Project-Level Cleanup (--project)

When using `--project` flag, performs comprehensive cleanup:

### Safety First
1. Creates timestamped backup in `~/.claude/backups/crew/[project]/[timestamp]/`
2. Generates restoration script
3. Protects critical directories:
   - `.claude/` - Claude configurations
   - `.git/` - Version control
   - `node_modules/` (unless --aggressive)
   - Source files (unless explicitly targeted)

### Default Artifact Cleanup
- Log files: `*.log`, `*.debug`, `*.trace`
- Temporary files: `*.tmp`, `*.temp`, `*.cache`
- OS artifacts: `.DS_Store`, `Thumbs.db`, `desktop.ini`
- Python: `__pycache__/`, `*.pyc`, `*.pyo`, `.pytest_cache/`
- JavaScript/TypeScript: `dist/`, `build/`, `.next/`, `.nuxt/`
- Coverage: `coverage/`, `.nyc_output/`, `*.lcov`
- IDE: `.idea/`, `.vscode/workspace.settings`

### Aggressive Cleanup (--aggressive --project)
- Dependencies: `node_modules/`, `vendor/`, virtual environments
- Generated docs: `docs/generated/`, `api-docs/`
- Large binaries: `*.exe`, `*.dll`, `*.so` (with size check)
- Build caches: `.gradle/`, `.maven/`, `target/`

## Code Cleanup

### Dead Code Detection
- Unused functions and variables
- Unreachable code after return/throw
- Commented-out code blocks
- Empty catch blocks
- Redundant conditions

### Import Optimization
- Remove unused imports
- Consolidate duplicate imports
- Sort imports by convention
- Fix import paths

## Execution Flow

### 1. Analysis Phase
```bash
# Scan for cleanup opportunities
/crew:cleanup --dry-run --project

# Check specific types
/crew:cleanup --type artifacts --dry-run
```

### 2. Backup Phase (--project)
```bash
# Automatic backup location
~/.claude/backups/crew/[project-name]/[timestamp]/
├── manifest.json      # List of all changes
├── removed_files/     # Copies of deleted files
├── restore.sh        # One-click restoration
└── cleanup.log       # Detailed operation log
```

### 3. Cleanup Phase
- Executes cleanup operations
- Validates each change
- Maintains atomic operations
- Reports progress

### 4. Verification Phase
- Ensures project still builds/runs
- Validates no critical files affected
- Generates summary report

## Examples

### Basic code cleanup
```
/crew:cleanup src/ --type code
```

### Project-wide artifact cleanup with preview
```
/crew:cleanup --project --type artifacts --dry-run
```

### Aggressive full cleanup
```
/crew:cleanup --project --aggressive --type all
```

### Exclude patterns
```
/crew:cleanup --project --exclude "*.test.js" --exclude "temp_*"
```

### Clean specific language artifacts
```
# Python project
/crew:cleanup --project --type artifacts --exclude "venv/"

# Node.js project  
/crew:cleanup --project --aggressive --type artifacts
```

## Configuration

Create `.claude/cleanup.json` for project-specific rules:
```json
{
  "artifacts": {
    "include": ["*.custom", "temp/"],
    "exclude": ["important.tmp", "*.dev.log"]
  },
  "protected": ["my-special-dir/"],
  "backup_retention_days": 14,
  "code_cleanup": {
    "remove_console_logs": false,
    "remove_todo_comments": false
  }
}
```

## Restoration

If cleanup removes something important:
```bash
# List available backups
ls ~/.claude/backups/crew/[project]/

# Restore from backup
~/.claude/backups/crew/[project]/[timestamp]/restore.sh

# Selective restoration
~/.claude/backups/crew/[project]/[timestamp]/restore.sh --file path/to/file
```

## Best Practices

1. Always run `--dry-run` first, especially with `--project`
2. Commit changes before major cleanup operations
3. Review cleanup report before proceeding
4. Keep backups for at least 7 days
5. Use `.claude/cleanup.json` for project-specific rules
6. Run tests after cleanup to ensure nothing broke

## Integration

Works well with other commands:
- `/crew:analyze` - Find what needs cleanup
- `/crew:build` - Verify build after cleanup
- `/crew:test` - Ensure tests pass post-cleanup
- `/crew:git` - Commit after successful cleanup