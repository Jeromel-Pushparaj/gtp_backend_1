#!/bin/bash

# Simple Chat Client with Session Management
# This script maintains conversation context automatically

SESSION_FILE="/tmp/chat_session_id.txt"
SERVER_URL="http://localhost:8082/api/v1/chat"

# Colors for better UX
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🤖 Chat Agent Client${NC}"
echo -e "${GREEN}=====================${NC}"
echo ""

# Load existing session if available
if [ -f "$SESSION_FILE" ]; then
    SESSION_ID=$(cat "$SESSION_FILE")
    echo -e "${YELLOW}📌 Resuming session: $SESSION_ID${NC}"
    echo ""
else
    echo -e "${YELLOW}📌 Starting new conversation${NC}"
    echo ""
fi

# Function to send message
send_message() {
    local message="$1"

    # Build JSON payload
    if [ -z "$SESSION_ID" ]; then
        # No session ID yet
        JSON_PAYLOAD="{\"message\": \"$message\"}"
    else
        # Include session ID
        JSON_PAYLOAD="{\"message\": \"$message\", \"session_id\": \"$SESSION_ID\"}"
    fi

    # Create temp files for headers and body
    HEADER_FILE=$(mktemp)
    BODY_FILE=$(mktemp)

    # Send request and save response
    curl -s -D "$HEADER_FILE" -o "$BODY_FILE" -X POST "$SERVER_URL" \
        -H "Content-Type: application/json" \
        -d "$JSON_PAYLOAD"

    # Extract session ID from headers
    NEW_SESSION_ID=$(grep -i "X-Session-ID:" "$HEADER_FILE" | cut -d' ' -f2 | tr -d '\r\n')

    # Save session ID if we got a new one
    if [ -n "$NEW_SESSION_ID" ]; then
        SESSION_ID="$NEW_SESSION_ID"
        echo "$SESSION_ID" > "$SESSION_FILE"
    fi

    # Display response body
    echo -e "${BLUE}🤖 Assistant:${NC}"
    cat "$BODY_FILE"
    echo ""

    # Clean up temp files
    rm -f "$HEADER_FILE" "$BODY_FILE"
}

# Interactive chat loop
echo "Type your messages below. Type 'exit' to quit, 'new' to start a new conversation."
echo ""

while true; do
    echo -ne "${GREEN}You: ${NC}"
    read -r USER_MESSAGE
    
    # Handle special commands
    if [ "$USER_MESSAGE" = "exit" ]; then
        echo "Goodbye! 👋"
        break
    elif [ "$USER_MESSAGE" = "new" ]; then
        rm -f "$SESSION_FILE"
        SESSION_ID=""
        echo -e "${YELLOW}📌 Started new conversation${NC}"
        echo ""
        continue
    elif [ -z "$USER_MESSAGE" ]; then
        continue
    fi
    
    # Send message
    send_message "$USER_MESSAGE"
done

