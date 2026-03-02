# Score Card Service

A microservice for calculating and managing service quality scorecards based on various metrics.

## Overview

The Score Card Service evaluates services based on multiple quality metrics and provides an overall score and grade.

### Metrics Evaluated

- **Code Quality** (25% weight)
- **Test Coverage** (25% weight)
- **Security Score** (20% weight)
- **Performance Score** (15% weight)
- **Documentation Score** (15% weight)

### Grading Scale

- **A**: 90-100
- **B**: 80-89
- **C**: 70-79
- **D**: 60-69
- **F**: Below 60

## Architecture

```
score-card-service/
├── cmd/                    # Application entry point
├── api/v1/                 # HTTP API handlers and routes
├── internal/
│   ├── models/            # Data models
│   ├── service/           # Business logic
│   └── repository/        # Database operations
├── kafka/                 # Kafka producer/consumer
└── config/                # Configuration management
```

## API Endpoints

### Create ScoreCard
```http
POST /api/v1/scorecards
Content-Type: application/json

{
  "service_name": "my-service",
  "metrics": {
    "code_quality": 85.5,
    "test_coverage": 90.0,
    "security_score": 88.0,
    "performance_score": 92.0,
    "documentation_score": 75.0
  }
}
```

### Get ScoreCard by ID
```http
GET /api/v1/scorecards/:id
```

### Get All ScoreCards for a Service
```http
GET /api/v1/scorecards/service/:name
```

### Get Latest ScoreCard for a Service
```http
GET /api/v1/scorecards/service/:name/latest
```

## Configuration

Create a `.env` file in the service directory:

```bash
# Service Configuration
SERVICE_NAME=score-card-service
SERVICE_PORT=8085
SERVICE_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=scorecard_db

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_SCORE_CALCULATED=score.calculated
KAFKA_TOPIC_SCORE_REQUESTED=score.requested
KAFKA_GROUP_ID=score-card-service

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Logging
LOG_LEVEL=debug
ENVIRONMENT=development
```

## Running the Service

### Using Make (from project root)
```bash
make score-card
```

### Manually
```bash
cd services/score-card-service
go mod tidy
go run cmd/main.go
```

### Using Docker
```bash
docker build -t score-card-service .
docker run -p 8085:8085 --env-file .env score-card-service
```

## Development

### Install Dependencies
```bash
go mod tidy
```

### Run Tests
```bash
go test ./...
```

### Build
```bash
go build -o bin/score-card-service cmd/main.go
```

## Kafka Integration

### Published Events
- **Topic**: `score.calculated`
- **Event**: Score calculation completed

### Consumed Events
- **Topic**: `score.requested`
- **Event**: Request to calculate score for a service

## Database Schema

```sql
CREATE TABLE scorecards (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    score DECIMAL(5,2) NOT NULL,
    code_quality DECIMAL(5,2) NOT NULL,
    test_coverage DECIMAL(5,2) NOT NULL,
    security_score DECIMAL(5,2) NOT NULL,
    performance_score DECIMAL(5,2) NOT NULL,
    documentation_score DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Health Check

```http
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "score-card-service"
}
```

## Port

Default: **8085**

