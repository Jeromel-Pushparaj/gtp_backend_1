#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================="
echo "🔍 Testing Network Connectivity"
echo "========================================="
echo ""

API_HOST="10.140.16.78"
API_PORT="8080"
API_URL="http://10.140.16.78:8080/api/v1"

echo -e "${BLUE}Target Server:${NC} $API_URL"
echo ""

echo "========================================="
echo "Step 1: Ping Test"
echo "========================================="
echo ""

echo "Pinging $API_HOST..."
if ping -c 3 $API_HOST > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Host is reachable via ping${NC}"
else
    echo -e "${RED}❌ Host is NOT reachable via ping${NC}"
    echo ""
    echo "Possible issues:"
    echo "  - Host is down"
    echo "  - Firewall blocking ICMP"
    echo "  - Wrong IP address"
    echo "  - Not on same network/VPN"
fi

echo ""
echo "========================================="
echo "Step 2: Port Test"
echo "========================================="
echo ""

echo "Testing port $API_PORT on $API_HOST..."
if nc -z -w 5 $API_HOST $API_PORT 2>/dev/null; then
    echo -e "${GREEN}✅ Port $API_PORT is open${NC}"
else
    echo -e "${RED}❌ Port $API_PORT is NOT accessible${NC}"
    echo ""
    echo "Possible issues:"
    echo "  - Server not running on port $API_PORT"
    echo "  - Firewall blocking port $API_PORT"
    echo "  - Wrong port number"
fi

echo ""
echo "========================================="
echo "Step 3: HTTP Test"
echo "========================================="
echo ""

echo "Testing HTTP connection to $API_URL..."
HTTP_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$API_URL" 2>&1)

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ HTTP connection successful (Status: $HTTP_RESPONSE)${NC}"
else
    echo -e "${RED}❌ HTTP connection failed${NC}"
    echo "Error: $HTTP_RESPONSE"
fi

echo ""
echo "========================================="
echo "Step 4: Test Specific Endpoint"
echo "========================================="
echo ""

echo "Testing GET $API_URL/employees..."
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" --connect-timeout 5 "$API_URL/employees" 2>&1)

if echo "$RESPONSE" | grep -q "HTTP_CODE:"; then
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE:/d')
    
    echo -e "${GREEN}✅ Endpoint is accessible (HTTP $HTTP_CODE)${NC}"
    echo ""
    echo "Response preview:"
    echo "$BODY" | head -20
else
    echo -e "${RED}❌ Cannot connect to endpoint${NC}"
    echo "Error: $RESPONSE"
fi

echo ""
echo "========================================="
echo "Step 5: Network Route Check"
echo "========================================="
echo ""

echo "Checking route to $API_HOST..."
echo ""
traceroute -m 5 $API_HOST 2>/dev/null || echo "traceroute not available"

echo ""
echo "========================================="
echo "Diagnostics Summary"
echo "========================================="
echo ""

echo "If all tests pass:"
echo "  ✅ Your network connection is working"
echo "  ✅ The API server is accessible"
echo "  ✅ Tests should work"
echo ""

echo "If tests fail, check:"
echo "  1. Are you on the same network/VPN as the server?"
echo "  2. Is the server running on 10.140.16.78:8080?"
echo "  3. Is there a firewall blocking the connection?"
echo "  4. Can you access the API from a browser?"
echo "     Try: http://10.140.16.78:8080/api/v1/employees"
echo ""

echo "========================================="
echo "Quick Fixes"
echo "========================================="
echo ""

echo "1. Test from browser:"
echo "   Open: http://10.140.16.78:8080/api/v1/employees"
echo ""

echo "2. Check if you're on VPN (if required):"
echo "   ifconfig | grep -A 1 'utun\\|tun'"
echo ""

echo "3. Check your IP address:"
echo "   ifconfig | grep 'inet ' | grep -v 127.0.0.1"
echo ""

echo "4. Test with curl directly:"
echo "   curl -v http://10.140.16.78:8080/api/v1/employees"
echo ""

echo "========================================="
echo "Done!"
echo "========================================="

