# API Endpoints Documentation

This document provides a comprehensive list of all API endpoints available in the GTP Backend system.

## Base URL
```
http://localhost:8080
```

## Authentication
- **Type**: Bearer Token (Optional)
- **Header**: `Authorization: Bearer <API_KEY>`
- **Note**: Authentication is optional unless `API_KEY` environment variable is set

---

## Health Check

### GET /health
Check the health status of the service.

**Response:**
```json
{
  "status": "healthy",
  "service": "sonar-automation",
  "organization": "your-org-name"
}
```

---

## SonarCloud Management

### GET /api/v1/secrets/list
List all secrets for repositories in the organization.

**Response:**
```json
{
  "success": true,
  "message": "Found N repositories",
  "data": [
    {
      "repository": "repo-name",
      "secrets": ["SECRET_1", "SECRET_2"],
      "env_secrets": ["ENV_SECRET_1"]
    }
  ]
}
```

### POST /api/v1/secrets/add
Add environment secrets to all repositories.

**Response:**
```json
{
  "success": true,
  "message": "Processed N repositories: X success, Y errors",
  "data": [
    {
      "repository": "repo-name",
      "status": "success|error",
      "error": "error message (if any)"
    }
  ]
}
```

### POST /api/v1/workflows/update
Update GitHub workflows to use environment secrets.

**Response:**
```json
{
  "success": true,
  "message": "Processed N repositories: X success, Y errors",
  "data": [
    {
      "repository": "repo-name",
      "status": "success|error|skipped",
      "error": "error message (if any)"
    }
  ]
}
```

### POST /api/v1/setup/full
Perform full SonarCloud setup for all repositories.

**Response:**
```json
{
  "success": true,
  "message": "Processed N repositories: X success, Y errors",
  "data": [
    {
      "repository": "repo-name",
      "status": "success|error",
      "error": "error message (if any)"
    }
  ]
}
```

### GET /api/v1/results/fetch
Fetch SonarCloud analysis results for all repositories.

**Response:**
```json
{
  "success": true,
  "message": "Fetched results for N repositories",
  "data": [
    {
      "repository": "repo-name",
      "project_key": "org_repo-name",
      "quality_gate": "OK|ERROR",
      "metrics": {
        "bugs": "0",
        "vulnerabilities": "0",
        "code_smells": "5"
      },
      "issues_count": 10,
      "error": "error message (if any)"
    }
  ]
}
```

### GET /api/v1/sonar/metrics
Get SonarCloud metrics for a specific repository.

**Query Parameters:**
- `repo` (required): Repository name
- `include_issues` (optional): Set to "true" to include issue details

**Response:**
```json
{
  "success": true,
  "data": {
    "repository": "repo-name",
    "project_key": "org_repo-name",
    "quality_gate_status": "OK",
    "metrics": {
      "bugs": "0",
      "vulnerabilities": "0",
      "code_smells": "5",
      "coverage": "80.5"
    },
    "issues_count": 10,
    "issues": []
  }
}
```

### POST /api/v1/repository/process
Process a single repository for SonarCloud setup.

**Request Body:**
```json
{
  "repository_name": "repo-name"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully processed repository: repo-name"
}
```

---

## GitHub Metrics

### Pull Requests

#### GET /api/v1/github/pulls
List pull requests for a repository.

**Query Parameters:**
- `repo` (required): Repository name
- `state` (optional): PR state (open, closed, all) - default: "all"

**Response:**
```json
{
  "success": true,
  "message": "Found N pull requests",
  "data": [
    {
      "number": 123,
      "title": "PR Title",
      "state": "open",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z",
      "closed_at": null,
      "merged_at": null,
      "user": "username",
      "mergeable": true,
      "has_conflicts": false,
      "additions": 100,
      "deletions": 50,
      "changed_files": 5,
      "commits": 3,
      "comments": 2,
      "review_comments": 1,
      "head": "feature-branch",
      "base": "main"
    }
  ]
}
```

#### GET /api/v1/github/pulls/get
Get details of a specific pull request.

**Query Parameters:**
- `repo` (required): Repository name
- `number` (required): PR number

