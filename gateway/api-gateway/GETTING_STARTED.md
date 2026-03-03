# 🚀 Getting Started with GTP Backend API Gateway

Welcome! This guide will help you get the API Gateway up and running in minutes.

## 📋 Prerequisites

Before you begin, ensure you have:

- ✅ Go 1.21 or higher installed
- ✅ All backend services running (or at least the ones you want to use)
- ✅ Basic understanding of REST APIs

## ⚡ Quick Start (3 Steps)

### Step 1: Setup Environment

```bash
cd gateway/api-gateway
cp .env.example .env
```

Edit `.env` and configure your service URLs:

```bash
# The gateway will run on this port
GATEWAY_PORT=8000

# Update these URLs to match your running services
JIRA_TRIGGER_SERVICE_URL=http://localhost:8086
CHAT_AGENT_SERVICE_URL=http://localhost:8082
APPROVAL_SERVICE_URL=http://localhost:8083
ONBOARDING_SERVICE_URL=http://localhost:8084
SCORECARD_SERVICE_URL=http://localhost:8085
SONAR_SHELL_SERVICE_URL=http://localhost:8080
```

### Step 2: Install Dependencies

```bash
make install
# or
go mod download
```

### Step 3: Run the Gateway

```bash
make run
# or
go run main.go
```

You should see:

```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     🌟  GTP Backend API Gateway  🌟                      ║
║                                                           ║
║     Connecting all your microservices with style! 🚀     ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

🚀 API Gateway starting on 0.0.0.0:8000 (Environment: development)
🎯 Rate Limit: 100 requests per 1m0s
✨ All systems operational!
```

## 🧪 Test It Out

### Test Gateway Health

```bash
curl http://localhost:8000/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "message": "🚀 Gateway is running smoothly!"
}
```

### Test Service Proxying

Try creating a Jira issue:

```bash
curl -X POST http://localhost:8000/jira/api/create-issue \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Test issue from gateway",
    "issueType": "Task"
  }'
```

## 📚 What's Next?

### For Backend Developers

1. **Read the README** - `README.md` has complete documentation
2. **Explore the Code** - Check out `main.go`, `handlers/`, `middleware/`
3. **Customize** - Add authentication, modify rate limits, etc.

### For Frontend Developers

1. **Read the Integration Guide** - `FRONTEND_INTEGRATION_GUIDE.md`
2. **Check the OpenAPI Spec** - `openapi.yaml` has all endpoints documented
3. **Import to Swagger** - View the spec at [editor.swagger.io](https://editor.swagger.io/)

## 🎯 Available Services

Once the gateway is running, you can access:

| Service | Path | Backend Port | Description |
|---------|------|--------------|-------------|
| 🎫 Jira | `/jira/*` | 8086 | Create and manage Jira issues |
| 🤖 Chat | `/chat/*` | 8082 | AI-powered chat agent |
| ✅ Approval | `/approval/*` | 8083 | Slack approval workflows |
| 📦 Onboarding | `/onboarding/*` | 8084 | Service catalog |
| 📊 ScoreCard | `/scorecard/*` | 8085 | Service quality metrics |
| 🔍 SonarShell | `/sonar/*` | 8080 | SonarCloud automation |

## 🔧 Common Commands

```bash
# Run the gateway
make run

# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean

# Setup environment
make setup

# Check health
make health
```

## 🐳 Docker Deployment

### Build Docker Image

```bash
make docker-build
# or
docker build -t gtp-api-gateway .
```

### Run with Docker

```bash
make docker-run
# or
docker run -p 8000:8000 --env-file .env gtp-api-gateway
```

## 🚨 Troubleshooting

### Port Already in Use

If you see `bind: address already in use`:

1. Change `GATEWAY_PORT` in `.env` to a different port
2. Or stop the process using port 8000

### Service Unavailable Errors

If you get `502 Bad Gateway`:

1. Verify the backend service is running
2. Check the service URL in `.env`
3. Test the service directly (e.g., `curl http://localhost:8086/health`)

### CORS Issues

If frontend gets CORS errors:

1. Check `CORS_ALLOWED_ORIGINS` in `.env`
2. For development, set it to `*`
3. For production, specify your frontend URL

## 📖 Documentation Files

- **README.md** - Complete documentation for developers
- **FRONTEND_INTEGRATION_GUIDE.md** - Guide for frontend developers
- **openapi.yaml** - OpenAPI 3.0 specification
- **GETTING_STARTED.md** - This file!

## 💡 Tips

1. **Development Mode** - Use `ENVIRONMENT=development` for detailed logs
2. **Rate Limiting** - Adjust `RATE_LIMIT_REQUESTS` for your needs
3. **Health Checks** - Monitor `/health` endpoint for uptime
4. **Logs** - Watch the console for emoji-coded request logs

## 🎉 You're Ready!

The gateway is now running and routing requests to your microservices. 

**Next steps:**
- Integrate with your frontend application
- Customize middleware and handlers
- Deploy to production

Need help? Check the other documentation files or contact the team!

---

**Made with ❤️ by the GTP Backend Team**

