# 🤖 OpenTest: Autonomous AI Agent Testing Platform

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Event--Driven-green.svg)](AUTONOMOUS_AGENTS.md)

## Overview

**OpenTest** is an advanced API testing platform powered by **autonomous AI agents** that collaborate to automatically generate, execute, and analyze comprehensive API tests from OpenAPI/Swagger specifications. The platform uses event-driven architecture with AI-powered decision making at every step.

**🎉 Status**: ✅ **FULLY OPERATIONAL** - Autonomous AI agents working collaboratively!

### Autonomous AI Agents

- **🔍 Discovery Agent** - Analyzes OpenAPI specs using GPT-4
- **🎨 Designer Agent** - Creates intelligent test strategies using GPT-4
- **🎲 Payload Agent** - Generates realistic test data using GPT-3.5
- **⚡ Executor Agent** - Executes HTTP tests with validation
- **📊 Analyzer Agent** - Analyzes results and provides feedback using GPT-4

### Key Features

- ✅ **AI-Powered Decision Making**: Every step uses LLM for intelligent analysis
- ✅ **Collaborative Consensus**: Agents vote on important decisions
- ✅ **Event-Driven Architecture**: Scalable and loosely coupled via Redis pub/sub
- ✅ **Continuous Learning**: Agents improve from feedback over time
- ✅ **Real-time Monitoring**: Track agent activities and test progress
- ✅ **Swagger 2.0 & OpenAPI 3.x Support**: Automatic spec analysis

## Architecture

```
User Upload → Backend → Event Bus (Redis Pub/Sub)
                              ↓
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
Discovery Agent (GPT-4)  Designer Agent (GPT-4)  Payload Agent (GPT-3.5)
    │                         │                         │
    └─────────────────────────┼─────────────────────────┘
                              ↓
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
Executor Agent          Analyzer Agent (GPT-4)
    │                         │
    └─────────────────────────┘
                ↓
        Feedback & Learning
```

**Event Flow:**
```
spec_uploaded → spec_analyzed → strategy_proposed → consensus_request
    ↓               ↓                ↓                    ↓
Backend      Discovery Agent   Designer Agent      All Agents Vote
                                     ↓
                            strategy_approved
                                     ↓
                              Payload Agent
                                     ↓
                              payloads_ready
                                     ↓
                              Executor Agent
                                     ↓
                              tests_complete
                                     ↓
                              Analyzer Agent
                                     ↓
                            analysis_complete
```

## Features

### 🤖 AI-Powered Test Generation
- Automatic test generation from OpenAPI specifications
- Intelligent scenario building (happy path, edge cases, error conditions)
- Smart test data generation based on JSON schemas
- Coverage gap analysis and recommendations

### 🎯 Deterministic Test Execution
- Pure HTTP request execution based on JSON manifests
- Comprehensive response validation (status, headers, body, schema)
- Baseline comparison for regression detection
- Performance assertion validation

### 🔄 Intelligent Workflow Orchestration
- State machine-based workflow management
- Priority-based job scheduling
- Parallel and sequential execution strategies
- Automatic retry with configurable backoff

### 📊 Advanced Analysis & Reporting
- AI-powered regression detection
- Baseline drift analysis
- Trend analysis across test runs
- Detailed execution reports (JSON, HTML, JUnit XML)

### 🏢 Enterprise-Grade Features
- Multi-tenant architecture with team isolation
- RBAC and API key authentication
- Resource quotas and rate limiting
- Comprehensive audit logging

### 🔍 Full Observability
- Structured logging with correlation IDs
- Prometheus metrics for all components
- Distributed tracing with OpenTelemetry
- Health checks and readiness probes

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 7+
- S3-compatible storage (AWS S3, MinIO, etc.)
- OpenAI or Anthropic API key

### Installation

```bash
# Clone the repository (SSH - recommended)
git clone git@gitlab.com:tekion/development/toc/poc/opentest.git
cd opentest

# Or using HTTPS
# git clone https://gitlab.com/tekion/development/toc/poc/opentest.git
# cd opentest

# Install dependencies
go mod download

# Run database migrations
make migrate-up

# Start the services
make run-server    # API server
make run-worker    # Background worker
make run-scheduler # Job scheduler
```

### Configuration

Create a `configs/config.yaml` file:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  
database:
  host: localhost
  port: 5432
  name: auto_regression
  
llm:
  provider: openai
  api_key: ${OPENAI_API_KEY}
  model: gpt-4-turbo-preview
```

See [Configuration Guide](docs/guides/configuration.md) for full options.

## Quick Start

### Test a Real Service in 30 Seconds

```bash
# 1. Start the platform
cd deployments/docker
docker-compose up -d

