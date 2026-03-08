#!/bin/bash

# Kafka Monitor Script - Prettified JSON output
# Usage: ./kafka-monitor.sh [topic-name]

TOPIC=${1:-approval.requested}

echo "=========================================="
echo "Monitoring Kafka Topic: $TOPIC"
echo "=========================================="
echo ""

docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic "$TOPIC" \
  --from-beginning \
  --property print.timestamp=true \
  --property print.key=true | while IFS= read -r line; do
    
    # Extract timestamp, key, and value
    if [[ $line =~ CreateTime:([0-9]+)[[:space:]]+([^[:space:]]*)[[:space:]]+(.*) ]]; then
        timestamp="${BASH_REMATCH[1]}"
        key="${BASH_REMATCH[2]}"
        json="${BASH_REMATCH[3]}"
        
        # Convert timestamp to readable format
        readable_time=$(date -r $((timestamp / 1000)) '+%Y-%m-%d %H:%M:%S')
        
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "⏰ Time: $readable_time"
        echo "🔑 Key:  $key"
        echo "📦 Message:"
        echo "$json" | jq '.' 2>/dev/null || echo "$json"
        echo ""
    else
        # If no timestamp, just prettify the JSON
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "$line" | jq '.' 2>/dev/null || echo "$line"
        echo ""
    fi
done

