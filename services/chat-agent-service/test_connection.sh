#!/bin/bash

echo "Testing GTP Backend API Connection"
echo "===================================="
echo ""

API_BASE_URL=${API_BASE_URL:-http://localhost:8080}

echo "Testing connection to: $API_BASE_URL"
echo ""

echo "1. Testing /health endpoint..."
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE_URL/health")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | grep -v "HTTP_CODE")

if [ "$http_code" = "200" ]; then
    echo "   SUCCESS: Backend is healthy"
    echo "   Response: $body"
else
    echo "   FAILED: HTTP $http_code"
    echo "   Response: $body"
    echo ""
    echo "   Make sure the backend API is running:"
    echo "   cd sonar-shell-test && go run main.go -server -port=8080"
    exit 1
fi

echo ""
echo "2. Testing /api/v1/github/metrics/all endpoint..."
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE_URL/api/v1/github/metrics/all")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)

if [ "$http_code" = "200" ]; then
    echo "   SUCCESS: GitHub metrics endpoint working"
else
    echo "   FAILED: HTTP $http_code"
    echo "   This might be due to missing credentials or API issues"
fi

echo ""
echo "3. Testing /api/v1/orgs endpoint..."
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE_URL/api/v1/orgs")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)

if [ "$http_code" = "200" ]; then
    echo "   SUCCESS: Organizations endpoint working"
else
    echo "   FAILED: HTTP $http_code"
fi

echo ""
echo "===================================="
echo "Connection test complete!"
echo ""
echo "If all tests passed, your MCP server should work correctly."
echo "Configure Claude Desktop with:"
echo ""
echo "  Command: $(pwd)/mcp-server"
echo "  API_BASE_URL: $API_BASE_URL"
echo ""

