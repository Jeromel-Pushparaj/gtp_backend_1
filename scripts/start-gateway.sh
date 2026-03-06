#!/bin/bash

# 🚀 Start API Gateway for Network Access
# This script starts the API Gateway so friends on your local network can access all services

echo "════════════════════════════════════════════════════════════"
echo "🚀 Starting API Gateway for Network Access"
echo "════════════════════════════════════════════════════════════"
echo ""

# Get local IP address
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    LOCAL_IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo "Unable to detect IP")
else
    # Linux
    LOCAL_IP=$(hostname -I | awk '{print $1}' || echo "Unable to detect IP")
fi

echo "📍 Your Local IP Address: $LOCAL_IP"
echo ""
echo "🌐 Gateway will be accessible at:"
echo "   - Local:   http://localhost:8089"
echo "   - Network: http://$LOCAL_IP:8089"
echo ""
echo "📋 Available Routes:"
echo "   - Health:      http://$LOCAL_IP:8089/health"
echo "   - Jira:        http://$LOCAL_IP:8089/jira/*"
echo "   - Chat:        http://$LOCAL_IP:8089/chat/*"
echo "   - Approval:    http://$LOCAL_IP:8089/approval/*"
echo "   - Onboarding:  http://$LOCAL_IP:8089/onboarding/*"
echo "   - ScoreCard:   http://$LOCAL_IP:8089/scorecard/*"
echo "   - Sonar:       http://$LOCAL_IP:8089/sonar/*"
echo ""
echo "════════════════════════════════════════════════════════════"
echo "💡 Share this URL with your friends on the same WiFi:"
echo "   http://$LOCAL_IP:8089"
echo "════════════════════════════════════════════════════════════"
echo ""

# Navigate to gateway directory
cd gateway/api-gateway || {
    echo "❌ Error: gateway/api-gateway directory not found"
    exit 1
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Start the gateway
echo "🚀 Starting gateway..."
echo ""
go run main.go

