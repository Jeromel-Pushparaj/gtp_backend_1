#!/bin/bash

echo "=========================================="
echo "Testing Chat Agent with Detailed Logs"
echo "=========================================="
echo ""

# Check if server is running
if ! curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "❌ Server is not running on port 8082"
    echo ""
    echo "Please start the server in another terminal:"
    echo "  cd services/chat-agent-service"
    echo "  GROQ_API_KEY=your_key ./chat-agent-server"
    echo ""
    echo "Watch the server logs to see what's happening!"
    exit 1
fi

echo "✓ Server is running"
echo ""

echo "Sending test message: 'What is the health status of the backend?'"
echo ""
echo "Watch the server terminal for detailed logs..."
echo ""

curl -X POST http://localhost:8082/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is the health status of the backend?"}' \
  2>/dev/null | jq '.'

echo ""
echo "=========================================="
echo "Check the server logs to see:"
echo "  1. If Groq API was called with tools"
echo "  2. If tool calls were detected"
echo "  3. If backend API was called"
echo "  4. What the final response was"
echo "=========================================="

