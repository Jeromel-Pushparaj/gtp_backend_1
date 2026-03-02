# Chat Agent Service - Deployment Guide

## Overview

The Chat Agent Service is now ready to be integrated with your frontend! It provides a REST API on port 8082 that allows users to chat with an AI assistant powered by Groq, which can access your GTP Backend API.

## Architecture

```
┌─────────────┐      HTTP      ┌──────────────────┐      Groq API      ┌─────────────┐
│   Frontend  │ ────────────▶  │  Chat Agent      │ ─────────────────▶ │   Groq AI   │
│  (Port ???) │                 │  Service         │                     │             │
└─────────────┘                 │  (Port 8082)     │                     └─────────────┘
                                └──────────────────┘
                                         │
                                         │ Tool Calls
                                         ▼
                                ┌──────────────────┐
                                │  GTP Backend API │
                                │  (Port 8080)     │
                                └──────────────────┘
```

## Prerequisites

1. **Groq API Key**
   - Sign up at https://console.groq.com
   - Create an API key
   - Keep it secure (never commit to git)

2. **Backend API Running**
   - The GTP Backend API must be running on port 8080
   - Test with: `curl http://localhost:8080/health`

3. **Go 1.21+**
   - Required to build and run the service

## Quick Start

### 1. Configure Environment

```bash
cd services/chat-agent-service
cp .env.example .env
```

Edit `.env` and add your Groq API key:

```env
PORT=8082
GROQ_API_KEY=your_actual_groq_api_key_here
BACKEND_URL=http://localhost:8080
BACKEND_API_KEY=
```

### 2. Build the Service

```bash
make build-http
```

This creates the `chat-agent-server` binary.

### 3. Run the Service

```bash
./chat-agent-server
```

Or with environment variables:

```bash
GROQ_API_KEY=your_key ./chat-agent-server
```

Or using make:

```bash
make run-http
```

### 4. Test the Service

```bash
# Health check
curl http://localhost:8082/health

# Chat request
curl -X POST http://localhost:8082/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is the health status of the backend?"}'
```

Or use the test script:

```bash
./test_http_server.sh
```

## Frontend Integration

### Simple HTML/JavaScript

Open `examples/frontend-example.html` in your browser for a working demo.

### API Endpoint

```
POST http://localhost:8082/api/v1/chat
Content-Type: application/json

{
  "message": "Your question here"
}
```

### Example Code

```javascript
async function chat(message) {
  const response = await fetch('http://localhost:8082/api/v1/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message }),
  });
  
  const data = await response.json();
  return data.response;
}

// Usage
const answer = await chat('List all repositories');
console.log(answer);
```

## Production Deployment

### Using systemd (Linux)

Create `/etc/systemd/system/chat-agent.service`:

```ini
[Unit]
Description=Chat Agent Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/services/chat-agent-service
Environment="GROQ_API_KEY=your_key"
Environment="BACKEND_URL=http://localhost:8080"
Environment="PORT=8082"
ExecStart=/path/to/services/chat-agent-service/chat-agent-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable chat-agent
sudo systemctl start chat-agent
sudo systemctl status chat-agent
```

### Using Docker

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o chat-agent-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/chat-agent-server .
EXPOSE 8082
CMD ["./chat-agent-server"]
```

Build and run:

```bash
docker build -t chat-agent-service .
docker run -p 8082:8082 \
  -e GROQ_API_KEY=your_key \
  -e BACKEND_URL=http://host.docker.internal:8080 \
  chat-agent-service
```

### Using Docker Compose

Add to your `docker-compose.yml`:

```yaml
services:
  chat-agent:
    build: ./services/chat-agent-service
    ports:
      - "8082:8082"
    environment:
      - GROQ_API_KEY=${GROQ_API_KEY}
      - BACKEND_URL=http://backend:8080
      - PORT=8082
    depends_on:
      - backend
```

## Security Considerations

1. **API Key Protection**
   - Never commit API keys to git
   - Use environment variables or secrets management
   - Rotate keys regularly

2. **CORS Configuration**
   - Update `server/http_server.go` to restrict origins
   - Change `Access-Control-Allow-Origin` from `*` to your domain

3. **Rate Limiting**
   - Implement rate limiting to prevent abuse
   - Consider using a reverse proxy (nginx, traefik)

4. **Authentication**
   - Add authentication to the chat endpoint
   - Verify user identity before processing requests

## Monitoring

### Logs

The service logs to stdout. Monitor with:

```bash
# If running directly
./chat-agent-server 2>&1 | tee chat-agent.log

# If using systemd
journalctl -u chat-agent -f

# If using Docker
docker logs -f chat-agent-service
```

### Health Check

```bash
curl http://localhost:8082/health
```

Expected response:

```json
{
  "status": "healthy",
  "service": "chat-agent-service",
  "version": "1.0.0"
}
```

## Troubleshooting

### Server won't start

- Check if port 8082 is already in use: `lsof -i :8082`
- Verify GROQ_API_KEY is set correctly
- Check logs for error messages

### Tool calls fail

- Verify Backend API is running: `curl http://localhost:8080/health`
- Check BACKEND_URL is correct
- Review server logs for API errors

### Groq API errors

- Verify API key is valid
- Check Groq API status
- Review rate limits and quotas

## Next Steps

1. Integrate with your frontend application
2. Add user authentication
3. Implement conversation history
4. Add more tools as needed
5. Set up monitoring and alerting
6. Configure production deployment

## Support

For issues or questions:
- Check the logs
- Review the documentation in `docs/`
- Test with the example frontend in `examples/`

