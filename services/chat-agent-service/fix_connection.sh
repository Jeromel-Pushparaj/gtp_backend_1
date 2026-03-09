#!/bin/bash

# Automatic fix script for chat-agent <-> MCP server connection
# This script stops conflicting services and starts the correct ones

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   Chat Agent <-> MCP Server Connection Fix            ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if .env file exists and has GROQ_API_KEY
if [ ! -f .env ]; then
    echo -e "${RED}✗ .env file not found${NC}"
    echo ""
    echo "Please create a .env file with:"
    echo "  GROQ_API_KEY=your_groq_api_key_here"
    echo ""
    exit 1
fi

if ! grep -q "GROQ_API_KEY" .env; then
    echo -e "${RED}✗ GROQ_API_KEY not found in .env${NC}"
    echo ""
    echo "Please add to .env file:"
    echo "  GROQ_API_KEY=your_groq_api_key_here"
    echo ""
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

if [ -z "$GROQ_API_KEY" ]; then
    echo -e "${RED}✗ GROQ_API_KEY is empty${NC}"
    exit 1
fi

echo -e "${BLUE}Step 1: Stopping conflicting services on ports 8082 and 8083${NC}"
echo ""

# Stop services on port 8082 (chat-agent)
if lsof -ti :8082 > /dev/null 2>&1; then
    echo -e "${YELLOW}  Stopping service on port 8082...${NC}"
    lsof -ti :8082 | xargs kill -9 2>/dev/null
    sleep 1
    echo -e "${GREEN}  ✓ Port 8082 freed${NC}"
else
    echo -e "${GREEN}  ✓ Port 8082 already free${NC}"
fi

# Stop services on port 8083 (should be MCP, but might be approval-service)
if lsof -ti :8083 > /dev/null 2>&1; then
    echo -e "${YELLOW}  Stopping service on port 8083...${NC}"
    lsof -ti :8083 | xargs kill -9 2>/dev/null
    sleep 1
    echo -e "${GREEN}  ✓ Port 8083 freed${NC}"
else
    echo -e "${GREEN}  ✓ Port 8083 already free${NC}"
fi

echo ""
echo -e "${BLUE}Step 2: Starting MCP Server on port 8083${NC}"
echo ""

# Start MCP server in background
MCP_PORT=8083 API_BASE_URL=http://localhost:8080 go run cmd/main.go > /tmp/mcp-server.log 2>&1 &
MCP_PID=$!

echo -e "${YELLOW}  Waiting for MCP server to start...${NC}"
sleep 3

# Verify MCP server is running
if curl -s http://localhost:8083/health > /dev/null 2>&1; then
    SERVICE_NAME=$(curl -s http://localhost:8083/health | grep -o '"service":"[^"]*"' | cut -d'"' -f4)
    if [ "$SERVICE_NAME" = "mcp-server" ]; then
        echo -e "${GREEN}  ✓ MCP Server started successfully (PID: $MCP_PID)${NC}"
        echo "    Logs: tail -f /tmp/mcp-server.log"
    else
        echo -e "${RED}  ✗ Wrong service on port 8083: $SERVICE_NAME${NC}"
        echo "    Expected: mcp-server"
        kill $MCP_PID 2>/dev/null
        exit 1
    fi
else
    echo -e "${RED}  ✗ MCP Server failed to start${NC}"
    echo "    Check logs: tail -f /tmp/mcp-server.log"
    kill $MCP_PID 2>/dev/null
    exit 1
fi

echo ""
echo -e "${BLUE}Step 3: Starting Chat Agent on port 8082${NC}"
echo ""

# Check if chat-agent-server binary exists
if [ ! -f ./chat-agent-server ]; then
    echo -e "${YELLOW}  Building chat-agent-server...${NC}"
    make build-http
    if [ $? -ne 0 ]; then
        echo -e "${RED}  ✗ Failed to build chat-agent-server${NC}"
        kill $MCP_PID 2>/dev/null
        exit 1
    fi
fi

# Start chat agent in background
MCP_SERVER_URL=http://localhost:8083 ./chat-agent-server > /tmp/chat-agent.log 2>&1 &
CHAT_PID=$!

echo -e "${YELLOW}  Waiting for Chat Agent to start...${NC}"
sleep 3

# Verify chat agent is running
if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo -e "${GREEN}  ✓ Chat Agent started successfully (PID: $CHAT_PID)${NC}"
    echo "    Logs: tail -f /tmp/chat-agent.log"
else
    echo -e "${RED}  ✗ Chat Agent failed to start${NC}"
    echo "    Check logs: tail -f /tmp/chat-agent.log"
    kill $MCP_PID $CHAT_PID 2>/dev/null
    exit 1
fi

echo ""
echo -e "${BLUE}Step 4: Testing connection${NC}"
echo ""

# Test MCP server tool execution
echo -e "${YELLOW}  Testing MCP server tool execution...${NC}"
MCP_RESPONSE=$(curl -s -X POST http://localhost:8083/execute \
    -H "Content-Type: application/json" \
    -d '{"tool":"health_check","arguments":{}}')

if echo "$MCP_RESPONSE" | grep -q '"result"'; then
    echo -e "${GREEN}  ✓ MCP server can execute tools${NC}"
else
    echo -e "${RED}  ✗ MCP server tool execution failed${NC}"
    echo "    Response: $MCP_RESPONSE"
fi

# Test chat agent
echo -e "${YELLOW}  Testing Chat Agent...${NC}"
CHAT_RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/chat \
    -H "Content-Type: application/json" \
    -d '{"message":"What is the health status of the backend?"}')

if echo "$CHAT_RESPONSE" | grep -q '"status":"success"'; then
    echo -e "${GREEN}  ✓ Chat Agent is responding${NC}"
    
    # Check if it can fetch data
    if echo "$CHAT_RESPONSE" | grep -qi "can't fetch\|cannot fetch\|outside my domain"; then
        echo -e "${YELLOW}  ⚠ Chat Agent cannot fetch data from MCP server${NC}"
        echo "    This might be due to Groq API rate limits"
    else
        echo -e "${GREEN}  ✓ Chat Agent can communicate with MCP server${NC}"
    fi
else
    echo -e "${RED}  ✗ Chat Agent returned error${NC}"
    echo "    Response: $CHAT_RESPONSE"
fi

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                  Setup Complete!                       ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${GREEN}Services Running:${NC}"
echo "  • MCP Server (port 8083) - PID: $MCP_PID"
echo "  • Chat Agent (port 8082) - PID: $CHAT_PID"
echo ""

echo -e "${BLUE}Logs:${NC}"
echo "  • MCP Server: tail -f /tmp/mcp-server.log"
echo "  • Chat Agent: tail -f /tmp/chat-agent.log"
echo ""

echo -e "${BLUE}Test Commands:${NC}"
echo "  • Health check: curl http://localhost:8082/health"
echo "  • Chat test: curl -X POST http://localhost:8082/api/v1/chat \\"
echo "      -H 'Content-Type: application/json' \\"
echo "      -d '{\"message\":\"List all organizations\"}'"
echo "  • Full diagnostic: ./diagnose_connection.sh"
echo ""

echo -e "${BLUE}To stop services:${NC}"
echo "  kill $MCP_PID $CHAT_PID"
echo ""

# Save PIDs to file for easy cleanup
echo "$MCP_PID" > /tmp/mcp-server.pid
echo "$CHAT_PID" > /tmp/chat-agent.pid

echo -e "${GREEN}✓ All done! Services are running in the background.${NC}"
echo ""

