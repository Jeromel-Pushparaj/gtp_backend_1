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
echo "🐙 GitHub Integration Test"
echo "========================================="
echo ""

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

# Check if agents are running
echo "Checking if agents are running..."
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Redis is not running - agents may not work${NC}"
    echo ""
    echo "Start Redis and agents:"
    echo "  brew services start redis"
    echo "  ./start_agents_no_feedback.sh"
    echo ""
fi

# Configuration - Read from environment variables
GITHUB_URL="${GITHUB_URL:-https://github.com/OAI/OpenAPI-Specification}"
BRANCH="${GITHUB_BRANCH:-main}"
PAT_TOKEN="${GITHUB_PAT_TOKEN}"

# Validate required environment variables
if [ -z "$PAT_TOKEN" ]; then
    echo -e "${RED}❌ GITHUB_PAT_TOKEN not set!${NC}"
    echo ""
    echo "Set it in .env file:"
    echo "  GITHUB_PAT_TOKEN=your_token_here"
    exit 1
fi

echo "========================================="
echo "Test Configuration:"
echo "========================================="
echo -e "${BLUE}GitHub URL:${NC} $GITHUB_URL"
echo -e "${BLUE}Branch:${NC} $BRANCH"
echo -e "${BLUE}PAT Token:${NC} ${PAT_TOKEN:0:10}..."
echo ""

# Allow command-line arguments to override
if [ ! -z "$1" ]; then
    GITHUB_URL="$1"
fi

if [ ! -z "$2" ]; then
    PAT_TOKEN="$2"
fi

if [ ! -z "$3" ]; then
    BRANCH="$3"
fi

echo "========================================="
echo "Triggering GitHub Integration Test..."
echo "========================================="
echo ""

# Make the API call
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/github/test \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }")

# Check if request was successful
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✅ Test completed successfully!${NC}"
    echo ""
    
    # Extract workflow ID
    WORKFLOW_ID=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('workflow_id', 'N/A'))" 2>/dev/null || echo "N/A")
    SPEC_ID=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('spec_id', 'N/A'))" 2>/dev/null || echo "N/A")
    
    echo -e "${BLUE}Workflow ID:${NC} $WORKFLOW_ID"
    echo -e "${BLUE}Spec ID:${NC} $SPEC_ID"
    echo ""
    
    # Pretty print results
    echo "========================================="
    echo "Test Results:"
    echo "========================================="
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
    
else
    echo -e "${RED}❌ Test failed${NC}"
    echo ""
    echo "Response:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
fi

echo ""
echo "========================================="
echo "Done!"
echo "========================================="

