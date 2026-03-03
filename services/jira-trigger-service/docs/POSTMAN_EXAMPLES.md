# Postman Test Examples

Copy these JSON bodies directly into Postman for testing!

**Endpoint:** `POST http://localhost:8080/api/create-issue`  
**Headers:** `Content-Type: application/json`  
**Authorization:** No Auth

---

## Example 1: Minimal (Only Summary)
```json
{
  "summary": "Fix login bug"
}
```
**Creates:** Task with auto-generated sprint

---

## Example 2: With Sprint Name
```json
{
  "summary": "Update user profile",
  "sprintName": "Sprint March 2024"
}
```
**Creates:** Task in existing sprint "Sprint March 2024"

---

## Example 3: Bug with High Priority
```json
{
  "summary": "Critical security vulnerability",
  "issueType": "Bug",
  "priority": "Highest"
}
```
**Creates:** Bug with Highest priority, auto-generated sprint

---

## Example 4: Story with Description
```json
{
  "summary": "User can reset password",
  "issueType": "Story",
  "description": "As a user, I want to reset my password via email so that I can regain access to my account"
}
```
**Creates:** Story with description, auto-generated sprint

---

## Example 5: Task with Assignee (by ID)
```json
{
  "summary": "Implement login API",
  "assigneeId": "712020:89339710-f1ec-4142-afc6-eec8f5044097"
}
```
**Creates:** Task assigned to specific user by account ID, auto-generated sprint

---

## Example 5b: Task with Assignee (by Name)
```json
{
  "summary": "Implement login API",
  "assigneeName": "Ganesh Sriramulu"
}
```
**Creates:** Task assigned to user by display name, auto-generated sprint

---

## Example 6: Task with Labels
```json
{
  "summary": "Refactor authentication module",
  "labels": ["backend", "refactoring", "tech-debt"]
}
```
**Creates:** Task with 3 labels, auto-generated sprint

---

## Example 7: Story with Multiple Labels and Priority
```json
{
  "summary": "Implement OAuth2 integration",
  "issueType": "Story",
  "priority": "High",
  "labels": ["backend", "security", "oauth", "v2.0"]
}
```
**Creates:** Story with High priority and 4 labels, auto-generated sprint

---

## Example 8: Full Request (All Fields)
```json
{
  "projectKey": "SCRUM",
  "summary": "Implement user authentication",
  "description": "Add login and signup functionality with JWT tokens. Include password reset and email verification.",
  "issueType": "Story",
  "sprintName": "Sprint March 2024",
  "assigneeId": "712020:89339710-f1ec-4142-afc6-eec8f5044097",
  "priority": "High",
  "labels": ["backend", "security", "authentication"]
}
```
**Creates:** Story in SCRUM project, existing sprint, assigned, High priority, 3 labels

---

## Example 9: Bug with All Optional Fields
```json
{
  "summary": "Login fails with special characters in password",
  "description": "Users cannot login when their password contains special characters like @, #, $",
  "issueType": "Bug",
  "priority": "Highest",
  "assigneeId": "712020:89339710-f1ec-4142-afc6-eec8f5044097",
  "labels": ["bug", "login", "urgent", "hotfix"]
}
```
**Creates:** Bug with Highest priority, assigned, 4 labels, auto-generated sprint

---

## Example 10: Task in Specific Project and Sprint
```json
{
  "projectKey": "SCRUM",
  "summary": "Update API documentation",
  "description": "Update Swagger docs for new authentication endpoints",
  "issueType": "Task",
  "sprintName": "Sprint March 2024",
  "priority": "Medium",
  "labels": ["documentation", "api"]
}
```
**Creates:** Task in SCRUM project, existing sprint, Medium priority, 2 labels

---

## Quick Copy-Paste for Testing

### Test 1: Minimal
```json
{"summary": "Test issue 1"}
```

### Test 2: With Sprint
```json
{"summary": "Test issue 2", "sprintName": "Sprint March 2024"}
```

### Test 3: Bug
```json
{"summary": "Test bug", "issueType": "Bug", "priority": "High"}
```

### Test 4: With Assignee Name
```json
{"summary": "Test with assignee", "assigneeName": "Ganesh Sriramulu"}
```

### Test 5: Story with Labels
```json
{"summary": "Test story", "issueType": "Story", "labels": ["test", "demo"]}
```

### Test 6: Full
```json
{
  "summary": "Full test",
  "description": "Testing all fields",
  "issueType": "Task",
  "priority": "Medium",
  "assigneeName": "Ganesh Sriramulu",
  "labels": ["test"]
}
```

---

## Expected Success Response
```json
{
  "success": true,
  "issueKey": "SCRUM-XX",
  "issueUrl": "https://intern-poc-proj.atlassian.net/browse/SCRUM-XX",
  "sprintId": 75,
  "message": "Issue SCRUM-XX created and added to sprint 'Sprint Name' successfully"
}
```

---

**Pro Tip:** Save these as separate requests in a Postman collection for easy testing!

