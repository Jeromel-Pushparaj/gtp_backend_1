package resources

import "time"

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ServiceResponse represents a comprehensive service object for frontend
type ServiceResponse struct {
	ID                string                 `json:"id"`
	Title             string                 `json:"title"`
	RepositoryURL     string                 `json:"repositoryUrl"`
	Owner             string                 `json:"owner"`
	DefaultBranch     string                 `json:"defaultBranch"`
	Language          string                 `json:"language,omitempty"`
	Organization      OrganizationInfo       `json:"organization"`
	JiraProjectKey    string                 `json:"jiraProjectKey,omitempty"`
	OnCall            string                 `json:"onCall"`
	Metrics           MetricsInfo            `json:"metrics"`
	EvaluationMetrics *EvaluationMetricsInfo `json:"evaluationMetrics,omitempty"`
	PullRequests      []PullRequestInfo      `json:"pullRequests"`
	JiraIssues        []JiraIssueInfo        `json:"jiraIssues"`
}

// OrganizationInfo represents organization information
type OrganizationInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// MetricsInfo represents aggregated metrics from GitHub and Jira
type MetricsInfo struct {
	OpenPullRequests  int `json:"openPullRequests"`
	CommitsLast90Days int `json:"commitsLast90Days"`
	Contributors      int `json:"contributors"`
	JiraOpenBugs      int `json:"jiraOpenBugs"`
	JiraOpenTasks     int `json:"jiraOpenTasks"`
	JiraActiveSprints int `json:"jiraActiveSprints"`
}

// EvaluationMetricsInfo represents scorecard evaluation metrics
type EvaluationMetricsInfo struct {
	ServiceName            string  `json:"serviceName"`
	Coverage               float64 `json:"coverage"`
	CodeSmells             int     `json:"codeSmells"`
	Vulnerabilities        int     `json:"vulnerabilities"`
	DuplicatedLinesDensity float64 `json:"duplicatedLinesDensity"`
	HasReadme              int     `json:"hasReadme"`
	DeploymentFrequency    int     `json:"deploymentFrequency"`
	MTTR                   int     `json:"mttr"`
}

// PullRequestInfo represents a pull request
type PullRequestInfo struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	URL       string    `json:"url"`
}

// JiraIssueInfo represents a Jira issue
type JiraIssueInfo struct {
	Key       string `json:"key"`
	Summary   string `json:"summary"`
	IssueType string `json:"issueType"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	Assignee  string `json:"assignee"`
}

// ServicesResponse represents a list of services for frontend
type ServicesResponse struct {
	Total    int               `json:"total"`
	Services []ServiceResponse `json:"services"`
}
