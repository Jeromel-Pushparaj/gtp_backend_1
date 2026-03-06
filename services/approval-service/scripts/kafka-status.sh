#!/bin/bash

# Kafka Status Check Script

echo "=========================================="
echo "Kafka Infrastructure Status"
echo "=========================================="
echo ""

# Check if containers are running
echo "Container Status:"
echo "----------------------------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "NAME|kafka|zookeeper"

echo ""
echo "=========================================="
echo "Health Checks"
echo "=========================================="

# Check Zookeeper
echo ""
echo "Zookeeper (port 2181):"
if docker exec zookeeper nc -z localhost 2181 2>/dev/null; then
    echo "  ✓ Zookeeper is healthy"
else
    echo "  ✗ Zookeeper is DOWN"
fi

# Check Kafka
echo ""
echo "Kafka (port 9092):"
if docker exec kafka kafka-broker-api-versions --bootstrap-server localhost:9092 2>/dev/null | grep -q "ApiVersion"; then
    echo "  ✓ Kafka is healthy"
else
    echo "  ✗ Kafka is DOWN"
fi

# Check Kafka UI
echo ""
echo "Kafka UI (port 8090):"
if curl -s http://localhost:8090 > /dev/null 2>&1; then
    echo "  ✓ Kafka UI is accessible at http://localhost:8090"
else
    echo "  ✗ Kafka UI is not accessible"
fi

# List topics
echo ""
echo "=========================================="
echo "Kafka Topics"
echo "=========================================="
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092 2>/dev/null || echo "Cannot list topics - Kafka may be down"

# Count messages in each topic
echo ""
echo "=========================================="
echo "Message Counts"
echo "=========================================="

for topic in approval.requested approval.completed action.executed action.rejected; do
    count=$(docker exec kafka kafka-run-class kafka.tools.GetOffsetShell \
        --broker-list localhost:9092 \
        --topic "$topic" 2>/dev/null | awk -F':' '{sum += $3} END {print sum}')
    
    if [ -z "$count" ]; then
        count="N/A"
    fi
    
    echo "$topic: $count messages"
done

echo ""
echo "=========================================="
echo "Consumer Groups"
echo "=========================================="
docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list 2>/dev/null || echo "Cannot list consumer groups"

echo ""

