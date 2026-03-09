# MCP Tools Reference

This document lists all available MCP tools exposed by the chat-agent-service.

## Health Check

### health_check
Check the health status of the service.

**Arguments:** None

**Example:**
```
Check the health of the service
```

## SonarCloud Management

### list_secrets
List all secrets for repositories in the organization.

**Arguments:** None

### add_env_secrets
Add environment secrets to all repositories.

**Arguments:** None

### update_workflows
Update GitHub workflows to use environment secrets.

**Arguments:** None

### full_setup
Perform full SonarCloud setup for all repositories.

**Arguments:** None

### fetch_results
Fetch SonarCloud analysis results for all repositories.

**Arguments:** None

### get_sonar_metrics
Get SonarCloud metrics for a specific repository.

**Arguments:**
- `repo` (required): Repository name
- `include_issues` (optional): Set to "true" to include issue details

### process_repository
Process a single repository for SonarCloud setup.

**Arguments:**
- `repository_name` (required): Repository name to process

## GitHub Metrics

### list_pull_requests
List pull requests for a repository.

**Arguments:**
- `repo` (required): Repository name
- `state` (optional): PR state (open, closed, all)

### get_pull_request
Get details of a specific pull request.

**Arguments:**
- `repo` (required): Repository name
- `number` (required): PR number

### list_commits
List commits for a repository.

**Arguments:**
- `repo` (required): Repository name
- `since` (optional): ISO 8601 date string to filter commits

### get_commit_activity
Get commit activity analysis for a repository.

**Arguments:**
- `repo` (required): Repository name

### list_issues
List issues for a repository.

**Arguments:**
- `repo` (required): Repository name
- `state` (optional): Issue state (open, closed, all)

### list_issue_comments
List comments for a specific issue.

**Arguments:**
- `repo` (required): Repository name
- `issue` (required): Issue number

### check_readme
Check if a repository has a README file.

**Arguments:**
- `repo` (required): Repository name

### list_branches
List branches for a repository.

**Arguments:**
- `repo` (required): Repository name

### list_org_members
List organization members.

**Arguments:** None

### list_org_teams
List organization teams.

**Arguments:** None

### get_repository_metrics
Get comprehensive metrics for a specific repository.

**Arguments:**
- `repo` (required): Repository name

### get_all_repositories_metrics
Get metrics for all repositories in the organization.

**Arguments:** None

## Jira Metrics

### get_jira_issue_stats
Get issue statistics for a Jira project.

**Arguments:**
- `project` (required): Jira project key

### get_jira_open_bugs
Get open bugs for a Jira project.

**Arguments:**
- `project` (required): Jira project key

### get_jira_open_tasks
Get open tasks for a Jira project.

**Arguments:**
- `project` (required): Jira project key

### get_jira_issues_by_assignee
Get issues grouped by assignee.

**Arguments:**
- `project` (required): Jira project key

### get_jira_sprint_stats
Get sprint statistics for a Jira project.

**Arguments:**
- `project` (required): Jira project key

### get_jira_project_metrics
Get comprehensive project metrics.

**Arguments:**
- `project` (required): Jira project key

### search_jira_issues
Search for Jira issues using JQL.

**Arguments:**
- `jql` (required): JQL query string
- `max_results` (optional): Maximum number of results

## Database-Backed Metrics

### collect_github_metrics
Collect and store GitHub metrics for a repository.

**Arguments:**
- `repo` (required): Repository name

### get_stored_github_metrics
Get stored GitHub metrics for a repository.

**Arguments:**
- `repo` (required): Repository name

### collect_sonar_metrics
Collect and store SonarCloud metrics for a repository.

**Arguments:**
- `repo` (required): Repository name

### get_stored_sonar_metrics
Get stored SonarCloud metrics for a repository.

**Arguments:**
- `repo` (required): Repository name

## Organization Management

### fetch_orgs
Get all organizations.

**Arguments:** None

### create_org
Create a new organization.

**Arguments:**
- `name` (required): Organization name
- `github_pat` (required): GitHub Personal Access Token
- `sonar_token` (required): SonarCloud token
- `sonar_org_key` (required): SonarCloud organization key
- `jira_token` (required): Jira API token
- `jira_domain` (required): Jira domain
- `jira_email` (required): Jira user email

## Repository Management

### fetch_repos_by_org
Fetch repositories for an organization from GitHub.

**Arguments:**
- `org_id` (required): Organization ID

### update_repo
Update repository configuration.

**Arguments:**
- `repo_id` (required): Repository ID
- `jira_project_key` (optional): Jira project key
- `environment_name` (optional): Environment name

### fetch_github_metrics_by_repo
Fetch and store GitHub metrics for a repository.

**Arguments:**
- `repo_id` (required): Repository ID

### fetch_jira_metrics_by_repo
Fetch and store Jira metrics for a repository.

**Arguments:**
- `repo_id` (required): Repository ID

### fetch_sonar_metrics_by_repo
Fetch and store SonarCloud metrics for a repository.

**Arguments:**
- `repo_id` (required): Repository ID

