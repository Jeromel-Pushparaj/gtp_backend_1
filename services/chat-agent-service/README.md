# Chat Agent Service

This service provides two modes of operation:

1. **HTTP Server Mode** - A chatbot API powered by Groq AI that can be integrated with your frontend
2. **MCP Server Mode** - A Model Context Protocol server for AI assistants like Claude and Cursor

## Overview

### HTTP Server Mode (Recommended for Frontend Integration)

The HTTP server exposes a REST API on port 8082 that allows your frontend to interact with an AI chatbot. The chatbot uses Groq's API and can access your GTP Backend API through tool calling.

**Use this mode when:**
- You want to integrate a chatbot into your web application
- You need a REST API for your frontend
- You want to expose the chatbot to end users

### MCP Server Mode (For AI Development Tools)

The MCP server acts as a bridge between AI assistants (like Claude Desktop or Cursor) and the GTP Backend API, allowing AI tools to interact with GitHub, SonarCloud, and Jira metrics through a standardized protocol.

**Use this mode when:**
- You want to use Claude Desktop or Cursor with your backend
- You're developing with AI-assisted tools
- You need stdio-based MCP protocol support

## Prerequisites

- Go 1.21 or higher
- Access to the GTP Backend API (default: http://localhost:8080)
- Groq API key (for HTTP server mode) - Get one at https://console.groq.com
- Optional: Backend API key for authentication

## Quick Start

### HTTP Server Mode (For Frontend Integration)

1. **Get a Groq API Key**
   - Visit https://console.groq.com
   - Sign up and create an API key

2. **Configure the service**
   ```bash
   cd services/chat-agent-service
   cp .env.example .env
   # Edit .env and add your GROQ_API_KEY
   ```

3. **Build and run**
   ```bash
   go build -o chat-agent-server cmd/server/main.go
   ./chat-agent-server
   ```

4. **Test it**
   ```bash
   curl -X POST http://localhost:8082/api/v1/chat \
     -H "Content-Type: application/json" \
     -d '{"message": "What is the health status of the backend?"}'
   ```

See [HTTP Server Guide](docs/HTTP_SERVER_GUIDE.md) for detailed documentation.

### MCP Server Mode (For AI Development Tools)

1. **Build the MCP server**
   ```bash
   cd services/chat-agent-service
   go build -o mcp-server cmd/main.go
   ```

2. **Configure Claude Desktop or Cursor**
   See [Client Setup Guide](docs/CLIENT_SETUP.md) for detailed instructions.

## Configuration

The server can be configured using environment variables or command-line flags.

### Environment Variables

Create a `.env` file in the service directory:

```env
API_BASE_URL=http://localhost:8080
API_KEY=your_api_key_here
```

### Command-Line Flags

```bash
./mcp-server -base-url=http://localhost:8080 -api-key=your_api_key
```

## Running the Server

```bash
go run cmd/main.go
```

Or using the built binary:

```bash
./mcp-server
```

## Available Tools

The MCP server exposes the following categories of tools:

### Health Check
- `health_check` - Check the health status of the service

### SonarCloud Management
- `list_secrets` - List all secrets for repositories
- `add_env_secrets` - Add environment secrets to repositories
- `update_workflows` - Update GitHub workflows
- `full_setup` - Perform full SonarCloud setup
- `fetch_results` - Fetch SonarCloud analysis results
- `get_sonar_metrics` - Get SonarCloud metrics for a repository
- `process_repository` - Process a single repository

### GitHub Metrics
- `list_pull_requests` - List pull requests for a repository
- `get_pull_request` - Get details of a specific pull request
- `list_commits` - List commits for a repository
- `get_commit_activity` - Get commit activity analysis
- `list_issues` - List issues for a repository
- `list_issue_comments` - List comments for an issue
- `check_readme` - Check if a repository has a README
- `list_branches` - List branches for a repository
- `list_org_members` - List organization members
- `list_org_teams` - List organization teams
- `get_repository_metrics` - Get comprehensive repository metrics
- `get_all_repositories_metrics` - Get metrics for all repositories

### Jira Metrics
- `get_jira_issue_stats` - Get issue statistics
- `get_jira_open_bugs` - Get open bugs
- `get_jira_open_tasks` - Get open tasks
- `get_jira_issues_by_assignee` - Get issues by assignee
- `get_jira_sprint_stats` - Get sprint statistics
- `get_jira_project_metrics` - Get comprehensive project metrics
- `search_jira_issues` - Search for Jira issues using JQL

### Database-Backed Metrics
- `collect_github_metrics` - Collect and store GitHub metrics
- `get_stored_github_metrics` - Get stored GitHub metrics
- `collect_sonar_metrics` - Collect and store SonarCloud metrics
- `get_stored_sonar_metrics` - Get stored SonarCloud metrics

### Organization Management
- `fetch_orgs` - Get all organizations
- `create_org` - Create a new organization

### Repository Management
- `fetch_repos_by_org` - Fetch repositories for an organization
- `update_repo` - Update repository configuration
- `fetch_github_metrics_by_repo` - Fetch GitHub metrics by repository ID
- `fetch_jira_metrics_by_repo` - Fetch Jira metrics by repository ID
- `fetch_sonar_metrics_by_repo` - Fetch SonarCloud metrics by repository ID

## Configuring MCP Clients

### Claude Desktop

Edit your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/path/to/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": "your_api_key"
      }
    }
  }
}
```

### Cursor IDE

Add to your `mcp.json`:

```json
{
  "mcpServers": {
    "gtp-backend": {
      "command": "/path/to/services/chat-agent-service/mcp-server",
      "args": [],
      "env": {
        "API_BASE_URL": "http://localhost:8080",
        "API_KEY": "your_api_key"
      }
    }
  }
}
```

## Development

The MCP server is organized into the following files:

- `cmd/main.go` - Main entry point
- `mcp/server.go` - Core server implementation
- `mcp/sonarcloud_tools.go` - SonarCloud-related tools
- `mcp/github_tools.go` - GitHub-related tools
- `mcp/jira_tools.go` - Jira-related tools
- `mcp/metrics_tools.go` - Metrics collection tools
- `mcp/org_tools.go` - Organization management tools
- `mcp/repo_tools.go` - Repository management tools

## API Documentation

For detailed API endpoint documentation, see [docs/API_ENDPOINTS.md](../../docs/API_ENDPOINTS.md)

