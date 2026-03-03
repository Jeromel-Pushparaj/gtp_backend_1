# OpenAPI Quick Start Guide

## 📋 What's Included

This directory now contains a complete OpenAPI 3.0.3 specification for the Sonar Shell Test API.

### Files

- **`openapi.yaml`** - Complete OpenAPI specification (1,191 lines, 30 endpoints)
- **`API.md`** - Comprehensive API documentation with examples
- **`scripts/validate-openapi.sh`** - Validation script

## 🚀 Quick Start

### 1. View the API Documentation

#### Option A: Online (Easiest)
1. Go to [Swagger Editor](https://editor.swagger.io/)
2. Copy the contents of `openapi.yaml`
3. Paste into the editor
4. Explore the interactive documentation

#### Option B: Local with Docker
```bash
docker run -p 8081:8080 \
  -v $(pwd)/openapi.yaml:/openapi.yaml \
  -e SWAGGER_JSON=/openapi.yaml \
  swaggerapi/swagger-ui
```
Then visit: http://localhost:8081

### 2. Validate the Specification

```bash
./scripts/validate-openapi.sh
```

### 3. Test the API

Start the server:
```bash
# Set required environment variables
export GITHUB_PAT="your-github-token"
export GITHUB_ORG="your-org"
export API_KEY="your-secret-key"  # Optional

# Start the server
./sonar-automation -server -port=8080
```

Test an endpoint:
```bash
curl -H "Authorization: Bearer your-secret-key" \
  http://localhost:8080/api/v1/orgs
```

## 📊 API Overview

### Endpoint Categories

| Category | Endpoints | Description |
|----------|-----------|-------------|
| Health | 1 | Server health check |
| Organizations | 2 | Manage organizations |
| Repositories | 2 | Manage repositories |
| GitHub Metrics | 12 | GitHub analytics |
| SonarCloud | 1 | Code quality metrics |
| Jira | 7 | Project management metrics |
| Metrics Collection | 4 | Database-backed metrics |
| Repo Metrics | 3 | Fetch metrics by repo |
| SonarCloud Setup | 6 | Automation & setup |

**Total: 38 endpoints**

### Key Endpoints

```
GET  /health                              - Health check
GET  /api/v1/orgs                         - List organizations
POST /api/v1/orgs/create                  - Create organization
GET  /api/v1/repos/fetch?org_id=<id>      - Fetch repositories
GET  /api/v1/github/metrics?repo=<name>   - Get repo metrics
GET  /api/v1/sonar/metrics?repo=<name>    - Get SonarCloud metrics
GET  /api/v1/jira/metrics?project_key=<k> - Get Jira metrics
```

## 🔐 Authentication

The API uses Bearer token authentication (optional):

```bash
# In requests
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:8080/api/v1/endpoint
```

The `/health` endpoint does not require authentication.

## 📝 Example Requests

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

## 🛠️ Generate Client SDKs

You can generate client libraries in various languages:

```bash
# JavaScript/TypeScript
npx @openapitools/openapi-generator-cli generate \
  -i openapi.yaml \
  -g typescript-fetch \
  -o ./client/typescript

# Python
npx @openapitools/openapi-generator-cli generate \
  -i openapi.yaml \
  -g python \
  -o ./client/python

# Go
npx @openapitools/openapi-generator-cli generate \
  -i openapi.yaml \
  -g go \
  -o ./client/go
```

## 📚 Data Models

The specification includes complete schemas for:

- **Organization** - Organization details
- **Repository** - Repository information
- **RepositoryMetrics** - Comprehensive GitHub metrics
- **GitHubMetrics** - Database-stored GitHub metrics
- **SonarMetrics** - SonarCloud code quality metrics
- **JiraMetrics** - Jira project metrics
- **JiraIssue** - Jira issue details
- **PullRequestInfo** - Pull request information
- And more...

## 🔍 Response Format

All endpoints return JSON with this structure:

**Success:**
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "error": "Error message"
}
```

## 📖 More Information

- See `API.md` for detailed documentation
- See `openapi.yaml` for the complete specification
- Run `./scripts/validate-openapi.sh` to validate the spec

## 🎯 Next Steps

1. ✅ View the spec in Swagger Editor
2. ✅ Start the API server
3. ✅ Test endpoints with curl or Postman
4. ✅ Generate client SDKs if needed
5. ✅ Integrate with your applications

---

**Need Help?** Check the main project documentation or the API.md file.

