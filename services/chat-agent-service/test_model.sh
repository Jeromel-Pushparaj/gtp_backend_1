#!/bin/bash

# Test script to verify Groq model configuration
# This script checks what model is being used and tests it

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         Groq Model Configuration Test                 ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}✗ .env file not found${NC}"
    echo ""
    echo "Please create a .env file with:"
    echo "  GROQ_API_KEY=your_api_key_here"
    echo "  GROQ_MODEL=meta-llama/llama-4-maverick-17b-128e-instruct"
    echo ""
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

echo -e "${BLUE}Step 1: Check environment configuration${NC}"
echo ""

# Check GROQ_API_KEY
if [ -z "$GROQ_API_KEY" ]; then
    echo -e "${RED}✗ GROQ_API_KEY not set${NC}"
    exit 1
else
    echo -e "${GREEN}✓ GROQ_API_KEY is set${NC}"
    echo "  Value: ${GROQ_API_KEY:0:8}...${GROQ_API_KEY: -4}"
fi

# Check GROQ_MODEL
if [ -z "$GROQ_MODEL" ]; then
    echo -e "${YELLOW}⚠ GROQ_MODEL not set (will use default)${NC}"
    echo "  Default: meta-llama/llama-4-maverick-17b-128e-instruct"
    GROQ_MODEL="meta-llama/llama-4-maverick-17b-128e-instruct"
else
    echo -e "${GREEN}✓ GROQ_MODEL is set${NC}"
    echo "  Value: $GROQ_MODEL"
fi

echo ""
echo -e "${BLUE}Step 2: Check if chat-agent service is running${NC}"
echo ""

if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Chat Agent service is running on port 8082${NC}"
    
    # Check logs for model initialization
    if [ -f /tmp/chat-agent.log ]; then
        echo ""
        echo -e "${BLUE}Step 3: Check service logs for model initialization${NC}"
        echo ""
        
        MODEL_LOG=$(grep "Initializing Groq client with model" /tmp/chat-agent.log | tail -1)
        if [ -n "$MODEL_LOG" ]; then
            echo -e "${GREEN}✓ Found model initialization in logs:${NC}"
            echo "  $MODEL_LOG"
            
            # Extract model name from log
            ACTUAL_MODEL=$(echo "$MODEL_LOG" | grep -o 'model: .*' | cut -d' ' -f2)
            
            if [ "$ACTUAL_MODEL" = "$GROQ_MODEL" ]; then
                echo -e "${GREEN}✓ Service is using the correct model!${NC}"
            else
                echo -e "${YELLOW}⚠ Model mismatch!${NC}"
                echo "  Expected: $GROQ_MODEL"
                echo "  Actual:   $ACTUAL_MODEL"
                echo ""
                echo -e "${YELLOW}  The service may have been started before .env was updated.${NC}"
                echo -e "${YELLOW}  Restart the service to use the new model.${NC}"
            fi
        else
            echo -e "${YELLOW}⚠ Model initialization not found in logs${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ Log file not found at /tmp/chat-agent.log${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Step 4: Test the model with a simple query${NC}"
    echo ""
    
    echo -e "${YELLOW}Sending test message: 'Hello, what model are you?'${NC}"
    
    RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/chat \
        -H "Content-Type: application/json" \
        -d '{"message":"Hello, what model are you?"}')
    
    if echo "$RESPONSE" | grep -q '"status":"success"'; then
        echo -e "${GREEN}✓ Model responded successfully${NC}"
        
        # Extract response text
        RESPONSE_TEXT=$(echo "$RESPONSE" | jq -r '.response' 2>/dev/null)
        if [ -n "$RESPONSE_TEXT" ] && [ "$RESPONSE_TEXT" != "null" ]; then
            echo ""
            echo -e "${CYAN}Response:${NC}"
            echo "$RESPONSE_TEXT" | fold -w 70 -s | sed 's/^/  /'
        fi
    else
        echo -e "${RED}✗ Model failed to respond${NC}"
        echo "  Response: $RESPONSE"
    fi
    
else
    echo -e "${RED}✗ Chat Agent service is not running${NC}"
    echo ""
    echo "Please start the service first:"
    echo "  ./fix_connection.sh"
    echo "  # OR"
    echo "  GROQ_MODEL=$GROQ_MODEL GROQ_API_KEY=$GROQ_API_KEY ./chat-agent-server"
    exit 1
fi

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                  Test Complete                         ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${BLUE}Configuration Summary:${NC}"
echo "  Model (configured): $GROQ_MODEL"
if [ -n "$ACTUAL_MODEL" ]; then
    echo "  Model (actual):     $ACTUAL_MODEL"
fi
echo ""

echo -e "${BLUE}Available Models:${NC}"
echo "  • meta-llama/llama-4-maverick-17b-128e-instruct (Default, 6K TPM)"
echo "  • meta-llama/llama-4-scout-17b-16e-instruct (6K TPM)"
echo "  • llama-3.3-70b-versatile (30K TPM - Higher rate limit)"
echo "  • llama-3.1-8b-instant (30K TPM - Fastest)"
echo ""

echo -e "${BLUE}To change the model:${NC}"
echo "  1. Edit .env file: nano .env"
echo "  2. Set GROQ_MODEL=your_desired_model"
echo "  3. Restart service: ./fix_connection.sh"
echo ""

echo -e "${BLUE}Documentation:${NC}"
echo "  See MODEL_CONFIGURATION.md for complete guide"
echo ""

