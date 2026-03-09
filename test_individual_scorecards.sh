#!/bin/bash

# Test script for individual scorecard endpoints
# Usage: ./test_individual_scorecards.sh [service_name] [jira_project_key]
# Note: jira_project_key is optional

# Configuration
SERVICE="${1:-delivery-management-frontend}"
JIRA_KEY="${2:-}"  # Optional, empty by default
BASE_URL="http://localhost:8085/api/v2/scorecards"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Testing Individual Scorecard Endpoints"
echo "=========================================="
echo "Service: $SERVICE"
if [ -n "$JIRA_KEY" ]; then
    echo "Jira Project: $JIRA_KEY"
else
    echo "Jira Project: Not provided (optional)"
fi
echo "Base URL: $BASE_URL"
echo "=========================================="
echo ""

# Function to test an endpoint
test_endpoint() {
    local name=$1
    local endpoint=$2
    local params=$3
    
    echo -e "${BLUE}Testing: $name${NC}"
    echo "Endpoint: $endpoint"
    echo "Parameters: $params"
    echo ""
    
    local url="$BASE_URL/$endpoint?$params"
    local response=$(curl -s "$url")
    
    if [ $? -eq 0 ]; then
        echo "$response" | jq '.evaluation | {scorecard_name, achieved_level_name, pass_percentage, rules_passed, rules_total}'
        echo ""
        echo -e "${GREEN}✓ Success${NC}"
    else
        echo -e "${YELLOW}✗ Failed to fetch${NC}"
    fi
    
    echo "=========================================="
    echo ""
}

# Build query parameters
if [ -n "$JIRA_KEY" ]; then
    JIRA_PARAM="&jira_project_key=$JIRA_KEY"
else
    JIRA_PARAM=""
fi

# Test 1: Code Quality
test_endpoint \
    "Code Quality Scorecard" \
    "code-quality" \
    "service_name=$SERVICE$JIRA_PARAM"

# Test 2: Service Health
test_endpoint \
    "Service Health Scorecard" \
    "service-health" \
    "service_name=$SERVICE$JIRA_PARAM"

# Test 3: Security Maturity
test_endpoint \
    "Security Maturity Scorecard" \
    "security-maturity" \
    "service_name=$SERVICE"

# Test 4: Production Readiness
test_endpoint \
    "Production Readiness Scorecard" \
    "production-readiness" \
    "service_name=$SERVICE"

# Test 5: PR Metrics
test_endpoint \
    "PR Metrics Scorecard" \
    "pr-metrics" \
    "service_name=$SERVICE"

# Summary
echo ""
echo "=========================================="
echo "Summary: All Individual Scorecards"
echo "=========================================="
echo ""

echo -e "${BLUE}Fetching all scorecards in one view...${NC}"
echo ""

curl -s "$BASE_URL/code-quality?service_name=$SERVICE$JIRA_PARAM" | jq -r '"1. Code Quality: " + .evaluation.achieved_level_name + " (" + (.evaluation.pass_percentage | tostring) + "%)"'

curl -s "$BASE_URL/service-health?service_name=$SERVICE$JIRA_PARAM" | jq -r '"2. Service Health: " + .evaluation.achieved_level_name + " (" + (.evaluation.pass_percentage | tostring) + "%)"'

curl -s "$BASE_URL/security-maturity?service_name=$SERVICE" | jq -r '"3. Security Maturity: " + .evaluation.achieved_level_name + " (" + (.evaluation.pass_percentage | tostring) + "%)"'

curl -s "$BASE_URL/production-readiness?service_name=$SERVICE" | jq -r '"4. Production Readiness: " + .evaluation.achieved_level_name + " (" + (.evaluation.pass_percentage | tostring) + "%)"'

curl -s "$BASE_URL/pr-metrics?service_name=$SERVICE" | jq -r '"5. PR Metrics: " + .evaluation.achieved_level_name + " (" + (.evaluation.pass_percentage | tostring) + "%)"'

echo ""
echo "=========================================="
echo -e "${GREEN}✓ All tests completed!${NC}"
echo "=========================================="

