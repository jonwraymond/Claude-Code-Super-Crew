#!/bin/bash
#
# Claude Code Super Crew - Lint on Save Hook
#
# This hook automatically runs linters after file modifications.
# It supports multiple languages and can auto-fix issues.
#
# Installation:
#   crew install hooks --enable lint-on-save
#
# Configuration:
#   SUPERCREW_LINT_AUTOFIX=true  # Enable auto-fixing
#   SUPERCREW_LINT_QUIET=true    # Suppress output unless errors
#

set -euo pipefail

# Configuration
AUTOFIX="${SUPERCREW_LINT_AUTOFIX:-false}"
QUIET="${SUPERCREW_LINT_QUIET:-false}"

# Read tool output from stdin
TOOL_OUTPUT=$(cat)

# Extract file path
FILE_PATH=$(echo "$TOOL_OUTPUT" | jq -r '.result.file_path // .file_path // empty' 2>/dev/null)

if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Get file extension
EXT="${FILE_PATH##*.}"

# Function to run linter
run_linter() {
    local linter="$1"
    local args="$2"
    local file="$3"
    
    if command -v "$linter" &> /dev/null; then
        if [[ "$QUIET" == "true" ]]; then
            $linter $args "$file" > /dev/null 2>&1 || {
                echo "⚠️  Linting issues found in $file"
                $linter $args "$file" 2>&1 | tail -n 10
            }
        else
            echo "🔍 Running $linter on $file"
            $linter $args "$file"
        fi
    fi
}

# Language-specific linting
case "$EXT" in
    go)
        run_linter "gofmt" "-w" "$FILE_PATH"
        run_linter "golint" "" "$FILE_PATH"
        ;;
    js|jsx|ts|tsx)
        if [[ "$AUTOFIX" == "true" ]]; then
            run_linter "eslint" "--fix" "$FILE_PATH"
        else
            run_linter "eslint" "" "$FILE_PATH"
        fi
        ;;
    py)
        if [[ "$AUTOFIX" == "true" ]]; then
            run_linter "black" "" "$FILE_PATH"
            run_linter "isort" "" "$FILE_PATH"
        fi
        run_linter "flake8" "" "$FILE_PATH"
        ;;
    rs)
        run_linter "rustfmt" "" "$FILE_PATH"
        ;;
    sh|bash)
        run_linter "shellcheck" "" "$FILE_PATH"
        ;;
    yml|yaml)
        run_linter "yamllint" "" "$FILE_PATH"
        ;;
    json)
        # Validate JSON
        if command -v jq &> /dev/null; then
            jq . "$FILE_PATH" > /dev/null 2>&1 || echo "⚠️  Invalid JSON in $FILE_PATH"
        fi
        ;;
esac

exit 0