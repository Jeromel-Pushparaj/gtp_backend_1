#!/bin/bash

# Kafka Startup Script
# This script ensures Kafka starts properly and waits for it to be ready

echo "=========================================="
echo "Starting Kafka Infrastructure"
echo "=========================================="
echo ""

# Navigate to project root
cd /Users/sarumathis/Desktop/slack/gtp_backend_1

# Stop any existing containers
echo "Stopping existing containers..."
docker-compose down

# Start Zookeeper first
echo ""
echo "Starting Zookeeper..."
docker-compose up -d zookeeper

# Wait for Zookeeper to be healthy
echo "Waiting for Zookeeper to be ready..."
until docker exec zookeeper nc -z localhost 2181 2>/dev/null; do
    echo -n "."
    sleep 1
done
echo ""
echo "✓ Zookeeper is ready!"

# Start Kafka
echo ""
echo "Starting Kafka..."
docker-compose up -d kafka

# Wait for Kafka to be healthy
echo "Waiting for Kafka to be ready (this may take 30-60 seconds)..."
attempt=0
max_attempts=60
until docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092 2>/dev/null | grep -q "ApiVersion"; do
    echo -n "."
    sleep 1
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo ""
        echo "ERROR: Kafka failed to start after $max_attempts seconds"
        echo "Check logs with: docker logs kafka"
        exit 1
    fi
done
echo ""
echo "✓ Kafka is ready!"

# Start Kafka UI
echo ""
echo "Starting Kafka UI..."
docker-compose up -d kafka-ui

# Wait for Kafka UI
sleep 5
echo "✓ Kafka UI is ready at http://localhost:8090"

# Start other services
echo ""
echo "Starting PostgreSQL and Redis..."
docker-compose up -d postgres redis

echo ""
echo "=========================================="
echo "All services started successfully!"
echo "=========================================="
echo ""
echo "Service Status:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "kafka|zookeeper|postgres|redis"

echo ""
echo "Kafka Topics:"
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092

echo ""
echo "=========================================="
echo "Ready to start approval-service!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  cd services/approval-service"
echo "  go run cmd/main.go"
echo ""

