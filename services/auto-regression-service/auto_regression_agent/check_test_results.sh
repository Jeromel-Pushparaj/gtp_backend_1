#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api/v1"

echo "========================================="
echo "OpenTest - Test Results Checker"
echo "========================================="
echo ""

# Check if workflow_id is provided
if [ -z "$1" ]; then
    echo -e "${YELLOW}Usage: $0 <workflow_id>${NC}"
    echo ""
    echo "Example:"
    echo "  $0 wf_123456789"
    echo ""
    echo "Or list all runs:"
    echo "  $0 list"
    echo ""
    exit 1
fi

# List all runs
if [ "$1" = "list" ]; then
    echo "Fetching all test runs..."
    echo ""
    
    response=$(curl -s "${API_BASE}/runs")
    
    if [ $? -eq 0 ]; then
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        echo -e "${RED}❌ Failed to fetch runs${NC}"
    fi
    exit 0
fi

WORKFLOW_ID="$1"

echo -e "${BLUE}Checking results for workflow: ${WORKFLOW_ID}${NC}"
echo ""

# 1. Get Run Details
echo "========================================="
echo "1. Run Details"
echo "========================================="
echo ""

run_details=$(curl -s "${API_BASE}/runs/${WORKFLOW_ID}")

if echo "$run_details" | grep -q '"error"'; then
    echo -e "${RED}❌ Workflow not found or error occurred${NC}"
    echo "$run_details" | python3 -m json.tool 2>/dev/null || echo "$run_details"
    exit 1
fi

echo "$run_details" | python3 -m json.tool 2>/dev/null || echo "$run_details"

# Extract status
status=$(echo "$run_details" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('state', 'unknown'))" 2>/dev/null)
progress=$(echo "$run_details" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('progress', 0))" 2>/dev/null)

echo ""
echo -e "Status: ${BLUE}${status}${NC}"
echo -e "Progress: ${BLUE}${progress}%${NC}"
echo ""

# 2. Get Test Report
echo "========================================="
echo "2. Test Report"
echo "========================================="
echo ""

report=$(curl -s "${API_BASE}/runs/${WORKFLOW_ID}/report")

if echo "$report" | grep -q '"error"'; then
    echo -e "${YELLOW}⚠️  Report not yet available${NC}"
else
    echo "$report" | python3 -m json.tool 2>/dev/null || echo "$report"
    
    # Extract summary
    total=$(echo "$report" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('summary', {}).get('total_tests', 0))" 2>/dev/null)
    passed=$(echo "$report" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('summary', {}).get('passed', 0))" 2>/dev/null)
    failed=$(echo "$report" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('summary', {}).get('failed', 0))" 2>/dev/null)
    
    echo ""
    echo -e "${GREEN}✅ Passed: ${passed}${NC}"
    echo -e "${RED}❌ Failed: ${failed}${NC}"
    echo -e "${BLUE}📊 Total: ${total}${NC}"
fi

echo ""

# 3. Get Test Cases
echo "========================================="
echo "3. Test Cases"
echo "========================================="
echo ""

test_cases=$(curl -s "${API_BASE}/runs/${WORKFLOW_ID}/test-cases")

if echo "$test_cases" | grep -q '"error"'; then
    echo -e "${YELLOW}⚠️  Test cases not yet available${NC}"
else
    echo "$test_cases" | python3 -m json.tool 2>/dev/null || echo "$test_cases"
fi

echo ""

# 4. Get Logs
echo "========================================="
echo "4. Execution Logs"
echo "========================================="
echo ""

logs=$(curl -s "${API_BASE}/runs/${WORKFLOW_ID}/logs")

if echo "$logs" | grep -q '"error"'; then
    echo -e "${YELLOW}⚠️  Logs not yet available${NC}"
else
    echo "$logs" | python3 -m json.tool 2>/dev/null || echo "$logs"
fi

echo ""

# 5. Check Output Files
echo "========================================="
echo "5. Output Files"
echo "========================================="
echo ""

OUTPUT_DIR="./output"

if [ -d "$OUTPUT_DIR" ]; then
    echo "Checking output directory: $OUTPUT_DIR"
    echo ""
    
    # Check for spec files
    if [ -d "$OUTPUT_DIR/specs" ]; then
        echo -e "${BLUE}Spec files:${NC}"
        ls -lh "$OUTPUT_DIR/specs/" 2>/dev/null || echo "  (none)"
        echo ""
    fi
    
    # Check for results
    if [ -d "$OUTPUT_DIR/results" ]; then
        echo -e "${BLUE}Result files:${NC}"
        ls -lh "$OUTPUT_DIR/results/" 2>/dev/null || echo "  (none)"
        echo ""
    fi
    
    # Check for reports
    if [ -d "$OUTPUT_DIR/reports" ]; then
        echo -e "${BLUE}Report files:${NC}"
        ls -lh "$OUTPUT_DIR/reports/" 2>/dev/null || echo "  (none)"
        echo ""
    fi
else
    echo -e "${YELLOW}⚠️  Output directory not found${NC}"
fi

# 6. Download Report (if available)
echo "========================================="
echo "6. Download Report"
echo "========================================="
echo ""

download_url="${API_BASE}/runs/${WORKFLOW_ID}/download"
echo "Download URL: $download_url"
echo ""

echo "To download the report, run:"
echo -e "${GREEN}curl -O ${download_url}${NC}"
echo ""

# 7. Summary
echo "========================================="
echo "Summary"
echo "========================================="
echo ""

if [ "$status" = "completed" ]; then
    echo -e "${GREEN}✅ Test run completed!${NC}"
elif [ "$status" = "running" ] || [ "$status" = "pending" ]; then
    echo -e "${YELLOW}⏳ Test run in progress (${progress}%)${NC}"
    echo ""
    echo "Run this script again to check updated status:"
    echo -e "${BLUE}./check_test_results.sh ${WORKFLOW_ID}${NC}"
elif [ "$status" = "failed" ]; then
    echo -e "${RED}❌ Test run failed${NC}"
else
    echo -e "${YELLOW}⚠️  Status: ${status}${NC}"
fi

echo ""

