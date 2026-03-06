# Service Catalog (Onboarding Service)

A microservice for managing and cataloging services within the GTP Backend Platform. This service provides a centralized registry for onboarding, tracking, and managing all services in the ecosystem.

## 📋 Overview

The Service Catalog acts as a **service registry** that allows teams to:
- **Onboard new services** with metadata (team, repository, lifecycle, etc.)
- **Track service information** across different environments
- **Query services** by ID or retrieve all registered services
- **Monitor service health** and count

## 🏗️ Architecture

### Clean Architecture Pattern
```
┌─────────────────────────────────────────────────┐
│              HTTP Layer (Gin)                   │
│                 routes/                         │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│           Controller Layer                      │
│         controller/onboarding_controller.go     │
│  (Handles HTTP requests/responses)              │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│           Service Layer                         │
│         service/onboarding_service.go           │
│  (Business logic & validation)                  │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│         Repository Layer                        │
│         db/service_repository.go                │
│  (In-memory data storage with mutex)            │
└─────────────────────────────────────────────────┘
```

### Key Components

- **`main.go`** - Application entry point, server initialization
- **`routes/`** - Route definitions and dependency injection
- **`controller/`** - HTTP request handlers
- **`service/`** - Business logic and validation
- **`db/`** - Data access layer (in-memory storage)
- **`models/`** - Domain models
- **`resources/`** - DTOs (Data Transfer Objects)
- **`config/`** - Configuration management
- **`utils/`** - Utility functions (ID generation)

## 🚀 Features

### ✅ Service Onboarding
Register new services with comprehensive metadata including:
- Service name and description
- Team ownership
- Repository URL
- Lifecycle stage (development, staging, production)
- Programming language
- Integration points
- Custom tags

### 📊 Service Query
- Retrieve all registered services
- Get specific service by ID
- Count total services

### 🏥 Health Monitoring
- Health check endpoint
- Service count metrics

## 📡 API Endpoints

### Base URL
```
http://localhost:8084/api
```

### Endpoints

#### 1. **Onboard a Service**
```http
POST /api/onboard
Content-Type: application/json

{
  "name": "payment-service",
  "description": "Handles payment processing",
  "team": "payments-team",
  "repository": "https://github.com/org/payment-service",
  "lifecycle": "production",
  "language": "Go",
  "integrations": {
    "kafka": "enabled",
    "postgres": "enabled"
  },
  "tags": ["payment", "critical"]
}
```

**Response (201 Created):**
```json
{
  "status": "success",
  "message": "Service onboarded successfully",
  "data": {
    "service_id": "SVC-20240315-ABC123",
    "name": "payment-service",
    "description": "Handles payment processing",
    "team": "payments-team",
    "repository": "https://github.com/org/payment-service",
    "lifecycle": "production",
    "language": "Go",
    "integrations": {
      "kafka": "enabled",
      "postgres": "enabled"
    },
    "tags": ["payment", "critical"],
    "onboarded_at": "2024-03-15T10:30:00Z"
  }
}
```

#### 2. **Get All Services**
```http
GET /api/services
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Services retrieved successfully",
  "data": [
    {
      "service_id": "SVC-20240315-ABC123",
      "name": "payment-service",
      ...
    }
  ]
}
```

#### 3. **Get Service by ID**
```http
GET /api/services/:id
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Service retrieved successfully",
  "data": {
    "service_id": "SVC-20240315-ABC123",
    ...
  }
}
```

#### 4. **Health Check**
```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "service": "onboarding-service",
  "total_services": 5
}
```

## 🛠️ Setup & Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone git@github.com:Jeromel-Pushparaj/gtp_backend_1.git
   cd gtp_backend_1/services/service-catelog
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables** (optional)
   Create a `.env` file:
   ```env
   SERVICE_NAME=onboarding-service
   SERVICE_PORT=8084
   SERVICE_HOST=0.0.0.0
   ENVIRONMENT=development
   LOG_LEVEL=debug
   ```

