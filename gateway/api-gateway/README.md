# 🌟 GTP Backend API Gateway

A unified API Gateway that connects all microservices in the GTP Backend ecosystem. Built with Go and Gin framework, featuring rate limiting, CORS, logging, and intelligent request routing.

## 🎯 Overview

The API Gateway serves as a single entry point for all backend microservices, providing:

- **🔄 Intelligent Routing** - Routes requests to appropriate backend services
- **⚡ Rate Limiting** - Protects services from overload (configurable)
- **🔐 CORS Support** - Handles cross-origin requests
- **📝 Request Logging** - Comprehensive logging with emojis for easy debugging
- **🛡️ Error Handling** - Graceful error recovery and meaningful error messages
- **🏥 Health Checks** - Monitor gateway and service health

## 🏗️ Architecture

```
┌─────────────────┐
│   Frontend      │
│  Applications   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│      🌟 API Gateway (Port 8000)        │
│  ┌─────────────────────────────────┐   │
│  │  Middleware Layer               │   │
│  │  • Rate Limiting                │   │
│  │  • CORS                         │   │
│  │  • Logging                      │   │
│  │  • Recovery                     │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  Routing Layer                  │   │
│  │  • /jira/*                      │   │
│  │  • /chat/*                      │   │
│  │  • /approval/*                  │   │
│  │  • /onboarding/*                │   │
│  │  • /scorecard/*                 │   │
│  │  • /sonar/*                     │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
         │
         ├──────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
         ▼              ▼              ▼              ▼              ▼              ▼
    ┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐
    │ Jira   │    │ Chat   │    │Approval│    │Onboard │    │ScoreCard│   │ Sonar  │
    │ :8086  │    │ :8082  │    │ :8083  │    │ :8084  │    │ :8085  │    │ :8080  │
    └────────┘    └────────┘    └────────┘    └────────┘    └────────┘    └────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- All backend services running on their respective ports

### Installation

1. **Navigate to the gateway directory:**
   ```bash
   cd gateway/api-gateway
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the gateway:**
   ```bash
   go run main.go
   ```

The gateway will start on `http://localhost:8000` (configurable via `GATEWAY_PORT`)

## ⚙️ Configuration

All configuration is done via environment variables. See `.env.example` for all available options.

### Key Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_PORT` | `8000` | Port for the gateway |
| `GATEWAY_HOST` | `0.0.0.0` | Host to bind to |
| `RATE_LIMIT_REQUESTS` | `100` | Max requests per duration |
| `RATE_LIMIT_DURATION` | `1m` | Rate limit time window |
| `CORS_ALLOWED_ORIGINS` | `*` | Allowed CORS origins |
| `JIRA_TRIGGER_SERVICE_URL` | `http://localhost:8086` | Jira service URL |
| `CHAT_AGENT_SERVICE_URL` | `http://localhost:8082` | Chat service URL |
| `APPROVAL_SERVICE_URL` | `http://localhost:8083` | Approval service URL |
| `ONBOARDING_SERVICE_URL` | `http://localhost:8084` | Onboarding service URL |
| `SCORECARD_SERVICE_URL` | `http://localhost:8085` | ScoreCard service URL |
| `SONAR_SHELL_SERVICE_URL` | `http://localhost:8080` | SonarShell service URL |

## 📡 API Routes

### Gateway Endpoints

- **GET /health** - Gateway health check

### Service Routes

All service routes are prefixed with the service name:

