# Development Setup Guide

This guide will help you set up the development environment for the GTP Backend Platform.

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, but recommended)

## Quick Start

### 1. Initial Setup

Run the setup command to create environment files:

```bash
make setup
```

This will copy all `.env.example` files to `.env` files. Update these files with your configuration.

### 2. Start Infrastructure

Start Kafka, PostgreSQL, and Redis:

```bash
make infra-up
```

This will start:

- **Kafka** on `localhost:9092`
- **Kafka UI** on `http://localhost:8090`
- **PostgreSQL** on `localhost:5432`
- **Redis** on `localhost:6379`

### 3. Verify Infrastructure

Check that all services are running:

```bash
docker-compose ps
```

Access Kafka UI at: <http://localhost:8090>

## Project Structure

```
gtp_backend_1/
├── services/
│   ├── jira-trigger-service/    # Port 8081
│   ├── chat-agent-service/      # Port 8082
│   ├── approval-service/        # Port 8083
│   ├── service-catelog/         # Port 8084
│   └── score-card-service/      # Port 8085
├── gateway/
│   └── api-gateway/             # Port 8080
├── shared/
│   ├── contracts/
│   ├── middleware/
│   ├── auth/
│   └── utils/
└── infra/
    ├── kafka/
    ├── docker/
    └── terraform/
```

## Running Services

Each service has been initialized with Go modules. To run a service:

```bash
# Using Make
make jira-trigger    # Run jira-trigger-service
make chat-agent      # Run chat-agent-service
make approval        # Run approval-service
make onboarding      # Run onboarding-service
make gateway         # Run api-gateway

# Or manually
cd services/jira-trigger-service
go run cmd/main.go
```

## Environment Configuration

Each service has its own `.env` file. Key configurations:

### Jira Trigger Service (Port 8081)

- Jira API credentials
- Kafka topic: `jira.trigger.created`

### Chat Agent Service (Port 8082)

- OpenAI API key
- Kafka topics: `chat.request`, `chat.response`

### Approval Service (Port 8083)

- Slack credentials
- Kafka topics: `approval.requested`, `approval.completed`

### Onboarding Service (Port 8084)

- Email configuration
- Kafka topic: `service.onboarded`

### API Gateway (Port 8080)

- Service URLs
- JWT configuration
- CORS settings

## Development Workflow

### Install Dependencies

```bash
make tidy
```

### Run Tests

```bash
make test
```

### View Infrastructure Logs

```bash
make infra-logs
```

### Stop Infrastructure

```bash
make infra-down
```

### Clean Up (removes volumes)

```bash
make clean
```

## Kafka Topics

The platform uses the following Kafka topics:

- `jira.trigger.created`
- `approval.requested`
- `approval.completed`
- `service.onboarded`
- `chat.request`
- `chat.response`

Topics are auto-created when first used.

## Next Steps

1. Update `.env` files with your credentials
2. Start implementing service logic in each service's directories
3. Define shared contracts in `shared/contracts/`
4. Implement API handlers in `api/v1/` directories

## Useful Commands

```bash
make help              # Show all available commands
make setup             # Initial environment setup
make infra-up          # Start infrastructure
make infra-down        # Stop infrastructure
make infra-logs        # View logs
make test              # Run all tests
make tidy              # Update dependencies
```

## Team Ownership

| Service      | Owner     | Directory                    |
| ------------ | --------- | ---------------------------- |
| Jira Trigger | Keerthana | services/jira-trigger-service |
| Chat Agent   | Jeromel   | services/chat-agent-service   |
| Approval     | Sarumathi | services/approval-service     |
| Onboarding   | Rashmi    | services/onboarding-service   |
