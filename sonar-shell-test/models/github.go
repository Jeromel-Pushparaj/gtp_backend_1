package models

import "time"

// PullRequestInfo represents pull request information
type PullRequestInfo struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	MergedAt     *time.Time `json:"merged_at,omitempty"`
	User         string    `json:"user"`
	Mergeable    *bool     `json:"mergeable,omitempty"`
	HasConflicts bool      `json:"has_conflicts"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	ChangedFiles int       `json:"changed_files"`
	Commits      int       `json:"commits"`
	Comments     int       `json:"comments"`
	ReviewComments int     `json:"review_comments"`
	Head         string    `json:"head"`
	Base         string    `json:"base"`
}

// CommitInfo represents commit information
type CommitInfo struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Date      time.Time `json:"date"`
	Additions int       `json:"additions"`
	Deletions int       `json:"deletions"`
	Total     int       `json:"total"`
}

// CommitActivity represents commit activity analysis
type CommitActivity struct {
	TotalCommits      int       `json:"total_commits"`
	CommitsLast90Days int       `json:"commits_last_90_days"`
	IsActive          bool      `json:"is_active"`
	LastCommitDate    time.Time `json:"last_commit_date"`
	Contributors      []string  `json:"contributors"`
}

// IssueInfo represents issue information
type IssueInfo struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	User         string    `json:"user"`
	Comments     int       `json:"comments"`
	Labels       []string  `json:"labels"`
	IsPullRequest bool     `json:"is_pull_request"`
}

// CommentInfo represents comment information
type CommentInfo struct {
	ID        int64     `json:"id"`
	User      string    `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventInfo represents event information
type EventInfo struct {
	ID        int64     `json:"id"`
	Event     string    `json:"event"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

// BranchInfo represents branch information
type BranchInfo struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	SHA       string `json:"sha"`
}

// OrganizationMember represents organization member information
type OrganizationMember struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

// TeamInfo represents team information
type TeamInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
	MembersCount int   `json:"members_count"`
}

// RepositoryMetrics represents comprehensive repository metrics
type RepositoryMetrics struct {
	Repository       string          `json:"repository"`
	HasReadme        bool            `json:"has_readme"`
	DefaultBranch    string          `json:"default_branch"`
	OpenPRs          int             `json:"open_prs"`
	ClosedPRs        int             `json:"closed_prs"`
	MergedPRs        int             `json:"merged_prs"`
	PRsWithConflicts int             `json:"prs_with_conflicts"`
	OpenIssues       int             `json:"open_issues"`
	ClosedIssues     int             `json:"closed_issues"`
	TotalCommits     int             `json:"total_commits"`
	CommitsLast90Days int            `json:"commits_last_90_days"`
	IsActive         bool            `json:"is_active"`
	LastCommitDate   *time.Time      `json:"last_commit_date,omitempty"`
	Contributors     int             `json:"contributors"`
	Branches         int             `json:"branches"`
	Score            float64         `json:"score"`
}

// ScoreCalculation represents the score calculation breakdown
type ScoreCalculation struct {
	Repository        string  `json:"repository"`
	HasReadmeScore    float64 `json:"has_readme_score"`
	ActivityScore     float64 `json:"activity_score"`
	PRScore           float64 `json:"pr_score"`
	IssueScore        float64 `json:"issue_score"`
	ConflictPenalty   float64 `json:"conflict_penalty"`
	TotalScore        float64 `json:"total_score"`
	MaxScore          float64 `json:"max_score"`
	PercentageScore   float64 `json:"percentage_score"`
}

// ReadmeInfo represents README file information
type ReadmeInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
	Size     int    `json:"size"`
	Content  string `json:"content,omitempty"`
	Exists   bool   `json:"exists"`
}

