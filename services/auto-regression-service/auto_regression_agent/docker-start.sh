#!/bin/bash

# Docker Start Script - Complete System with Redis
# This script starts the entire OpenTest system using Docker Compose

set -e

echo "🐳 Starting OpenTest System with Docker..."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed!${NC}"
    echo "Please install Docker from: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed!${NC}"
    echo "Please install Docker Compose from: https://docs.docker.com/compose/install/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker daemon is not running!${NC}"
    echo ""

    # Check if Colima is installed
    if command -v colima &> /dev/null; then
        echo -e "${YELLOW}📦 Colima detected! Starting Colima...${NC}"
        colima start

        # Wait for Colima to be ready
        echo -n "Waiting for Colima to start..."
        for i in {1..30}; do
            if docker info &> /dev/null; then
                echo -e " ${GREEN}✓${NC}"
                break
            fi
            if [ $i -eq 30 ]; then
                echo -e " ${RED}✗ Timeout${NC}"
                echo "Please start Colima manually: colima start"
                exit 1
            fi
            sleep 1
            echo -n "."
        done
    else
        echo "Please start Docker Desktop, Colima, or Docker daemon"
        echo ""
        echo "To install Colima:"
        echo "  brew install colima"
        echo "  colima start"
        exit 1
    fi
fi

echo -e "${GREEN}✓ Docker is installed and running${NC}"
echo ""

# Stop any existing containers
echo -e "${BLUE}1️⃣  Stopping existing containers...${NC}"
docker-compose down 2>/dev/null || true
echo -e "${GREEN}   ✓ Stopped${NC}"
echo ""

# Clean up old data (optional - comment out if you want to preserve data)
echo -e "${BLUE}2️⃣  Cleaning up old data...${NC}"
rm -rf output/specs/* output/manifests/* output/discovery/* output/reports/* logs/* 2>/dev/null || true
echo -e "${GREEN}   ✓ Cleaned${NC}"
echo ""

# Build images
echo -e "${BLUE}3️⃣  Building Docker images...${NC}"
echo ""

# Build backend
echo -n "   Building backend... "
if docker-compose build backend 2>&1 | tee /tmp/backend-build.log | grep -q "ERROR"; then
    echo -e "${RED}✗ Failed${NC}"
    echo ""
    echo -e "${RED}Build errors:${NC}"
    cat /tmp/backend-build.log | grep -A 5 "ERROR"
    exit 1
else
    echo -e "${GREEN}✓${NC}"
fi

# Build frontend
echo -n "   Building frontend... "
if docker-compose build frontend 2>&1 | tee /tmp/frontend-build.log | grep -q "ERROR"; then
    echo -e "${RED}✗ Failed${NC}"
    echo ""
    echo -e "${RED}Build errors:${NC}"
    cat /tmp/frontend-build.log | grep -A 5 "ERROR"
    exit 1
else
    echo -e "${GREEN}✓${NC}"
fi

# Pull Redis image
echo -n "   Pulling Redis... "
docker-compose pull redis &> /dev/null
echo -e "${GREEN}✓${NC}"

echo ""
echo -e "${GREEN}   ✓ All images built successfully${NC}"
echo ""

# Start services
echo -e "${BLUE}4️⃣  Starting services...${NC}"
docker-compose up -d
if [ $? -eq 0 ]; then
    echo -e "${GREEN}   ✓ Services started${NC}"
else
    echo -e "${RED}   ✗ Failed to start services${NC}"
    exit 1
fi
echo ""

# Wait for services to be healthy
echo -e "${BLUE}5️⃣  Waiting for services to be healthy...${NC}"

# Wait for Redis
echo -n "   Redis: "
for i in {1..30}; do
    if docker-compose exec -T redis redis-cli ping &> /dev/null; then
        echo -e "${GREEN}✓ Ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ Timeout${NC}"
        exit 1
    fi
    sleep 1
done

# Wait for Backend
echo -n "   Backend: "
for i in {1..60}; do
    if curl -s http://localhost:8080/health &> /dev/null; then
        echo -e "${GREEN}✓ Ready${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${RED}✗ Timeout${NC}"
        docker-compose logs backend
        exit 1
    fi
    sleep 1
done

# Wait for Frontend
echo -n "   Frontend: "
for i in {1..30}; do
    if curl -s http://localhost:5173/health &> /dev/null; then
        echo -e "${GREEN}✓ Ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ Timeout${NC}"
        exit 1
    fi
    sleep 1
done

echo ""
echo -e "${GREEN}✅ All services are healthy!${NC}"
echo ""

# Display service information
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}📊 Service Status:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker-compose ps
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}🌐 Access URLs:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  Frontend:  ${GREEN}http://localhost:5173${NC}"
echo -e "  Backend:   ${GREEN}http://localhost:8080${NC}"
echo -e "  Redis:     ${GREEN}localhost:6379${NC}"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}📝 Quick Commands:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  View logs:        docker-compose logs -f"
echo "  View backend:     docker-compose logs -f backend"
echo "  View frontend:    docker-compose logs -f frontend"
echo "  Stop services:    docker-compose down"
echo "  Restart:          docker-compose restart"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}🧪 Next Steps:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  1. Open http://localhost:5173 in your browser"
echo "  2. Upload openAPISample.json"
echo "  3. View workflow progress"
echo "  4. Click 'View Test Cases' to see generated tests"
echo ""

echo -e "${GREEN}🚀 System is ready!${NC}"

