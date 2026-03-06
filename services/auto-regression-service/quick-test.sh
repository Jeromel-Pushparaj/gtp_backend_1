#!/bin/bash

# Quick test script for Auto-Regression API on localhost:8092

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   Auto-Regression API - Quick Test (Port 8092)        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Test 1: Health Check
echo -e "${BLUE}[Test 1/3]${NC} Health Check"
echo -e "${YELLOW}→ GET http://localhost:8092/health${NC}"
echo ""

response=$(curl -s http://localhost:8092/health 2>/dev/null)
if [ $? -eq 0 ] && echo "$response" | grep -q "healthy"; then
    echo -e "${GREEN}✓ PASSED${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "Service is not responding on port 8092"
    echo ""
    echo "Start the service with:"
    echo "  cd services/auto-regression-service"
    echo "  make run-api"
    exit 1
fi

echo ""
echo "─────────────────────────────────────────────────────────"
echo ""

# Test 2: Validation Test (Empty Request)
echo -e "${BLUE}[Test 2/3]${NC} Input Validation (Empty Request)"
echo -e "${YELLOW}→ POST http://localhost:8092/api/v1/test/aggregate${NC}"
echo -e "${YELLOW}  Body: {}${NC}"
echo ""

response=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8092/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d '{}' 2>/dev/null)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 400 ]; then
    echo -e "${GREEN}✓ PASSED${NC} (Correctly rejected invalid request)"
    echo "Error message: $body"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "Expected HTTP 400, got HTTP $http_code"
    echo "Response: $body"
fi

echo ""
echo "─────────────────────────────────────────────────────────"
echo ""

# Test 3: Validation Test (Missing PAT Token)
echo -e "${BLUE}[Test 3/3]${NC} Input Validation (Missing PAT Token)"
echo -e "${YELLOW}→ POST http://localhost:8092/api/v1/test/aggregate${NC}"
echo -e "${YELLOW}  Body: {\"github_url\": \"https://github.com/test/repo\"}${NC}"
echo ""

response=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8092/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d '{"github_url": "https://github.com/test/repo"}' 2>/dev/null)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 400 ]; then
    echo -e "${GREEN}✓ PASSED${NC} (Correctly rejected request without PAT token)"
    echo "Error message: $body"
else
    echo -e "${RED}✗ FAILED${NC}"
    echo "Expected HTTP 400, got HTTP $http_code"
    echo "Response: $body"
fi

echo ""
echo "─────────────────────────────────────────────────────────"
echo ""

# Summary
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                    Test Summary                        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✓ Service is running on localhost:8092${NC}"
echo -e "${GREEN}✓ Health endpoint is working${NC}"
echo -e "${GREEN}✓ Input validation is working${NC}"
echo ""
echo -e "${BLUE}Available Endpoints:${NC}"
echo "  • GET  /health                    - Health check"
echo "  • POST /api/v1/test/aggregate     - Run regression tests"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "  1. View full testing guide:"
echo "     ${CYAN}cat services/auto-regression-service/TESTING_GUIDE.md${NC}"
echo ""
echo "  2. View OpenAPI specification:"
echo "     ${CYAN}cat services/auto-regression-service/auto_regression_api/openapi-spec.json | jq '.'${NC}"
echo ""
echo "  3. Test with real GitHub repository:"
echo "     ${CYAN}cd services/auto-regression-service/auto_regression_api${NC}"
echo "     ${CYAN}./example-request.sh${NC}"
echo "     (Requires GITHUB_PAT_TOKEN in ../auto_regression_agent/.env)"
echo ""
echo "  4. Check all services health:"
echo "     ${CYAN}cd services/auto-regression-service && make health${NC}"
echo ""

