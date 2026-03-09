#!/bin/bash

# Diagnostic script to test chat-agent <-> MCP server connection
# This script verifies that both services can communicate properly

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   Chat Agent <-> MCP Server Connection Diagnostic     ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Configuration
CHAT_AGENT_URL="http://localhost:8082"
MCP_SERVER_URL="http://localhost:8083"
BACKEND_API_URL="http://localhost:8080"

ERRORS=0
WARNINGS=0

# Function to test endpoint
test_endpoint() {
    local url=$1
    local description=$2
    local expected_status=${3:-200}
    
    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${YELLOW}  URL:${NC} $url"
    
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$url" 2>&1)
    http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
    body=$(echo "$response" | grep -v "HTTP_CODE")
    
    if [ -z "$http_code" ]; then
        echo -e "${RED}  ✗ FAILED - Cannot connect${NC}"
        echo -e "${RED}    Error: Connection refused or timeout${NC}"
        ((ERRORS++))
        return 1
    elif [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}  ✓ SUCCESS${NC} (HTTP $http_code)"
        echo "    Response: ${body:0:100}..."
        return 0
    else
        echo -e "${RED}  ✗ FAILED${NC} (HTTP $http_code, expected $expected_status)"
        echo "    Response: $body"
        ((ERRORS++))
        return 1
    fi
}

