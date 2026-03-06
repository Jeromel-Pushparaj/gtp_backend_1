#!/bin/bash

# Start Auto-Regression Services Without Docker
# This script starts both the Agent API (port 8080) and Auto-Regression API (port 8092)

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   Starting Auto-Regression Services (No Docker)       ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if ports are already in use
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Kill process on port
kill_port() {
    local port=$1
    local pid=$(lsof -ti:$port)
    if [ ! -z "$pid" ]; then
        echo -e "${YELLOW}Killing process on port $port (PID: $pid)${NC}"
        kill -9 $pid 2>/dev/null
        sleep 1
    fi
}

# Check and clean ports
echo -e "${BLUE}Checking ports...${NC}"

if check_port 8080; then
    echo -e "${YELLOW}⚠️  Port 8080 is already in use${NC}"
    read -p "Kill the process and continue? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kill_port 8080
    else
        echo -e "${RED}Cannot start Agent API on port 8080${NC}"
        exit 1
    fi
fi

if check_port 8092; then
    echo -e "${YELLOW}⚠️  Port 8092 is already in use${NC}"
    read -p "Kill the process and continue? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kill_port 8092
    else
        echo -e "${RED}Cannot start Auto-Regression API on port 8092${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}✓ Ports are available${NC}"
echo ""

# Create log directory
mkdir -p logs

# Start Agent API (port 8080)
echo -e "${BLUE}Starting Agent API on port 8080...${NC}"
cd auto_regression_agent
go run cmd/server/main.go > ../logs/agent-api.log 2>&1 &
AGENT_PID=$!
cd ..

echo -e "${GREEN}✓ Agent API started (PID: $AGENT_PID)${NC}"
sleep 3

# Check if agent started successfully
if ! check_port 8080; then
    echo -e "${RED}✗ Failed to start Agent API${NC}"
    echo "Check logs: tail -f logs/agent-api.log"
    kill $AGENT_PID 2>/dev/null
    exit 1
fi

# Start Auto-Regression API (port 8092)
echo -e "${BLUE}Starting Auto-Regression API on port 8092...${NC}"
cd auto_regression_api
PORT=8092 go run main.go > ../logs/auto-regression-api.log 2>&1 &
API_PID=$!
cd ..

echo -e "${GREEN}✓ Auto-Regression API started (PID: $API_PID)${NC}"
sleep 2

# Check if API started successfully
if ! check_port 8092; then
    echo -e "${RED}✗ Failed to start Auto-Regression API${NC}"
    echo "Check logs: tail -f logs/auto-regression-api.log"
    kill $AGENT_PID $API_PID 2>/dev/null
    exit 1
fi

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║              Services Started Successfully!            ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✓ Agent API:           http://localhost:8080${NC}"
echo -e "${GREEN}✓ Auto-Regression API: http://localhost:8092${NC}"
echo ""
echo -e "${BLUE}Process IDs:${NC}"
echo "  • Agent API:           PID $AGENT_PID"
echo "  • Auto-Regression API: PID $API_PID"
echo ""
echo -e "${BLUE}Logs:${NC}"
echo "  • Agent API:           tail -f logs/agent-api.log"
echo "  • Auto-Regression API: tail -f logs/auto-regression-api.log"
echo ""
echo -e "${YELLOW}To stop services:${NC}"
echo "  kill $AGENT_PID $API_PID"
echo "  or run: ./stop-services.sh"
echo ""
echo -e "${YELLOW}To test:${NC}"
echo "  ./quick-test.sh"
echo "  ./verify-endpoints.sh"
echo ""

# Save PIDs to file for easy cleanup
echo "$AGENT_PID" > logs/agent-api.pid
echo "$API_PID" > logs/auto-regression-api.pid

echo -e "${GREEN}Services are running in the background!${NC}"
echo ""

