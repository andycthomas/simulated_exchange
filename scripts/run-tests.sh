#!/bin/bash

# run-tests.sh - Comprehensive test execution script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-output"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker is not running"
        exit 1
    fi
    print_success "Docker is available"

    # Check if Docker Compose is available
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not available"
        exit 1
    fi
    print_success "Docker Compose is available"

    # Check if we're in the correct directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        print_error "Not in a Go project directory"
        exit 1
    fi
    print_success "Go project detected"
}

# Clean up previous test runs
cleanup() {
    print_header "Cleaning Up Previous Test Runs"

    # Stop and remove containers
    if [[ -f "$DOCKER_COMPOSE_FILE" ]]; then
        docker-compose -f "$DOCKER_COMPOSE_FILE" down --volumes --remove-orphans 2>/dev/null || true
        print_success "Containers cleaned up"
    fi

    # Clean test results directory
    if [[ -d "$TEST_RESULTS_DIR" ]]; then
        rm -rf "$TEST_RESULTS_DIR"
        print_success "Previous test results cleaned"
    fi

    mkdir -p "$TEST_RESULTS_DIR"
    print_success "Test results directory created"
}

# Run unit tests locally
run_unit_tests() {
    print_header "Running Unit Tests"

    cd "$PROJECT_ROOT"

    # Run tests with coverage
    print_info "Running unit tests with coverage..."
    go test -v -race -coverprofile="$TEST_RESULTS_DIR/unit-coverage.out" -covermode=atomic ./internal/... 2>&1 | tee "$TEST_RESULTS_DIR/unit-tests.log"

    if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
        print_success "Unit tests passed"

        # Generate coverage report
        go tool cover -html="$TEST_RESULTS_DIR/unit-coverage.out" -o "$TEST_RESULTS_DIR/unit-coverage.html"
        go tool cover -func="$TEST_RESULTS_DIR/unit-coverage.out" > "$TEST_RESULTS_DIR/unit-coverage-summary.txt"

        # Display coverage summary
        echo -e "\n${BLUE}Unit Test Coverage Summary:${NC}"
        tail -1 "$TEST_RESULTS_DIR/unit-coverage-summary.txt"

    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Build test environment
build_test_environment() {
    print_header "Building Test Environment"

    cd "$PROJECT_ROOT"

    print_info "Building Docker test images..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" build --no-cache

    if [[ $? -eq 0 ]]; then
        print_success "Test environment built successfully"
    else
        print_error "Failed to build test environment"
        return 1
    fi
}

# Run E2E tests in Docker
run_e2e_tests() {
    print_header "Running E2E Tests in Docker"

    cd "$PROJECT_ROOT"

    print_info "Starting test environment..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" up --abort-on-container-exit --exit-code-from results-collector

    local exit_code=$?

    # Always collect logs
    print_info "Collecting container logs..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs app > "$TEST_RESULTS_DIR/app.log" 2>&1 || true
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs test-runner > "$TEST_RESULTS_DIR/e2e-tests.log" 2>&1 || true
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs load-tester > "$TEST_RESULTS_DIR/load-tests.log" 2>&1 || true
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs demo-tester > "$TEST_RESULTS_DIR/demo-tests.log" 2>&1 || true

    # Cleanup containers
    docker-compose -f "$DOCKER_COMPOSE_FILE" down --volumes

    if [[ $exit_code -eq 0 ]]; then
        print_success "E2E tests completed successfully"
    else
        print_error "E2E tests failed"
        return 1
    fi
}

# Run performance benchmarks
run_benchmarks() {
    print_header "Running Performance Benchmarks"

    cd "$PROJECT_ROOT"

    print_info "Running benchmarks..."
    go test -v -bench=. -benchmem -run=^$ ./test/e2e/... 2>&1 | tee "$TEST_RESULTS_DIR/benchmarks.log"

    if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
        print_success "Benchmarks completed"
    else
        print_warning "Some benchmarks may have failed"
    fi
}

# Generate test reports
generate_reports() {
    print_header "Generating Test Reports"

    # Create a comprehensive test report
    local report_file="$TEST_RESULTS_DIR/test-report.md"

    cat > "$report_file" << EOF
# Test Execution Report

Generated on: $(date)

## Test Summary

$(if [[ -f "$TEST_RESULTS_DIR/summary.txt" ]]; then cat "$TEST_RESULTS_DIR/summary.txt"; else echo "No summary available"; fi)

## Unit Test Results

$(if [[ -f "$TEST_RESULTS_DIR/unit-coverage-summary.txt" ]]; then
    echo "### Coverage Summary"
    echo "\`\`\`"
    cat "$TEST_RESULTS_DIR/unit-coverage-summary.txt"
    echo "\`\`\`"
fi)

## E2E Test Results

$(if [[ -f "$TEST_RESULTS_DIR/e2e-tests.log" ]]; then
    echo "### E2E Test Log (Last 50 lines)"
    echo "\`\`\`"
    tail -50 "$TEST_RESULTS_DIR/e2e-tests.log"
    echo "\`\`\`"
fi)

## Load Test Results

$(if [[ -f "$TEST_RESULTS_DIR/load-tests.log" ]]; then
    echo "### Load Test Log (Last 30 lines)"
    echo "\`\`\`"
    tail -30 "$TEST_RESULTS_DIR/load-tests.log"
    echo "\`\`\`"
fi)

## Demo Test Results

$(if [[ -f "$TEST_RESULTS_DIR/demo-tests.log" ]]; then
    echo "### Demo Test Log (Last 30 lines)"
    echo "\`\`\`"
    tail -30 "$TEST_RESULTS_DIR/demo-tests.log"
    echo "\`\`\`"
fi)

## Benchmark Results

$(if [[ -f "$TEST_RESULTS_DIR/benchmarks.log" ]]; then
    echo "### Benchmark Results"
    echo "\`\`\`"
    grep -A 10 -B 5 "Benchmark" "$TEST_RESULTS_DIR/benchmarks.log" || echo "No benchmark results found"
    echo "\`\`\`"
fi)

## Files Generated

- Unit Test Coverage: [unit-coverage.html](./unit-coverage.html)
- Detailed Coverage: [coverage-detailed.html](./coverage-detailed.html) (if available)
- Test Logs: Available in individual .log files

EOF

    print_success "Test report generated: $report_file"
}

# Display results summary
show_summary() {
    print_header "Test Execution Summary"

    echo -e "${BLUE}Test Results Location: ${NC}$TEST_RESULTS_DIR"
    echo -e "${BLUE}Available Reports:${NC}"

    if [[ -f "$TEST_RESULTS_DIR/test-report.md" ]]; then
        echo -e "  📄 Comprehensive Report: test-report.md"
    fi

    if [[ -f "$TEST_RESULTS_DIR/unit-coverage.html" ]]; then
        echo -e "  📊 Unit Test Coverage: unit-coverage.html"
    fi

    if [[ -f "$TEST_RESULTS_DIR/coverage-detailed.html" ]]; then
        echo -e "  📈 Detailed Coverage: coverage-detailed.html"
    fi

    echo -e "${BLUE}Log Files:${NC}"
    find "$TEST_RESULTS_DIR" -name "*.log" -type f | while read -r log_file; do
        echo -e "  📝 $(basename "$log_file")"
    done

    # Show quick coverage summary if available
    if [[ -f "$TEST_RESULTS_DIR/unit-coverage-summary.txt" ]]; then
        echo -e "\n${BLUE}Unit Test Coverage:${NC}"
        tail -1 "$TEST_RESULTS_DIR/unit-coverage-summary.txt"
    fi

    print_success "Test execution completed!"
}

# Main execution
main() {
    local run_unit=true
    local run_e2e=true
    local run_benchmarks=false
    local cleanup_only=false

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --unit-only)
                run_e2e=false
                shift
                ;;
            --e2e-only)
                run_unit=false
                shift
                ;;
            --with-benchmarks)
                run_benchmarks=true
                shift
                ;;
            --cleanup-only)
                cleanup_only=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --unit-only       Run only unit tests"
                echo "  --e2e-only        Run only E2E tests"
                echo "  --with-benchmarks Include benchmark tests"
                echo "  --cleanup-only    Only cleanup previous runs"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Start execution
    print_header "Simulated Exchange - Comprehensive Test Suite"

    check_prerequisites
    cleanup

    if [[ "$cleanup_only" == "true" ]]; then
        print_success "Cleanup completed"
        exit 0
    fi

    local overall_success=true

    # Run unit tests
    if [[ "$run_unit" == "true" ]]; then
        if ! run_unit_tests; then
            overall_success=false
        fi
    fi

    # Run E2E tests
    if [[ "$run_e2e" == "true" ]]; then
        if ! build_test_environment; then
            overall_success=false
        elif ! run_e2e_tests; then
            overall_success=false
        fi
    fi

    # Run benchmarks
    if [[ "$run_benchmarks" == "true" ]]; then
        run_benchmarks  # Don't fail overall on benchmark issues
    fi

    # Generate reports
    generate_reports
    show_summary

    if [[ "$overall_success" == "true" ]]; then
        print_success "All tests completed successfully!"
        exit 0
    else
        print_error "Some tests failed. Check the reports for details."
        exit 1
    fi
}

# Handle script interruption
trap 'echo -e "\n${YELLOW}Test execution interrupted${NC}"; cleanup; exit 130' INT TERM

# Run main function
main "$@"