# 2. Test the Petstore API
cd ../..
./scripts/test_real_service.sh https://petstore.swagger.io/v2/swagger.json

# 3. Verify the workflow
./scripts/verify_workflow.sh
```

**That's it!** The platform will:
1. ✅ Download and convert the Swagger 2.0 spec to OpenAPI 3.0
2. ✅ Extract the base URL (`https://petstore.swagger.io/v2`)
3. ✅ Analyze the spec and discover 20 endpoints
4. ✅ Generate 20 test manifests
5. ✅ Execute tests against the real Petstore API
6. ✅ Generate detailed test reports with real API responses

See [WORKFLOW_SUCCESS.md](docs/WORKFLOW_SUCCESS.md) for detailed results.

## Usage

### 1. Upload an OpenAPI Specification

```bash
# Upload Swagger 2.0 or OpenAPI 3.x spec
curl -X POST http://localhost:8080/api/v1/specs \
  -H "Content-Type: application/x-yaml" \
  --data-binary @examples/specs/petstore-api.yaml
```

### 2. View Test Results

```bash
# View discovery results
docker exec auto-regression-worker-1 cat /app/output/discovery/{spec_id}-discovery.json | python3 -m json.tool

# View test manifests
docker exec auto-regression-worker-1 find /app/output/manifests/{spec_id} -name "*.json"

# View test reports
docker exec auto-regression-worker-1 find /app/output/reports -name "*.json" | head -10
```

### 3. Test Your Own Service

```bash
# Option 1: Direct URL
./scripts/test_real_service.sh https://your-api.com/swagger.json

# Option 2: With custom base URL override
./scripts/test_real_service.sh \
  https://api.example.com/swagger.json \
  https://staging.example.com

# Option 3: Local service
./scripts/test_real_service.sh http://localhost:8000/openapi.json
```

See [TESTING_REAL_SERVICES.md](docs/TESTING_REAL_SERVICES.md) for more examples.

### 4. Advanced: Generate Tests from Spec (Future)

```bash
curl -X POST http://localhost:8080/api/v1/analysis/generate-tests \
  -H "Content-Type: application/json" \
  -d '{
    "spec_id": "spec-123",
    "coverage_level": "comprehensive"
  }'
```

### 3. Execute Tests

```bash
curl -X POST http://localhost:8080/api/v1/executions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "suite_id": "suite-456",
    "environment": "staging"
  }'
```

### 4. View Results

```bash
curl http://localhost:8080/api/v1/executions/exec-789/results \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## Test Manifest Format

Tests are defined as declarative JSON manifests:

```json
{
  "version": "1.0",
  "metadata": {
    "name": "User API Tests",
    "service": "user-service",
    "team": "platform"
  },
  "tests": [
    {
      "id": "test-001",
      "name": "Get user by ID",
      "request": {
        "method": "GET",
        "path": "/users/{user_id}",
        "path_params": {"user_id": "123"}
      },
      "assertions": {
        "status_code": 200,
        "body": {
          "schema": {
            "type": "object",
            "required": ["id", "email"]
          }
        }
      }
    }
  ]
}
```

See [Test Manifest Schema](api/schemas/test-manifest-schema.json) for full specification.

## Documentation

- [Architecture Overview](docs/architecture/overview.md)
- [Getting Started Guide](docs/guides/getting-started.md)
- [Writing Test Manifests](docs/guides/writing-tests.md)
- [AI Agent Guide](docs/guides/ai-agents.md)
- [API Reference](docs/api/rest-api.md)
- [Deployment Guide](docs/guides/deployment.md)

## Development

### Project Structure

```
auto-regression/
├── cmd/              # Application entrypoints
├── pkg/              # Public APIs and domain models
├── internal/         # Private implementation
├── api/              # API definitions (proto, OpenAPI)
├── migrations/       # Database migrations
├── configs/          # Configuration files
├── deployments/      # Deployment configurations
└── docs/             # Documentation
```

### Running Tests

```bash
make test           # Run unit tests
make test-integration # Run integration tests
make test-e2e       # Run end-to-end tests
```

### Building

```bash
make build          # Build all binaries
make docker-build   # Build Docker images
```

## Contributing

Please read [CONTRIBUTING.md](docs/development/contributing.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📧 Email: support@auto-regression.io
- 💬 Slack: [Join our community](https://slack.auto-regression.io)
- 🐛 Issues: [GitLab Issues](https://gitlab.com/tekion/development/toc/poc/opentest/-/issues)

