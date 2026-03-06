# Postman Test Examples

Quick copy-paste JSON examples for testing the Jira Trigger Service in Postman!

**Endpoint:** `POST http://localhost:8086/api/create-issue`
**Headers:** `Content-Type: application/json`
**Authorization:** No Auth

---

## 🎯 Main Scenarios

### Scenario 1: Create New Scrum Project + Board + Sprint + Issue
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
**Creates:** New Scrum project, board, auto-named sprint, and first issue

---

### Scenario 2: Create New Kanban Project + Board + Issue
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
**Creates:** New Kanban project, board, and first issue (no sprint)

---

### Scenario 3: Add Issue to Existing Project + Create New Sprint
```json
{
  "summary": "Implement new feature",
  "projectKey": "JIRATEST",
  "issueType": "Story",
  "description": "Add user profile page",
  "priority": "Medium"
}
```
**Creates:** Issue in existing project with NEW auto-named sprint

---

### Scenario 4: Add Issue to Existing Project + Existing Sprint
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
**Creates:** Issue in existing project and existing sprint

---

## 📝 Additional Examples

### Example 1: Issue with Assignment (by Name)
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

### Example 2: Issue with Assignment (by ID)
```json
{
  "summary": "Implement login API",
  "projectKey": "JIRATEST",
  "assigneeId": "712020:89339710-f1ec-4142-afc6-eec8f5044097",
  "issueType": "Task"
}
```

---

### Example 3: Issue with Labels
```json
{
  "summary": "Refactor authentication module",
  "projectKey": "JIRATEST",
  "issueType": "Task",
  "labels": ["backend", "refactoring", "tech-debt"]
}
```

---

### Example 4: Bug with High Priority
```json
{
  "summary": "Critical security vulnerability",
  "projectKey": "JIRATEST",
  "issueType": "Bug",
  "priority": "Highest",
  "description": "SQL injection vulnerability in login form",
  "labels": ["security", "urgent", "hotfix"]
}
```

---

### Example 5: Full Request with All Fields
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

## 🚀 Quick Copy-Paste for Testing

### Test 1: Create New Project
```json
{"summary": "First task", "projectKey": "TESTPROJ", "projectName": "Test Project"}
```

### Test 2: Add to Existing Project
```json
{"summary": "Second task", "projectKey": "TESTPROJ", "issueType": "Story"}
```

### Test 3: Use Existing Sprint
```json
{"summary": "Third task", "projectKey": "TESTPROJ", "sprintName": "Auto Sprint 2026-03-03 11:15"}
```

### Test 4: Bug with Priority
```json
{"summary": "Critical bug", "projectKey": "TESTPROJ", "issueType": "Bug", "priority": "Highest"}
```

### Test 5: With Assignee
```json
{"summary": "Assigned task", "projectKey": "TESTPROJ", "assigneeName": "Keerthana U"}
```

### Test 6: With Labels
```json
{"summary": "Labeled task", "projectKey": "TESTPROJ", "labels": ["test", "demo", "api"]}
```

### Test 7: Kanban Project
```json
{"summary": "Kanban card", "projectKey": "KANTEST", "projectType": "kanban"}
```

---

## ✅ Expected Success Response

```json
{
  "success": true,
  "issueKey": "TESTPROJ-1",
  "issueUrl": "https://your-domain.atlassian.net/browse/TESTPROJ-1",
  "sprintId": 177,
  "message": "Issue TESTPROJ-1 created and added to sprint 'Auto Sprint 2026-03-03 11:15' successfully"
}
```

---

## ❌ Common Error Responses

### Missing Required Field
```json
{
  "success": false,
  "error": "projectKey is required"
}
```

### Invalid Project Key
```json
{
  "success": false,
  "error": "Failed to create project: Project keys must start with an uppercase letter..."
}
```

### Sprint Not Found
```json
{
  "success": false,
  "error": "Failed to find sprint: sprint 'Sprint March 2026' not found"
}
```

---

## 📋 Field Reference

### Required (2)
- `summary` - Issue title
- `projectKey` - Project key (max 10 chars, uppercase + numbers only)

### Optional (9)
- `projectName` - Project display name
- `projectType` - `"scrum"` or `"kanban"` (default: `"scrum"`)
- `issueType` - `"Task"`, `"Story"`, `"Bug"`, `"Epic"` (default: `"Task"`)
- `sprintName` - Existing sprint name (auto-creates if empty)
- `description` - Issue description
- `assigneeId` - Jira account ID
- `assigneeName` - Jira display name
- `priority` - `"Highest"`, `"High"`, `"Medium"`, `"Low"`, `"Lowest"`
- `labels` - Array of strings

---

## 💡 Pro Tips

1. **Project Key Rules:** Max 10 chars, uppercase letters + numbers only, must start with letter
2. **Valid Keys:** `JIRATEST`, `BACKEND`, `PROJ123`, `TEST`
3. **Invalid Keys:** `jira-test`, `test_proj`, `my.project`, `TOOLONGPROJECTKEY`
4. **Sprint Behavior:**
   - Empty `sprintName` → Creates new auto-named sprint
   - Provided `sprintName` → Uses existing sprint (fails if not found)
5. **Save these as a Postman collection** for easy testing!

---

## 🎉 Summary

This service provides **complete Jira automation**:
- ✅ Creates projects (Scrum/Kanban) automatically
- ✅ Creates boards with proper filters
- ✅ Manages sprints (auto-create or use existing)
- ✅ Creates issues with full metadata
- ✅ All through simple HTTP requests

**No need to manually set up projects in Jira!** 🚀

