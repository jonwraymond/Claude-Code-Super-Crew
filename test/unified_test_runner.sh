#!/bin/bash

# Unified Claude Code Super Crew Test Runner
# Consolidates functionality from multiple test scripts
# Usage: ./unified_test_runner.sh [--mode essential|comprehensive|full] [--format json|markdown|console]

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CREW_BINARY="$PROJECT_ROOT/crew"
TEST_RESULTS_DIR="$SCRIPT_DIR/results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

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
SKIPPED_TESTS=0

# Configuration defaults
MODE="essential"
FORMAT="console"
VERBOSE=false
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --mode essential|comprehensive|full    Test mode (default: essential)"
            echo "  --format console|json|markdown          Output format (default: console)"
            echo "  --verbose, -v                          Verbose output"
            echo "  --dry-run                              Show what would be run"
            echo "  --help, -h                             Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Initialize test environment
init_test_env() {
    echo -e "${BLUE}Initializing unified test environment...${NC}"
    
    # Create results directory
    mkdir -p "$TEST_RESULTS_DIR"
    
    # Check crew binary
    if [[ ! -f "$CREW_BINARY" ]]; then
        echo -e "${RED}Error: Crew binary not found at $CREW_BINARY${NC}"
        echo "Please build the project first: make build"
        exit 1
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        echo "Project root: $PROJECT_ROOT"
        echo "Test results: $TEST_RESULTS_DIR"
        echo "Crew binary: $CREW_BINARY"
        echo "Test mode: $MODE"
        echo "Output format: $FORMAT"
    fi
}

# Execute a single test
run_test() {
    local test_id="$1"
    local test_cmd="$2"
    local expected_exit_code="${3:-0}"
    local test_description="$4"
    local test_category="${5:-General}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BLUE}[DRY RUN] Test $test_id:${NC} $test_description"
        echo "Command: $test_cmd"
        return 0
    fi
    
    if [[ "$VERBOSE" == "true" ]] || [[ "$FORMAT" == "console" ]]; then
        echo -e "${BLUE}[Test $test_id]${NC} $test_description"
    fi
    
    # Create test directory
    local test_dir="$TEST_RESULTS_DIR/test_${test_id}"
    mkdir -p "$test_dir"
    
    # Execute test
    local start_time=$(date +%s)
    local output_file="$test_dir/output.log"
    local exit_code_file="$test_dir/exit_code"
    
    # Replace crew with full path and add test directories
    local actual_cmd=$(echo "$test_cmd" | sed "s|^crew |\"$CREW_BINARY\" |g")
    if [[ "$actual_cmd" == *"install"* ]] && [[ "$actual_cmd" != *"--install-dir"* ]]; then
        actual_cmd="$actual_cmd --install-dir \"$HOME/crew-test-$test_id\""
    fi
    
    set +e
    (cd "$PROJECT_ROOT" && timeout 60 bash -c "$actual_cmd" > "$output_file" 2>&1)
    local actual_exit_code=$?
    set -e
    
    echo "$actual_exit_code" > "$exit_code_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Analyze results
    local status="UNKNOWN"
    if [[ "$actual_exit_code" == "$expected_exit_code" ]]; then
        status="PASS"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        if [[ "$FORMAT" == "console" ]]; then
            echo -e "${GREEN}✓ PASS${NC} (Exit: $actual_exit_code, Duration: ${duration}s)"
        fi
    else
        status="FAIL"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        if [[ "$FORMAT" == "console" ]]; then
            echo -e "${RED}✗ FAIL${NC} (Expected: $expected_exit_code, Got: $actual_exit_code, Duration: ${duration}s)"
            
            # Show error details in verbose mode
            if [[ "$VERBOSE" == "true" ]] && [[ -f "$output_file" ]] && [[ -s "$output_file" ]]; then
                echo -e "${YELLOW}Error Output:${NC}"
                head -3 "$output_file" | sed 's/^/  /'
            fi
        fi
    fi
    
    # Store test result for reporting
    echo "$test_id|$test_description|$test_category|$actual_cmd|$expected_exit_code|$actual_exit_code|$status|$duration" >> "$TEST_RESULTS_DIR/test_results.csv"
    
    # Cleanup test directories (keep logs for failed tests)
    if [[ "$status" == "PASS" ]]; then
        rm -rf "$HOME/crew-test-$test_id" 2>/dev/null || true
    fi
    
    if [[ "$FORMAT" == "console" ]]; then
        echo ""
    fi
}

