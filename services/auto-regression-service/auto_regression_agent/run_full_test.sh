#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api/v1"

echo "========================================="
echo "OpenTest - Full Test Workflow"
echo "========================================="
echo ""

# Check if backend is running
echo "Checking if backend is running..."
if ! curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running at http://localhost:8080${NC}"
    echo ""
    echo "Start the backend with:"
    echo "  go run cmd/server/main.go"
    exit 1
fi
echo -e "${GREEN}✓ Backend is running${NC}"
echo ""

# Check if Redis is running
echo "Checking if Redis is running..."
if redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Redis is running${NC}"
else
    echo -e "${YELLOW}⚠️  Redis is not running (tests will work but no persistence)${NC}"
fi
echo ""

# Upload spec
echo "========================================="
echo "Step 1: Upload OpenAPI Spec"
echo "========================================="
echo ""

if [ ! -f "openAPISample.json" ]; then
    echo -e "${RED}❌ openAPISample.json not found${NC}"
    exit 1
fi

echo "Uploading openAPISample.json..."
echo ""

response=$(curl -s -X POST "${API_BASE}/specs" \
  -F "spec=@openAPISample.json" \
  -F "name=Petstore API Full Test" \
  -F "service=petstore" \
  -F "team_id=test-team" \
  -F "run_mode=full")

# Check if upload was successful (check for workflow_id instead of success field)
if echo "$response" | grep -q '"workflow_id"'; then
    echo -e "${GREEN}✅ Upload successful!${NC}"
    echo ""
    echo "$response" | python3 -m json.tool
    
    # Extract workflow_id
    workflow_id=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('workflow_id', ''))" 2>/dev/null)
    spec_id=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('spec_id', ''))" 2>/dev/null)
    endpoints=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('endpoints_extracted', 0))" 2>/dev/null)
    jobs=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('jobs_created', 0))" 2>/dev/null)
    
    echo ""
    echo -e "${BLUE}Workflow ID: ${workflow_id}${NC}"
    echo -e "${BLUE}Spec ID: ${spec_id}${NC}"
    echo -e "${BLUE}Endpoints: ${endpoints}${NC}"
    echo -e "${BLUE}Jobs Created: ${jobs}${NC}"
else
    echo -e "${RED}❌ Upload failed${NC}"
    echo "$response"
    exit 1
fi

echo ""

# Monitor progress
echo "========================================="
echo "Step 2: Monitor Test Execution"
echo "========================================="
echo ""

if [ -z "$workflow_id" ]; then
    echo -e "${RED}❌ No workflow_id received${NC}"
    exit 1
fi

echo "Monitoring workflow: $workflow_id"
echo ""
echo "Checking status every 5 seconds..."
echo "(Press Ctrl+C to stop monitoring)"
echo ""

max_checks=60  # Monitor for up to 5 minutes
check_count=0

while [ $check_count -lt $max_checks ]; do
    # Get current status
    status_response=$(curl -s "${API_BASE}/runs/${workflow_id}")
    
    if echo "$status_response" | grep -q '"state"'; then
        state=$(echo "$status_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('state', 'unknown'))" 2>/dev/null)
        progress=$(echo "$status_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('progress', 0))" 2>/dev/null)
        
        echo -e "[$(date +%H:%M:%S)] State: ${BLUE}${state}${NC} | Progress: ${BLUE}${progress}%${NC}"
        
        # Check if completed
        if [ "$state" = "completed" ]; then
            echo ""
            echo -e "${GREEN}✅ Test execution completed!${NC}"
            break
        elif [ "$state" = "failed" ]; then
            echo ""
            echo -e "${RED}❌ Test execution failed${NC}"
            break
        fi
    fi
    
    sleep 5
    check_count=$((check_count + 1))
done

echo ""

# Get final results
echo "========================================="
echo "Step 3: Test Results"
echo "========================================="
echo ""

echo "Fetching detailed results..."
echo ""

./check_test_results.sh "$workflow_id"

echo ""
echo "========================================="
echo "Complete!"
echo "========================================="
echo ""

echo "To check results again later, run:"
echo -e "${GREEN}./check_test_results.sh ${workflow_id}${NC}"
echo ""

echo "To list all test runs:"
echo -e "${GREEN}./check_test_results.sh list${NC}"
echo ""

