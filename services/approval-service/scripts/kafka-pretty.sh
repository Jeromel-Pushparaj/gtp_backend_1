#!/bin/bash

# Pretty Kafka Consumer
# Usage: ./kafka-pretty.sh [topic-name]

TOPIC=${1:-approval.requested}

echo "=========================================="
echo "📊 Monitoring: $TOPIC"
echo "=========================================="
echo ""

docker exec -i kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic "$TOPIC" \
  --from-beginning 2>/dev/null | while IFS= read -r line; do
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "$line" | jq '.'
    echo ""
done

