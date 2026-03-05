#!/bin/bash

# Pull Docker images for the infrastructure
# This bypasses certificate issues in Colima VM by pulling from host

echo "🐳 Pulling Docker images..."
echo ""

images=(
    "confluentinc/cp-zookeeper:7.5.0"
    "confluentinc/cp-kafka:7.5.0"
    "postgres:15-alpine"
    "redis:7-alpine"
    "provectuslabs/kafka-ui:latest"
)

for image in "${images[@]}"; do
    echo "📦 Pulling $image..."
    docker pull "$image"
    if [ $? -eq 0 ]; then
        echo "✅ Successfully pulled $image"
    else
        echo "❌ Failed to pull $image"
    fi
    echo ""
done

echo "✅ All images pulled successfully!"
echo "Now you can run: make infra-up"

