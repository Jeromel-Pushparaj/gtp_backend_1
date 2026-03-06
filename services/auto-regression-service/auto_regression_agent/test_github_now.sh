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
echo "🐙 Testing GitHub Integration"
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

# Configuration - Read from environment variables
PAT_TOKEN="${GITHUB_PAT_TOKEN}"
GITHUB_URL="${GITHUB_URL:-https://github.com/swagger-api/swagger-petstore}"
BRANCH="${GITHUB_BRANCH:-master}"

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
echo "=========echo "=======echo "===echo "=========ec-eecho "========= Uecho "=========echRL"
echo -e "${BLUE}echo -e "${BLUE}echo -echo -e "${BLUE}PAT Token:${NC} ${PAT_TOKEN:0:20}..."
echo ""

read -p "Press Enter to starread -p "Press Enter to starread -p "Press Ente"========================================="
echo "Triggering GitHub Integration Test..."
echo "========================================="
echo ""

# Mak# Mak# Mak# Mak# Mak# Mak# Mak# Mak# Mak# Mak# Makocalhost:8080/# Mak# Mak# Mak# Mak# Mak# Mak# Ma-T# Mak# Mak# Mak# Mak# Mak# Mak# Mak# \"github_url\#: # Mak# Mak# Mak# Mak# Mak# Mak# Mak# Mak# MaOKEN\",
    \"branch\": \"$B    \"branch\": \"$B    \"branch\": \"$B    \"branch\": \"$B    \"branch\": \"$B    \"branch\": \"$en
    echo -e "${GREEN}✅ Test completed successfully!${NC}"
    echo ""
    
    # Extract workflow ID
    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKn.    WORKFLOW_I).get('workflow_id', 'N/A'))" 2>/dev/null || echo "N    WORKFLOW_ID=$    WORK$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('spec_id', 'N/A'))" 2>/    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKn.    WORKFLOW-e    WLUE}    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKFLOW_ID=$    WORKn.    WORKFLOW_I).get('workflow_id', 'N/A=="
    echo "Test Results:"
    echo "========    echo "========    echo "========    echo "=ESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
    
else
    echo -e "${RED}❌ Test failed or spec not found${NC}"
    echo ""
    echo "Response:"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
    echo ""
    echo -e "${YELLOW}Note: The repository might not have openAPISpec.json/yml in the root${NC}"
    echo ""
    echo "To test with your own reposito   "
    ech  "      ech  "      ech  "  as openAPISpec.json or openAPISpec.yml in the root"
    echo "  2. Edit this script and change GITHUB_URL to your repository"
    echo "  3. Run the script again"
fi

echo ""
echo "========================================="
echo "Done!"
echo "========================================="
