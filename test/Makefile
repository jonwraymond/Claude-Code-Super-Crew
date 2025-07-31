# Claude Code Super Crew Test Suite Makefile
# Author: Jonathan Raymond
# Date: July 29, 2025
# Purpose: Comprehensive testing infrastructure for crew commands

# Go environment variables
GO := go
GOFLAGS := -v
GOTEST := $(GO) test $(GOFLAGS)
GOBUILD := $(GO) build $(GOFLAGS)

# Project paths
PROJECT_ROOT := $(shell dirname $(shell pwd))
INTERNAL_CLI := $(PROJECT_ROOT)/internal/cli
INTERNAL_CORE := $(PROJECT_ROOT)/internal/core
INTERNAL_ORCHESTRATOR := $(PROJECT_ROOT)/internal/orchestrator
PKG_BACKUP := $(PROJECT_ROOT)/pkg/backup
PKG_LOGGER := $(PROJECT_ROOT)/pkg/logger

# Test configuration
TEST_TIMEOUT := 30s
COVERAGE_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
TEST_REPORT := test_report_$(shell date +%Y%m%d_%H%M%S).txt

# Color codes for output
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: all help clean test test-all test-install test-backup test-cli test-core test-orchestrator test-pkg test-integration test-unit coverage coverage-html lint vet fmt test-verbose test-short test-race benchmark

# Default target
all: test-all

# Help target - shows available commands
help:
	@echo "$(BLUE)Claude Code Super Crew Test Suite$(NC)"
	@echo "=================================="
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  make test-all         - Run all tests (unit + integration)"
	@echo "  make test-unit        - Run only unit tests"
	@echo "  make test-integration - Run only integration tests"
	@echo ""
	@echo "$(GREEN)Component-specific tests:$(NC)"
	@echo "  make test-install     - Test install command"
	@echo "  make test-backup      - Test backup command"
	@echo "  make test-cli         - Test all CLI commands"
	@echo "  make test-core        - Test core components"
	@echo "  make test-orchestrator- Test orchestrator components"
	@echo "  make test-pkg         - Test package utilities"
	@echo ""
	@echo "$(GREEN)Testing options:$(NC)"
	@echo "  make test-verbose     - Run tests with verbose output"
	@echo "  make test-short       - Run only short tests"
	@echo "  make test-race        - Run tests with race detector"
	@echo "  make coverage         - Generate test coverage report"
	@echo "  make coverage-html    - Generate HTML coverage report"
	@echo "  make benchmark        - Run benchmark tests"
	@echo ""
	@echo "$(GREEN)Code quality:$(NC)"
	@echo "  make lint             - Run golangci-lint"
	@echo "  make vet              - Run go vet"
	@echo "  make fmt              - Format code with gofmt"
	@echo ""
	@echo "$(GREEN)Maintenance:$(NC)"
	@echo "  make clean            - Clean test artifacts"

# Clean test artifacts
clean:
	@echo "$(YELLOW)Cleaning test artifacts...$(NC)"
	@rm -rf $(COVERAGE_DIR)
	@rm -f test_report_*.txt
	@rm -f test_report_*.md
	@rm -f test_report_*.json
	@rm -rf /tmp/install-test-*
	@rm -rf /tmp/backup-test-*
	@rm -rf /tmp/crew_test_*
	@rm -rf results/
	@rm -rf test_backup_*/
	@echo "$(GREEN)✓ Clean complete$(NC)"

# Run all tests
test-all: clean
	@echo "$(BLUE)Running all tests...$(NC)"
	@echo "===================="
	@$(GOTEST) -timeout $(TEST_TIMEOUT) ./... 2>&1 | tee $(TEST_REPORT)
	@if grep -q "FAIL" $(TEST_REPORT); then \
		echo "$(RED)✗ Some tests failed$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✓ All tests passed$(NC)"; \
	fi

# Run unit tests only
test-unit: clean
	@echo "$(BLUE)Running unit tests...$(NC)"
	@$(GOTEST) -short -timeout $(TEST_TIMEOUT) ./...

# Run integration tests only
test-integration: clean
	@echo "$(BLUE)Running integration tests...$(NC)"
	@$(GOTEST) -run Integration -timeout $(TEST_TIMEOUT) ./...

# Test install command specifically
test-install:
	@echo "$(BLUE)Testing install command...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(INTERNAL_CLI) -run TestInstall

