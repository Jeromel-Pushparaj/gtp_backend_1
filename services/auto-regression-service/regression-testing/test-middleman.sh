#!/bin/bash

# Test script for the middleman service
# Usage: ./test-middleman.sh

# Configuration
MIDDLEMAN_URL="http://localhost:8081/api/v1/test/aggregate"
HEALTH_URL="http://localhost:8081/health"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Regression Testing Middleman Service Test ===${NC}\n"

# Check if service is running
echo -e "${YELLOW}1. Checking if middleman service is healthy...${NC}"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" $HEALTH_URL)
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
HEALTH_BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Service is healthy${NC}"
    echo "$HEALTH_BODY" | jq '.'
else
    echo -e "${RED}✗ Service is not responding. Please start the service first.${NC}"
    echo "Run: go run main.go"
    exit 1
fi

echo ""

# Get user input
echo -e "${YELLOW}2. Enter test parameters:${NC}"
read -p "GitHub URL: " GITHUB_URL
read -p "PAT Token: " PAT_TOKEN
read -p "Branch (default: main): " BRANCH
BRANCH=${BRANCH:-main}

echo ""
echo -e "${YELLOW}3. Calling middleman service...${NC}"
echo "URL: $MIDDLEMAN_URL"
echo "GitHub URL: $GITHUB_URL"
echo "Branch: $BRANCH"
echo ""

# Make the request
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $MIDDLEMAN_URL \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo -e "${YELLOW}4. Response:${NC}"
echo "HTTP Status Code: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✓ Request successful${NC}\n"
    
    # Parse and display summary
    echo -e "${YELLOW}=== Test Summary ===${NC}"
    echo "$BODY" | jq '{
        total_tests: .total_tests,
        tests_passed: .tests_passed,
        tests_failed: .tests_failed,
        tests_skipped: .tests_skipped,
        pass_rate: .pass_rate,
        executed_at: .executed_at
    }'
    
    echo ""
    echo -e "${YELLOW}=== Test Cases by Status ===${NC}"
    
    # Count by status code
    echo "$BODY" | jq -r '.unique_test_cases | group_by(.status_code) | .[] | "Status \(.[0].status_code): \(length) tests"'
    
    echo ""
    echo -e "${YELLOW}=== Failed Tests ===${NC}"
    echo "$BODY" | jq -r '.unique_test_cases | .[] | select(.passed == false and .skipped == false) | "- \(.name) (Status: \(.status_code))"'
    
    echo ""
    echo -e "${YELLOW}=== Full Response ===${NC}"
    echo "$BODY" | jq '.'
else
    echo -e "${RED}✗ Request failed${NC}\n"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
fi

echo ""
echo -e "${YELLOW}=== Test Complete ===${NC}"

