#!/bin/bash

echo "========================================="
echo "OpenTest API - File Upload Test"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api/v1"

# Check if backend is running
echo "Checking if backend is running..."
if ! curl -s -f "${API_BASE%/api/v1}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running at http://localhost:8080${NC}"
    echo "Start the backend with: docker-compose up -d backend"
    exit 1
fi
echo -e "${GREEN}✓ Backend is running${NC}"
echo ""

# Test 1: Upload OpenAPI spec file (multipart/form-data)
echo "========================================="
echo "Test 1: Upload OpenAPI Spec (Multipart)"
echo "========================================="
echo ""

if [ ! -f "openAPISample.json" ]; then
    echo -e "${RED}❌ openAPISample.json not found${NC}"
    exit 1
fi

echo "Uploading openAPISample.json..."
echo ""

response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/specs" \
  -F "spec=@openAPISample.json" \
  -F "name=Petstore API Test" \
  -F "service=petstore" \
  -F "team_id=test-team" \
  -F "run_mode=full")

http_status=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS/d')

if [ "$http_status" = "200" ] || [ "$http_status" = "201" ]; then
    echo -e "${GREEN}✅ Upload successful!${NC}"
    echo ""
    echo "Response:"
    echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    
    # Extract workflow_id and spec_id for later use
    workflow_id=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('workflow_id', ''))" 2>/dev/null)
    spec_id=$(echo "$body" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('spec_id', ''))" 2>/dev/null)
    
    echo ""
    echo -e "${GREEN}Workflow ID: $workflow_id${NC}"
    echo -e "${GREEN}Spec ID: $spec_id${NC}"
else
    echo -e "${RED}❌ Upload failed with status: $http_status${NC}"
    echo "Response:"
    echo "$body"
    exit 1
fi

echo ""

# Test 2: Validate spec without creating jobs
echo "========================================="
echo "Test 2: Validate Spec (No Job Creation)"
echo "========================================="
echo ""

echo "Validating openAPISample.json..."
echo ""

response2=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/specs/validate" \
  -F "spec=@openAPISample.json")

http_status2=$(echo "$response2" | grep "HTTP_STATUS" | cut -d: -f2)
body2=$(echo "$response2" | sed '/HTTP_STATUS/d')

if [ "$http_status2" = "200" ]; then
    echo -e "${GREEN}✅ Validation successful!${NC}"
    echo ""
    echo "Response:"
    echo "$body2" | python3 -m json.tool 2>/dev/null || echo "$body2"
else
    echo -e "${RED}❌ Validation failed with status: $http_status2${NC}"
    echo "Response:"
    echo "$body2"
fi

echo ""

# Test 3: Upload raw YAML/JSON (without multipart)
echo "========================================="
echo "Test 3: Upload Raw JSON (Direct Body)"
echo "========================================="
echo ""

echo "Uploading raw JSON content..."
echo ""

response3=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${API_BASE}/specs" \
  -H "Content-Type: application/json" \
  --data-binary @openAPISample.json)

http_status3=$(echo "$response3" | grep "HTTP_STATUS" | cut -d: -f2)
body3=$(echo "$response3" | sed '/HTTP_STATUS/d')

if [ "$http_status3" = "200" ] || [ "$http_status3" = "201" ]; then
    echo -e "${GREEN}✅ Raw upload successful!${NC}"
    echo ""
    echo "Response:"
    echo "$body3" | python3 -m json.tool 2>/dev/null || echo "$body3"
else
    echo -e "${RED}❌ Raw upload failed with status: $http_status3${NC}"
    echo "Response:"
    echo "$body3"
fi

echo ""

# Test 4: Check spec status (if we have a spec_id)
if [ -n "$spec_id" ]; then
    echo "========================================="
    echo "Test 4: Check Spec Status"
    echo "========================================="
    echo ""
    
    echo "Checking status for spec: $spec_id"
    echo ""
    
    response4=$(curl -s "${API_BASE}/specs/${spec_id}/status")
    
    echo "Response:"
    echo "$response4" | python3 -m json.tool 2>/dev/null || echo "$response4"
    echo ""
fi

echo "========================================="
echo -e "${GREEN}✅ All upload tests completed!${NC}"
echo "========================================="
echo ""
echo "Summary:"
echo "- Test 1 (Multipart Upload): $([ "$http_status" = "200" ] || [ "$http_status" = "201" ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${RED}FAILED${NC}")"
echo "- Test 2 (Validation): $([ "$http_status2" = "200" ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${RED}FAILED${NC}")"
echo "- Test 3 (Raw Upload): $([ "$http_status3" = "200" ] || [ "$http_status3" = "201" ] && echo -e "${GREEN}PASSED${NC}" || echo -e "${RED}FAILED${NC}")"

