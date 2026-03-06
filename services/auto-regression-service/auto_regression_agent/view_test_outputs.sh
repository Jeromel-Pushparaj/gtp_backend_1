#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "📁 Test Output Files Viewer"
echo "========================================="
echo ""

SPEC_ID="${1:-7fc2339c-e201-46b8-8d05-3d6a458a125a}"

echo -e "${BLUE}Viewing outputs for spec: ${SPEC_ID}${NC}"
echo ""

# 1. Discovery Analysis
echo "========================================="
echo "1. Discovery Agent Analysis"
echo "========================================="
DISCOVERY_FILE="./output/discovery/${SPEC_ID}-ai-analysis.json"
if [ -f "$DISCOVERY_FILE" ]; then
    echo -e "${GREEN}✓ Found: $DISCOVERY_FILE${NC}"
    echo ""
    cat "$DISCOVERY_FILE" | python3 -m json.tool | head -50
    echo ""
    echo "(Showing first 50 lines. Full file: $DISCOVERY_FILE)"
else
    echo -e "${RED}✗ Not found: $DISCOVERY_FILE${NC}"
fi
echo ""

# 2. Test Strategy
echo "========================================="
echo "2. Designer Agent Strategy"
echo "========================================="
STRATEGY_FILE="./output/strategy/${SPEC_ID}-test-strategy.json"
if [ -f "$STRATEGY_FILE" ]; then
    echo -e "${GREEN}✓ Found: $STRATEGY_FILE${NC}"
    echo ""
    cat "$STRATEGY_FILE" | python3 -m json.tool | head -50
    echo ""
    echo "(Showing first 50 lines. Full file: $STRATEGY_FILE)"
else
    echo -e "${RED}✗ Not found: $STRATEGY_FILE${NC}"
fi
echo ""

# 3. Test Payloads
echo "========================================="
echo "3. Payload Agent Payloads"
echo "========================================="
PAYLOAD_FILE="./output/payloads/${SPEC_ID}-test-payloads.json"
if [ -f "$PAYLOAD_FILE" ]; then
    echo -e "${GREEN}✓ Found: $PAYLOAD_FILE${NC}"
    echo ""
    cat "$PAYLOAD_FILE" | python3 -m json.tool | head -50
    echo ""
    echo "(Showing first 50 lines. Full file: $PAYLOAD_FILE)"
else
    echo -e "${RED}✗ Not found: $PAYLOAD_FILE${NC}"
fi
echo ""

# 4. Execution Plan
echo "========================================="
echo "4. Executor Agent Plan"
echo "========================================="
PLAN_FILE="./output/plans/${SPEC_ID}-execution-plan.json"
if [ -f "$PLAN_FILE" ]; then
    echo -e "${GREEN}✓ Found: $PLAN_FILE${NC}"
    echo ""
    cat "$PLAN_FILE" | python3 -m json.tool
    echo ""
else
    echo -e "${RED}✗ Not found: $PLAN_FILE${NC}"
fi
echo ""

# 5. Test Results
echo "========================================="
echo "5. Test Results"
echo "========================================="
RESULTS_FILE="./output/results/${SPEC_ID}-test-results.json"
if [ -f "$RESULTS_FILE" ]; then
    echo -e "${GREEN}✓ Found: $RESULTS_FILE${NC}"
    echo ""
    cat "$RESULTS_FILE" | python3 -m json.tool
    echo ""
else
    echo -e "${RED}✗ Not found: $RESULTS_FILE${NC}"
fi
echo ""

# 6. Test Suites
echo "========================================="
echo "6. Test Suites"
echo "========================================="
if [ -d "./test_suites" ]; then
    echo -e "${BLUE}Test suite files:${NC}"
    ls -lht ./test_suites/ | head -5
    echo ""
    
    # Show latest test suite
    LATEST_SUITE=$(ls -t ./test_suites/*.json 2>/dev/null | head -1)
    if [ -n "$LATEST_SUITE" ]; then
        echo -e "${GREEN}Latest test suite: $LATEST_SUITE${NC}"
        echo ""
        cat "$LATEST_SUITE" | python3 -m json.tool | head -30
        echo ""
        echo "(Showing first 30 lines)"
    fi
else
    echo -e "${RED}✗ test_suites directory not found${NC}"
fi
echo ""

# Summary
echo "========================================="
echo "Summary"
echo "========================================="
echo ""
echo "To view full files:"
echo -e "${GREEN}cat $DISCOVERY_FILE | python3 -m json.tool | less${NC}"
echo -e "${GREEN}cat $STRATEGY_FILE | python3 -m json.tool | less${NC}"
echo -e "${GREEN}cat $PAYLOAD_FILE | python3 -m json.tool | less${NC}"
echo -e "${GREEN}cat $PLAN_FILE | python3 -m json.tool | less${NC}"
echo -e "${GREEN}cat $RESULTS_FILE | python3 -m json.tool | less${NC}"
echo ""

