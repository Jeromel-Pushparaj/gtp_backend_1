#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "========================================="
echo "Testing Groq API Integration"
echo "========================================="
echo ""

# Check if GROQ_API_KEY is set
if [ -z "$GROQ_API_KEY" ]; then
    echo "❌ GROQ_API_KEY is not set in .env file"
    exit 1
fi

echo "✓ GROQ_API_KEY found: ${GROQ_API_KEY:0:10}...${GROQ_API_KEY: -10}"
echo "✓ API Base URL: ${GROQ_API_BASE:-https://api.groq.com/openai/v1}"
echo "✓ Model: ${GROQ_MODEL:-openai/gpt-oss-120b}"
echo ""

# Test 1: Simple completion
echo "========================================="
echo "Test 1: Simple Completion"
echo "========================================="
echo "Prompt: What is 2+2? Answer in one sentence."
echo ""

response=$(curl -s "https://api.groq.com/openai/v1/chat/completions" \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${GROQ_API_KEY}" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "What is 2+2? Answer in one sentence."
      }
    ],
    "model": "'"${GROQ_MODEL:-openai/gpt-oss-120b}"'",
    "temperature": 0.7,
    "max_completion_tokens": 100
  }')

# Check if response contains error
if echo "$response" | grep -q '"error"'; then
    echo "❌ API Error:"
    echo "$response" | python3 -m json.tool
    exit 1
fi

# Extract and display response
answer=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['choices'][0]['message']['content'])" 2>/dev/null)
tokens=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['usage']['total_tokens'])" 2>/dev/null)

if [ -n "$answer" ]; then
    echo "✅ Response: $answer"
    echo "   Tokens used: $tokens"
else
    echo "❌ Failed to parse response"
    echo "$response"
fi

echo ""

# Test 2: More complex prompt
echo "========================================="
echo "Test 2: Complex Prompt"
echo "========================================="
echo "Prompt: Explain what API testing is in 2 sentences."
echo ""

response2=$(curl -s "https://api.groq.com/openai/v1/chat/completions" \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${GROQ_API_KEY}" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Explain what API testing is in 2 sentences."
      }
    ],
    "model": "'"${GROQ_MODEL:-openai/gpt-oss-120b}"'",
    "temperature": 0.7,
    "max_completion_tokens": 200
  }')

# Check if response contains error
if echo "$response2" | grep -q '"error"'; then
    echo "❌ API Error:"
    echo "$response2" | python3 -m json.tool
    exit 1
fi

# Extract and display response
answer2=$(echo "$response2" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['choices'][0]['message']['content'])" 2>/dev/null)
tokens2=$(echo "$response2" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['usage']['total_tokens'])" 2>/dev/null)

if [ -n "$answer2" ]; then
    echo "✅ Response: $answer2"
    echo "   Tokens used: $tokens2"
else
    echo "❌ Failed to parse response"
    echo "$response2"
fi

echo ""
echo "========================================="
echo "✅ All Groq API tests completed!"
echo "========================================="