- **🎫 /jira/*** - Jira Trigger Service (Port 8086)
  - Create and manage Jira issues

- **🤖 /chat/*** - Chat Agent Service (Port 8082)
  - AI-powered chat interactions

- **✅ /approval/*** - Approval Service (Port 8083)
  - Slack approval workflows

- **📦 /onboarding/*** - Onboarding Service (Port 8084)
  - Service catalog and onboarding

- **📊 /scorecard/*** - ScoreCard Service (Port 8085)
  - Service quality scorecards

- **🔍 /sonar/*** - SonarShell Service (Port 8080)
  - SonarCloud automation

## 📖 API Documentation

Complete API documentation is available in OpenAPI 3.0 format:

- **File:** `openapi.yaml`
- **View online:** Import `openapi.yaml` into [Swagger Editor](https://editor.swagger.io/)

### Example Requests

#### Create a Jira Issue
```bash
curl -X POST http://localhost:8000/jira/api/create-issue \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Fix login bug",
    "issueType": "Bug",
    "priority": "High"
  }'
```

#### Chat with AI Agent
```bash
curl -X POST http://localhost:8000/chat/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What services are available?"
  }'
```

#### Onboard a Service
```bash
curl -X POST http://localhost:8000/onboarding/api/onboard \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "payment-service",
    "team": "Platform Team",
    "repositoryUrl": "https://github.com/company/payment-service"
  }'
```

## 🔧 Development

### Project Structure

```
gateway/api-gateway/
├── main.go              # 🚀 Application entry point
├── config/
│   └── config.go        # ⚙️ Configuration management
├── handlers/
│   └── proxy.go         # 🔄 Request proxy handlers
├── middleware/
│   └── middleware.go    # 🛡️ Middleware (CORS, rate limiting, logging)
├── openapi.yaml         # 📖 OpenAPI specification
├── .env.example         # 📝 Environment variables template
├── go.mod               # 📦 Go module definition
└── README.md            # 📚 This file
```

### Building for Production

```bash
# Build binary
go build -o api-gateway main.go

# Run binary
./api-gateway
```

### Running with Docker

```bash
# Build Docker image
docker build -t gtp-api-gateway .

# Run container
docker run -p 8000:8000 --env-file .env gtp-api-gateway
```

## 🔐 Security Features

### Rate Limiting

The gateway implements IP-based rate limiting to protect backend services:

- Default: 100 requests per minute per IP
- Configurable via `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_DURATION`
- Returns `429 Too Many Requests` when limit exceeded

### CORS

Cross-Origin Resource Sharing is enabled and configurable:

- Set `CORS_ALLOWED_ORIGINS=*` for development (allow all)
- Set specific origins for production: `CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com`

### JWT Authentication (Optional)

The gateway supports JWT authentication:

- Set `JWT_SECRET` to enable JWT validation
- Include `Authorization: Bearer <token>` header in requests

## 📊 Monitoring & Logging

### Health Checks

Check gateway health:
```bash
curl http://localhost:8000/health
```

Response:
```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "message": "🚀 Gateway is running smoothly!"
}
```

### Logging

The gateway logs all requests with emojis for easy visual parsing:

- ✅ 2xx - Success
- 🔄 3xx - Redirect
- ⚠️ 4xx - Client Error
- ❌ 5xx - Server Error

Example log output:
```
✅ GET /health 200 1.234ms
🔄 POST /jira/api/create-issue -> http://localhost:8086/api/create-issue
✅ POST /jira/api/create-issue 201 45.678ms
```

## 🚨 Error Handling

The gateway provides meaningful error responses:

### Service Unavailable (502)
```json
{
  "error": "Service unavailable",
  "message": "Failed to connect to backend service: connection refused"
}
```

### Rate Limit Exceeded (429)
```json
{
  "error": "Rate limit exceeded",
  "message": "Maximum 100 requests per 1m0s allowed"
}
```

### Internal Server Error (500)
```json
{
  "error": "Internal server error",
  "message": "An unexpected error occurred"
}
```

## 🧪 Testing

### Manual Testing

Test the gateway with curl:

```bash
# Test health endpoint
curl http://localhost:8000/health

# Test proxying to Jira service
curl -X POST http://localhost:8000/jira/api/create-issue \
  -H "Content-Type: application/json" \
  -d '{"summary": "Test issue"}'

# Test rate limiting (send 101 requests quickly)
for i in {1..101}; do
  curl http://localhost:8000/health
done
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

MIT License - see LICENSE file for details

## 🆘 Troubleshooting

### Gateway won't start

**Problem:** Port already in use
```
Failed to start server: listen tcp :8000: bind: address already in use
```

**Solution:** Change `GATEWAY_PORT` in `.env` or stop the process using port 8000

### Service unavailable errors

**Problem:** Backend service not responding
```
Service unavailable: Failed to connect to backend service
```

**Solution:**
1. Verify the backend service is running
2. Check the service URL in `.env`
3. Ensure firewall allows connections

### Rate limit issues

**Problem:** Getting rate limited during development

**Solution:** Increase `RATE_LIMIT_REQUESTS` or `RATE_LIMIT_DURATION` in `.env`

## 📞 Support

For issues and questions:
- Create an issue in the repository
- Contact the GTP Backend Team
- Check the OpenAPI documentation

---

**Made with ❤️ and lots of ☕ by the GTP Backend Team**