# Essential test suite (13 core tests)
run_essential_tests() {
    echo -e "${YELLOW}=== Essential CLI Tests (13 tests) ===${NC}"
    
    # Basic functionality
    run_test "E01" "crew install --help" 0 "Help command works" "Basic"
    run_test "E02" "crew version" 0 "Version command works" "Basic"
    run_test "E03" "crew install --dry-run --minimal --yes" 0 "Dry run minimal installation" "Installation"
    run_test "E04" "crew install --minimal --yes" 0 "Minimal installation" "Installation"
    run_test "E05" "crew install --components core --yes" 0 "Core component installation" "Installation"
    
    # Update/uninstall tests (require setup)
    local perm_test_dir="$HOME/crew-test-permanent"
    if [[ "$DRY_RUN" != "true" ]]; then
        # Ensure directory exists
        mkdir -p "$perm_test_dir"
        # Try to install minimal setup for update tests
        "$CREW_BINARY" install --install-dir "$perm_test_dir" --minimal --yes >/dev/null 2>&1
        install_result=$?
        if [[ $install_result -ne 0 ]]; then
            echo -e "${YELLOW}Warning: Could not create test installation for update tests${NC}"
        fi
    fi
    
    run_test "E06" "crew update --check --install-dir $perm_test_dir" 0 "Check for updates" "Update"
    run_test "E07" "crew update --dry-run --install-dir $perm_test_dir" 0 "Update dry run" "Update"
    run_test "E08" "crew uninstall --dry-run --install-dir $perm_test_dir" 0 "Uninstall dry run" "Uninstall"
    run_test "E09" "crew uninstall --install-dir $perm_test_dir --yes" 0 "Basic uninstall" "Uninstall"
    
    # Error conditions (should fail)
    run_test "E10" "crew install --components invalid --yes" 1 "Invalid component (should fail)" "Error"
    run_test "E11" "crew install --minimal --quick --yes" 1 "Conflicting flags (should fail)" "Error"
    
    # Claude integration
    run_test "E12" "crew claude --help" 0 "Claude help command" "Claude"
    run_test "E13" "crew claude --shell bash" 0 "Generate bash completion" "Claude"
}

# Comprehensive test suite (additional tests)
run_comprehensive_tests() {
    echo -e "${YELLOW}=== Comprehensive CLI Tests (additional) ===${NC}"
    
    # Additional installation tests
    run_test "C01" "crew install --quick --yes" 0 "Quick installation" "Installation"
    run_test "C02" "crew install --verbose --yes" 0 "Verbose installation" "Installation"
    run_test "C03" "crew install --components core,commands --yes" 0 "Multiple components" "Installation"
    run_test "C04" "crew install --profile developer --yes" 0 "Developer profile" "Installation"
    
    # Claude integration comprehensive
    run_test "C05" "crew claude --status" 0 "Claude status check" "Claude"
    run_test "C06" "crew claude --list" 0 "List Claude commands" "Claude"
    run_test "C07" "crew claude --shell zsh" 0 "ZSH completion" "Claude" 
    run_test "C08" "crew claude --shell fish" 0 "Fish completion" "Claude"
    
    # Additional error conditions
    run_test "C09" "crew install --profile invalid --yes" 1 "Invalid profile" "Error"
    run_test "C10" "crew claude --shell invalid" 1 "Invalid shell" "Error"
}

# Full test suite (Go tests + shell tests)
run_go_tests() {
    echo -e "${YELLOW}=== Go Integration Tests ===${NC}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BLUE}[DRY RUN]${NC} Would run: go test ./test/integration/... -v"
        echo -e "${BLUE}[DRY RUN]${NC} Would run: go test ./test/... -short"
        return 0
    fi
    
    # Run integration tests
    echo -e "${BLUE}Running Go integration tests...${NC}"
    local go_output="$TEST_RESULTS_DIR/go_integration_tests.log"
    
    set +e
    (cd "$PROJECT_ROOT" && go test ./test/integration/... -v > "$go_output" 2>&1)
    local go_exit_code=$?
    set -e
    
    if [[ "$go_exit_code" == "0" ]]; then
        echo -e "${GREEN}✓ Go integration tests PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ Go integration tests FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        if [[ "$VERBOSE" == "true" ]]; then
            echo -e "${YELLOW}Go test output:${NC}"
            tail -10 "$go_output" | sed 's/^/  /'
        fi
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Generate test report
generate_report() {
    local success_rate=0
    if [[ "$TOTAL_TESTS" -gt 0 ]]; then
        success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    case "$FORMAT" in
        "json")
            generate_json_report "$success_rate"
            ;;
        "markdown")
            generate_markdown_report "$success_rate"
            ;;
        "console"|*)
            generate_console_report "$success_rate"
            ;;
    esac
}