# Test backup command specifically
test-backup:
	@echo "$(BLUE)Testing backup command...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(INTERNAL_CLI) -run TestBackup

# Test all CLI commands
test-cli:
	@echo "$(BLUE)Testing CLI commands...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(INTERNAL_CLI)/...

# Test core components
test-core:
	@echo "$(BLUE)Testing core components...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(INTERNAL_CORE)/...

# Test orchestrator components
test-orchestrator:
	@echo "$(BLUE)Testing orchestrator components...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(INTERNAL_ORCHESTRATOR)/...

# Test package utilities
test-pkg:
	@echo "$(BLUE)Testing package utilities...$(NC)"
	@$(GOTEST) -timeout $(TEST_TIMEOUT) $(PROJECT_ROOT)/pkg/...

# Run tests with verbose output
test-verbose:
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	@$(GO) test -v -timeout $(TEST_TIMEOUT) ./...

# Run only short tests
test-short:
	@echo "$(BLUE)Running short tests...$(NC)"
	@$(GO) test -short -timeout $(TEST_TIMEOUT) ./...

# Run tests with race detector
test-race:
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	@$(GO) test -race -timeout $(TEST_TIMEOUT) ./...

# Generate test coverage
coverage: clean
	@echo "$(BLUE)Generating test coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GO) tool cover -func=$(COVERAGE_FILE)

# Generate HTML coverage report
coverage-html: coverage
	@echo "$(BLUE)Generating HTML coverage report...$(NC)"
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Coverage report generated at $(COVERAGE_DIR)/coverage.html$(NC)"

# Run benchmarks
benchmark:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@$(GO) test -bench=. -benchmem ./...

# Run linter (requires golangci-lint)
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with:$(NC)"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	@$(GO) vet ./...

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

# Continuous Integration target
ci: clean fmt vet lint test-all coverage
	@echo "$(GREEN)✓ CI checks complete$(NC)"

# Watch mode for development (requires entr)
watch:
	@echo "$(BLUE)Watching for changes...$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make test-unit; \
	else \
		echo "$(YELLOW)entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)$(NC)"; \
	fi

# Test specific function or test case
test-func:
	@if [ -z "$(FUNC)" ]; then \
		echo "$(RED)Error: FUNC not specified. Usage: make test-func FUNC=TestInstallCommand$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running test function: $(FUNC)$(NC)"
	@$(GOTEST) -run $(FUNC) ./...

# Generate test report in markdown format
test-report: test-all
	@echo "$(BLUE)Generating test report...$(NC)"
	@./unified_test_runner.sh --mode comprehensive --format markdown

# Run unified test runner
test-scripts:
	@echo "$(BLUE)Running unified test suite...$(NC)"
	@./unified_test_runner.sh --mode essential --format console

test-scripts-comprehensive:
	@echo "$(BLUE)Running comprehensive test suite...$(NC)"
	@./unified_test_runner.sh --mode comprehensive --format console

test-scripts-full:
	@echo "$(BLUE)Running full test suite...$(NC)"
	@./unified_test_runner.sh --mode full --format console

# Dependency check
deps:
	@echo "$(BLUE)Checking dependencies...$(NC)"
	@$(GO) mod verify
	@$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies verified$(NC)"

# Quick test - runs fast tests for rapid feedback
quick: test-short

# Full test - runs everything including benchmarks
full: ci benchmark

# Test matrix - runs tests with different Go versions (requires Go version manager)
test-matrix:
	@echo "$(BLUE)Running test matrix...$(NC)"
	@for version in 1.21 1.22; do \
		echo "Testing with Go $$version..."; \
		go$$version test ./... || exit 1; \
	done

# Debug mode - runs tests with delve debugger support
debug:
	@echo "$(BLUE)Running tests in debug mode...$(NC)"
	@$(GO) test -gcflags="all=-N -l" -timeout $(TEST_TIMEOUT) ./...

# Profile tests - generates CPU and memory profiles
profile:
	@echo "$(BLUE)Running tests with profiling...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -cpuprofile=$(COVERAGE_DIR)/cpu.prof -memprofile=$(COVERAGE_DIR)/mem.prof -timeout $(TEST_TIMEOUT) ./...
	@echo "$(GREEN)✓ Profiles generated in $(COVERAGE_DIR)/$(NC)"