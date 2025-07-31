#!/bin/bash
#
# Claude Code Super Crew - Git Auto-Commit Hook
# 
# This hook automatically commits changes made by Claude Code.
# It runs after Write, Edit, or MultiEdit tools are used.
#
# Installation:
#   crew install hooks --enable git-auto-commit
#
# Configuration:
#   Set SUPERCREW_GIT_AUTO_COMMIT=false to disable temporarily
#

set -euo pipefail

# Check if auto-commit is enabled
if [[ "${SUPERCREW_GIT_AUTO_COMMIT:-true}" == "false" ]]; then
    exit 0
fi

# Read tool output from stdin
TOOL_OUTPUT=$(cat)

# Extract file path from tool output
FILE_PATH=$(echo "$TOOL_OUTPUT" | jq -r '.result.file_path // .file_path // empty' 2>/dev/null)

if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Get the project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT_DIR"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    exit 0
fi

# Check if the file has changes
if ! git diff --quiet "$FILE_PATH" 2>/dev/null && ! git diff --cached --quiet "$FILE_PATH" 2>/dev/null; then
    # Add the file to git
    git add "$FILE_PATH"
    
    # Extract tool name for commit message
    TOOL_NAME=$(echo "$TOOL_OUTPUT" | jq -r '.tool // "Claude Code"' 2>/dev/null)
    
    # Create commit message
    COMMIT_MSG="🤖 Auto-commit: Updated $(basename "$FILE_PATH")

Tool: $TOOL_NAME
File: $FILE_PATH

🤖 Generated with Claude Code Super Crew
Co-Authored-By: Claude <noreply@anthropic.com>"
    
    # Commit the changes
    git commit -m "$COMMIT_MSG" --no-verify > /dev/null 2>&1 || true
    
    echo "✅ Auto-committed changes to $FILE_PATH"
fi