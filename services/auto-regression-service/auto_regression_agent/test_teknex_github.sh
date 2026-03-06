#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "========================================="
echo "🐙 Testing Teknex GitHub Integration"
echo "========================================="
echo ""

# Configuration - Read from environment variables
GITHUB_URL="${GITHUB_URL:-https://github.com/teknex-poc/test-backend}"
PAT_TOKEN="${GITHUB_PAT_TOKEN}"
BRANCH="${GITHUB_BRANCH:-main}"

# Validate required environment variables
if [ -z "$PAT_TOKEN" ]; then
    echo -e "${RED}❌ GITHUB_PAT_TOKEN not set!${NC}"
    echo ""
    echo "Set it in .env file:"
    echo "  GITHUB_PAT_TOKEN=your_token_here"
    exit 1
fi

# Check if backend is running
echo "Checking if backend is running..."
if ! curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running${NC}"
    echo ""
    echo "Start the backend first:"
    echo "  go run cmd/server/main.go"
    exit 1
fi
echo -e "${GREEN}✅ Backend is running${NC}"
echo ""

# Check if Redis is running
echo "Checking if Redis is running..."
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Redis is not running${NC}"
    echo ""
    echo "Start Redis:"
    echo "  brew services start redis"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✅ Redis is running${NC}"
fi
echo ""

echo "========================================="
echo "Test Configuration:"
echo "========================================="
echo -e "${BLUE}GitHub URL:${NC} $GITHUB_URL"
echo -e "${BLUE}Branch:${NC} $BRANCH"
echo -e "${BLUE}PAT Token:${NC} ${PAT_TOKEN:0:20}..."
echo -e "${BLUE}Expected Spec File:${NC} openapi-spec.json"
echo ""

echo "========================================="
echo "Triggering GitHub Integration Test..."
echo "========================================="
echo ""

# Make the API call
echo "Sending request to: POST http://localhost:8080/api/v1/github/test"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/github/test \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }")

# Extract HTTP status code and body
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP Status Code: $HTTP_CODE"
echo ""

# Check if request was successful
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ Test completed successfully!${NC}"
    echo ""
    
    # Extract workflow ID and spec ID
    WORKFLOW_ID=$(echo "$BODY" | python3 -c "import sys, json; print(json.load(sys.stdin).get('workflow_id', 'N/A'))" 2>/dev/null || echo "N/A")
    SPEC_ID=$(echo "$BODY" | python3 -c "import sys, json; print(json.load(sys.stdin).get('spec_id', 'N/A'))" 2>/dev/null || echo "N/A")
    
    echo -e "${BLUE}Workflow ID:${NC} $WORKFLOW_ID"
    echo -e "${BLUE}Spec ID:${NC} $SPEC_ID"
    echo ""
    
    # Extract test summary
    TOTAL=$(echo "$BODY" | python3 -c "import sys, json; r=json.load(sys.stdin).get('results',{}); print(r.get('summary',{}).get('total_tests','N/A'))" 2>/dev/null || echo "N/A")
    PASSED=$(echo "$BODY" | python3 -c "import sys, json; r=json.load(sys.stdin).get('results',{}); print(r.get('summary',{}).get('passed','N/A'))" 2>/dev/null || echo "N/A")
    FAILED=$(echo "$BODY" | python3 -c "import sys, json; r=json.load(sys.stdin).get('results',{}); print(r.get('summary',{}).get('failed','N/A'))" 2>/dev/null || echo "N/A")
    
    echo "========================================="
    echo "Test Summary:"
    echo "========================================="
    echo -e "${BLUE}Total Tests:${NC} $TOTAL"
    echo -e "${GREEN}Passed:${NC} $PASSED"
    echo -e "${RED}Failed:${NC} $FAILED"
    echo ""
    
    # Pretty print full results
    echo "========================================="
    echo "Full Response:"
    echo "========================================="
    echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
    
    # Show where to find detailed results
    echo ""
    echo "========================================="
    echo "Detailed Results Location:"
    echo "========================================="
    echo "Results: ./output/results/${SPEC_ID}-test-results.json"
    echo "Strategy: ./output/strategy/${SPEC_ID}-test-strategy.json"
    echo "Payloads: ./output/payloads/${SPEC_ID}-test-payloads.json"
    echo "Plan: ./output/plans/${SPEC_ID}-execution-plan.json"
    
else
    echo -e "${RED}❌ Test failed (HTTP $HTTP_CODE)${NC}"
    echo ""
    echo "Response:"
    echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
    echo ""
    
    # Check for common errors
    if echo "$BODY" | grep -q "OpenAPI spec not found"; then
        echo -e "${YELLOW}💡 Tip: Make sure 'openapi-spec.json' exists in the repository root${NC}"
    elif echo "$BODY" | grep -q "401\|403\|404"; then
        echo -e "${YELLOW}💡 Tip: Check your PAT token permissions and repository access${NC}"
    elif echo "$BODY" | grep -q "timeout"; then
        echo -e "${YELLOW}💡 Tip: Make sure agents are running (./start_agents_no_feedback.sh)${NC}"
    fi
fi

echo ""
echo "========================================="
echo "Done!"
echo "========================================="

