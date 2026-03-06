#!/bin/bash

# Script to run the middleman service and test it

echo "=== Starting Middleman Service ==="
echo ""

# Kill any existing middleman process
pkill -f "go run main.go" 2>/dev/null
pkill -f "./middleman" 2>/dev/null
sleep 1

# Start the service in the background
echo "Starting service on port 8081..."
./middleman > middleman.log 2>&1 &
SERVICE_PID=$!

echo "Service PID: $SERVICE_PID"
echo "Logs: tail -f middleman.log"
echo ""

# Wait for service to start
sleep 2

# Check if service is running
if ! kill -0 $SERVICE_PID 2>/dev/null; then
    echo "❌ Service failed to start. Check middleman.log"
    cat middleman.log
    exit 1
fi

# Test health endpoint
echo "Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8081/health)
if [ $? -eq 0 ]; then
    echo "✅ Service is healthy"
    echo "$HEALTH" | jq '.'
else
    echo "❌ Service is not responding"
    kill $SERVICE_PID
    exit 1
fi

echo ""
echo "=== Service is ready! ==="
echo ""
echo "To test, run:"
echo "  ./example-request.sh"
echo ""
echo "To view logs:"
echo "  tail -f middleman.log"
echo ""
echo "To stop the service:"
echo "  kill $SERVICE_PID"
echo ""
echo "Service PID: $SERVICE_PID"

