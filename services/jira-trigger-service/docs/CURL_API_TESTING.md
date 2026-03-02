# cURL API Testing Guide

Complete guide for testing the Jira Issue Creation API using cURL commands.

---

## 🚀 Server Setup

**Start the server:**
```bash
go run main.go
```

**Server will run on:** `http://localhost:8080`

---

## 📋 All Endpoints

### 1. Health Check

```bash
curl --request GET \
  --url 'http://localhost:8080/health'
```

**Expected Response:**
```json
{"status":"ok"}
```

---

### 2. Create Issue - Minimal (Only Summary)

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Fix login bug"
}'
```

**What happens:**
- Creates Task issue
- Auto-creates sprint with name like "Auto Sprint 2026-03-01 14:30"
- Default description: "Created via API"

---

### 3. Create Issue - With Description and Issue Type

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Fix login bug",
  "description": "Users cannot login with special characters in password",
  "issueType": "Bug",
  "priority": "High"
}'
```

**What happens:**
- Creates Bug issue with High priority
- Auto-creates new sprint
- Custom description

---

### 4. Create Issue - Add to Existing Sprint

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Fix dashboard bug",
  "description": "Dashboard not loading properly",
  "issueType": "Bug",
  "sprintName": "Sprint March 2026",
  "priority": "Medium"
}'
```

**What happens:**
- Creates Bug issue
- Searches for existing sprint "Sprint March 2026"
- Adds issue to that sprint (does NOT create new sprint)

---

### 5. Create Issue - Assign by Display Name

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Implement user authentication",
  "description": "Add JWT-based authentication system",
  "issueType": "Story",
  "assigneeName": "saru",
  "priority": "High"
}'
```

**What happens:**
- Creates Story issue
- Searches for user with display name "saru"
- Assigns issue to that user
- Auto-creates sprint

---

### 6. Create Issue - With Labels

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Update API documentation",
  "description": "Update Swagger docs for authentication endpoints",
  "issueType": "Task",
  "priority": "Medium",
  "labels": ["documentation", "api", "backend"]
}'
```

**What happens:**
- Creates Task issue
- Adds 3 labels: documentation, api, backend
- Auto-creates sprint

---

### 7. Create Issue - Full Request (All Fields)

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "projectKey": "SCRUM",
  "summary": "Implement user authentication system",
  "description": "Add comprehensive user authentication with JWT tokens, password reset, email verification, and OAuth2 integration",
  "issueType": "Story",
  "sprintName": "Sprint March 2026",
  "assigneeName": "saru",
  "priority": "High",
  "labels": ["backend", "security", "authentication", "v2.0"]
}'
```

**What happens:**
- Creates Story issue in SCRUM project
- Adds to existing sprint "Sprint March 2026"
- Assigns to user "saru"
- Sets priority to High
- Adds 4 labels

---

### 8. Create Bug - Urgent with All Fields

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "projectKey": "SCRUM",
  "summary": "Login fails with special characters in password",
  "description": "Users cannot login when their password contains special characters like @, #, $, %. This affects approximately 15% of users based on error logs.",
  "issueType": "Bug",
  "sprintName": "Sprint March 2026",
  "assigneeName": "saru",
  "priority": "Highest",
  "labels": ["bug", "login", "urgent", "hotfix", "production"]
}'
```

---

### 9. Create Task - Documentation

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "projectKey": "SCRUM",
  "summary": "Update API documentation for authentication endpoints",
  "description": "Update Swagger/OpenAPI documentation to include new authentication endpoints: /login, /signup, /reset-password, /verify-email",
  "issueType": "Task",
  "assigneeName": "saru",
  "priority": "Medium",
  "labels": ["documentation", "api", "swagger", "backend"]
}'
```

---

## 🔍 Pretty Print JSON Response

Add `| jq` to format the response:

```bash
curl --request POST \
  --url 'http://localhost:8080/api/create-issue' \
  --header 'Content-Type: application/json' \
  --data '{
  "summary": "Fix login bug"
}' | jq
```

**Note:** Requires `jq` to be installed. Install with:
- **Mac:** `brew install jq`
- **Linux:** `sudo apt-get install jq`

---

## 📝 Field Reference

| Field | Type | Required? | Example | Description |
|-------|------|-----------|---------|-------------|
| `summary` | string | **YES** | `"Fix login bug"` | Issue title |
| `projectKey` | string | No | `"SCRUM"` | Project key (uses .env default if empty) |
| `description` | string | No | `"Users cannot login..."` | Issue description (default: "Created via API") |
| `issueType` | string | No | `"Story"` | Task, Story, Bug, Epic (default: "Task") |
| `sprintName` | string | No | `"Sprint March 2026"` | Sprint name (auto-creates if empty) |
| `assigneeName` | string | No | `"saru"` | Display name of assignee |
| `assigneeId` | string | No | `"712020:xxxxx..."` | Account ID (takes priority over assigneeName) |
| `priority` | string | No | `"High"` | Highest, High, Medium, Low, Lowest |
| `labels` | array | No | `["backend", "security"]` | Array of label strings |

---

## ✅ Expected Success Response

```json
{
  "success": true,
  "issueKey": "SCRUM-22",
  "issueUrl": "https://intern-poc-proj.atlassian.net/browse/SCRUM-22",
  "sprintId": 106,
  "message": "Issue SCRUM-22 created and added to sprint 'Sprint March 2026' successfully"
}
```

---

## ❌ Expected Error Responses

### Missing Summary
```json
{
  "success": false,
  "error": "summary is required"
}
```

### Invalid JSON
```json
{
  "success": false,
  "error": "Invalid JSON body: ..."
}
```

---

## 🎯 Quick Tips

1. **Only `summary` is required** - all other fields are optional
2. **No authentication needed** for localhost API (server handles Jira auth)
3. **Sprint auto-creation:** If `sprintName` is empty, creates new sprint automatically
4. **Assign by name:** Use `assigneeName` with exact display name (case-sensitive)
5. **Labels are case-sensitive** - use lowercase for consistency
6. **Use `\` for multi-line commands** in bash/zsh
7. **Use `| jq` for pretty JSON output** (requires jq installation)

---

## 🧪 Quick Test Sequence

```bash
# 1. Check server health
curl http://localhost:8080/health

# 2. Create minimal issue
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{"summary": "Test issue"}'

# 3. Create full issue
curl -X POST http://localhost:8080/api/create-issue \
  -H 'Content-Type: application/json' \
  -d '{
  "summary": "Full test",
  "description": "Testing all fields",
  "issueType": "Story",
  "assigneeName": "saru",
  "priority": "High",
  "labels": ["test", "demo"]
}' | jq
```