**Response:**
```json
{
  "success": true,
  "data": {
    "number": 123,
    "title": "PR Title",
    "state": "open",
    "created_at": "2024-01-01T00:00:00Z",
    "user": "username",
    "mergeable": true,
    "has_conflicts": false,
    "additions": 100,
    "deletions": 50
  }
}
```

### Commits

#### GET /api/v1/github/commits
List commits for a repository.

**Query Parameters:**
- `repo` (required): Repository name
- `since` (optional): ISO 8601 date string to filter commits

**Response:**
```json
{
  "success": true,
  "message": "Found N commits",
  "data": [
    {
      "sha": "abc123...",
      "message": "Commit message",
      "author": "Author Name",
      "date": "2024-01-01T00:00:00Z",
      "additions": 50,
      "deletions": 20,
      "total": 70
    }
  ]
}
```

#### GET /api/v1/github/commits/activity
Get commit activity analysis for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "data": {
    "total_commits": 500,
    "commits_last_90_days": 50,
    "is_active": true,
    "last_commit_date": "2024-01-01T00:00:00Z",
    "contributors": ["user1", "user2", "user3"]
  }
}
```

### Issues

#### GET /api/v1/github/issues
List issues for a repository.

**Query Parameters:**
- `repo` (required): Repository name
- `state` (optional): Issue state (open, closed, all) - default: "all"

**Response:**
```json
{
  "success": true,
  "message": "Found N issues",
  "data": [
    {
      "number": 456,
      "title": "Issue Title",
      "state": "open",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z",
      "closed_at": null,
      "user": "username",
      "comments": 5,
      "labels": ["bug", "priority-high"],
      "is_pull_request": false
    }
  ]
}
```

#### GET /api/v1/github/issues/comments
List comments for a specific issue.

**Query Parameters:**
- `repo` (required): Repository name
- `issue` (required): Issue number

**Response:**
```json
{
  "success": true,
  "message": "Found N comments",
  "data": [
    {
      "id": 123456,
      "user": "username",
      "body": "Comment text",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Repository

#### GET /api/v1/github/readme
Check if a repository has a README file.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "data": {
    "name": "README.md",
    "path": "README.md",
    "sha": "abc123...",
    "size": 1024,
    "exists": true
  }
}
```

#### GET /api/v1/github/branches
List branches for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "message": "Found N branches",
  "data": [
    {
      "name": "main",
      "protected": true,
      "sha": "abc123..."
    }
  ]
}
```

### Organization

#### GET /api/v1/github/org/members
List organization members.

**Response:**
```json
{
  "success": true,
  "message": "Found N members",
  "data": [
    {
      "login": "username",
      "id": 12345,
      "type": "User",
      "site_admin": false
    }
  ]
}
```

#### GET /api/v1/github/org/teams
List organization teams.

**Response:**
```json
{
  "success": true,
  "message": "Found N teams",
  "data": [
    {
      "id": 12345,
      "name": "Team Name",
      "slug": "team-name",
      "description": "Team description",
      "privacy": "closed",
      "members_count": 5
    }
  ]
}
```

### Repository Metrics

#### GET /api/v1/github/metrics
Get comprehensive metrics for a specific repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "data": {
    "repository": "repo-name",
    "has_readme": true,
    "default_branch": "main",
    "open_prs": 5,
    "closed_prs": 100,
    "merged_prs": 95,
    "prs_with_conflicts": 2,
    "open_issues": 10,
    "closed_issues": 50,
    "total_commits": 500,
    "commits_last_90_days": 50,
    "is_active": true,
    "last_commit_date": "2024-01-01T00:00:00Z",
    "contributors": 10,
    "branches": 5,
    "score": 85.5
  }
}
```

#### GET /api/v1/github/metrics/all
Get metrics for all repositories in the organization.

**Response:**
```json
{
  "success": true,
  "message": "Fetched metrics for N repositories",
  "data": [
    {
      "repository": "repo-name",
      "has_readme": true,
      "open_prs": 5,
      "score": 85.5
    }
  ]
}
```

---

## Jira Metrics

### GET /api/v1/jira/issues/stats
Get issue statistics for a Jira project.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "data": {
    "total_issues": 100,
    "open_issues": 20,
    "closed_issues": 80,
    "bugs": 15,
    "tasks": 50,
    "stories": 35,
    "avg_time_to_resolve": 48.5
  }
}
```

### GET /api/v1/jira/bugs/open
Get open bugs for a Jira project.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "message": "Found N open bugs",
  "data": [
    {
      "key": "PROJ-123",
      "summary": "Bug description",
      "status": "Open",
      "priority": "High",
      "assignee": "username",
      "created": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/jira/tasks/open
Get open tasks for a Jira project.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "message": "Found N open tasks",
  "data": [
    {
      "key": "PROJ-456",
      "summary": "Task description",
      "status": "In Progress",
      "assignee": "username"
    }
  ]
}
```

### GET /api/v1/jira/issues/by-assignee
Get issues grouped by assignee.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "data": {
    "username1": [
      {
        "key": "PROJ-123",
        "summary": "Issue summary",
        "status": "In Progress"
      }
    ],
    "username2": []
  }
}
```

### GET /api/v1/jira/sprints/stats
Get sprint statistics for a Jira project.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "data": {
    "active_sprints": 1,
    "completed_sprints": 10,
    "avg_sprint_duration": 14.5,
    "total_story_points": 100,
    "completed_story_points": 85
  }
}
```

### GET /api/v1/jira/metrics
Get comprehensive project metrics.

**Query Parameters:**
- `project` (required): Jira project key

**Response:**
```json
{
  "success": true,
  "data": {
    "project_key": "PROJ",
    "issue_stats": {
      "total_issues": 100,
      "open_issues": 20,
      "closed_issues": 80
    },
    "sprint_stats": {
      "active_sprints": 1,
      "completed_sprints": 10
    }
  }
}
```

### GET /api/v1/jira/issues/search
Search for Jira issues.

**Query Parameters:**
- `jql` (required): JQL query string
- `max_results` (optional): Maximum number of results (default: 50)

**Response:**
```json
{
  "success": true,
  "message": "Found N issues",
  "data": [
    {
      "key": "PROJ-123",
      "summary": "Issue summary",
      "status": "Open",
      "priority": "High"
    }
  ]
}
```

---

## Database-Backed Metrics

### GitHub Metrics

#### POST /api/v1/metrics/github/collect
Collect and store GitHub metrics for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "message": "GitHub metrics collected and stored successfully",
  "data": {
    "repo_id": 1,
    "open_prs": 5,
    "merged_prs": 95,
    "commits_last_90_days": 50,
    "score": 85.5,
    "collected_at": "2024-01-01T00:00:00Z"
  }
}
```

#### GET /api/v1/metrics/github/stored
Get stored GitHub metrics for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "repo_id": 1,
      "open_prs": 5,
      "merged_prs": 95,
      "score": 85.5,
      "collected_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### SonarCloud Metrics

#### POST /api/v1/metrics/sonar/collect
Collect and store SonarCloud metrics for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "message": "SonarCloud metrics collected and stored successfully",
  "data": {
    "repo_id": 1,
    "project_key": "org_repo-name",
    "quality_gate_status": "OK",
    "bugs": 0,
    "vulnerabilities": 0,
    "code_smells": 5,
    "coverage": 80.5,
    "lines_of_code": 10000,
    "collected_at": "2024-01-01T00:00:00Z"
  }
}
```

#### GET /api/v1/metrics/sonar/stored
Get stored SonarCloud metrics for a repository.

**Query Parameters:**
- `repo` (required): Repository name

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "repo_id": 1,
      "project_key": "org_repo-name",
      "quality_gate_status": "OK",
      "bugs": 0,
      "coverage": 80.5,
      "collected_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

## Organization Management

### GET /api/v1/orgs
Get all organizations.

**Response:**
```json
{
  "success": true,
  "message": "Found N organizations",
  "data": [
    {
      "id": 1,
      "name": "org-name",
      "sonar_org_key": "org-key",
      "jira_domain": "company.atlassian.net",
      "jira_email": "user@example.com",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### POST /api/v1/orgs/create
Create a new organization.

**Request Body:**
```json
{
  "name": "org-name",
  "github_pat": "ghp_...",
  "sonar_token": "sonar_token",
  "sonar_org_key": "org-key",
  "jira_token": "jira_token",
  "jira_domain": "company.atlassian.net",
  "jira_email": "user@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Organization 'org-name' created successfully",
  "data": {
    "id": 1,
    "name": "org-name",
    "sonar_org_key": "org-key",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Repository Management

### GET /api/v1/repos/fetch
Fetch repositories for an organization from GitHub.

**Query Parameters:**
- `org_id` (required): Organization ID

**Response:**
```json
{
  "success": true,
  "message": "Found N repositories",
  "data": [
    {
      "id": 1,
      "org_id": 1,
      "name": "repo-name",
      "github_url": "https://github.com/org/repo",
      "owner": "username",
      "last_commit_time": "2024-01-01T00:00:00Z",
      "last_commit_by": "Author Name",
      "is_active": true,
      "default_branch": "main",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### PUT /api/v1/repos/update
Update repository configuration.

**Query Parameters:**
- `repo_id` (required): Repository ID

**Request Body:**
```json
{
  "jira_project_key": "PROJ",
  "environment_name": "production"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Repository updated successfully",
  "data": {
    "id": 1,
    "name": "repo-name",
    "jira_project_key": "PROJ",
    "environment_name": "production",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Repository Metrics (Fetch and Store)

### GET /api/v1/repos/metrics/github
Fetch and store GitHub metrics for a repository.

**Query Parameters:**
- `repo_id` (required): Repository ID

**Response:**
```json
{
  "success": true,
  "message": "GitHub metrics fetched and saved successfully",
  "data": {
    "repo_id": 1,
    "stars": 100,
    "forks": 50,
    "watchers": 75,
    "open_prs": 5,
    "merged_prs": 95,
    "commits_last_90_days": 50,
    "contributors": 10,
    "has_readme": true,
    "score": 85.5,
    "collected_at": "2024-01-01T00:00:00Z"
  }
}
```

### GET /api/v1/repos/metrics/jira
Fetch and store Jira metrics for a repository.

**Query Parameters:**
- `repo_id` (required): Repository ID

**Response:**
```json
{
  "success": true,
  "message": "Jira metrics fetched and saved successfully",
  "data": {
    "repo_id": 1,
    "project_key": "PROJ",
    "open_bugs": 5,
    "open_tasks": 10,
    "open_issues": 20,
    "closed_issues": 80,
    "avg_time_to_resolve": 48.5,
    "active_sprints": 1,
    "completed_sprints": 10,
    "collected_at": "2024-01-01T00:00:00Z"
  }
}
```

### GET /api/v1/repos/metrics/sonar
Fetch and store SonarCloud metrics for a repository.

**Query Parameters:**
- `repo_id` (required): Repository ID

**Response:**
```json
{
  "success": true,
  "message": "SonarCloud metrics fetched and saved successfully",
  "data": {
    "repo_id": 1,
    "project_key": "org_repo-name",
    "quality_gate_status": "OK",
    "bugs": 0,
    "vulnerabilities": 0,
    "code_smells": 5,
    "coverage": 80.5,
    "duplicated_lines_density": 2.5,
    "lines_of_code": 10000,
    "security_rating": "A",
    "reliability_rating": "A",
    "maintainability_rating": "A",
    "technical_debt": "2h",
    "collected_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Standard Response Format

All API endpoints follow a standard response format:

### Success Response
```json
{
  "success": true,
  "message": "Optional success message",
  "data": {}
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

---

## HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request parameters
- **401 Unauthorized**: Missing or invalid API key
- **404 Not Found**: Resource not found
- **405 Method Not Allowed**: HTTP method not supported
- **409 Conflict**: Resource already exists
- **500 Internal Server Error**: Server error

---

## Notes

1. All timestamps are in ISO 8601 format (UTC)
2. Query parameters are case-sensitive
3. Request bodies must be valid JSON with `Content-Type: application/json`
4. The `/health` endpoint does not require authentication
5. Metrics are calculated and stored in the database for historical tracking
6. Repository names should match the GitHub repository name exactly
7. Organization and repository IDs are auto-generated database IDs


