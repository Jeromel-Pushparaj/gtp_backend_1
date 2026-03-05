# Score Card Service - Quick API Reference

## Endpoints Overview

| Method | Endpoint | Description | Version |
|--------|----------|-------------|---------|
| GET | `/health` | Health check | - |
| POST | `/api/v1/scorecards` | Create scorecard | V1 |
| GET | `/api/v1/scorecards/{id}` | Get scorecard by ID | V1 |
| GET | `/api/v1/scorecards/service/{name}` | Get all scorecards for service | V1 |
| GET | `/api/v1/scorecards/service/{name}/latest` | Get latest scorecard for service | V1 |
| GET | `/api/v2/scorecards/definitions` | Get all scorecard definitions | V2 |
| GET | `/api/v2/scorecards/definitions/{name}` | Get specific scorecard definition | V2 |
| POST | `/api/v2/scorecards/evaluate` | Evaluate against all scorecards | V2 |
| POST | `/api/v2/scorecards/evaluate/{name}` | Evaluate against specific scorecard | V2 |

## Quick Examples

### V1: Create Simple Scorecard

```bash
curl -X POST http://localhost:8085/api/v1/scorecards \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "my-service",
    "metrics": {
      "code_quality": 85.5,
      "test_coverage": 90.0,
      "security_score": 88.0,
      "performance_score": 92.0,
      "documentation_score": 75.0
    }
  }'
```

### V1: Get Latest Scorecard

```bash
curl http://localhost:8085/api/v1/scorecards/service/my-service/latest
```

### V2: Get All Scorecard Definitions

```bash
curl http://localhost:8085/api/v2/scorecards/definitions
```

### V2: Get Code Quality Scorecard Definition

```bash
curl http://localhost:8085/api/v2/scorecards/definitions/CodeQuality
```

### V2: Evaluate Service Against All Scorecards

```bash
curl -X POST http://localhost:8085/api/v2/scorecards/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "my-service",
    "service_data": {
      "coverage": 85.5,
      "code_smells": 12,
      "vulnerabilities": 2,
      "duplicated_lines_density": 3.2,
      "has_readme": 1,
      "deployment_frequency": 5,
      "mttr": 8
    }
  }'
```

### V2: Evaluate Service Against Code Quality Only

```bash
curl -X POST http://localhost:8085/api/v2/scorecards/evaluate/CodeQuality \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "my-service",
    "service_data": {
      "coverage": 85.5,
      "code_smells": 12,
      "vulnerabilities": 2,
      "duplicated_lines_density": 3.2,
      "has_readme": 1
    }
  }'
```

## Available Scorecard Names (V2)

- `CodeQuality` - Code quality metrics (Bronze/Silver/Gold)
- `DORA_Metrics` - DevOps performance (Low/Medium/High/Elite)
- `Security_Maturity` - Security posture (Basic/Good/Great)
- `Production_Readiness` - Production readiness checks
- `Service_Health` - Service health metrics
- `PR_Metrics` - Pull request metrics

## Common Service Data Properties

### Code Quality Metrics
- `coverage` - Test coverage percentage
- `code_smells` - Number of code smells
- `vulnerabilities` - Number of vulnerabilities
- `duplicated_lines_density` - Percentage of duplicated lines
- `has_readme` - Whether README exists (0 or 1)

### DORA Metrics
- `deployment_frequency` - Deployments per week
- `mttr` - Mean time to resolve (hours)
- `lead_time` - Lead time for changes (hours)
- `change_failure_rate` - Percentage of failed changes

### Security Metrics
- `security_hotspots` - Number of security hotspots
- `security_rating` - Security rating (1-5)
- `has_security_policy` - Whether security policy exists (0 or 1)

### Production Readiness
- `has_monitoring` - Whether monitoring is configured (0 or 1)
- `has_logging` - Whether logging is configured (0 or 1)
- `has_health_check` - Whether health check exists (0 or 1)
- `has_documentation` - Whether documentation exists (0 or 1)

## Response Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |
| 503 | Service Unavailable - Database not available |

## Documentation Files

- **[openapi.yaml](./openapi.yaml)** - Complete OpenAPI 3.0 specification
- **[API_ENDPOINTS.md](./API_ENDPOINTS.md)** - Detailed endpoint documentation
- **[API_V2_EXAMPLES.md](./API_V2_EXAMPLES.md)** - V2 API examples
- **[README.md](./README.md)** - Service overview and setup

## Testing

Run the test scripts to verify the API:

```bash
# Test V2 API
./test_v2_api.sh

# Test auto-evaluation
./test_auto_evaluate.sh
```

