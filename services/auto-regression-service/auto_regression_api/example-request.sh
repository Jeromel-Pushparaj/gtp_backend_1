#!/bin/bash

# Example request to the middleman service
# Loads configuration from .env file in auto_regression_agent directory

# Load environment variables from .env file
if [ -f ../auto_regression_agent/.env ]; then
    export $(cat ../auto_regression_agent/.env | grep -v '^#' | xargs)
elif [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Configuration - Read from environment variables
GITHUB_URL="${GITHUB_URL:-https://github.com/teknex-poc/test-backend}"
PAT_TOKEN="${GITHUB_PAT_TOKEN}"
BRANCH="${GITHUB_BRANCH:-main}"

# Validate required environment variables
if [ -z "$PAT_TOKEN" ]; then
    echo "❌ GITHUB_PAT_TOKEN not set!"
    echo ""
    echo "Set it in ../auto_regression_agent/.env file:"
    echo "  GITHUB_PAT_TOKEN=your_token_here"
    exit 1
fi

echo "Calling Auto-Regression API service on port 8092..."
echo "GitHub URL: $GITHUB_URL"
echo "Branch: $BRANCH"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8092/api/v1/test/aggregate \
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

