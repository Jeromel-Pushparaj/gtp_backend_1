# Postman API Testing Guide

Complete guide for testing the Jira Trigger Service API using Postman.

---

## 🚀 Server Setup

**Start the server:**
```bash
cd services/jira-trigger-service
go run main.go
```

**Server will run on:** `http://localhost:8086`

---

## 📋 Endpoints

### 1. Health Check

**Method:** `GET`
**URL:** `http://localhost:8086/health`
**Authorization:** No Auth
**Headers:** None

**Expected Response:**
```json
{
  "status": "ok"
}
```

---

## 🎯 Usage Scenarios

### Scenario 1: Create New Scrum Project + Board + Sprint + Issue

**Use Case:** You want to create a brand new Scrum project from scratch.

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Authorization:** No Auth
**Headers:**
```
Content-Type: application/json
```

**Body (raw JSON):**
```json
{
  "summary": "First task in new project",
  "projectKey": "NEWPROJ",
  "projectName": "My New Project",
  "projectType": "scrum",
  "issueType": "Story",
  "description": "Setting up the new project",
  "priority": "High"
}
```

**What happens:**
1. ✅ Creates project "NEWPROJ" with Scrum template
2. ✅ Creates board "My New Project Board"
3. ✅ Creates auto-named sprint (e.g., "Auto Sprint 2026-03-03 14:30")
4. ✅ Creates Story issue "NEWPROJ-1"
5. ✅ Adds issue to the sprint

**Expected Response:**
```json
{
  "success": true,
  "issueKey": "NEWPROJ-1",
  "issueUrl": "https://your-domain.atlassian.net/browse/NEWPROJ-1",
  "sprintId": 123,
  "message": "Issue NEWPROJ-1 created and added to sprint 'Auto Sprint 2026-03-03 14:30' successfully"
}
```

---

### Scenario 2: Create New Kanban Project + Board + Issue

**Use Case:** You want to create a Kanban board instead of Scrum.

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Authorization:** No Auth
**Headers:**
```
Content-Type: application/json
```

**Body (raw JSON):**
```json
{
  "summary": "First kanban card",
  "projectKey": "KANBAN1",
  "projectName": "My Kanban Board",
  "projectType": "kanban",
  "issueType": "Task",
  "description": "Setting up kanban workflow"
}
```

