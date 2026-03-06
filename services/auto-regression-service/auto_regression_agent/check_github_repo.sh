#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "========================================="
echo "🔍 Checking GitHub Repository Contents"
echo "========================================="
echo ""

# Configuration - Read from environment variables
GITHUB_URL="${GITHUB_URL:-https://github.com/teknex-poc/test-backend}"
PAT_TOKEN="${GITHUB_PAT_TOKEN}"
BRANCH="${GITHUB_BRANCH:-main}"

# Validate required environment variables
if [ -z "$PAT_TOKEN" ]; then
    echo -e "${RED}❌ GITHUB_PAT_TOKEN not set!${NC}"
    echo ""
    echo "Set it in .env file:"
    echo "  GITHUB_PAT_TOKEN=your_token_here"
    exit 1
fi

# Extract owner and repo
OWNER="teknex-poc"
REPO="test-backend"

echo -e "${BLUE}Repository:${NC} $GITHUB_URL"
echo -e "${BLUE}Branch:${NC} $BRANCH"
echo ""

echo "========================================="
echo "Step 1: List all files in repository root"
echo "========================================="
echo ""

RESPONSE=$(curl -s -H "Authorization: token $PAT_TOKEN" \
  "https://api.github.com/repos/$OWNER/$REPO/contents?ref=$BRANCH")

echo "$RESPONSE" | python3 -c "
import sys, json
try:
    files = json.load(sys.stdin)
    if isinstance(files, list):
        print('Files in repository root:')
        print('-' * 50)
        for f in files:
            if f.get('type') == 'file':
                print(f'  📄 {f.get(\"name\")}')
            elif f.get('type') == 'dir':
                print(f'  📁 {f.get(\"name\")}/')
    else:
        print('Error:', files.get('message', 'Unknown error'))
except Exception as e:
    print('Error parsing response:', e)
    print('Raw response:', sys.stdin.read())
" || echo "$RESPONSE"

echo ""
echo "========================================="
echo "Step 2: Check for OpenAPI spec files"
echo "========================================="
echo ""

# List of possible spec file names
SPEC_FILES=(
    "openapi-spec.json"
    "openAPISpec.json"
    "openapi-spec.yml"
    "openAPISpec.yml"
    "openapi-spec.yaml"
    "openAPISpec.yaml"
    "openapi.json"
    "openapi.yml"
    "openapi.yaml"
    "swagger.json"
    "swagger.yml"
    "swagger.yaml"
)

FOUND_FILES=()

for filename in "${SPEC_FILES[@]}"; do
    echo -n "Checking for $filename... "
    
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: token $PAT_TOKEN" \
        "https://api.github.com/repos/$OWNER/$REPO/contents/$filename?ref=$BRANCH")
    
    if [ "$STATUS" == "200" ]; then
        echo -e "${GREEN}✅ FOUND${NC}"
        FOUND_FILES+=("$filename")
    else
        echo -e "${RED}❌ Not found (HTTP $STATUS)${NC}"
    fi
done

echo ""

if [ ${#FOUND_FILES[@]} -eq 0 ]; then
    echo -e "${RED}❌ No OpenAPI spec files found in repository root${NC}"
    echo ""
    echo "Suggestions:"
    echo "  1. Check if the spec file is in a subdirectory"
    echo "  2. Verify the file name matches one of the expected formats"
    echo "  3. Make sure the file is committed to the '$BRANCH' branch"
    echo ""
    echo "Expected file names:"
    for filename in "${SPEC_FILES[@]}"; do
        echo "  - $filename"
    done
else
    echo -e "${GREEN}✅ Found ${#FOUND_FILES[@]} OpenAPI spec file(s):${NC}"
    for filename in "${FOUND_FILES[@]}"; do
        echo "  - $filename"
    done
    echo ""
    echo "You can now use the GitHub integration endpoint!"
fi

echo ""
echo "========================================="
echo "Step 3: Search for spec files in subdirectories"
echo "========================================="
echo ""

echo "Searching for files containing 'openapi' or 'swagger'..."
echo ""

SEARCH_RESPONSE=$(curl -s -H "Authorization: token $PAT_TOKEN" \
  "https://api.github.com/search/code?q=filename:openapi+OR+filename:swagger+repo:$OWNER/$REPO")

echo "$SEARCH_RESPONSE" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    items = data.get('items', [])
    if items:
        print(f'Found {len(items)} file(s):')
        print('-' * 50)
        for item in items:
            print(f'  📄 {item.get(\"path\")}')
    else:
        print('No files found matching \"openapi\" or \"swagger\"')
except Exception as e:
    print('Error:', e)
" || echo "Search failed"

echo ""
echo "========================================="
echo "Done!"
echo "========================================="

