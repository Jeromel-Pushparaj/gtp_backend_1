#!/bin/bash

# Stop Auto-Regression Services

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}Stopping Auto-Regression Services...${NC}"
echo ""

# Function to kill process
kill_process() {
    local pid=$1
    local name=$2
    
    if [ ! -z "$pid" ] && kill -0 $pid 2>/dev/null; then
        echo -e "${YELLOW}Stopping $name (PID: $pid)...${NC}"
        kill $pid 2>/dev/null
        sleep 1
        
        # Force kill if still running
        if kill -0 $pid 2>/dev/null; then
            echo -e "${YELLOW}Force stopping $name...${NC}"
            kill -9 $pid 2>/dev/null
        fi
        
        echo -e "${GREEN}✓ $name stopped${NC}"
    else
        echo -e "${YELLOW}$name is not running${NC}"
    fi
}

# Read PIDs from files
if [ -f logs/agent-api.pid ]; then
    AGENT_PID=$(cat logs/agent-api.pid)
    kill_process $AGENT_PID "Agent API"
    rm -f logs/agent-api.pid
fi

if [ -f logs/auto-regression-api.pid ]; then
    API_PID=$(cat logs/auto-regression-api.pid)
    kill_process $API_PID "Auto-Regression API"
    rm -f logs/auto-regression-api.pid
fi

# Also kill by port if PIDs don't work
echo ""
echo -e "${BLUE}Checking ports...${NC}"

# Kill process on port 8080
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    PID=$(lsof -ti:8080)
    echo -e "${YELLOW}Found process on port 8080 (PID: $PID)${NC}"
    kill -9 $PID 2>/dev/null
    echo -e "${GREEN}✓ Killed process on port 8080${NC}"
fi

# Kill process on port 8092
if lsof -Pi :8092 -sTCP:LISTEN -t >/dev/null 2>&1; then
    PID=$(lsof -ti:8092)
    echo -e "${YELLOW}Found process on port 8092 (PID: $PID)${NC}"
    kill -9 $PID 2>/dev/null
    echo -e "${GREEN}✓ Killed process on port 8092${NC}"
fi

echo ""
echo -e "${GREEN}✓ All services stopped${NC}"
echo ""

