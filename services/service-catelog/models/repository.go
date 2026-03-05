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

// CachedRepository represents a cached repository in local database
type CachedRepository struct {
	ID              int64      `db:"id"`
	RepositoryID    int64      `db:"repository_id"`
	Name            string     `db:"name"`
	GitHubURL       string     `db:"github_url"`
	Owner           string     `db:"owner"`
	LastCommitTime  *time.Time `db:"last_commit_time"`
	LastCommitBy    string     `db:"last_commit_by"`
	DefaultBranch   string     `db:"default_branch"`
	EnvironmentName string     `db:"environment_name"`
	JiraProjectKey  string     `db:"jira_project_key"`
	CachedAt        time.Time  `db:"cached_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

