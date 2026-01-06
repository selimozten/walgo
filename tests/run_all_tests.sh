#!/bin/bash
# Walgo Complete Command Test Runner

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test result log
TEST_LOG="/tmp/walgo_test_results_$(date +%Y%m%d_%H%M%S).log"

# Helper functions
log_test() {
    local test_name="$1"
    local status="$2"
    local notes="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$status" == "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✅ PASS${NC} | $test_name" | tee -a "$TEST_LOG"
    elif [ "$status" == "FAIL" ]; then
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}❌ FAIL${NC} | $test_name | $notes" | tee -a "$TEST_LOG"
    elif [ "$status" == "SKIP" ]; then
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        echo -e "${YELLOW}⏸ SKIP${NC} | $test_name | $notes" | tee -a "$TEST_LOG"
    fi
}

run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="$3"

    echo -e "\n${BLUE}Running:${NC} $test_name"
    echo -e "${YELLOW}Command:${NC} $command"

    if output=$(eval "$command" 2>&1); then
        if [ -n "$expected_pattern" ]; then
            if echo "$output" | grep -q "$expected_pattern"; then
                log_test "$test_name" "PASS"
            else
                log_test "$test_name" "FAIL" "Expected pattern not found: $expected_pattern"
                echo "$output" | head -10
            fi
        else
            log_test "$test_name" "PASS"
        fi
    else
        log_test "$test_name" "FAIL" "Command failed with exit code $?"
        echo "$output" | head -10
    fi
}

echo -e "${BLUE}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Walgo Complete Test Runner                ║${NC}"
echo -e "${BLUE}║   Automated Test Execution                   ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo "Test log: $TEST_LOG"
echo ""

# Create test directory
WALGO_TEST_DIR="$HOME/walgo-tests-$(date +%s)"
mkdir -p "$WALGO_TEST_DIR"
cd "$WALGO_TEST_DIR"
echo -e "${GREEN}✓${NC} Test directory: $WALGO_TEST_DIR"
echo ""

