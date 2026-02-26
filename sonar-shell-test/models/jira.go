package models

import "time"

// JiraIssue represents a Jira issue
type JiraIssue struct {
	Key         string     `json:"key"`
	ID          string     `json:"id"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	IssueType   string     `json:"issue_type"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Assignee    string     `json:"assignee"`
	Reporter    string     `json:"reporter"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
	Resolved    *time.Time `json:"resolved,omitempty"`
	StoryPoints float64    `json:"story_points,omitempty"`
	Labels      []string   `json:"labels"`
	Components  []string   `json:"components"`
}

// JiraSprint represents a Jira sprint
type JiraSprint struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	State         string     `json:"state"` // future, active, closed
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	CompleteDate  *time.Time `json:"complete_date,omitempty"`
	Goal          string     `json:"goal,omitempty"`
	IssuesCount   int        `json:"issues_count"`
	CompletedCount int       `json:"completed_count"`
}

// JiraProject represents a Jira project
type JiraProject struct {
	Key         string `json:"key"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lead        string `json:"lead"`
	ProjectType string `json:"project_type"`
}

// JiraBoard represents a Jira board
type JiraBoard struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // scrum, kanban
	Location struct {
		ProjectKey string `json:"project_key"`
	} `json:"location"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	AccountID    string `json:"account_id"`
	DisplayName  string `json:"display_name"`
	EmailAddress string `json:"email_address"`
	Active       bool   `json:"active"`
}

// JiraIssueStats represents statistics for issues
type JiraIssueStats struct {
	TotalIssues      int     `json:"total_issues"`
	OpenIssues       int     `json:"open_issues"`
	InProgressIssues int     `json:"in_progress_issues"`
	ClosedIssues     int     `json:"closed_issues"`
	Bugs             int     `json:"bugs"`
	Tasks            int     `json:"tasks"`
	Stories          int     `json:"stories"`
	Epics            int     `json:"epics"`
	AvgTimeToResolve float64 `json:"avg_time_to_resolve"` // in hours
}

// JiraSprintStats represents statistics for sprints
type JiraSprintStats struct {
	TotalSprints      int     `json:"total_sprints"`
	ActiveSprints     int     `json:"active_sprints"`
	CompletedSprints  int     `json:"completed_sprints"`
	AvgSprintDuration float64 `json:"avg_sprint_duration"` // in days
	AvgVelocity       float64 `json:"avg_velocity"`        // story points per sprint
}

// JiraAssigneeStats represents statistics per assignee
type JiraAssigneeStats struct {
	Assignee         string  `json:"assignee"`
	TotalIssues      int     `json:"total_issues"`
	OpenIssues       int     `json:"open_issues"`
	ClosedIssues     int     `json:"closed_issues"`
	Bugs             int     `json:"bugs"`
	Tasks            int     `json:"tasks"`
	AvgTimeToResolve float64 `json:"avg_time_to_resolve"` // in hours
}

// JiraProjectMetrics represents comprehensive Jira project metrics
type JiraProjectMetrics struct {
	ProjectKey       string              `json:"project_key"`
	ProjectName      string              `json:"project_name"`
	IssueStats       JiraIssueStats      `json:"issue_stats"`
	SprintStats      JiraSprintStats     `json:"sprint_stats"`
	AssigneeStats    []JiraAssigneeStats `json:"assignee_stats"`
	RecentIssues     []JiraIssue         `json:"recent_issues,omitempty"`
	ActiveSprints    []JiraSprint        `json:"active_sprints,omitempty"`
	CollectedAt      time.Time           `json:"collected_at"`
}

// JiraSearchRequest represents a Jira search request
type JiraSearchRequest struct {
	JQL        string   `json:"jql"`
	MaxResults int      `json:"max_results"`
	StartAt    int      `json:"start_at"`
	Fields     []string `json:"fields"`
}

// JiraSearchResponse represents a Jira search response
type JiraSearchResponse struct {
	Total      int         `json:"total"`
	StartAt    int         `json:"start_at"`
	MaxResults int         `json:"max_results"`
	Issues     []JiraIssue `json:"issues"`
}

// JiraIssuesByType represents issues grouped by type
type JiraIssuesByType struct {
	Bugs    []JiraIssue `json:"bugs"`
	Tasks   []JiraIssue `json:"tasks"`
	Stories []JiraIssue `json:"stories"`
	Epics   []JiraIssue `json:"epics"`
	Others  []JiraIssue `json:"others"`
}

// JiraIssuesByStatus represents issues grouped by status
type JiraIssuesByStatus struct {
	Open       []JiraIssue `json:"open"`
	InProgress []JiraIssue `json:"in_progress"`
	Resolved   []JiraIssue `json:"resolved"`
	Closed     []JiraIssue `json:"closed"`
}

// JiraIssuesByPriority represents issues grouped by priority
type JiraIssuesByPriority struct {
	Highest []JiraIssue `json:"highest"`
	High    []JiraIssue `json:"high"`
	Medium  []JiraIssue `json:"medium"`
	Low     []JiraIssue `json:"low"`
	Lowest  []JiraIssue `json:"lowest"`
}

