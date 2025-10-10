#!/bin/bash
# Race detection test script for xmldot concurrency safety
# This script runs tests with Go's race detector to identify data races

set -e

echo "=========================================="
echo "xmldot Concurrency Safety Testing"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
RACE_COUNT=10
BENCH_TIME="1s"

echo "Configuration:"
echo "  Race test iterations: ${RACE_COUNT}"
echo "  Benchmark time: ${BENCH_TIME}"
echo ""

# Function to run tests with race detector
run_race_tests() {
    local pattern=$1
    local description=$2

    echo "----------------------------------------"
    echo "${description}"
    echo "----------------------------------------"

    if go test -race -count=${RACE_COUNT} -run "${pattern}" -v 2>&1 | tee /tmp/race-test.log; then
        echo -e "${GREEN}✓ PASSED${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED - Race conditions detected!${NC}"
        return 1
    fi
    echo ""
}

# Track failures
FAILED_TESTS=()

# Run concurrency tests
echo "=========================================="
echo "1. Testing Concurrent Read Operations"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentReads" "Testing concurrent Get operations"; then
    FAILED_TESTS+=("Concurrent reads")
fi

if ! run_race_tests "TestConcurrentReadsWithFilters" "Testing concurrent filter queries"; then
    FAILED_TESTS+=("Concurrent reads with filters")
fi

if ! run_race_tests "TestConcurrentReadsWithModifiers" "Testing concurrent modifier usage"; then
    FAILED_TESTS+=("Concurrent reads with modifiers")
fi

echo ""
echo "=========================================="
echo "2. Testing Concurrent Write Operations"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentWrites/safe_concurrent_writes_with_mutex" "Testing synchronized writes"; then
    FAILED_TESTS+=("Synchronized writes")
fi

if ! run_race_tests "TestConcurrentDeletes/safe_concurrent_deletes_with_mutex" "Testing synchronized deletes"; then
    FAILED_TESTS+=("Synchronized deletes")
fi

echo ""
echo "=========================================="
echo "3. Testing Modifier Registry"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentModifierRegistration" "Testing concurrent modifier registration"; then
    FAILED_TESTS+=("Modifier registration")
fi

if ! run_race_tests "TestConcurrentModifierUsage" "Testing concurrent modifier usage"; then
    FAILED_TESTS+=("Modifier usage")
fi

if ! run_race_tests "TestConcurrentBuiltinModifierUsage" "Testing concurrent built-in modifiers"; then
    FAILED_TESTS+=("Built-in modifiers")
fi

echo ""
echo "=========================================="
echo "4. Testing Options Thread-Safety"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentOptionsUsage" "Testing concurrent options usage"; then
    FAILED_TESTS+=("Options usage")
fi

if ! run_race_tests "TestConcurrentGetWithDifferentOptions" "Testing GetWithOptions concurrency"; then
    FAILED_TESTS+=("GetWithOptions")
fi

echo ""
echo "=========================================="
echo "5. Testing Result Immutability"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentResultAccess" "Testing concurrent Result access"; then
    FAILED_TESTS+=("Result access")
fi

echo ""
echo "=========================================="
echo "6. Testing Mixed Operations"
echo "=========================================="
echo ""

if ! run_race_tests "TestConcurrentMixedOperations" "Testing concurrent mixed operations"; then
    FAILED_TESTS+=("Mixed operations")
fi

echo ""
echo "=========================================="
echo "7. Running Benchmarks with Race Detector"
echo "=========================================="
echo ""

echo "Running benchmarks with race detector (limited iterations)..."
if go test -race -bench=. -benchtime=${BENCH_TIME} -run=^$ 2>&1 | grep -E "(Benchmark|PASS|FAIL|WARNING)"; then
    echo -e "${GREEN}✓ Benchmarks completed${NC}"
else
    echo -e "${YELLOW}⚠ Benchmark race detection had warnings${NC}"
fi

echo ""
echo "=========================================="
echo "Race Detection Summary"
echo "=========================================="
echo ""

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
    echo ""
    echo "No race conditions detected in any test!"
    echo "The library is safe for concurrent use following documented patterns."
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo ""
    echo "Failed tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - ${test}"
    done
    echo ""
    echo "Review the output above for race condition details."
    exit 1
fi