4. **Run the service**
   ```bash
   # From the service directory
   go run main.go

   # Or from the project root using Makefile
   make service-catelog
   ```

5. **Verify the service is running**
   ```bash
   curl http://localhost:8084/health
   ```

## 🧪 Testing

### Manual Testing with cURL

**Onboard a service:**
```bash
curl -X POST http://localhost:8084/api/onboard \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-service",
    "description": "A test service",
    "team": "platform-team",
    "repository": "https://github.com/org/test-service",
    "lifecycle": "development",
    "language": "Go",
    "tags": ["test"]
  }'
```

**Get all services:**
```bash
curl http://localhost:8084/api/services
```

**Get service by ID:**
```bash
curl http://localhost:8084/api/services/SVC-20240315-ABC123
```

## 📦 Data Model

### Service Model
```go
type Service struct {
    ID           string            // Auto-generated (e.g., SVC-20240315-ABC123)
    Name         string            // Service name (required)
    Description  string            // Service description
    Team         string            // Owning team (required)
    Repository   string            // Git repository URL (required)
    Lifecycle    string            // development|staging|production (required)
    Language     string            // Programming language
    Integrations map[string]string // Integration configurations
    Tags         []string          // Custom tags
    OnboardedAt  time.Time         // Timestamp of onboarding
}
```

### Validation Rules
- **Name**: Required, non-empty
- **Team**: Required, non-empty
- **Repository**: Required, non-empty
- **Lifecycle**: Required, must be one of: `development`, `staging`, `production`

## 🔧 Configuration

### Environment Variables

| Variable       | Default              | Description                    |
|----------------|----------------------|--------------------------------|
| `SERVICE_NAME` | `onboarding-service` | Name of the service            |
| `SERVICE_PORT` | `8084`               | Port to run the service on     |
| `SERVICE_HOST` | `0.0.0.0`            | Host to bind the service to    |
| `ENVIRONMENT`  | `development`        | Environment (dev/staging/prod) |
| `LOG_LEVEL`    | `debug`              | Logging level                  |

## 🏛️ Design Patterns

### 1. **Dependency Injection**
All dependencies are injected through constructors, making the code testable and maintainable.

### 2. **Repository Pattern**
Data access is abstracted through the repository layer, allowing easy swapping of storage backends.

### 3. **DTO Pattern**
Separate request/response DTOs from domain models for better API contract management.

### 4. **Clean Architecture**
Clear separation of concerns across layers (HTTP → Controller → Service → Repository).

## 🔄 Integration with GTP Platform

The Service Catalog integrates with the broader GTP Backend Platform:

- **API Gateway** (Port 8080) - Routes requests to this service
- **Future Kafka Integration** - Can publish service onboarding events
- **Service Discovery** - Acts as the source of truth for all services

## 📊 Storage

Currently uses **in-memory storage** with thread-safe operations (sync.RWMutex).

### Future Enhancements
- [ ] PostgreSQL integration for persistent storage
- [ ] Redis caching layer
- [ ] Kafka event publishing on service onboarding
- [ ] Service health monitoring integration
- [ ] API versioning support
- [ ] Service update/delete endpoints
- [ ] Search and filtering capabilities

## 🚦 CORS Configuration

Configured to allow requests from:
- `http://localhost:5173` (Vite dev server)
- `http://localhost:3000` (React dev server)

## 📝 Logging

The service logs:
- Server startup information
- Available endpoints
- Request/response details (via Gin middleware)

## 🤝 Contributing

1. Follow the existing code structure
2. Maintain clean architecture principles
3. Add validation for new fields
4. Update this README for new features

## 📄 License

Part of the GTP Backend Platform project.

## 👥 Team

Maintained by the Platform Engineering Team.

---

**Service Port**: 8084
**Health Check**: `GET /health`
**API Version**: v1 (implicit)