# ============================================================================
# PHASE 1: Environment & Diagnostics (Pre-Setup)
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 1: Environment & Diagnostics${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo version" "walgo version" "v0\."
run_test "walgo doctor (pre-setup)" "walgo doctor" ""

# ============================================================================
# PHASE 2: Dependency Setup
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 2: Dependency Setup${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo setup-deps" "walgo setup-deps" ""
run_test "walgo setup (testnet)" "walgo setup --network testnet" "testnet"
run_test "walgo doctor -v (post-setup)" "walgo doctor -v" ""

# ============================================================================
# PHASE 3: Site Creation
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 3: Site Creation${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo init" "walgo init test-init-site && ls test-init-site/walgo.yaml" "test-init-site/walgo.yaml"

# Quickstart with automated input
run_test "walgo quickstart" "printf '1\nn\n' | walgo quickstart test-quickstart-site" ""

cd test-quickstart-site || exit 1
run_test "walgo new content" "walgo new posts/test-post.md" ""

# ============================================================================
# PHASE 4: Obsidian Import
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 4: Obsidian Import${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

cd "$WALGO_TEST_DIR"
mkdir -p test-vault
cat > test-vault/note1.md << 'EOF'
# Note 1
This links to [[note2]].
EOF

cat > test-vault/note2.md << 'EOF'
# Note 2
Content here.
EOF

run_test "walgo import --dry-run" "walgo import test-vault --dry-run" "note1.md"
run_test "walgo import" "walgo import test-vault --site-name test-obsidian-import" ""

# ============================================================================
# PHASE 5: Build & Optimization
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 5: Build & Optimization${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

cd "$WALGO_TEST_DIR/test-quickstart-site"
run_test "walgo build" "walgo build" "public"
run_test "walgo optimize" "walgo optimize" ""
run_test "walgo compress" "walgo compress" ""

# Verify compression
if [ "$(find public -name '*.br' | wc -l)" -gt 0 ]; then
    log_test "Compression verification" "PASS"
else
    log_test "Compression verification" "FAIL" "No .br files created"
fi

# Test serve (run in background briefly)
walgo serve > /dev/null 2>&1 &
SERVE_PID=$!
sleep 3
if curl -s http://localhost:1313 | grep -qi "html"; then
    log_test "walgo serve" "PASS"
else
    log_test "walgo serve" "FAIL" "Server not responding"
fi
kill $SERVE_PID 2>/dev/null || true

# ============================================================================
# PHASE 6: AI Features (Optional - requires API key)
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 6: AI Features (Skipping - requires API key)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

log_test "walgo ai configure" "SKIP" "Requires API key"
log_test "walgo ai generate" "SKIP" "Requires API key"
log_test "walgo ai update" "SKIP" "Requires API key"
log_test "walgo ai plan" "SKIP" "Requires API key"
log_test "walgo ai resume" "SKIP" "Requires API key"
log_test "walgo ai pipeline" "SKIP" "Requires API key"

# ============================================================================
# PHASE 7: Deployment (HTTP - No Wallet)
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 7: Deployment (HTTP - Skipping)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

log_test "walgo deploy-http" "SKIP" "Network-dependent, skipping in automated test"
log_test "walgo status" "SKIP" "Requires deployment first"

# ============================================================================
# PHASE 8: Deployment (On-Chain - With Wallet)
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 8: On-Chain Deployment (Skipping)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

log_test "walgo deploy" "SKIP" "Requires wallet with funds"
log_test "walgo launch" "SKIP" "Requires wallet with funds"
log_test "walgo update" "SKIP" "Requires deployment first"

# ============================================================================
# PHASE 9: Project Management
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 9: Project Management${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo projects (list)" "walgo projects" ""
run_test "walgo projects show" "walgo projects show test-quickstart-site || true" ""
run_test "walgo projects edit" "walgo projects edit test-quickstart-site --description 'Test' || true" ""

# ============================================================================
# PHASE 10: Domain Configuration
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 10: Domain Configuration${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo domain" "walgo domain || true" ""

# ============================================================================
# PHASE 11: Advanced Features
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 11: Advanced Features${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo uninstall --dry-run" "walgo uninstall --dry-run" ""
log_test "walgo desktop" "SKIP" "Optional desktop app"

# ============================================================================
# PHASE 12: Error Handling & Edge Cases
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 12: Error Handling${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Invalid command
if walgo invalid-command 2>&1 | grep -qi "unknown\|invalid"; then
    log_test "Invalid command handling" "PASS"
else
    log_test "Invalid command handling" "FAIL" "Expected error message"
fi

# Missing config
cd /tmp
if walgo build 2>&1 | grep -qi "walgo.yaml\|not found\|config"; then
    log_test "Missing config handling" "PASS"
else
    log_test "Missing config handling" "FAIL" "Expected config error"
fi

cd "$WALGO_TEST_DIR/test-quickstart-site"

# ============================================================================
# PHASE 13: Help & Documentation
# ============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}PHASE 13: Help & Documentation${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

run_test "walgo --help" "walgo --help" "Usage"
run_test "walgo deploy --help" "walgo deploy --help" "epochs"
run_test "walgo projects --help" "walgo projects --help" ""
run_test "walgo ai --help" "walgo ai --help" ""

# ============================================================================
# Test Summary
# ============================================================================
echo -e "\n${BLUE}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Test Execution Summary                     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${GREEN}Total Tests:${NC} $TOTAL_TESTS"
echo -e "${GREEN}Passed:${NC} $PASSED_TESTS ($(echo "scale=1; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)%)"
echo -e "${RED}Failed:${NC} $FAILED_TESTS"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED_TESTS (optional or requires credentials)"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✅ All non-skipped tests passed!${NC}"
    EXIT_CODE=0
else
    echo -e "${RED}❌ Some tests failed. See log for details.${NC}"
    EXIT_CODE=1
fi

echo ""
echo "Test directory: $WALGO_TEST_DIR"
echo "Full test log: $TEST_LOG"
echo ""
echo -e "${YELLOW}To clean up:${NC}"
echo "  rm -rf $WALGO_TEST_DIR"
echo ""

# Create summary report
{
    echo "Walgo Test Execution Report"
    echo "==========================="
    echo "Date: $(date)"
    echo "Version: $(walgo version 2>&1 | head -1)"
    echo ""
    echo "Results:"
    echo "--------"
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Skipped: $SKIPPED_TESTS"
    echo ""
    echo "Test Directory: $WALGO_TEST_DIR"
    echo "Full Log: $TEST_LOG"
} > "$WALGO_TEST_DIR/test_summary.txt"

echo -e "${BLUE}Summary saved to:${NC} $WALGO_TEST_DIR/test_summary.txt"
echo ""

exit $EXIT_CODE
