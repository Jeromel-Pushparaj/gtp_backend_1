# Score Card Service - API Endpoints

This document provides a comprehensive overview of all active endpoints in the Score Card Service.

## Base URLs

- **Local Development**: `http://localhost:8085`
- **API Gateway**: `http://localhost:8080/score-card-service`

## API Versions

The service provides two API versions:

- **V1**: Simple scorecard creation and retrieval with weighted average scoring
- **V2**: Advanced rule-based evaluation with level progression (Bronze/Silver/Gold, etc.)

---

## Health Check

### GET /health

Check if the service is running and healthy.

**Response**: `200 OK`
```json
{
  "status": "healthy",
  "service": "score-card-service"
}
```

---

## API V1 - Simple Scorecards

### POST /api/v1/scorecards

Create a new scorecard with weighted average scoring.

**Request Body**:
```json
{
  "service_name": "authentication-service",
  "metrics": {
    "code_quality": 85.5,
    "test_coverage": 90.0,
    "security_score": 88.0,
    "performance_score": 92.0,
    "documentation_score": 75.0
  }
}
```

**Response**: `201 Created`
```json
{
  "id": 1,
  "service_name": "authentication-service",
  "score": 86.25,
  "grade": "B",
  "metrics": { ... },
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Scoring Weights**:
- Code Quality: 25%
- Test Coverage: 25%
- Security Score: 20%
- Performance Score: 15%
- Documentation Score: 15%

### GET /api/v1/scorecards/{id}

Retrieve a specific scorecard by ID.

**Parameters**:
- `id` (path, integer): Scorecard ID

**Response**: `200 OK`
```json
{
  "id": 1,
  "service_name": "authentication-service",
  "score": 86.25,
  "grade": "B",
  "metrics": { ... },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### GET /api/v1/scorecards/service/{name}

Retrieve all scorecards for a specific service.

**Parameters**:
- `name` (path, string): Service name

**Response**: `200 OK`
```json
{
  "service": "authentication-service",
  "count": 5,
  "scorecards": [ ... ]
}
```

### GET /api/v1/scorecards/service/{name}/latest

Retrieve the most recent scorecard for a service.

**Parameters**:
- `name` (path, string): Service name

**Response**: `200 OK`
```json
{
  "id": 5,
  "service_name": "authentication-service",
  "score": 88.5,
  "grade": "B",
  "metrics": { ... },
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

## API V2 - Advanced Rule-Based Scorecards

### GET /api/v2/scorecards/definitions

Get all available scorecard definitions with their levels and rules.

**Response**: `200 OK`
```json
{
  "count": 6,
  "scorecards": [
    {
      "name": "CodeQuality",
      "display_name": "Code Quality",
      "category": "code_quality",
      "level_pattern": "metal",
      "levels": [ ... ]
    }
  ]
}
```

**Available Scorecards**:
1. **CodeQuality** - Code quality metrics (Bronze/Silver/Gold)
2. **DORA_Metrics** - DevOps performance (Low/Medium/High/Elite)
3. **Security_Maturity** - Security posture (Basic/Good/Great)
4. **Production_Readiness** - Production readiness checks
5. **Service_Health** - Service health metrics
6. **PR_Metrics** - Pull request metrics

### GET /api/v2/scorecards/definitions/{name}

Get a specific scorecard definition.

**Parameters**:
- `name` (path, string): One of: `CodeQuality`, `DORA_Metrics`, `Security_Maturity`, `Production_Readiness`, `Service_Health`, `PR_Metrics`

**Response**: `200 OK`
```json
{
  "name": "CodeQuality",
  "display_name": "Code Quality",
  "category": "code_quality",
  "description": "Evaluates code quality based on test coverage, vulnerabilities, code smells, and duplications",
  "level_pattern": "metal",
  "levels": [
    {
      "name": "Bronze",
      "display_name": "Bronze",
      "order_index": 1,
      "color": "#CD7F32",
      "rules": [ ... ]
    }
  ],
  "is_active": true
}
```

### POST /api/v2/scorecards/evaluate

Evaluate a service against all active scorecard definitions.

**Request Body**:
```json
{
  "service_name": "authentication-service",
  "service_data": {
    "coverage": 85.5,
    "code_smells": 12,
    "vulnerabilities": 2,
    "duplicated_lines_density": 3.2,
    "has_readme": 1,
    "deployment_frequency": 5,
    "mttr": 8
  }
}
```

**Response**: `200 OK`
```json
{
  "service_name": "authentication-service",
  "overall_percentage": 75.5,
  "total_rules_passed": 45,
  "total_rules": 60,
  "scorecards": [
    {
      "service_name": "authentication-service",
      "scorecard_name": "Code Quality",
      "achieved_level_name": "Silver",
      "rules_passed": 8,
      "rules_total": 10,
      "pass_percentage": 80.0,
      "rule_results": [ ... ],
      "evaluated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "strengths": [
    "Code Quality: Silver",
    "Security Maturity: Good"
  ],
  "improvement_areas": [
    "DORA Metrics: Needs attention",
    "Production Readiness: Needs attention"
  ],
  "evaluated_at": "2024-01-15T10:30:00Z"
}
```

### POST /api/v2/scorecards/evaluate/{name}

Evaluate a service against a specific scorecard definition.

**Parameters**:
- `name` (path, string): One of: `CodeQuality`, `DORA_Metrics`, `Security_Maturity`, `Production_Readiness`, `Service_Health`, `PR_Metrics`

**Request Body**:
```json
{
  "service_name": "authentication-service",
  "service_data": {
    "coverage": 85.5,
    "code_smells": 12,
    "vulnerabilities": 2,
    "duplicated_lines_density": 3.2,
    "has_readme": 1
  }
}
```

**Response**: `200 OK`
```json
{
  "service_name": "authentication-service",
  "scorecard_name": "Code Quality",
  "achieved_level_name": "Silver",
  "rules_passed": 8,
  "rules_total": 10,
  "pass_percentage": 80.0,
  "rule_results": [
    {
      "rule_name": "Coverage >= 60%",
      "passed": true,
      "actual_value": 85.5,
      "expected_value": 60,
      "operator": ">=",
      "message": "Passed",
      "evaluated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "evaluated_at": "2024-01-15T10:30:00Z"
}
```

---

## Level Patterns

The V2 API uses different level patterns for different scorecards:

### Metal Pattern (Code Quality)
- Bronze (Order: 1)
- Silver (Order: 2)
- Gold (Order: 3)

### Performance Pattern (DORA Metrics)
- Low (Order: 1)
- Medium (Order: 2)
- High (Order: 3)
- Elite (Order: 4)

### Traffic Light Pattern
- Red (Order: 1)
- Yellow (Order: 2)
- Orange (Order: 3)
- Green (Order: 4)

### Descriptive Pattern (Security Maturity)
- Basic (Order: 1)
- Good (Order: 2)
- Great (Order: 3)

---

## Rule Operators

The following operators are supported for rule evaluation:

- `>=` - Greater than or equal to
- `<=` - Less than or equal to
- `==` - Equal to
- `!=` - Not equal to
- `>` - Greater than
- `<` - Less than

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request body",
  "details": "service_name is required"
}
```

### 404 Not Found
```json
{
  "error": "Scorecard not found",
  "details": "No scorecard with ID 123"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to create scorecard",
  "details": "Database connection error"
}
```

### 503 Service Unavailable
```json
{
  "error": "Database not available"
}
```

---

## OpenAPI Specification

For the complete OpenAPI 3.0 specification, see [openapi.yaml](./openapi.yaml).

You can use this specification with tools like:
- Swagger UI
- Postman
- Insomnia
- OpenAPI Generator

---

## Summary

**Total Endpoints**: 9

### By Version:
- **Health**: 1 endpoint
- **V1 API**: 4 endpoints (Simple scorecards)
- **V2 API**: 4 endpoints (Advanced rule-based evaluation)

### By HTTP Method:
- **GET**: 5 endpoints
- **POST**: 4 endpoints