# Function to test POST endpoint
test_post_endpoint() {
    local url=$1
    local data=$2
    local description=$3
    local expected_status=${4:-200}
    
    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${YELLOW}  URL:${NC} $url"
    echo -e "${YELLOW}  Data:${NC} $data"
    
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$data" 2>&1)
    
    http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
    body=$(echo "$response" | grep -v "HTTP_CODE")
    
    if [ -z "$http_code" ]; then
        echo -e "${RED}  ✗ FAILED - Cannot connect${NC}"
        echo -e "${RED}    Error: Connection refused or timeout${NC}"
        ((ERRORS++))
        return 1
    elif [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}  ✓ SUCCESS${NC} (HTTP $http_code)"
        echo "    Response: ${body:0:200}..."
        return 0
    else
        echo -e "${RED}  ✗ FAILED${NC} (HTTP $http_code, expected $expected_status)"
        echo "    Response: $body"
        ((ERRORS++))
        return 1
    fi
}

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}STEP 1: Check if all services are running${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

# Test 1: Backend API
test_endpoint "$BACKEND_API_URL/health" "Backend API (port 8080)" 200
BACKEND_STATUS=$?
echo ""

# Test 2: MCP Server
test_endpoint "$MCP_SERVER_URL/health" "MCP Server (port 8083)" 200
MCP_STATUS=$?
echo ""

# Test 3: Chat Agent
test_endpoint "$CHAT_AGENT_URL/health" "Chat Agent Service (port 8082)" 200
CHAT_STATUS=$?
echo ""

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}STEP 2: Test MCP Server tool execution${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

if [ $MCP_STATUS -eq 0 ]; then
    # Test MCP server can execute a simple tool
    test_post_endpoint "$MCP_SERVER_URL/execute" \
        '{"tool":"health_check","arguments":{}}' \
        "MCP Server - Execute health_check tool" 200
    echo ""
else
    echo -e "${YELLOW}  ⚠ Skipping - MCP Server not running${NC}"
    ((WARNINGS++))
    echo ""
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}STEP 3: Test Chat Agent -> MCP Server communication${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

if [ $CHAT_STATUS -eq 0 ]; then
    # Test chat agent with a question that requires MCP tool
    test_post_endpoint "$CHAT_AGENT_URL/api/v1/chat" \
        '{"message":"What is the health status of the backend?"}' \
        "Chat Agent - Question requiring MCP tool" 200
    echo ""
else
    echo -e "${YELLOW}  ⚠ Skipping - Chat Agent not running${NC}"
    ((WARNINGS++))
    echo ""
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}STEP 4: Check environment configuration${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${BLUE}Checking environment variables...${NC}"

# Check if .env file exists
if [ -f .env ]; then
    echo -e "${GREEN}  ✓ .env file found${NC}"

    # Check for required variables
    if grep -q "GROQ_API_KEY" .env; then
        echo -e "${GREEN}  ✓ GROQ_API_KEY is set in .env${NC}"
    else
        echo -e "${RED}  ✗ GROQ_API_KEY not found in .env${NC}"
        ((ERRORS++))
    fi

    if grep -q "MCP_SERVER_URL" .env; then
        MCP_URL=$(grep "MCP_SERVER_URL" .env | cut -d= -f2)
        echo -e "${GREEN}  ✓ MCP_SERVER_URL is set: $MCP_URL${NC}"
        if [ "$MCP_URL" != "$MCP_SERVER_URL" ]; then
            echo -e "${YELLOW}  ⚠ Warning: MCP_SERVER_URL in .env ($MCP_URL) differs from expected ($MCP_SERVER_URL)${NC}"
            ((WARNINGS++))
        fi
    else
        echo -e "${YELLOW}  ⚠ MCP_SERVER_URL not set in .env (will use default: $MCP_SERVER_URL)${NC}"
        ((WARNINGS++))
    fi
else
    echo -e "${YELLOW}  ⚠ No .env file found${NC}"
    echo -e "${YELLOW}    Services will use default configuration${NC}"
    ((WARNINGS++))
fi
echo ""

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}STEP 5: Test end-to-end flow${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

if [ $CHAT_STATUS -eq 0 ] && [ $MCP_STATUS -eq 0 ] && [ $BACKEND_STATUS -eq 0 ]; then
    echo -e "${BLUE}Testing complete flow: Chat Agent -> MCP Server -> Backend API${NC}"
    echo ""

    # Test with a question that requires fetching data
    echo -e "${YELLOW}Sending question: 'List all organizations'${NC}"
    response=$(curl -s -X POST "$CHAT_AGENT_URL/api/v1/chat" \
        -H "Content-Type: application/json" \
        -d '{"message":"List all organizations"}')

    if echo "$response" | grep -q '"status":"success"'; then
        echo -e "${GREEN}  ✓ Chat Agent responded successfully${NC}"

        # Check if response contains actual data
        if echo "$response" | grep -q '"response"'; then
            response_text=$(echo "$response" | grep -o '"response":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}  ✓ Response contains data${NC}"
            echo "    Preview: ${response_text:0:150}..."

            # Check if it says "can't fetch" or similar error
            if echo "$response_text" | grep -qi "can't fetch\|cannot fetch\|unable to fetch\|failed to fetch"; then
                echo -e "${RED}  ✗ Response indicates data fetch failure${NC}"
                echo -e "${RED}    This means Chat Agent cannot communicate with MCP Server properly${NC}"
                ((ERRORS++))
            else
                echo -e "${GREEN}  ✓ Data fetched successfully - Connection is working!${NC}"
            fi
        else
            echo -e "${RED}  ✗ Response missing data${NC}"
            ((ERRORS++))
        fi
    else
        echo -e "${RED}  ✗ Chat Agent returned error${NC}"
        echo "    Response: $response"
        ((ERRORS++))
    fi
    echo ""
else
    echo -e "${YELLOW}  ⚠ Skipping - Not all services are running${NC}"
    echo -e "${YELLOW}    Required: Backend API (8080), MCP Server (8083), Chat Agent (8082)${NC}"
    ((WARNINGS++))
    echo ""
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}DIAGNOSTIC SUMMARY${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${BLUE}Service Status:${NC}"
if [ $BACKEND_STATUS -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} Backend API (port 8080) - Running"
else
    echo -e "  ${RED}✗${NC} Backend API (port 8080) - Not running"
fi

if [ $MCP_STATUS -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} MCP Server (port 8083) - Running"
else
    echo -e "  ${RED}✗${NC} MCP Server (port 8083) - Not running"
fi

if [ $CHAT_STATUS -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} Chat Agent (port 8082) - Running"
else
    echo -e "  ${RED}✗${NC} Chat Agent (port 8082) - Not running"
fi

echo ""
echo -e "${BLUE}Results:${NC}"
echo -e "  Errors: ${RED}$ERRORS${NC}"
echo -e "  Warnings: ${YELLOW}$WARNINGS${NC}"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ✓ ALL TESTS PASSED - Connection is working!          ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║  ⚠ Tests passed with warnings                         ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  ✗ TESTS FAILED - Connection issues detected          ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Troubleshooting Steps:${NC}"
    echo ""

    if [ $BACKEND_STATUS -ne 0 ]; then
        echo -e "${YELLOW}1. Start Backend API:${NC}"
        echo "   cd sonar-shell-test"
        echo "   go run main.go -server -port=8080"
        echo ""
    fi

    if [ $MCP_STATUS -ne 0 ]; then
        echo -e "${YELLOW}2. Start MCP Server:${NC}"
        echo "   cd services/chat-agent-service"
        echo "   make run-mcp"
        echo "   # OR"
        echo "   MCP_PORT=8083 API_BASE_URL=http://localhost:8080 go run cmd/main.go"
        echo ""
    fi

    if [ $CHAT_STATUS -ne 0 ]; then
        echo -e "${YELLOW}3. Start Chat Agent:${NC}"
        echo "   cd services/chat-agent-service"
        echo "   make run-http"
        echo "   # OR"
        echo "   GROQ_API_KEY=your_key MCP_SERVER_URL=http://localhost:8083 ./chat-agent-server"
        echo ""
    fi

    echo -e "${YELLOW}4. Check logs for errors:${NC}"
    echo "   - MCP Server logs: Check terminal where MCP server is running"
    echo "   - Chat Agent logs: Check terminal where chat agent is running"
    echo ""

    echo -e "${YELLOW}5. Verify environment variables:${NC}"
    echo "   - GROQ_API_KEY must be set for Chat Agent"
    echo "   - MCP_SERVER_URL should be http://localhost:8083"
    echo "   - API_BASE_URL should be http://localhost:8080"
    echo ""

    exit 1
fi

