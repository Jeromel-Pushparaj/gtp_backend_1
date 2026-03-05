#!/bin/bash

# ═══════════════════════════════════════════════════════════════
# GTP Backend Platform - Complete Setup Script
# ═══════════════════════════════════════════════════════════════

set -e  # Exit on error

echo "╔══════════════════════════════════════════════════════════╗"
echo "║   GTP Backend Platform - Automated Setup                ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ═══════════════════════════════════════════════════════════════
# Step 1: Fix Colima Docker Environment
# ═══════════════════════════════════════════════════════════════

echo -e "${YELLOW}Step 1: Setting up Docker environment (Colima)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if Colima is installed
if ! command -v colima &> /dev/null; then
    echo -e "${RED}❌ Colima is not installed${NC}"
    echo "Install it with: brew install colima"
    exit 1
fi

# Stop and delete existing Colima instance
echo "🔄 Stopping and cleaning up existing Colima instance..."
colima stop 2>/dev/null || true
colima delete -f 2>/dev/null || true

# Start fresh Colima instance
echo "🚀 Starting fresh Colima instance..."
colima start --cpu 4 --memory 8 --disk 60

# Verify Docker is working
echo "✅ Verifying Docker..."
if docker info &> /dev/null; then
    echo -e "${GREEN}✅ Docker is running successfully${NC}"
else
    echo -e "${RED}❌ Docker is not responding${NC}"
    exit 1
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# Step 2: Setup Environment Files
# ═══════════════════════════════════════════════════════════════

echo -e "${YELLOW}Step 2: Setting up environment files${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Create .env files from examples if they don't exist
create_env_if_missing() {
    local example_file=$1
    local env_file="${example_file%.example}"
    
    if [ -f "$example_file" ]; then
        if [ ! -f "$env_file" ]; then
            cp "$example_file" "$env_file"
            echo "✅ Created: $env_file"
        else
            echo "⏭️  Already exists: $env_file"
        fi
    fi
}

# Main project .env
create_env_if_missing ".env.example"

# Service .env files
create_env_if_missing "services/jira-trigger-service/.env.example"
create_env_if_missing "services/chat-agent-service/.env.example"
create_env_if_missing "services/approval-service/.env.example"
create_env_if_missing "services/service-catelog/.env.example"
create_env_if_missing "services/score-card-service/.env.example"
create_env_if_missing "gateway/api-gateway/.env.example"
create_env_if_missing "sonar-shell-test/.env.example"

echo ""

# ═══════════════════════════════════════════════════════════════
# Step 3: Start Infrastructure Services
# ═══════════════════════════════════════════════════════════════

echo -e "${YELLOW}Step 3: Starting infrastructure services${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "🐳 Starting Docker Compose services..."
docker-compose up -d

echo "⏳ Waiting for services to be healthy (30 seconds)..."
sleep 30

echo ""

# ═══════════════════════════════════════════════════════════════
# Step 4: Verify Services
# ═══════════════════════════════════════════════════════════════

echo -e "${YELLOW}Step 4: Verifying services${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check running containers
RUNNING_CONTAINERS=$(docker-compose ps --services --filter "status=running" | wc -l)
echo "📊 Running containers: $RUNNING_CONTAINERS"

docker-compose ps

echo ""

# ═══════════════════════════════════════════════════════════════
# Setup Complete
# ═══════════════════════════════════════════════════════════════

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║                  ✅ SETUP COMPLETE!                      ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}Infrastructure Services:${NC}"
echo "  🌐 Kafka UI:      http://localhost:8090"
echo "  🗄️  PostgreSQL:   localhost:5432 (user: postgres, pass: postgres)"
echo "  🔴 Redis:         localhost:6379"
echo "  📨 Kafka:         localhost:9092"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "  1. Update .env files with your credentials"
echo "  2. Run services:"
echo "     make gateway        # API Gateway (Port 8080)"
echo "     make jira-trigger   # Jira Service (Port 8081)"
echo "     make chat-agent     # Chat Service (Port 8082)"
echo "     make approval       # Approval Service (Port 8083)"
echo ""
echo -e "${YELLOW}Useful Commands:${NC}"
echo "  make infra-logs     # View infrastructure logs"
echo "  make infra-down     # Stop infrastructure"
echo "  make help           # Show all available commands"
echo ""