**What happens:**
1. ✅ Creates project "KANBAN1" with Kanban template
2. ✅ Creates board "My Kanban Board Board"
3. ✅ Creates Task issue "KANBAN1-1"
4. ⚠️ No sprint created (Kanban doesn't use sprints)

---

### Scenario 3: Add Issue to Existing Project + Create New Sprint

**Use Case:** Project "JIRATEST" already exists, but you want a new sprint.

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Authorization:** No Auth
**Headers:**
```
Content-Type: application/json
```

**Body (raw JSON):**
```json
{
  "summary": "Implement new feature",
  "projectKey": "JIRATEST",
  "issueType": "Story",
  "description": "Add user profile page",
  "priority": "Medium"
}
```

**What happens:**
1. ✅ Uses existing project "JIRATEST"
2. ✅ Creates NEW auto-named sprint (e.g., "Auto Sprint 2026-03-03 15:30")
3. ✅ Creates Story issue "JIRATEST-3"
4. ✅ Adds issue to the new sprint

---

### Scenario 4: Add Issue to Existing Project + Existing Sprint

**Use Case:** Both project "JIRATEST" and sprint "Auto Sprint 2026-03-03 11:15" exist.

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Authorization:** No Auth
**Headers:**
```
Content-Type: application/json
```

**Body (raw JSON):**
```json
{
  "summary": "Fix critical bug",
  "projectKey": "JIRATEST",
  "sprintName": "Auto Sprint 2026-03-03 11:15",
  "issueType": "Bug",
  "description": "Login fails for some users",
  "priority": "Highest"
}
```

**What happens:**
1. ✅ Uses existing project "JIRATEST"
2. ✅ Finds existing sprint "Auto Sprint 2026-03-03 11:15"
3. ✅ Creates Bug issue "JIRATEST-4"
4. ✅ Adds issue to the existing sprint (does NOT create new sprint)

---

## 📝 Additional Examples

### Example 1: Create Issue with Assignment

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Headers:** `Content-Type: application/json`

**Body:**
```json
{
  "summary": "Implement user authentication",
  "projectKey": "JIRATEST",
  "issueType": "Story",
  "description": "Add JWT-based authentication system",
  "assigneeName": "Keerthana U",
  "priority": "High"
}
```

---

### Example 2: Create Issue with Labels

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Headers:** `Content-Type: application/json`

**Body:**
```json
{
  "summary": "Update API documentation",
  "projectKey": "JIRATEST",
  "issueType": "Task",
  "description": "Update Swagger docs for authentication endpoints",
  "priority": "Medium",
  "labels": ["documentation", "api", "backend"]
}
```

---

### Example 3: Full Request with All Fields

**Method:** `POST`
**URL:** `http://localhost:8086/api/create-issue`
**Headers:** `Content-Type: application/json`

**Body:**
```json
{
  "summary": "Implement user authentication system",
  "projectKey": "BACKEND",
  "projectName": "Backend Services Platform",
  "projectType": "scrum",
  "issueType": "Story",
  "sprintName": "Sprint March 2026",
  "description": "Add comprehensive JWT authentication with refresh tokens, password reset, and email verification",
  "assigneeName": "Keerthana U",
  "priority": "High",
  "labels": ["backend", "security", "authentication", "api"]
}
```

---

## 📝 Complete Field Reference

### Required Fields (2)

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `summary` | string | Issue title | `"Fix login bug"` |
| `projectKey` | string | Project key (max 10 chars, uppercase + numbers only) | `"JIRATEST"` |

### Optional Fields (9)

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `projectName` | string | `"{projectKey} Project"` | Project display name (used when creating new project) | `"Jira Integration Test"` |
| `projectType` | string | `"scrum"` | Project type: `"scrum"` or `"kanban"` | `"scrum"` |
| `issueType` | string | `"Task"` | Issue type | `"Task"`, `"Story"`, `"Bug"`, `"Epic"` |
| `sprintName` | string | Auto-generated | Existing sprint name to use | `"Sprint March 2026"` |
| `description` | string | `"Created via API"` | Issue description | `"Add JWT authentication"` |
| `assigneeId` | string | Unassigned | Jira account ID (takes priority) | `"5b10a2844c20165700ede21g"` |
| `assigneeName` | string | Unassigned | Jira display name (fallback) | `"Keerthana U"` |
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

---

## ✅ Expected Success Response

```json
{
  "success": true,
  "issueKey": "JIRATEST-1",
  "issueUrl": "https://your-domain.atlassian.net/browse/JIRATEST-1",
  "sprintId": 177,
  "message": "Issue JIRATEST-1 created and added to sprint 'Auto Sprint 2026-03-03 11:15' successfully"
}
```

---

## ❌ Common Error Responses

### Missing Required Field: summary
```json
{
  "success": false,
  "error": "summary is required"
}
```

### Missing Required Field: projectKey
```json
{
  "success": false,
  "error": "projectKey is required"
}
```

### Invalid Project Key Format
```json
{
  "success": false,
  "error": "Failed to create project: Project keys must start with an uppercase letter..."
}
```

### Project Key Too Long
```json
{
  "success": false,
  "error": "Failed to create project: The project key must not exceed 10 characters in length."
}
```

### Sprint Not Found
```json
{
  "success": false,
  "error": "Failed to find sprint: sprint 'Sprint March 2026' not found"
}
```

### User Not Found
```json
{
  "success": false,
  "error": "Failed to create issue: user 'John Doe' not found or inactive"
}
```

---

## 🎯 Quick Tips

1. **Both `summary` and `projectKey` are required**
2. **Project key rules:** Max 10 chars, uppercase letters + numbers only, must start with letter
3. **No authentication needed** for localhost API (server handles Jira auth)
4. **Project creation:** If project doesn't exist, it's created automatically
5. **Sprint behavior:**
   - If `sprintName` is empty → Creates new auto-named sprint
   - If `sprintName` is provided → Uses existing sprint (fails if not found)
6. **Assignment:**
   - Use `assigneeId` for direct assignment (recommended)
   - Use `assigneeName` to search by display name (case-sensitive)
7. **Labels are case-sensitive** - use lowercase for consistency

---

## 📚 Summary

This service provides **complete Jira automation**:
- ✅ Creates projects (Scrum/Kanban) automatically
- ✅ Creates boards with proper filters
- ✅ Manages sprints (auto-create or use existing)
- ✅ Creates issues with full metadata
- ✅ All through simple HTTP requests

**No need to manually set up projects in Jira!** 🎉

