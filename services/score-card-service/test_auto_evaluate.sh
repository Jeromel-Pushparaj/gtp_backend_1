#!/bin/bash

# Test script for V2 API Auto-Evaluate endpoints
# This script demonstrates how to use the auto-evaluate endpoints that fetch metrics automatically

BASE_URL="http://localhost:8085"

echo "=========================================="
echo "V2 API Auto-Evaluate Test Script"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Auto-evaluate service against ALL scorecards
echo -e "${BLUE}Test 1: Auto-Evaluate Service (All Scorecards)${NC}"
echo "POST $BASE_URL/api/v2/scorecards/auto-evaluate"
echo ""
echo "Request Body:"
cat << 'EOF' | jq '.'
{
  "service_name": "gtp_backend_1",
  "jira_project_key": "GTP"
}
EOF
echo ""

echo -e "${YELLOW}Sending request...${NC}"
curl -X POST "$BASE_URL/api/v2/scorecards/auto-evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "gtp_backend_1",
    "jira_project_key": "GTP"
  }' | jq '.'

echo ""
echo "=========================================="
echo ""

# Test 2: Auto-evaluate service against Code Quality scorecard
echo -e "${BLUE}Test 2: Auto-Evaluate Service (Code Quality Only)${NC}"
echo "POST $BASE_URL/api/v2/scorecards/auto-evaluate/CodeQuality"
echo ""
echo "Request Body:"
cat << 'EOF' | jq '.'
{
  "service_name": "gtp_backend_1",
  "jira_project_key": "GTP"
}
EOF
echo ""

echo -e "${YELLOW}Sending request...${NC}"
curl -X POST "$BASE_URL/api/v2/scorecards/auto-evaluate/CodeQuality" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "gtp_backend_1",
    "jira_project_key": "GTP"
  }' | jq '.'

echo ""
echo "=========================================="
echo ""

# Test 3: Auto-evaluate service against DORA Metrics scorecard
echo -e "${BLUE}Test 3: Auto-Evaluate Service (DORA Metrics Only)${NC}"
echo "POST $BASE_URL/api/v2/scorecards/auto-evaluate/DORA_Metrics"
echo ""
echo "Request Body:"
cat << 'EOF' | jq '.'
{
  "service_name": "gtp_backend_1",
  "jira_project_key": "GTP"
}
EOF
echo ""

echo -e "${YELLOW}Sending request...${NC}"
curl -X POST "$BASE_URL/api/v2/scorecards/auto-evaluate/DORA_Metrics" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "gtp_backend_1",
    "jira_project_key": "GTP"
  }' | jq '.'

echo ""
echo "=========================================="
echo ""

# Test 4: Auto-evaluate without Jira project key
echo -e "${BLUE}Test 4: Auto-Evaluate Service (Without Jira)${NC}"
echo "POST $BASE_URL/api/v2/scorecards/auto-evaluate"
echo ""
echo "Request Body:"
cat << 'EOF' | jq '.'
{
  "service_name": "gtp_backend_1"
}
EOF
echo ""

echo -e "${YELLOW}Sending request...${NC}"
curl -X POST "$BASE_URL/api/v2/scorecards/auto-evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "gtp_backend_1"
  }' | jq '.'

echo ""
echo "=========================================="
echo ""

echo -e "${GREEN}✅ All tests completed!${NC}"
echo ""
echo "Available Scorecards for auto-evaluate:"
echo "  - CodeQuality"
echo "  - DORA_Metrics"
echo "  - Security_Maturity"
echo "  - Production_Readiness"
echo "  - Service_Health"
echo "  - PR_Metrics"
echo ""
echo "Usage:"
echo "  POST /api/v2/scorecards/auto-evaluate"
echo "  POST /api/v2/scorecards/auto-evaluate/:name"
echo ""
echo "Request format:"
echo '  {"service_name": "repo-name", "jira_project_key": "PROJECT"}'
echo ""

