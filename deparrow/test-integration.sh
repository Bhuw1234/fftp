#!/bin/bash

# DEparrow Integration Test Script
# Comprehensive test runner for all DEparrow components

set -e

echo "========================================="
echo "DEparrow Platform Integration Tests"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗${NC} $2"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

print_skip() {
    echo -e "${YELLOW}⊘${NC} $1"
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Function to check if a command exists
check_command() {
    if command -v $1 &> /dev/null; then
        print_status 0 "$1 is installed"
    else
        print_status 1 "$1 is not installed"
    fi
}

# Function to run Go tests
run_go_tests() {
    local package=$1
    local pattern=$2
    local description=$3
    
    echo -e "${CYAN}Running: ${description}${NC}"
    
    if go test -v -tags=integration -count=1 -timeout=5m "./${package}" -run "${pattern}" 2>&1; then
        print_status 0 "${description}"
        return 0
    else
        print_status 1 "${description}"
        return 1
    fi
}

# Function to run tests with coverage
run_tests_with_coverage() {
    local package=$1
    local output_file=$2
    
    echo -e "${CYAN}Running tests with coverage...${NC}"
    
    go test -v -tags=integration -coverprofile="${output_file}" -covermode=atomic "./${package}" 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Coverage report saved to: ${output_file}${NC}"
        go tool cover -func="${output_file}" | tail -n 1
        return 0
    else
        return 1
    fi
}

# Parse arguments
RUN_COVERAGE=false
PARALLEL=false
VERBOSE=false
TEST_PATTERN=""
PACKAGES=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage|-c)
            RUN_COVERAGE=true
            shift
            ;;
        --parallel|-p)
            PARALLEL=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --pattern|-r)
            TEST_PATTERN="$2"
            shift 2
            ;;
        --package|-P)
            PACKAGES="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --coverage, -c     Run tests with coverage report"
            echo "  --parallel, -p     Run tests in parallel"
            echo "  --verbose, -v      Verbose output"
            echo "  --pattern, -r      Test pattern to run (regex)"
            echo "  --package, -P      Package to test (default: all)"
            echo "  --help, -h         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                          # Run all tests"
            echo "  $0 --coverage               # Run with coverage"
            echo "  $0 -p -c                    # Parallel with coverage"
            echo "  $0 -r 'TestLogin'           # Run tests matching 'TestLogin'"
            echo "  $0 -P './test-integration'  # Run specific package"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}=== Prerequisite Checks ===${NC}"
echo ""

# Check required commands
check_command go
check_command docker
check_command python3

echo ""
echo -e "${BLUE}=== File Structure Verification ===${NC}"
echo ""

# Verify test files exist
test_files=(
    "test-integration/picoclaw_integration_test.go"
    "test-integration/e2e_workflow_test.go"
    "test-integration/api_test.go"
    "test-integration/gui_e2e_test.go"
    "test-integration/testutil/mock_server.go"
    "test-integration/testutil/fixtures.go"
    "test-integration/testutil/helpers.go"
)

missing_files=0
for file in "${test_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} Found: $file"
    else
        echo -e "${RED}✗${NC} Missing: $file"
        missing_files=$((missing_files + 1))
    fi
done

if [ $missing_files -gt 0 ]; then
    echo -e "${RED}Error: $missing_files test files are missing${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}=== Go Module Verification ===${NC}"
echo ""

# Verify Go modules
echo "Checking Go module dependencies..."
if go mod download; then
    echo -e "${GREEN}✓${NC} Go modules downloaded"
else
    echo -e "${RED}✗${NC} Failed to download Go modules"
    exit 1
fi

# Verify test dependencies compile
echo "Verifying test compilation..."
if go build -tags=integration ./test-integration/...; then
    echo -e "${GREEN}✓${NC} Test code compiles"
else
    echo -e "${RED}✗${NC} Test code has compilation errors"
    exit 1
fi

echo ""
echo -e "${BLUE}=== Integration Test Suites ===${NC}"
echo ""

# Define test packages
if [ -z "$PACKAGES" ]; then
    PACKAGES="./test-integration/..."
fi

# Run tests
if [ "$RUN_COVERAGE" = true ]; then
    echo -e "${CYAN}Running tests with coverage...${NC}"
    
    coverage_file="coverage.out"
    
    if [ "$PARALLEL" = true ]; then
        # Parallel test execution with coverage (requires gocovmerge)
        echo -e "${YELLOW}Note: Parallel coverage requires gocovmerge for merging reports${NC}"
        
        # Run each test suite separately
        packages=(
            "test-integration"
        )
        
        pids=()
        for i in "${!packages[@]}"; do
            pkg="${packages[$i]}"
            cov_file="coverage.${i}.out"
            go test -tags=integration -coverprofile="${cov_file}" -covermode=atomic "./${pkg}" &
            pids+=($!)
        done
        
        # Wait for all tests
        for pid in "${pids[@]}"; do
            wait $pid
        done
        
        # Merge coverage files
        if command -v gocovmerge &> /dev/null; then
            gocovmerge coverage.*.out > coverage.out
            rm coverage.*.out
        else
            echo -e "${YELLOW}gocovmerge not found, using last coverage file${NC}"
            mv coverage.0.out coverage.out
        fi
    else
        go test -tags=integration -coverprofile="${coverage_file}" -covermode=atomic ${PACKAGES}
    fi
    
    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}=== Coverage Summary ===${NC}"
        go tool cover -func="${coverage_file}"
        
        # Generate HTML report
        go tool cover -html="${coverage_file}" -o coverage.html
        echo -e "${GREEN}HTML coverage report: coverage.html${NC}"
    fi
else
    # Run without coverage
    if [ "$PARALLEL" = true ]; then
        echo -e "${CYAN}Running tests in parallel...${NC}"
        
        if [ -n "$TEST_PATTERN" ]; then
            go test -v -tags=integration -count=1 -parallel 4 ${PACKAGES} -run "${TEST_PATTERN}"
        else
            go test -v -tags=integration -count=1 -parallel 4 ${PACKAGES}
        fi
    else
        echo -e "${CYAN}Running tests sequentially...${NC}"
        
        if [ -n "$TEST_PATTERN" ]; then
            go test -v -tags=integration -count=1 ${PACKAGES} -run "${TEST_PATTERN}"
        else
            go test -v -tags=integration -count=1 ${PACKAGES}
        fi
    fi
fi

TEST_EXIT_CODE=$?

echo ""
echo -e "${BLUE}=== Test Suite Summary ===${NC}"
echo ""

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}All integration tests passed!${NC}"
    echo -e "${GREEN}=========================================${NC}"
else
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}Some integration tests failed!${NC}"
    echo -e "${RED}=========================================${NC}"
fi

echo ""
echo "Test execution details:"
echo "  - Test files: ${#test_files[@]} checked"
echo "  - Build tags: integration"
echo "  - Parallel: $PARALLEL"
echo "  - Coverage: $RUN_COVERAGE"
if [ -n "$TEST_PATTERN" ]; then
    echo "  - Pattern: $TEST_PATTERN"
fi

echo ""
echo "Next steps:"
echo "1. Review any failed tests above"
echo "2. Check logs for error details"
echo "3. Run individual test suites for debugging:"
echo "   go test -v -tags=integration -run TestPicoClawIntegrationSuite ./test-integration"
echo "   go test -v -tags=integration -run TestE2EWorkflowSuite ./test-integration"
echo "   go test -v -tags=integration -run TestAPICompatibilitySuite ./test-integration"
echo "   go test -v -tags=integration -run TestGUIE2ESuite ./test-integration"
echo ""

exit $TEST_EXIT_CODE
