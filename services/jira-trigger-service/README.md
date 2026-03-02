# Jira Issue Creation API 🚀

A simple Go-based REST API that makes creating Jira issues super easy. No need to deal with complex Jira API authentication or sprint management - this server handles it all for you!

## What Does This Do?

This API lets you create Jira issues with a simple HTTP request. Just send a JSON with your issue details, and the server:
- ✅ Creates the issue in Jira
- ✅ Automatically creates and starts a sprint (if needed)
- ✅ Assigns the issue to a team member
- ✅ Adds labels and sets priority
- ✅ Returns the Jira issue link

## Quick Start

### 1. Setup Environment Variables

Create a `.env` file in the project root:

```env
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@example.com
JIRA_API_TOKEN=your-jira-api-token
JIRA_PROJECT_KEY=SCRUM
```

**How to get your Jira API Token:**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the token and paste it in your `.env` file

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8080`

## How to Use

### Simple Example - Create an Issue

```bash
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Fix login bug"
  }'
```

That's it! The server will:
- Create a Task in Jira
- Auto-create a new sprint
- Return the issue link

### Full Example - All Options

```bash
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Implement user authentication",
    "description": "Add JWT-based authentication system",
    "issueType": "Story",
    "sprintName": "Sprint March 2026",
    "assigneeName": "John Doe",
    "priority": "High",
    "labels": ["backend", "security", "authentication"]
  }'
```

### Response

```json
{
  "success": true,
  "issueKey": "SCRUM-22",
  "issueUrl": "https://your-domain.atlassian.net/browse/SCRUM-22",
  "sprintId": 106,
  "message": "Issue SCRUM-22 created and added to sprint successfully"
}
```

## API Reference

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Check if server is running |
| `/api/create-issue` | POST | Create a new Jira issue |

### Request Fields

| Field | Required? | Description | Example |
|-------|-----------|-------------|---------|
| `summary` | ✅ **YES** | Issue title | `"Fix login bug"` |
| `description` | No | Issue details | `"Users cannot login..."` |
| `issueType` | No | Task, Story, or Bug (default: Task) | `"Story"` |
| `sprintName` | No | Sprint name (auto-creates if empty) | `"Sprint March 2026"` |
| `assigneeName` | No | Display name of assignee | `"John Doe"` |
| `priority` | No | Highest, High, Medium, Low, Lowest | `"High"` |
| `labels` | No | Array of labels | `["backend", "urgent"]` |

## Examples

### Create a Bug

```bash
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Login fails with special characters",
    "description": "Users cannot login when password has @ or # symbols",
    "issueType": "Bug",
    "priority": "Highest",
    "labels": ["bug", "urgent", "production"]
  }'
```

### Create a Story and Assign It

```bash
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Add dark mode to dashboard",
    "description": "Implement dark mode theme for better UX",
    "issueType": "Story",
    "assigneeName": "Jane Smith",
    "priority": "Medium",
    "labels": ["frontend", "ui", "enhancement"]
  }'
```

## Project Structure

```
.
├── main.go              # Server entry point
├── config/              # Environment configuration
├── handlers/            # HTTP request handlers
├── services/            # Jira API business logic
├── models/              # Data structures
└── docs/                # API testing guides
```

## Need More Help?

Check out the detailed testing guides:
- [cURL Examples](docs/CURL_API_TESTING.md) - Complete cURL command reference
- [Postman Guide](docs/POSTMAN_API_TESTING.md) - How to test with Postman

## Tech Stack

- **Go 1.25.0** - Programming language
- **net/http** - HTTP server
- **godotenv** - Environment variable management
- **Jira REST API** - Issue creation and sprint management

---


