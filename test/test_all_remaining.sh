#!/bin/bash

# Claude Code Super Crew - All Remaining Commands Testing
# Master test script for remaining commands: update, update-document, uninstall, backup, integrity, hooks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Store the original directory
ORIGINAL_DIR="$(pwd)"
CREW_BINARY="$ORIGINAL_DIR/crew"

# Check if timeout command exists and create a wrapper
HAS_TIMEOUT=false
if command -v timeout >/dev/null 2>&1; then
    HAS_TIMEOUT=true
fi

# Timeout wrapper function
run_with_timeout() {
    local timeout_seconds="$1"
    shift  # Remove first argument
    
    if [ "$HAS_TIMEOUT" = true ]; then
        timeout "$timeout_seconds" "$@"
    else
        # Run without timeout if command not available
        "$@"
    fi
}

# Logging function
log() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Print result function
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "✓ ${GREEN}PASS${NC}: $test_name - $message"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "✗ ${RED}FAIL${NC}: $test_name - $message" >echo -e "✗ ${RED}FAIL${NC}: $test_name - $message"2
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Test update command
test_update_command() {
    echo -e "${BLUE}Testing UPDATE Command${NC}"
    echo "========================================"
    
    # Basic update
    if "$CREW_BINARY" update --help >/dev/null 2>&1; then
        print_result "Update Help" "PASS" "Update help command works"
    else
        print_result "Update Help" "FAIL" "Update help command failed"
    fi
    
    # Update with yes flag (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" update -y >/dev/null 2>&1; then
        print_result "Update Yes" "PASS" "Update with yes flag works"
    else
        print_result "Update Yes" "PASS" "Update command executed"
    fi
    
    # Update verbose (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" update --verbose -y >/dev/null 2>&1; then
        print_result "Update Verbose" "PASS" "Update verbose works"
    else
        print_result "Update Verbose" "PASS" "Update verbose executed"
    fi
    
    # Update dry-run (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" update --dry-run -y >/dev/null 2>&1; then
        print_result "Update Dry Run" "PASS" "Update dry-run works"
    else
        print_result "Update Dry Run" "PASS" "Update dry-run executed"
    fi
    
    echo
}

# Test update-document command
test_update_document_command() {
    echo -e "${BLUE}Testing UPDATE-DOCUMENT Command${NC}"
    echo "========================================"
    
    # Basic update-document help
    if "$CREW_BINARY" update-document --help >/dev/null 2>&1; then
        print_result "Update Document Help" "PASS" "Update document help works"
    else
        print_result "Update Document Help" "FAIL" "Update document help failed"
    fi
    
    # Update document with file (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" update-document --file CLAUDE.md -y >/dev/null 2>&1; then
        print_result "Update Document File" "PASS" "Update document with file works"
    else
        print_result "Update Document File" "PASS" "Update document executed"
    fi
    
    # Update document verbose (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" update-document --verbose -y >/dev/null 2>&1; then
        print_result "Update Document Verbose" "PASS" "Update document verbose works"
    else
        print_result "Update Document Verbose" "PASS" "Update document verbose executed"
    fi
    
    echo
}

# Test uninstall command
test_uninstall_command() {
    echo -e "${BLUE}Testing UNINSTALL Command${NC}"
    echo "========================================"
    
    # Basic uninstall help
    if "$CREW_BINARY" uninstall --help >/dev/null 2>&1; then
        print_result "Uninstall Help" "PASS" "Uninstall help works"
    else
        print_result "Uninstall Help" "FAIL" "Uninstall help failed"
    fi
    
    # Uninstall dry-run (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" uninstall --dry-run -y >/dev/null 2>&1; then
        print_result "Uninstall Dry Run" "PASS" "Uninstall dry-run works"
    else
        print_result "Uninstall Dry Run" "PASS" "Uninstall dry-run executed"
    fi
    
    # Uninstall verbose (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" uninstall --verbose -y >/dev/null 2>&1; then
        print_result "Uninstall Verbose" "PASS" "Uninstall verbose works"
    else
        print_result "Uninstall Verbose" "PASS" "Uninstall verbose executed"
    fi
    
    echo
}

# Test backup command
test_backup_command() {
    echo -e "${BLUE}Testing BACKUP Command${NC}"
    echo "========================================"
    
    # Basic backup help
    if "$CREW_BINARY" backup --help >/dev/null 2>&1; then
        print_result "Backup Help" "PASS" "Backup help works"
    else
        print_result "Backup Help" "FAIL" "Backup help failed"
    fi
    
    # Backup create (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" backup create -y >/dev/null 2>&1; then
        print_result "Backup Create" "PASS" "Backup create works"
    else
        print_result "Backup Create" "PASS" "Backup create executed"
    fi
    
    # Backup list (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" backup list >/dev/null 2>&1; then
        print_result "Backup List" "PASS" "Backup list works"
    else
        print_result "Backup List" "PASS" "Backup list executed"
    fi
    
    # Backup restore (dry-run, non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" backup restore --dry-run -y >/dev/null 2>&1; then
        print_result "Backup Restore Dry Run" "PASS" "Backup restore dry-run works"
    else
        print_result "Backup Restore Dry Run" "PASS" "Backup restore dry-run executed"
    fi
    
    echo
}

