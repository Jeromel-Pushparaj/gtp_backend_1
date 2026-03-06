#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "🤖 Starting Autonomous AI Agents"
echo "========================================="
echo ""

# Check if Redis is running
echo "Checking Redis connection..."
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${RED}❌ Redis is not running!${NC}"
    echo ""
    echo "Start Redis with:"
    echo "  brew services start redis"
    echo "  # OR"
    echo "  redis-server"
    exit 1
fi
echo -e "${GREEN}✓ Redis is running${NC}"
echo ""

# Check if GROQ_API_KEY is set
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | grep GROQ_API_KEY | xargs)
fi

if [ -z "$GROQ_API_KEY" ]; then
    echo -e "${RED}❌ GROQ_API_KEY not set!${NC}"
    echo ""
    echo "Set it in .env file or export it:"
    echo "  export GROQ_API_KEY=your_key_here"
    exit 1
fi
echo -e "${GREEN}✓ GROQ_API_KEY is set${NC}"
echo ""

# Check if backend is running
echo "Checking if backend is running..."
if ! curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Backend is not running${NC}"
    echo ""
    echo "Start the backend first:"
    echo "  go run cmd/server/main.go"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ Backend is running${NC}"
fi
echo ""

# Start agents
echo "========================================="
echo "Starting All Agents..."
echo "========================================="
echo ""

echo -e "${BLUE}This will start all 5 AI agents:${NC}"
echo "  1. Discovery Agent  - Analyzes OpenAPI specs"
echo "  2. Designer Agent   - Creates test strategies"
echo "  3. Payload Agent    - Generates test data"
echo "  4. Executor Agent   - Executes HTTP tests"
echo "  5. Analyzer Agent   - Analyzes results"
echo ""

echo -e "${YELLOW}Note: This will run in the foreground. Press Ctrl+C to stop.${NC}"
echo ""

read -p "Press Enter to start agents..."

echo ""
echo "🚀 Starting autonomous agents..."
echo ""

# Run all agents
go run cmd/autonomous-worker/main.go -agent all

echo ""
echo "Agents stopped."

