#!/bin/bash
#
# Claude Code Super Crew - Test on Change Hook
#
# This hook automatically runs relevant tests when code files are modified.
# It's smart about which tests to run based on what changed.
#
# Installation:
#   crew install hooks --enable test-on-change
#
# Configuration:
#   SUPERCREW_TEST_PATTERN=auto     # auto, unit, integration, all
#   SUPERCREW_TEST_COVERAGE=true    # Generate coverage reports
#

set -euo pipefail

# Configuration
TEST_PATTERN="${SUPERCREW_TEST_PATTERN:-auto}"
COVERAGE="${SUPERCREW_TEST_COVERAGE:-false}"

# Read tool output
TOOL_OUTPUT=$(cat)
FILE_PATH=$(echo "$TOOL_OUTPUT" | jq -r '.result.file_path // .file_path // empty' 2>/dev/null)

if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
cd "$PROJECT_DIR"

# Function to find related test files
find_test_files() {
    local source_file="$1"
    local dir=$(dirname "$source_file")
    local base=$(basename "$source_file" | sed 's/\.[^.]*$//')
    local ext="${source_file##*.}"
    
    case "$ext" in
        go)
            # Go test files
            echo "${dir}/${base}_test.go"
            ;;
        js|jsx|ts|tsx)
            # JavaScript/TypeScript test files
            echo "${dir}/__tests__/${base}.test.${ext}"
            echo "${dir}/${base}.test.${ext}"
            echo "${dir}/${base}.spec.${ext}"
            ;;
        py)
            # Python test files
            echo "${dir}/test_${base}.py"
            echo "${dir}/tests/test_${base}.py"
            ;;
        *)
            # Generic pattern
            echo "${dir}/*test*"
            ;;
    esac
}

# Determine what to test
if [[ "$TEST_PATTERN" == "auto" ]]; then
    # Find related test files
    for test_pattern in $(find_test_files "$FILE_PATH"); do
        if ls $test_pattern 2>/dev/null | head -n1 > /dev/null; then
            TEST_FILES=$(ls $test_pattern 2>/dev/null)
            break
        fi
    done
fi

# Run tests based on file type
EXT="${FILE_PATH##*.}"

case "$EXT" in
    go)
        echo "🧪 Running Go tests..."
        if [[ "$COVERAGE" == "true" ]]; then
            go test -cover -coverprofile=coverage.out $(dirname "$FILE_PATH")/...
        else
            go test $(dirname "$FILE_PATH")/...
        fi
        ;;
    js|jsx|ts|tsx)
        if [[ -f "package.json" ]]; then
            echo "🧪 Running JavaScript tests..."
            if command -v npm &> /dev/null; then
                if [[ -n "${TEST_FILES:-}" ]]; then
                    npm test -- "$TEST_FILES"
                else
                    npm test -- --findRelatedTests "$FILE_PATH" || true
                fi
            fi
        fi
        ;;
    py)
        echo "🧪 Running Python tests..."
        if command -v pytest &> /dev/null; then
            if [[ -n "${TEST_FILES:-}" ]]; then
                pytest $TEST_FILES
            else
                pytest -k "$(basename "$FILE_PATH" .py)" || true
            fi
        fi
        ;;
esac

exit 0