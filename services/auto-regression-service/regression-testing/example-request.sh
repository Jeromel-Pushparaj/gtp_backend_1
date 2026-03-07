#!/bin/bash

# Example request to the middleman service
# Replace these values with your actual GitHub repository details

GITHUB_URL="https://github.com/teknex-poc/test-backend"
PAT_TOKEN="YOUR_GITHUB_PAT_TOKEN_HERE"
BRANCH="main"

echo "Calling middleman service..."
echo "GitHub URL: $GITHUB_URL"
echo "Branch: $BRANCH"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8081/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d "{
    \"github_url\": \"$GITHUB_URL\",
    \"pat_token\": \"$PAT_TOKEN\",
    \"branch\": \"$BRANCH\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP Status Code: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "✓ Success!"
    echo ""
    echo "Response:"
    echo "$BODY" | jq '.'
else
    echo "✗ Error"
    echo ""
    echo "Response:"
    echo "$BODY"
fi

