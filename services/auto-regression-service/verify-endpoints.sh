#!/bin/bash

# Verification script for Auto-Regression Service
# Tests that all endpoints are accessible on the correct ports

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "🔍 Auto-Regression Service Verification"
echo "========================================="
echo ""

# Function to check if a service is running
check_service() {
    local name=$1
    local url=$2
    local port=$3
    
    echo -e "${BLUE}Checking $name on port $port...${NC}"
    
    response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✓ $name is running${NC}"
        echo "  Response: $body" | head -c 100
        echo ""
        return 0
    else
        echo -e "${RED}✗ $name is not responding (HTTP $http_code)${NC}"
        echo ""
        return 1
    fi
}

# Track results
all_passed=true

# Check Agent API (port 8080)
echo "========================================="
echo "1. Agent API Health Check"
echo "========================================="
if ! check_service "Agent API" "http://localhost:8080/health" "8080"; then
    all_passed=false
    echo -e "${YELLOW}⚠️  Agent API is not running. Start it with:${NC}"
    echo "  cd services/auto-regression-service"
    echo "  make up"
    echo ""
fi

# Check Auto-Regression API (port 8092)
echo "========================================="
echo "2. Auto-Regression API Health Check"
echo "========================================="
if ! check_service "Auto-Regression API" "http://localhost:8092/health" "8092"; then
    all_passed=false
    echo -e "${YELLOW}⚠️  Auto-Regression API is not running. Start it with:${NC}"
    echo "  cd services/auto-regression-service"
    echo "  make run-api"
    echo ""
fi

# Test the main endpoint (if both services are running)
echo "========================================="
echo "3. Testing Main Endpoint"
echo "========================================="

if curl -s http://localhost:8092/health > /dev/null 2>&1; then
    echo -e "${BLUE}Testing POST /api/v1/test/aggregate...${NC}"
    echo ""
    
    # Test with invalid request (should return 400)
    response=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8092/api/v1/test/aggregate \
      -H "Content-Type: application/json" \
      -d '{}' 2>/dev/null)
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" -eq 400 ]; then
        echo -e "${GREEN}✓ Endpoint validation working (returned 400 for invalid request)${NC}"
        echo ""
    else
        echo -e "${YELLOW}⚠️  Unexpected response: HTTP $http_code${NC}"
        echo ""
    fi
else
    echo -e "${YELLOW}⚠️  Skipping endpoint test (API not running)${NC}"
    echo ""
fi

# Summary
echo "========================================="
echo "📊 Verification Summary"
echo "========================================="

if [ "$all_passed" = true ]; then
    echo -e "${GREEN}✓ All services are running correctly!${NC}"
    echo ""
    echo "Available endpoints:"
    echo "  • Agent API:           http://localhost:8080"
    echo "  • Auto-Regression API: http://localhost:8092"
    echo ""
    echo "Next steps:"
    echo "  1. View OpenAPI spec: cat services/auto-regression-service/auto_regression_api/openapi-spec.json"
    echo "  2. Test with real data: cd services/auto-regression-service/auto_regression_api && ./example-request.sh"
    echo "  3. Check health: make health"
else
    echo -e "${RED}✗ Some services are not running${NC}"
    echo ""
    echo "To start all services:"
    echo "  cd services/auto-regression-service"
    echo "  make up          # Start agent services"
    echo "  make run-api     # Start API on port 8092"
fi

echo ""
echo "========================================="