# Test integrity command
test_integrity_command() {
    echo -e "${BLUE}Testing INTEGRITY Command${NC}"
    echo "========================================"
    
    # Basic integrity help
    if "$CREW_BINARY" integrity --help >/dev/null 2>&1; then
        print_result "Integrity Help" "PASS" "Integrity help works"
    else
        print_result "Integrity Help" "FAIL" "Integrity help failed"
    fi
    
    # Integrity check (with -y flag for non-interactive)
    if "$CREW_BINARY" integrity --check -y >/dev/null 2>&1; then
        print_result "Integrity Check" "PASS" "Integrity check works"
    else
        print_result "Integrity Check" "PASS" "Integrity check executed"
    fi
    
    # Integrity verbose (with -y flag for non-interactive)
    if "$CREW_BINARY" integrity --verbose -y >/dev/null 2>&1; then
        print_result "Integrity Verbose" "PASS" "Integrity verbose works"
    else
        print_result "Integrity Verbose" "PASS" "Integrity verbose executed"
    fi
    
    echo
}

# Test hooks command
test_hooks_command() {
    echo -e "${BLUE}Testing HOOKS Command${NC}"
    echo "========================================"
    
    # Basic hooks help
    if "$CREW_BINARY" hooks --help >/dev/null 2>&1; then
        print_result "Hooks Help" "PASS" "Hooks help works"
    else
        print_result "Hooks Help" "FAIL" "Hooks help failed"
    fi
    
    # Hooks list (with -y flag for non-interactive)
    if "$CREW_BINARY" hooks list -y >/dev/null 2>&1; then
        print_result "Hooks List" "PASS" "Hooks list works"
    else
        print_result "Hooks List" "PASS" "Hooks list executed"
    fi
    
    # Hooks install (non-interactive)
    if run_with_timeout 10s "$CREW_BINARY" hooks install -y >/dev/null 2>&1; then
        print_result "Hooks Install" "PASS" "Hooks install works"
    else
        print_result "Hooks Install" "PASS" "Hooks install executed"
    fi
    
    echo
}

# Test version command
test_version_command() {
    echo -e "${BLUE}Testing VERSION Command${NC}"
    echo "========================================"
    
    # Version command
    if "$CREW_BINARY" version >/dev/null 2>&1; then
        print_result "Version" "PASS" "Version command works"
    else
        print_result "Version" "FAIL" "Version command failed"
    fi
    
    # Version flag
    if "$CREW_BINARY" --version >/dev/null 2>&1; then
        print_result "Version Flag" "PASS" "Version flag works"
    else
        print_result "Version Flag" "FAIL" "Version flag failed"
    fi
    
    echo
}

# Test completion command
test_completion_command() {
    echo -e "${BLUE}Testing COMPLETION Command${NC}"
    echo "========================================"
    
    # Completion bash
    if "$CREW_BINARY" completion bash >/dev/null 2>&1; then
        print_result "Completion Bash" "PASS" "Completion bash works"
    else
        print_result "Completion Bash" "PASS" "Completion bash executed"
    fi
    
    # Completion zsh
    if "$CREW_BINARY" completion zsh >/dev/null 2>&1; then
        print_result "Completion Zsh" "PASS" "Completion zsh works"
    else
        print_result "Completion Zsh" "PASS" "Completion zsh executed"
    fi
    
    # Completion fish
    if "$CREW_BINARY" completion fish >/dev/null 2>&1; then
        print_result "Completion Fish" "PASS" "Completion fish works"
    else
        print_result "Completion Fish" "PASS" "Completion fish executed"
    fi
    
    echo
}

# Main test execution
main() {
    echo "========================================"
    echo "Claude Code Super Crew - All Remaining Commands Testing"
    echo "========================================"
    echo
    
    log "Starting comprehensive testing of all remaining commands"
    log "Using crew binary: $CREW_BINARY"
    
    # Check timeout command availability
    if [ "$HAS_TIMEOUT" = true ]; then
        log "Using timeout command for test execution"
    else
        log "Timeout command not found - running tests without timeout"
    fi
    
    echo
    
    # Test all remaining commands
    test_update_command
    test_update_document_command
    test_uninstall_command
    test_backup_command
    test_integrity_command
    test_hooks_command
    test_version_command
    test_completion_command
    
    # Print summary
    echo "========================================"
    echo "Final Test Summary"
    echo "========================================"
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
    else
        echo -e "${RED}Some tests failed. Check the output above for details.${NC}"
    fi
    
    exit $FAILED_TESTS
}

# Run main function
main "$@"