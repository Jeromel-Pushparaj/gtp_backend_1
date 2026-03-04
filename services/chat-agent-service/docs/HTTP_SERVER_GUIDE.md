# Chat Agent HTTP Server Guide

## Overview

The Chat Agent Service now provides an HTTP server that exposes a chatbot API powered by Groq AI. The chatbot can access your GTP Backend API through tool calling, allowing it to fetch GitHub, Jira, and SonarCloud metrics.

## Architecture

```
Frontend → HTTP Server (Port 8082) → Groq API → Tool Executor → Backend API (Port 8080)
```

1. **Frontend** sends chat messages to the HTTP server
2. **HTTP Server** receives the message and passes it to the Chat Agent
3. **Chat Agent** sends the message to Groq API with available tools
4. **Groq API** decides which tools to call and returns tool calls
5. **Tool Executor** executes the tools by calling the Backend API
6. **Chat Agent** sends tool results back to Groq for final response
7. **HTTP Server** returns the response to the frontend

## Prerequisites

- Go 1.21 or higher
- Groq API key (get one at https://console.groq.com)
- GTP Backend API running on port 8080

## Configuration

### Environment Variables

Create a `.env` file in the service directory:

```env
# HTTP Server Configuration
PORT=8082

# Groq API Configuration
GROQ_API_KEY=your_groq_api_key_here

# Backend API Configuration
BACKEND_URL=http://localhost:8080
BACKEND_API_KEY=
```

### Getting a Groq API Key

1. Go to https://console.groq.com
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key and add it to your `.env` file

## Running the Server

### Option 1: Using the binary

```bash
cd services/chat-agent-service
go build -o chat-agent-server cmd/server/main.go
./chat-agent-server
```

### Option 2: Using go run

```bash
cd services/chat-agent-service
go run cmd/server/main.go
```

### Option 3: With command-line flags

```bash
./chat-agent-server \
  -port=8082 \
  -groq-api-key=your_key \
  -backend-url=http://localhost:8080
```

## API Endpoints

### Health Check

```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "chat-agent-service",
  "version": "1.0.0"
}
```

### Chat

```bash
POST /api/v1/chat
Content-Type: application/json

{
  "message": "List all pull requests for the backend repository"
}
```

**Response:**
```json
{
  "response": "Here are the pull requests for the backend repository: ..."
}
```

## Example Usage

### Using curl

```bash
# Health check
curl http://localhost:8082/health

# Chat request
curl -X POST http://localhost:8082/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is the health status of the backend?"}'
```

### Using JavaScript (Frontend)

```javascript
async function chat(message) {
  const response = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ message }),
  });
  
  const data = await response.json();
  return data.response;
}

// Example usage
const answer = await chat('List all repositories');
console.log(answer);
```

### Using Python

```python
import requests

def chat(message):
    response = requests.post(
        'http://localhost:8082/api/v1/chat',
        json={'message': message}
    )
    return response.json()['response']

# Example usage
answer = chat('Show me open issues in the backend repo')
print(answer)
```

## Available Tools

The chatbot has access to the following tools:

- `health_check` - Check backend health
- `list_pull_requests` - List PRs for a repository
- `list_commits` - List commits for a repository
- `list_issues` - List issues for a repository
- `check_readme` - Check if a repo has a README
- `list_branches` - List branches for a repository
- `list_org_members` - List organization members
- `list_org_teams` - List organization teams
- `fetch_orgs` - Get all organizations
- `fetch_repos_by_org` - Fetch repos for an organization

## Example Conversations

**User:** "What is the health status of the backend?"
**Bot:** "The backend is healthy. Service: sonar-automation, Organization: teknex-poc"

**User:** "List all pull requests for the backend repository"
**Bot:** "Here are the pull requests for the backend repository: [lists PRs]"

**User:** "Show me all organizations"
**Bot:** "Here are the organizations: [lists orgs]"

## Troubleshooting

### Server won't start

- Check if port 8082 is already in use
- Verify GROQ_API_KEY is set correctly
- Ensure Backend API is running on port 8080

### Tool calls fail

- Verify Backend API is accessible at BACKEND_URL
- Check if BACKEND_API_KEY is correct (if required)
- Review server logs for detailed error messages

### Groq API errors

- Verify your API key is valid
- Check your Groq API quota/limits
- Review the error message in the response

## Production Deployment

For production deployment:

1. Set appropriate environment variables
2. Use a process manager (systemd, supervisor, etc.)
3. Configure reverse proxy (nginx, traefik, etc.)
4. Enable HTTPS
5. Set up monitoring and logging
6. Configure rate limiting

## Next Steps

- Integrate with your frontend application
- Add more tools as needed
- Implement authentication/authorization
- Add conversation history
- Implement streaming responses

