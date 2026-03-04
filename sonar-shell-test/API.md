# Sonar Shell Test API Documentation

## Overview

The Sonar Shell Test API provides comprehensive endpoints for managing SonarCloud integration, GitHub metrics, Jira metrics, and repository management. This service enables automated collection and analysis of software development metrics across multiple platforms.

## OpenAPI Specification

The complete API specification is available in OpenAPI 3.0.3 format:
- **File**: `openapi.yaml`
- **Version**: 1.0.0

## Getting Started

### Running the API Server

Start the server in API mode:

```bash
# Using the binary
./sonar-automation -server -port=8080

# Or using make
make server PORT=8080

# Or using go run
go run . -server -port=8080
```

### Authentication

The API supports optional Bearer token authentication:

```bash
# Set API key in environment
export API_KEY="your-secret-api-key"

# Make authenticated requests
curl -H "Authorization: Bearer your-secret-api-key" \
  http://localhost:8080/api/v1/orgs
```

**Note**: The `/health` endpoint does not require authentication.

## API Endpoints

### Health Check
- `GET /health` - Check if the API server is running

### Organization Management
- `GET /api/v1/orgs` - List all organizations
- `POST /api/v1/orgs/create` - Create a new organization

### Repository Management
- `GET /api/v1/repos/fetch?org_id=<id>` - Fetch repositories by organization
- `PUT /api/v1/repos/update?repo_id=<id>` - Update repository details

### GitHub Metrics
- `GET /api/v1/github/metrics?repo=<name>` - Get repository metrics
- `GET /api/v1/github/metrics/all` - Get all repositories metrics
- `GET /api/v1/github/pulls?repo=<name>&state=<state>` - List pull requests
- `GET /api/v1/github/pulls/get?repo=<name>&number=<num>` - Get pull request details
- `GET /api/v1/github/commits?repo=<name>` - List commits
- `GET /api/v1/github/commits/activity?repo=<name>` - Get commit activity
- `GET /api/v1/github/issues?repo=<name>&state=<state>` - List issues
- `GET /api/v1/github/issues/comments?repo=<name>&number=<num>` - List issue comments
- `GET /api/v1/github/readme?repo=<name>` - Check README existence
- `GET /api/v1/github/branches?repo=<name>` - List branches
- `GET /api/v1/github/org/members` - List organization members
- `GET /api/v1/github/org/teams` - List organization teams

### SonarCloud Metrics
- `GET /api/v1/sonar/metrics?repo=<name>` - Get SonarCloud metrics

### Jira Metrics
- `GET /api/v1/jira/issues/stats?project_key=<key>` - Get issue statistics
- `GET /api/v1/jira/bugs/open?project_key=<key>` - Get open bugs
- `GET /api/v1/jira/tasks/open?project_key=<key>` - Get open tasks
- `GET /api/v1/jira/issues/by-assignee?project_key=<key>` - Get issues by assignee
- `GET /api/v1/jira/sprints/stats?project_key=<key>` - Get sprint statistics
- `GET /api/v1/jira/metrics?project_key=<key>` - Get project metrics
- `GET /api/v1/jira/issues/search?jql=<query>&max_results=<num>` - Search issues

### Metrics Collection (Database-backed)
- `POST /api/v1/metrics/github/collect?repo=<name>` - Collect and store GitHub metrics
- `GET /api/v1/metrics/github/stored?repo=<name>` - Get stored GitHub metrics
- `POST /api/v1/metrics/sonar/collect?repo=<name>` - Collect and store SonarCloud metrics
- `GET /api/v1/metrics/sonar/stored?repo=<name>` - Get stored SonarCloud metrics

### Repository Metrics Fetch
- `GET /api/v1/repos/metrics/github?repo_id=<id>` - Fetch GitHub metrics by repo
- `GET /api/v1/repos/metrics/jira?repo_id=<id>` - Fetch Jira metrics by repo
- `GET /api/v1/repos/metrics/sonar?repo_id=<id>` - Fetch SonarCloud metrics by repo

### SonarCloud Setup & Automation
- `GET /api/v1/secrets/list` - List repository secrets
- `POST /api/v1/secrets/add` - Add environment secrets
- `POST /api/v1/workflows/update` - Update workflows
- `POST /api/v1/setup/full` - Full SonarCloud setup
- `GET /api/v1/results/fetch` - Fetch SonarCloud results
- `POST /api/v1/repository/process?repo=<name>` - Process repository

## Example Requests

### Create an Organization

```bash
curl -X POST http://localhost:8080/api/v1/orgs/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "name": "my-org",
    "sonar_org_key": "my-org-key",
    "jira_domain": "mycompany.atlassian.net",
    "jira_email": "admin@example.com"
  }'
```

### Get Repository Metrics

```bash
curl http://localhost:8080/api/v1/github/metrics?repo=my-repo \
  -H "Authorization: Bearer your-api-key"
```

### Collect and Store Metrics

```bash
curl -X POST http://localhost:8080/api/v1/metrics/github/collect?repo=my-repo \
  -H "Authorization: Bearer your-api-key"
```

## Response Format

All endpoints return JSON responses with the following structure:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

Error responses:

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

## Viewing the OpenAPI Spec

You can view and interact with the API specification using various tools:

### Swagger UI (Online)
1. Go to [Swagger Editor](https://editor.swagger.io/)
2. Copy the contents of `openapi.yaml`
3. Paste into the editor

### Swagger UI (Local with Docker)
```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/openapi.yaml \
  -v $(pwd)/openapi.yaml:/openapi.yaml \
  swaggerapi/swagger-ui
```

Then visit: http://localhost:8081

## Configuration

Required environment variables:
- `GITHUB_PAT` - GitHub Personal Access Token
- `GITHUB_ORG` - GitHub Organization name
- `SONAR_TOKEN` - SonarCloud token (optional)
- `SONAR_ORG_KEY` - SonarCloud organization key (optional)
- `JIRA_TOKEN` - Jira API token (optional)
- `JIRA_DOMAIN` - Jira domain (optional)
- `JIRA_EMAIL` - Jira email (optional)
- `API_KEY` - API authentication key (optional)
- `DATABASE_PATH` - Path to SQLite database (default: ./data/metrics.db)

## Support

For issues or questions, please refer to the main project documentation.

