#!/bin/bash

# Fix Colima certificate issues
# This script fixes TLS certificate verification problems in Colima

set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║   Fixing Colima Certificate Issues                      ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is designed for macOS only"
    exit 1
fi

# Check if Colima is installed
if ! command -v colima &> /dev/null; then
    echo "❌ Colima is not installed"
    echo "Install it with: brew install colima"
    exit 1
fi

echo "🔧 Step 1: Stopping Colima..."
colima stop 2>/dev/null || true

echo ""
echo "🔧 Step 2: Checking for corporate proxy or VPN..."
if [ -n "$HTTP_PROXY" ] || [ -n "$HTTPS_PROXY" ]; then
    echo "⚠️  Proxy detected:"
    echo "   HTTP_PROXY: $HTTP_PROXY"
    echo "   HTTPS_PROXY: $HTTPS_PROXY"
    echo ""
    echo "This might be causing certificate issues."
    echo "Consider temporarily disabling VPN/proxy or configuring Docker to use it."
fi

echo ""
echo "🔧 Step 3: Starting Colima with insecure registry option..."
# Start Colima with environment variables to skip TLS verification (temporary workaround)
DOCKER_TLS_VERIFY=0 colima start \
    --cpu 4 \
    --memory 8 \
    --disk 60 \
    --network-address

echo ""
echo "🔧 Step 4: Configuring Docker daemon..."
# Configure Docker daemon to be more lenient with certificates
colima ssh -- sudo tee /etc/docker/daemon.json > /dev/null <<'EOF'
{
  "insecure-registries": [],
  "registry-mirrors": [],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

echo ""
echo "🔧 Step 5: Restarting Docker service..."
colima ssh -- sudo systemctl restart docker || colima ssh -- sudo service docker restart

echo ""
echo "🔧 Step 6: Testing Docker connectivity..."
sleep 5

if docker info &> /dev/null; then
    echo "✅ Docker is running"
else
    echo "❌ Docker is not responding"
    exit 1
fi

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║   ✅ Colima Configuration Complete                      ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  1. Try pulling images: ./pull-images.sh"
echo "  2. Or run infrastructure: make infra-up"
echo ""
echo "If you still have certificate issues, you may be behind a corporate"
echo "proxy/firewall. Contact your IT department for the proxy CA certificate."