generate_console_report() {
    local success_rate="$1"
    
    echo -e "${BLUE}==============================${NC}"
    echo -e "${BLUE}Test Execution Complete${NC}"
    echo -e "${BLUE}==============================${NC}"
    echo -e "Total Tests: ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
    echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
    echo -e "${YELLOW}Skipped: ${SKIPPED_TESTS}${NC}"
    echo -e "Success Rate: ${success_rate}%"
    echo -e "Results Directory: ${TEST_RESULTS_DIR}"
    echo ""
    
    if [[ "$FAILED_TESTS" -gt 0 ]]; then
        echo -e "${YELLOW}Failed tests details available in: $TEST_RESULTS_DIR/${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed successfully!${NC}"
        exit 0
    fi
}

generate_json_report() {
    local success_rate="$1"
    local json_file="$TEST_RESULTS_DIR/test_report_${TIMESTAMP}.json"
    
    cat > "$json_file" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "mode": "$MODE",
    "summary": {
        "total": $TOTAL_TESTS,
        "passed": $PASSED_TESTS,
        "failed": $FAILED_TESTS,
        "skipped": $SKIPPED_TESTS,
        "success_rate": $success_rate
    },
    "environment": {
        "crew_binary": "$CREW_BINARY",
        "project_root": "$PROJECT_ROOT"
    }
}
EOF
    
    echo "JSON report generated: $json_file"
}

generate_markdown_report() {
    local success_rate="$1"
    local md_file="$TEST_RESULTS_DIR/test_report_${TIMESTAMP}.md"
    
    cat > "$md_file" << EOF
# Unified Test Report

**Generated**: $(date)  
**Mode**: $MODE  
**Success Rate**: ${success_rate}%

## Summary
- **Total Tests**: $TOTAL_TESTS
- **Passed**: $PASSED_TESTS  
- **Failed**: $FAILED_TESTS
- **Skipped**: $SKIPPED_TESTS

## Environment
- **Crew Binary**: $CREW_BINARY
- **Project Root**: $PROJECT_ROOT
- **Results Directory**: $TEST_RESULTS_DIR

## Test Results

EOF

    # Add detailed results if CSV exists
    if [[ -f "$TEST_RESULTS_DIR/test_results.csv" ]]; then
        echo "| Test ID | Description | Status | Duration |" >> "$md_file"
        echo "|---------|-------------|--------|----------|" >> "$md_file"
        
        while IFS='|' read -r test_id desc category cmd expected actual status duration; do
            echo "| $test_id | $desc | $status | ${duration}s |" >> "$md_file"
        done < "$TEST_RESULTS_DIR/test_results.csv"
    fi
    
    echo "" >> "$md_file"
    echo "---" >> "$md_file"
    echo "Generated by unified_test_runner.sh" >> "$md_file"
    
    echo "Markdown report generated: $md_file"
}

# Main execution
main() {
    echo -e "${BLUE}==============================${NC}"
    echo -e "${BLUE}Unified Test Runner${NC}"
    echo -e "${BLUE}==============================${NC}"
    echo ""
    
    # Initialize
    init_test_env
    
    # Create CSV header
    echo "test_id|description|category|command|expected|actual|status|duration" > "$TEST_RESULTS_DIR/test_results.csv"
    
    # Run tests based on mode
    case "$MODE" in
        "essential")
            run_essential_tests
            ;;
        "comprehensive")
            run_essential_tests
            run_comprehensive_tests
            ;;
        "full")
            run_essential_tests
            run_comprehensive_tests
            run_go_tests
            ;;
        *)
            echo -e "${RED}Invalid mode: $MODE${NC}"
            echo "Valid modes: essential, comprehensive, full"
            exit 1
            ;;
    esac
    
    # Generate report
    generate_report
}

# Execute main function
main "$@"