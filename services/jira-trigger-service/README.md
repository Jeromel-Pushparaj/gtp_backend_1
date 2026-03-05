# Jira Trigger Service 🚀

A powerful Go-based REST API that simplifies Jira project and issue management. This service automatically handles project creation, board setup, sprint management, and issue creation - all through a single API call!

## ✨ What Does This Do?

This API provides complete Jira automation with a simple HTTP request. Just send a JSON with your requirements, and the server:
- ✅ **Creates Jira projects** (Scrum or Kanban) if they don't exist
- ✅ **Creates boards** automatically for new projects
- ✅ **Creates and manages sprints** (auto-creates or uses existing)
- ✅ **Creates issues** with full metadata (type, priority, labels, description)
- ✅ **Assigns issues** to team members by name or account ID
- ✅ **Adds issues to sprints** automatically
- ✅ Returns the Jira issue link and complete details

## 🚀 Quick Start

### 1. Setup Environment Variables

Create a `.env` file in the project root:

```env
JIRA_BASE_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@example.com
JIRA_API_TOKEN=your-jira-api-token
```

**How to get your Jira API Token:**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the token and paste it in your `.env` file

**Note:** For project creation, your API token needs **admin permissions** in Jira.

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Server

```bash
go run main.go
```

Server will start on `http://localhost:8086`

## 📖 How to Use

### Example 1: Create New Scrum Project + Issue

```bash
curl -X POST http://localhost:8086/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "First task in new project",
    "projectKey": "JIRATEST",
    "projectName": "Jira Integration Test",
    "projectType": "scrum"
  }'
```

**What happens:**
- ✅ Creates project "JIRATEST" with Scrum template
- ✅ Creates "Jira Integration Test Board"
- ✅ Creates auto-named sprint (e.g., "Auto Sprint 2026-03-03 11:15")
- ✅ Creates issue and adds to sprint

### Example 2: Create New Kanban Project + Issue

```bash
curl -X POST http://localhost:8086/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "First kanban card",
    "projectKey": "KANBAN1",
    "projectName": "My Kanban Board",
    "projectType": "kanban"
  }'
```

### Example 3: Add Issue to Existing Project + Existing Sprint

```bash
curl -X POST http://localhost:8086/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Add feature to existing sprint",
    "projectKey": "JIRATEST",
    "sprintName": "Auto Sprint 2026-03-03 11:15",
    "issueType": "Story",
    "priority": "High"
  }'
```

**What happens:**
- ✅ Uses existing project "JIRATEST"
- ✅ Finds existing sprint "Auto Sprint 2026-03-03 11:15"
- ✅ Creates issue and adds to that sprint

### Example 4: Full Request with All Fields

```bash
curl -X POST http://localhost:8086/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
    "summary": "Implement user authentication",
    "projectKey": "BACKEND",
    "projectName": "Backend Services",
    "projectType": "scrum",
    "issueType": "Story",
    "sprintName": "Sprint March 2026",
    "description": "Add JWT-based authentication system",
    "assigneeName": "John Doe",
    "priority": "High",
    "labels": ["backend", "security", "authentication"]
  }'
```

### Success Response

```json
{
  "success": true,
  "issueKey": "JIRATEST-1",
  "issueUrl": "https://your-domain.atlassian.net/browse/JIRATEST-1",
  "sprintId": 177,
  "message": "Issue JIRATEST-1 created and added to sprint 'Auto Sprint 2026-03-03 11:15' successfully"
}
```

### Error Response

```json
{
  "success": false,
  "error": "projectKey is required"
}
```

## 📋 API Reference

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check - returns `{"status": "ok"}` |
| `/api/create-issue` | POST | Create project, sprint, and issue |

### Request Fields

#### **Required Fields** (2)

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `summary` | string | Issue title | `"Fix login bug"` |
| `projectKey` | string | Project key (max 10 chars, uppercase + numbers only) | `"JIRATEST"` |

