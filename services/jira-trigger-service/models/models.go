package models

// CreateIssueRequest represents the incoming request body
type CreateIssueRequest struct {
	// Required fields
	Summary    string `json:"summary"`    // Issue title (REQUIRED)
	ProjectKey string `json:"projectKey"` // Project key (REQUIRED, e.g., "SCRUM", "KANBAN1")

	// Optional fields
	ProjectName  string   `json:"projectName"`  // Project name (optional, used when creating new project)
	ProjectType  string   `json:"projectType"`  // Project type: "scrum" or "kanban" (optional, default: "scrum")
	IssueType    string   `json:"issueType"`    // Type: Task, Story, Bug (optional, default: Task)
	SprintName   string   `json:"sprintName"`   // Name of the sprint (optional - if empty, creates new sprint)
	Description  string   `json:"description"`  // Issue description (optional, default: "Created via API")
	AssigneeID   string   `json:"assigneeId"`   // Jira account ID (optional, takes priority over assigneeName)
	AssigneeName string   `json:"assigneeName"` // Jira display name (optional, e.g., "Ganesh Sriramulu")
	Priority     string   `json:"priority"`     // Priority: Highest, High, Medium, Low, Lowest (optional)
	Labels       []string `json:"labels"`       // Array of labels (optional)
}

// CreateIssueResponse represents the API response
type CreateIssueResponse struct {
	Success  bool   `json:"success"`
	IssueKey string `json:"issueKey,omitempty"`
	IssueURL string `json:"issueUrl,omitempty"`
	SprintID int    `json:"sprintId,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}
