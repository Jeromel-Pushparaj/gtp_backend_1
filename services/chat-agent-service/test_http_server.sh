#!/bin/bash

echo "=========================================="
echo "Chat Agent HTTP Server Test"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SERVER_URL="http://localhost:8082"

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${YELLOW}Testing: $description${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$SERVER_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X "$method" "$SERVER_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
    body=$(echo "$response" | grep -v "HTTP_CODE")
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ SUCCESS${NC}"
        echo "Response: $body"
    else
        echo -e "${RED}✗ FAILED (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
}

# Check if server is running
echo "1. Checking if server is running..."
if ! curl -s "$SERVER_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}✗ Server is not running on port 8082${NC}"
    echo ""
    echo "Please start the server first:"
    echo "  cd services/chat-agent-service"
    echo "  make run-http"
    echo ""
    echo "Or with environment variables:"
    echo "  GROQ_API_KEY=your_key ./chat-agent-server"
    exit 1
fi
echo -e "${GREEN}✓ Server is running${NC}"
echo ""

# Test health endpoint
test_endpoint "GET" "/health" "" "Health check endpoint"

# Test chat endpoint with simple message
test_endpoint "POST" "/api/v1/chat" \
    '{"message": "Hello, what can you help me with?"}' \
    "Simple chat message"

# Test chat endpoint with tool usage
test_endpoint "POST" "/api/v1/chat" \
    '{"message": "What is the health status of the backend?"}' \
    "Chat with tool usage (health check)"

# Test invalid request
echo -e "${YELLOW}Testing: Invalid request (missing message)${NC}"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$SERVER_URL/api/v1/chat" \
    -H "Content-Type: application/json" \
    -d '{}')
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
if [ "$http_code" = "400" ]; then
    echo -e "${GREEN}✓ Correctly rejected invalid request${NC}"
else
    echo -e "${RED}✗ Should have returned 400 for invalid request${NC}"
fi
echo ""

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "All basic tests completed!"
echo ""
echo "Next steps:"
echo "1. Integrate with your frontend application"
echo "2. Test with more complex queries"
echo "3. Monitor the server logs for any issues"
echo ""
echo "Example frontend integration:"
echo ""
echo "  fetch('http://localhost:8082/api/v1/chat', {"
echo "    method: 'POST',"
echo "    headers: { 'Content-Type': 'application/json' },"
echo "    body: JSON.stringify({ message: 'Your question here' })"
echo "  })"
echo ""

