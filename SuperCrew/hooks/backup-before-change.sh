#!/bin/bash
#
# Claude Code Super Crew - Backup Before Change Hook
#
# This hook creates backups before modifying files, allowing easy rollback.
# Useful for protecting against accidental overwrites or bad changes.
#
# Installation:
#   crew install hooks --enable backup-before-change
#
# Configuration:
#   SUPERCREW_BACKUP_DIR=.claude/backups  # Where to store backups
#   SUPERCREW_BACKUP_DAYS=7               # Days to keep backups
#

set -euo pipefail

# Configuration
BACKUP_DIR="${SUPERCREW_BACKUP_DIR:-.claude/backups}"
BACKUP_DAYS="${SUPERCREW_BACKUP_DAYS:-7}"

# Read tool input (PreToolUse hook)
TOOL_INPUT=$(cat)
TOOL_NAME=$(echo "$TOOL_INPUT" | jq -r '.tool // empty' 2>/dev/null)

# Only backup for file modification tools
case "$TOOL_NAME" in
    Write|Edit|MultiEdit)
        ;;
    *)
        echo "$TOOL_INPUT"
        exit 0
        ;;
esac

# Extract file path
FILE_PATH=$(echo "$TOOL_INPUT" | jq -r '.parameters.file_path // empty' 2>/dev/null)

if [[ -z "$FILE_PATH" ]]; then
    echo "$TOOL_INPUT"
    exit 0
fi

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT_DIR"

# Create backup directory
FULL_BACKUP_DIR="$PROJECT_DIR/$BACKUP_DIR"
mkdir -p "$FULL_BACKUP_DIR"

# If file exists, create backup
if [[ -f "$FILE_PATH" ]]; then
    # Generate backup filename with timestamp
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_NAME="$(basename "$FILE_PATH").${TIMESTAMP}.bak"
    BACKUP_PATH="$FULL_BACKUP_DIR/$BACKUP_NAME"
    
    # Copy file to backup
    cp "$FILE_PATH" "$BACKUP_PATH"
    
    # Create metadata file
    cat > "${BACKUP_PATH}.meta" <<EOF
{
  "original_path": "$FILE_PATH",
  "backup_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "tool": "$TOOL_NAME",
  "file_size": $(stat -f%z "$FILE_PATH" 2>/dev/null || stat -c%s "$FILE_PATH" 2>/dev/null || echo 0),
  "file_hash": "$(shasum -a 256 "$FILE_PATH" | cut -d' ' -f1)"
}
EOF
    
    echo "📦 Backed up $FILE_PATH to $BACKUP_PATH"
fi

# Clean old backups
find "$FULL_BACKUP_DIR" -name "*.bak" -mtime +$BACKUP_DAYS -delete 2>/dev/null || true
find "$FULL_BACKUP_DIR" -name "*.meta" -mtime +$BACKUP_DAYS -delete 2>/dev/null || true

# Pass through the original input
echo "$TOOL_INPUT"