#### **Optional Fields** (9)

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `projectName` | string | `"{projectKey} Project"` | Project display name | `"Jira Integration Test"` |
| `projectType` | string | `"scrum"` | Project type: `"scrum"` or `"kanban"` | `"scrum"` |
| `issueType` | string | `"Task"` | Issue type | `"Task"`, `"Story"`, `"Bug"`, `"Epic"` |
| `sprintName` | string | Auto-generated | Existing sprint name to use | `"Sprint March 2026"` |
| `description` | string | `"Created via API"` | Issue description | `"Add JWT authentication"` |
| `assigneeId` | string | Unassigned | Jira account ID (takes priority) | `"5b10a2844c20165700ede21g"` |
| `assigneeName` | string | Unassigned | Jira display name (fallback) | `"John Doe"` |
| `priority` | string | None | Priority level | `"Highest"`, `"High"`, `"Medium"`, `"Low"`, `"Lowest"` |
| `labels` | array | `[]` | Array of labels | `["backend", "security"]` |

### Project Key Rules

⚠️ **Important:** Project keys must follow these rules:
- ✅ Start with uppercase letter (A-Z)
- ✅ Only uppercase letters and numbers (A-Z, 0-9)
- ✅ Maximum 10 characters
- ❌ No hyphens, underscores, or special characters

**Valid:** `JIRATEST`, `BACKEND`, `PROJ123`, `TEST`
**Invalid:** `jira-test`, `test_proj`, `my.project`, `TOOLONGPROJECTKEY`

## 🎯 Workflow Scenarios

### Scenario 1: Create New Project + Auto Sprint
```json
{
  "summary": "First task",
  "projectKey": "NEWPROJ"
}
```
**Result:** Creates project, board, auto-named sprint, and issue

---

### Scenario 2: Use Existing Project + Create New Sprint
```json
{
  "summary": "New feature",
  "projectKey": "JIRATEST"
}
```
**Result:** Uses existing project, creates new auto-named sprint, adds issue

---

### Scenario 3: Use Existing Project + Existing Sprint
```json
{
  "summary": "Bug fix",
  "projectKey": "JIRATEST",
  "sprintName": "Auto Sprint 2026-03-03 11:15"
}
```
**Result:** Uses existing project and sprint, adds issue

---

### Scenario 4: Create Kanban Project
```json
{
  "summary": "Kanban card",
  "projectKey": "KANBAN1",
  "projectType": "kanban"
}
```
**Result:** Creates Kanban project and board (no sprints)

---

## 🔄 Sprint Behavior

### Auto-Generated Sprint Names
When `sprintName` is **not provided**, the service creates a new sprint with format:
```
Auto Sprint YYYY-MM-DD HH:MM
```

**Examples:**
- `Auto Sprint 2026-03-03 11:15`
- `Auto Sprint 2026-03-03 14:30`
- `Auto Sprint 2026-12-25 09:00`

**Sprint Details:**
- Start Date: Current date/time
- End Date: 14 days later
- State: Active (automatically started)

### Using Existing Sprints
To add issues to an existing sprint, provide the exact sprint name:
```json
{
  "summary": "Add to existing sprint",
  "projectKey": "JIRATEST",
  "sprintName": "Auto Sprint 2026-03-03 11:15"
}
```

---

## 👥 Assignment Options

### Option 1: By Account ID (Recommended)
```json
{
  "assigneeId": "5b10a2844c20165700ede21g"
}
```

### Option 2: By Display Name
```json
{
  "assigneeName": "John Doe"
}
```

**Note:** If both are provided, `assigneeId` takes priority.

---

## 🐛 Common Issues & Solutions

### Error: "projectKey is required"
**Solution:** Add `"projectKey": "YOURKEY"` to your request

### Error: "Project keys must start with an uppercase letter..."
**Solution:** Use only uppercase letters and numbers (e.g., `JIRATEST` not `jira-test`)

### Error: "The project key must not exceed 10 characters"
**Solution:** Shorten your project key (e.g., `JIRATESTPROJ` → `JIRATEST`)

### Error: "Failed to find sprint"
**Solution:** Either remove `sprintName` to auto-create, or use exact existing sprint name

### Error: "user 'Name' not found"
**Solution:** Use exact Jira display name or use `assigneeId` instead

### Error: "You do not have permission to create projects"
**Solution:** Your API token needs admin permissions in Jira

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


