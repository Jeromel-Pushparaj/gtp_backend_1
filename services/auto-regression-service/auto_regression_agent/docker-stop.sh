#!/bin/bash

# Docker Stop Script
# This script stops all OpenTest Docker containers

set -e

echo "🛑 Stopping OpenTest System..."
echo ""

# Stop and remove containers
docker-compose down

echo ""
echo "✅ All services stopped!"
echo ""
echo "To start again, run: ./docker-start.sh"

