#!/bin/bash
#
# Claude Code Super Crew - Security Scan Hook
#
# This hook runs security scans on code changes to catch vulnerabilities early.
# It blocks dangerous operations and alerts on security issues.
#
# Installation:
#   crew install hooks --enable security-scan
#
# Configuration:
#   SUPERCREW_SECURITY_BLOCK=true   # Block operations with security issues
#   SUPERCREW_SECURITY_LEVEL=medium # low, medium, high
#

set -euo pipefail

# Configuration
BLOCK="${SUPERCREW_SECURITY_BLOCK:-true}"
LEVEL="${SUPERCREW_SECURITY_LEVEL:-medium}"

# Read tool output
TOOL_OUTPUT=$(cat)
FILE_PATH=$(echo "$TOOL_OUTPUT" | jq -r '.result.file_path // .file_path // empty' 2>/dev/null)
CONTENT=$(echo "$TOOL_OUTPUT" | jq -r '.result.content // .content // empty' 2>/dev/null)

if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Security patterns to check
declare -A CRITICAL_PATTERNS=(
    ["hardcoded_secret"]='(api_key|apikey|password|secret|token)\s*=\s*["\x27][^"\x27]+["\x27]'
    ["sql_injection"]='(SELECT|INSERT|UPDATE|DELETE).*\+.*\$'
    ["command_injection"]='(exec|system|eval)\s*\('
    ["path_traversal"]='\.\./'
)

declare -A WARNING_PATTERNS=(
    ["weak_crypto"]='(MD5|SHA1)\s*\('
    ["insecure_random"]='rand\(\)'
    ["debug_enabled"]='(debug|DEBUG)\s*=\s*(true|True|1)'
)

# Function to scan content
scan_content() {
    local severity="$1"
    local -n patterns=$2
    local found_issues=0
    
    for issue in "${!patterns[@]}"; do
        if echo "$CONTENT" | grep -qE "${patterns[$issue]}" 2>/dev/null; then
            echo "🚨 $severity: Potential $issue detected in $FILE_PATH"
            found_issues=$((found_issues + 1))
        fi
    done
    
    return $found_issues
}

# Run security scans based on file type
EXT="${FILE_PATH##*.}"

echo "🔒 Security scanning $FILE_PATH..."

# Check for critical issues
CRITICAL_FOUND=0
scan_content "CRITICAL" CRITICAL_PATTERNS || CRITICAL_FOUND=$?

# Check for warnings if not blocking on critical
if [[ "$LEVEL" != "high" ]] || [[ $CRITICAL_FOUND -eq 0 ]]; then
    scan_content "WARNING" WARNING_PATTERNS || true
fi

# Language-specific security tools
case "$EXT" in
    go)
        if command -v gosec &> /dev/null; then
            gosec -quiet "$FILE_PATH" 2>&1 || {
                echo "⚠️  Go security issues found"
                CRITICAL_FOUND=$((CRITICAL_FOUND + 1))
            }
        fi
        ;;
    js|jsx|ts|tsx)
        if command -v semgrep &> /dev/null; then
            semgrep --config=auto --quiet "$FILE_PATH" 2>&1 || {
                echo "⚠️  JavaScript security issues found"
                CRITICAL_FOUND=$((CRITICAL_FOUND + 1))
            }
        fi
        ;;
    py)
        if command -v bandit &> /dev/null; then
            bandit -q "$FILE_PATH" 2>&1 || {
                echo "⚠️  Python security issues found"
                CRITICAL_FOUND=$((CRITICAL_FOUND + 1))
            }
        fi
        ;;
esac

# Block if critical issues found and blocking is enabled
if [[ "$BLOCK" == "true" ]] && [[ $CRITICAL_FOUND -gt 0 ]]; then
    echo "❌ Security scan failed. Operation blocked."
    echo "Fix the security issues before proceeding."
    exit 1
fi

exit 0