package models

import "time"

// Organization represents an organization in the database
type Organization struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	GitHubPAT    string    `json:"-" db:"github_pat"` // Hidden from JSON
	SonarToken   string    `json:"-" db:"sonar_token"` // Hidden from JSON
	SonarOrgKey  string    `json:"sonar_org_key" db:"sonar_org_key"`
	JiraToken    string    `json:"-" db:"jira_token"` // Hidden from JSON
	JiraDomain   string    `json:"jira_domain" db:"jira_domain"`
	JiraEmail    string    `json:"jira_email" db:"jira_email"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Repository represents a repository in the database
type Repository struct {
	ID              int64      `json:"id" db:"id"`
	OrgID           int64      `json:"org_id" db:"org_id"`
	Name            string     `json:"name" db:"name"`
	GitHubURL       string     `json:"github_url" db:"github_url"`
	Owner           string     `json:"owner" db:"owner"`
	LastCommitTime  *time.Time `json:"last_commit_time,omitempty" db:"last_commit_time"`
	LastCommitBy    string     `json:"last_commit_by,omitempty" db:"last_commit_by"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	DefaultBranch   string     `json:"default_branch" db:"default_branch"`
	EnvName         string     `json:"env_name" db:"env_name"`
	EnvironmentName string     `json:"environment_name,omitempty" db:"env_name"` // Alias for EnvName
	JiraSpaceKey    string     `json:"jira_space_key,omitempty" db:"jira_space_key"`
	JiraProjectKey  string     `json:"jira_project_key,omitempty" db:"jira_space_key"` // Alias for JiraSpaceKey
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// GitHubMetrics represents GitHub metrics for a repository
type GitHubMetrics struct {
	ID                 int64      `json:"id" db:"id"`
	RepoID             int64      `json:"repo_id" db:"repo_id"`
	OpenPRs            int64      `json:"open_prs" db:"open_prs"`
	ClosedPRs          int64      `json:"closed_prs" db:"closed_prs"`
	MergedPRs          int64      `json:"merged_prs" db:"merged_prs"`
	TotalPRs           int64      `json:"total_prs" db:"total_commits"` // Reusing total_commits column
	PRsWithConflicts   int64      `json:"prs_with_conflicts" db:"prs_with_conflicts"`
	OpenIssues         int64      `json:"open_issues" db:"open_issues"`
	ClosedIssues       int64      `json:"closed_issues" db:"closed_issues"`
	OpenIssuesCount    int64      `json:"open_issues_count" db:"open_issues"` // Alias
	TotalCommits       int64      `json:"total_commits" db:"total_commits"`
	CommitsLast90Days  int64      `json:"commits_last_90_days" db:"commits_last_90_days"`
	Contributors       int64      `json:"contributors" db:"contributors"`
	Branches           int64      `json:"branches" db:"branches"`
	TotalBranches      int64      `json:"total_branches" db:"branches"` // Alias
	HasReadme          bool       `json:"has_readme" db:"has_readme"`
	Stars              int64      `json:"stars" db:"score"`              // Reusing score column
	Forks              int64      `json:"forks" db:"prs_with_conflicts"` // Temporary mapping
	Watchers           int64      `json:"watchers" db:"closed_issues"`   // Temporary mapping
	Score              float64    `json:"score" db:"score"`
	LastCommitDate     *time.Time `json:"last_commit_date,omitempty" db:"last_commit_date"`
	CollectedAt        time.Time  `json:"collected_at" db:"collected_at"`
}

// SonarMetrics represents SonarCloud metrics for a repository
type SonarMetrics struct {
	ID                     int64     `json:"id" db:"id"`
	RepoID                 int64     `json:"repo_id" db:"repo_id"`
	ProjectKey             string    `json:"project_key" db:"project_key"`
	QualityGateStatus      string    `json:"quality_gate_status" db:"quality_gate_status"`
	Bugs                   int       `json:"bugs" db:"bugs"`
	Vulnerabilities        int       `json:"vulnerabilities" db:"vulnerabilities"`
	CodeSmells             int       `json:"code_smells" db:"code_smells"`
	Coverage               float64   `json:"coverage" db:"coverage"`
	DuplicatedLinesDensity float64   `json:"duplicated_lines_density" db:"duplicated_lines_density"`
	LinesOfCode            int       `json:"lines_of_code" db:"lines_of_code"`
	SecurityRating         string    `json:"security_rating" db:"security_rating"`
	ReliabilityRating      string    `json:"reliability_rating" db:"reliability_rating"`
	MaintainabilityRating  string    `json:"maintainability_rating" db:"maintainability_rating"`
	TechnicalDebt          string    `json:"technical_debt" db:"technical_debt"`
	CollectedAt            time.Time `json:"collected_at" db:"collected_at"`
}

// JiraMetrics represents Jira metrics for a repository/project
type JiraMetrics struct {
	ID                  int64     `json:"id" db:"id"`
	RepoID              int64     `json:"repo_id" db:"repo_id"`
	ProjectKey          string    `json:"project_key" db:"project_key"`
	OpenBugs            int       `json:"open_bugs" db:"open_bugs"`
	ClosedBugs          int       `json:"closed_bugs" db:"closed_bugs"`
	OpenTasks           int       `json:"open_tasks" db:"open_tasks"`
	ClosedTasks         int       `json:"closed_tasks" db:"closed_tasks"`
	OpenIssues          int       `json:"open_issues" db:"open_issues"`
	ClosedIssues        int       `json:"closed_issues" db:"closed_issues"`
	AvgTimeToResolve    float64   `json:"avg_time_to_resolve" db:"avg_time_to_resolve"` // in hours
	AvgSprintTime       float64   `json:"avg_sprint_time" db:"avg_sprint_time"`         // in days
	ActiveSprints       int       `json:"active_sprints" db:"active_sprints"`
	CompletedSprints    int       `json:"completed_sprints" db:"completed_sprints"`
	TotalStoryPoints    int       `json:"total_story_points" db:"total_story_points"`
	CompletedStoryPoints int      `json:"completed_story_points" db:"completed_story_points"`
	CollectedAt         time.Time `json:"collected_at" db:"collected_at"`
}

// JiraIssueAssignee represents issue assignee information
type JiraIssueAssignee struct {
	ID          int64     `json:"id" db:"id"`
	RepoID      int64     `json:"repo_id" db:"repo_id"`
	IssueKey    string    `json:"issue_key" db:"issue_key"`
	IssueType   string    `json:"issue_type" db:"issue_type"`
	Status      string    `json:"status" db:"status"`
	Priority    string    `json:"priority" db:"priority"`
	Assignee    string    `json:"assignee" db:"assignee"`
	Reporter    string    `json:"reporter" db:"reporter"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CollectedAt time.Time `json:"collected_at" db:"collected_at"`
}

// DatabaseStats represents overall database statistics
type DatabaseStats struct {
	TotalOrganizations int       `json:"total_organizations"`
	TotalRepositories  int       `json:"total_repositories"`
	ActiveRepositories int       `json:"active_repositories"`
	LastUpdated        time.Time `json:"last_updated"`
}

