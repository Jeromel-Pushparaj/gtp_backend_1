#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "🧹 Cleanup and Restart"
echo "========================================="
echo ""

# Clean old output files
echo "Cleaning old output files..."
rm -rf ./output/discovery/*
rm -rf ./output/strategy/*
rm -rf ./output/payloads/*
rm -rf ./output/plans/*
rm -rf ./output/results/*
rm -rf ./output/specs/*
rm -rf ./test_suites/*

echo -e "${GREEN}✓ Output directories cleaned${NC}"
echo ""

# Clear Redis (optional)
echo "Clearing Redis cache..."
if redis-cli ping > /dev/null 2>&1; then
    redis-cli FLUSHDB > /dev/null 2>&1
    echo -e "${GREEN}✓ Redis cache cleared${NC}"
else
    echo -e "${YELLOW}⚠️  Redis not running (skipped)${NC}"
fi
echo ""

echo "========================================="
echo "Configuration Updated:"
echo "========================================="
echo ""
echo -e "${BLUE}Model: llama-3.1-8b-instant${NC} (fastest, best rate limits)"
echo -e "${BLUE}Max Tokens: 1024${NC} (reduced from 8192)"
echo ""

echo "========================================="
echo "Next Steps:"
echo "========================================="
echo ""
echo "1. Stop the backend (Ctrl+C in Terminal 1)"
echo "2. Stop the agents (Ctrl+C in Terminal 2)"
echo ""
echo "3. Restart backend:"
echo -e "   ${GREEN}go run cmd/server/main.go${NC}"
echo ""
echo "4. Restart agents:"
echo -e "   ${GREEN}./start_agents.sh${NC}"
echo ""
echo "5. Run new test:"
echo -e "   ${GREEN}./run_full_test.sh${NC}"
echo ""

