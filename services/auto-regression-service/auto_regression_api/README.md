# Auto-Regression API

A middleware service for automated regression testing of APIs from GitHub repositories using AI-powered autonomous agents.

## 🎯 Overview

The Auto-Regression API provides a simple REST interface to execute comprehensive regression tests on APIs by analyzing their OpenAPI specifications stored in GitHub repositories. It acts as a proxy to the Auto-Regression Agent API, simplifying the testing workflow and aggregating results.

## ✨ Features

- 🔍 **Automatic OpenAPI Discovery** - Fetches specs from GitHub repositories
- 🤖 **AI-Powered Testing** - Uses 5 autonomous agents for comprehensive testing
- 📊 **Aggregated Results** - Simplified, easy-to-consume test reports
- 🌿 **Branch Support** - Test any branch (main, develop, feature branches)
- 🔐 **Secure** - Uses GitHub PAT for repository access
- ⚡ **Fast** - Parallel test execution with intelligent orchestration

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Auto-Regression Agent API running on port 8080
- GitHub Personal Access Token

### Installation

```bash
# Navigate to the service directory
cd services/auto-regression-service/auto_regression_api

# Install dependencies
go mod download
```

### Running the Service

```bash
# Option 1: Using default port (8092)
go run main.go

# Option 2: Using custom port
PORT=9000 go run main.go

# Option 3: Using the parent Makefile
cd ..
make run-api
```

### Testing

```bash
# Health check
curl http://localhost:8092/health

# Run quick tests
./quick-test.sh

# Run example request (requires .env setup)
./example-request.sh
```

## 📖 API Documentation

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |
| POST | `/api/v1/test/aggregate` | Execute regression tests from GitHub |

### Example Request

```bash
curl -X POST http://localhost:8092/api/v1/test/aggregate \
  -H "Content-Type: application/json" \
  -d '{
    "github_url": "https://github.com/example/api-repo",
    "pat_token": "ghp_xxxxxxxxxxxxxxxxxxxx",
    "branch": "main"
  }'
```

### Example Response

```json
{
  "unique_test_cases": [
    {
      "name": "GET /users - Happy Path",
      "status_code": 200,
      "passed": true,
      "skipped": false,
      "category": "happy_path",
      "method": "GET",
      "path": "/api/v1/users"
    }
  ],
  "total_tests": 15,
  "tests_passed": 13,
  "tests_failed": 2,
  "tests_skipped": 0,
  "pass_rate": 86.67,
  "executed_at": "2024-03-07T10:35:42Z",
  "duration_ns": 5234567890
}
```

## 📚 Documentation

- **[Quick Reference](./QUICK_REFERENCE.md)** - Quick command reference
- **[API Documentation](./API_DOCUMENTATION.md)** - Complete API documentation
- **[OpenAPI Spec](./openapi.yaml)** - OpenAPI 3.0 specification
- **[Postman Collection](./postman-collection.json)** - Import into Postman

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8092` |
| `GITHUB_URL` | Default GitHub repository URL | - |
| `GITHUB_PAT_TOKEN` | GitHub Personal Access Token | - |
| `GITHUB_BRANCH` | Default branch to test | `main` |

### Setting Up Environment

```bash
# Create .env file (or use ../auto_regression_agent/.env)
cat > .env << EOF
GITHUB_URL=https://github.com/your-org/your-repo
GITHUB_PAT_TOKEN=ghp_your_token_here
GITHUB_BRANCH=main
EOF

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)
```

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ POST /api/v1/test/aggregate
       │ {github_url, pat_token, branch}
       ▼
┌──────────────────────┐
│ Auto-Regression API  │ (Port 8092)
│   (This Service)     │
└──────┬───────────────┘
       │ Forwards request
       ▼
┌──────────────────────┐
│ Agent API            │ (Port 8080)
│ 5 Autonomous Agents  │
│ - Discovery          │
│ - Designer           │
│ - Payload            │
│ - Executor           │
│ - Analyzer           │
└──────┬───────────────┘
       │ Returns results
       ▼
┌──────────────────────┐
│ Aggregated Response  │
└──────────────────────┘
```

## 🧪 Testing

### Unit Tests

```bash
go test -v ./...
```

### Integration Tests

```bash
# Start both services
cd ..
./start-without-docker.sh

# Run tests
cd auto_regression_api
go test -v -tags=integration ./...
```

## 📦 Files

```
auto_regression_api/
├── main.go                    # Main service implementation
├── main_test.go              # Unit tests
├── README.md                 # This file
├── API_DOCUMENTATION.md      # Complete API docs
├── QUICK_REFERENCE.md        # Quick reference guide
├── openapi.yaml              # OpenAPI 3.0 specification
├── postman-collection.json   # Postman collection
├── example-request.sh        # Example request script
├── go.mod                    # Go module definition
└── go.sum                    # Go module checksums
```

## 🤝 Contributing

1. Make changes to `main.go`
2. Update tests in `main_test.go`
3. Update OpenAPI spec if endpoints change
4. Run tests: `go test -v ./...`
5. Format code: `go fmt ./...`

## 📝 License

Apache 2.0

## 🔗 Related

- [Auto-Regression Agent](../auto_regression_agent/README.md) - The core AI testing engine
- [Parent Service](../README.md) - Overall service documentation

