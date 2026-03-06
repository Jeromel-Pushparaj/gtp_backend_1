package models

import "time"

// Organization represents an organization from sonar-shell-test
type Organization struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SonarOrgKey string    `json:"sonar_org_key"`
	JiraDomain  string    `json:"jira_domain"`
	JiraEmail   string    `json:"jira_email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository represents a repository from sonar-shell-test
type Repository struct {
	ID              int64      `json:"id"`
	OrgID           int64      `json:"org_id"`
	Name            string     `json:"name"`
	GitHubURL       string     `json:"github_url"`
	Owner           string     `json:"owner"`
	LastCommitTime  *time.Time `json:"last_commit_time,omitempty"`
	LastCommitBy    string     `json:"last_commit_by,omitempty"`
	IsActive        bool       `json:"is_active"`
	DefaultBranch   string     `json:"default_branch"`
	EnvironmentName string     `json:"environment_name,omitempty"`
	JiraProjectKey  string     `json:"jira_project_key,omitempty"`
	PrimaryLanguage *string    `json:"primary_language,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// GitHubMetrics represents GitHub metrics from sonar-shell-test
type GitHubMetrics struct {
	ID                int64      `json:"id"`
	RepoID            int64      `json:"repo_id"`
	OpenPRs           int64      `json:"open_prs"`
	ClosedPRs         int64      `json:"closed_prs"`
	MergedPRs         int64      `json:"merged_prs"`
	PRsWithConflicts  int64      `json:"prs_with_conflicts"`
	OpenIssues        int64      `json:"open_issues"`
	ClosedIssues      int64      `json:"closed_issues"`
	TotalCommits      int64      `json:"total_commits"`
	CommitsLast90Days int64      `json:"commits_last_90_days"`
	Contributors      int64      `json:"contributors"`
	Branches          int64      `json:"branches"`
	HasReadme         bool       `json:"has_readme"`
	Score             float64    `json:"score"`
	LastCommitDate    *time.Time `json:"last_commit_date,omitempty"`
	CollectedAt       time.Time  `json:"collected_at"`
}

// JiraMetrics represents Jira metrics from sonar-shell-test
type JiraMetrics struct {
	ID               int64     `json:"id"`
	RepoID           int64     `json:"repo_id"`
	ProjectKey       string    `json:"project_key"`
	OpenBugs         int64     `json:"open_bugs"`
	ClosedBugs       int64     `json:"closed_bugs"`
	OpenTasks        int64     `json:"open_tasks"`
	ClosedTasks      int64     `json:"closed_tasks"`
	OpenIssues       int64     `json:"open_issues"`
	ClosedIssues     int64     `json:"closed_issues"`
	AvgTimeToResolve float64   `json:"avg_time_to_resolve"`
	AvgSprintTime    float64   `json:"avg_sprint_time"`
	ActiveSprints    int64     `json:"active_sprints"`
	CompletedSprints int64     `json:"completed_sprints"`
	CollectedAt      time.Time `json:"collected_at"`
}

// PullRequest represents a pull request from GitHub API
type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	User      string     `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
	URL       string     `json:"url"`
	Mergeable bool       `json:"mergeable"`
}

// Issue represents a GitHub issue
type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	User      string     `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	URL       string     `json:"url"`
	Labels    []string   `json:"labels"`
}

// JiraIssue represents a Jira issue
type JiraIssue struct {
	Key       string     `json:"key"`
	ID        string     `json:"id"`
	Summary   string     `json:"summary"`
	IssueType string     `json:"issue_type"`
	Status    string     `json:"status"`
	Priority  string     `json:"priority"`
	Assignee  string     `json:"assignee"`
	Reporter  string     `json:"reporter"`
	Created   time.Time  `json:"created"`
	Updated   time.Time  `json:"updated"`
	Resolved  *time.Time `json:"resolved,omitempty"`
}

// EvaluationMetrics represents scorecard evaluation metrics from sonar-shell-test
type EvaluationMetrics struct {
	ServiceName            string  `json:"service_name"`
	Coverage               float64 `json:"coverage"`
	CodeSmells             int     `json:"code_smells"`
	Vulnerabilities        int     `json:"vulnerabilities"`
	DuplicatedLinesDensity float64 `json:"duplicated_lines_density"`
	HasReadme              int     `json:"has_readme"`
	DeploymentFrequency    int     `json:"deployment_frequency"`
	MTTR                   int     `json:"mttr"`
}

// SonarMetrics represents SonarCloud metrics from sonar-shell-test database
type SonarMetrics struct {
	ID                     int64   `json:"id"`
	RepoID                 int64   `json:"repo_id"`
	ProjectKey             string  `json:"project_key"`
	QualityGateStatus      string  `json:"quality_gate_status"`
	Bugs                   int64   `json:"bugs"`
	Vulnerabilities        int64   `json:"vulnerabilities"`
	CodeSmells             int64   `json:"code_smells"`
	Coverage               float64 `json:"coverage"`
	DuplicatedLinesDensity float64 `json:"duplicated_lines_density"`
	LinesOfCode            int64   `json:"lines_of_code"`
	SecurityRating         string  `json:"security_rating"`
	ReliabilityRating      string  `json:"reliability_rating"`
	MaintainabilityRating  string  `json:"maintainability_rating"`
	TechnicalDebt          string  `json:"technical_debt"`
	CollectedAt            string  `json:"collected_at"`
}
