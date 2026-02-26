package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"sonar-automation/models"
)

// DatabaseService handles all database operations
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	service := &DatabaseService{db: db}

	// Initialize schema
	if err := service.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return service, nil
}

// Close closes the database connection
func (ds *DatabaseService) Close() error {
	return ds.db.Close()
}

// initSchema creates all necessary tables
func (ds *DatabaseService) initSchema() error {
	schema := `
	-- Organizations table
	CREATE TABLE IF NOT EXISTS organizations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		github_pat TEXT NOT NULL,
		sonar_token TEXT,
		sonar_org_key TEXT,
		jira_token TEXT,
		jira_domain TEXT,
		jira_email TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Repositories table
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		org_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		github_url TEXT,
		owner TEXT NOT NULL,
		last_commit_time DATETIME,
		last_commit_by TEXT,
		is_active BOOLEAN DEFAULT 1,
		default_branch TEXT DEFAULT 'main',
		env_name TEXT DEFAULT 'production',
		jira_space_key TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		UNIQUE(org_id, name)
	);

	-- GitHub metrics table
	CREATE TABLE IF NOT EXISTS github_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_id INTEGER NOT NULL,
		open_prs INTEGER DEFAULT 0,
		closed_prs INTEGER DEFAULT 0,
		merged_prs INTEGER DEFAULT 0,
		prs_with_conflicts INTEGER DEFAULT 0,
		open_issues INTEGER DEFAULT 0,
		closed_issues INTEGER DEFAULT 0,
		total_commits INTEGER DEFAULT 0,
		commits_last_90_days INTEGER DEFAULT 0,
		contributors INTEGER DEFAULT 0,
		branches INTEGER DEFAULT 0,
		has_readme BOOLEAN DEFAULT 0,
		score REAL DEFAULT 0.0,
		last_commit_date DATETIME,
		collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
	);

	-- SonarCloud metrics table
	CREATE TABLE IF NOT EXISTS sonar_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_id INTEGER NOT NULL,
		project_key TEXT NOT NULL,
		quality_gate_status TEXT,
		bugs INTEGER DEFAULT 0,
		vulnerabilities INTEGER DEFAULT 0,
		code_smells INTEGER DEFAULT 0,
		coverage REAL DEFAULT 0.0,
		duplicated_lines_density REAL DEFAULT 0.0,
		lines_of_code INTEGER DEFAULT 0,
		security_rating TEXT,
		reliability_rating TEXT,
		maintainability_rating TEXT,
		technical_debt TEXT,
		collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
	);

	-- Jira metrics table
	CREATE TABLE IF NOT EXISTS jira_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_id INTEGER NOT NULL,
		project_key TEXT NOT NULL,
		open_bugs INTEGER DEFAULT 0,
		closed_bugs INTEGER DEFAULT 0,
		open_tasks INTEGER DEFAULT 0,
		closed_tasks INTEGER DEFAULT 0,
		open_issues INTEGER DEFAULT 0,
		closed_issues INTEGER DEFAULT 0,
		avg_time_to_resolve REAL DEFAULT 0.0,
		avg_sprint_time REAL DEFAULT 0.0,
		active_sprints INTEGER DEFAULT 0,
		completed_sprints INTEGER DEFAULT 0,
		total_story_points INTEGER DEFAULT 0,
		completed_story_points INTEGER DEFAULT 0,
		collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
	);

	-- Jira issue assignees table
	CREATE TABLE IF NOT EXISTS jira_issue_assignees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_id INTEGER NOT NULL,
		issue_key TEXT NOT NULL,
		issue_type TEXT,
		status TEXT,
		priority TEXT,
		assignee TEXT,
		reporter TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		resolved_at DATETIME,
		collected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
	);

	-- Indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_repos_org_id ON repositories(org_id);
	CREATE INDEX IF NOT EXISTS idx_repos_name ON repositories(name);
	CREATE INDEX IF NOT EXISTS idx_github_metrics_repo_id ON github_metrics(repo_id);
	CREATE INDEX IF NOT EXISTS idx_sonar_metrics_repo_id ON sonar_metrics(repo_id);
	CREATE INDEX IF NOT EXISTS idx_jira_metrics_repo_id ON jira_metrics(repo_id);
	CREATE INDEX IF NOT EXISTS idx_jira_assignees_repo_id ON jira_issue_assignees(repo_id);
	CREATE INDEX IF NOT EXISTS idx_jira_assignees_issue_key ON jira_issue_assignees(issue_key);
	`

	_, err := ds.db.Exec(schema)
	return err
}

// ═══════════════════════════════════════════════════════════════
// Organization Operations
// ═══════════════════════════════════════════════════════════════

// CreateOrganization creates a new organization
func (ds *DatabaseService) CreateOrganization(org *models.Organization) error {
	query := `INSERT INTO organizations (name, github_pat, sonar_token, sonar_org_key, jira_token, jira_domain, jira_email)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, org.Name, org.GitHubPAT, org.SonarToken, org.SonarOrgKey,
		org.JiraToken, org.JiraDomain, org.JiraEmail)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	org.ID = id
	return nil
}

// GetOrganization gets an organization by ID
func (ds *DatabaseService) GetOrganization(id int64) (*models.Organization, error) {
	query := `SELECT id, name, github_pat, sonar_token, sonar_org_key, jira_token, jira_domain, jira_email, created_at, updated_at
			  FROM organizations WHERE id = ?`

	org := &models.Organization{}
	err := ds.db.QueryRow(query, id).Scan(&org.ID, &org.Name, &org.GitHubPAT, &org.SonarToken,
		&org.SonarOrgKey, &org.JiraToken, &org.JiraDomain, &org.JiraEmail, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

// GetOrganizationByID gets an organization by ID (alias for GetOrganization)
func (ds *DatabaseService) GetOrganizationByID(id int64) (*models.Organization, error) {
	return ds.GetOrganization(id)
}

// GetOrganizationByName gets an organization by name
func (ds *DatabaseService) GetOrganizationByName(name string) (*models.Organization, error) {
	query := `SELECT id, name, github_pat, sonar_token, sonar_org_key, jira_token, jira_domain, jira_email, created_at, updated_at
			  FROM organizations WHERE name = ?`

	org := &models.Organization{}
	err := ds.db.QueryRow(query, name).Scan(&org.ID, &org.Name, &org.GitHubPAT, &org.SonarToken,
		&org.SonarOrgKey, &org.JiraToken, &org.JiraDomain, &org.JiraEmail, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

// GetAllOrganizations gets all organizations
func (ds *DatabaseService) GetAllOrganizations() ([]*models.Organization, error) {
	query := `SELECT id, name, github_pat, sonar_token, sonar_org_key, jira_token, jira_domain, jira_email, created_at, updated_at
			  FROM organizations ORDER BY name`

	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(&org.ID, &org.Name, &org.GitHubPAT, &org.SonarToken,
			&org.SonarOrgKey, &org.JiraToken, &org.JiraDomain, &org.JiraEmail, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

// UpdateOrganization updates an organization
func (ds *DatabaseService) UpdateOrganization(org *models.Organization) error {
	query := `UPDATE organizations SET github_pat = ?, sonar_token = ?, sonar_org_key = ?,
			  jira_token = ?, jira_domain = ?, jira_email = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := ds.db.Exec(query, org.GitHubPAT, org.SonarToken, org.SonarOrgKey,
		org.JiraToken, org.JiraDomain, org.JiraEmail, org.ID)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════
// Repository Operations
// ═══════════════════════════════════════════════════════════════

// CreateRepository creates a new repository
func (ds *DatabaseService) CreateRepository(repo *models.Repository) error {
	query := `INSERT INTO repositories (org_id, name, github_url, owner, last_commit_time, last_commit_by,
			  is_active, default_branch, env_name, jira_space_key)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, repo.OrgID, repo.Name, repo.GitHubURL, repo.Owner,
		repo.LastCommitTime, repo.LastCommitBy, repo.IsActive, repo.DefaultBranch, repo.EnvName, repo.JiraSpaceKey)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	repo.ID = id
	return nil
}

// GetRepository gets a repository by ID
func (ds *DatabaseService) GetRepository(id int64) (*models.Repository, error) {
	query := `SELECT id, org_id, name, github_url, owner, last_commit_time, last_commit_by,
			  is_active, default_branch, env_name, jira_space_key, created_at, updated_at
			  FROM repositories WHERE id = ?`

	repo := &models.Repository{}
	err := ds.db.QueryRow(query, id).Scan(&repo.ID, &repo.OrgID, &repo.Name, &repo.GitHubURL,
		&repo.Owner, &repo.LastCommitTime, &repo.LastCommitBy, &repo.IsActive, &repo.DefaultBranch,
		&repo.EnvName, &repo.JiraSpaceKey, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	// Set aliases
	repo.JiraProjectKey = repo.JiraSpaceKey
	repo.EnvironmentName = repo.EnvName
	return repo, nil
}

// GetRepositoryByID gets a repository by ID (alias for GetRepository)
func (ds *DatabaseService) GetRepositoryByID(id int64) (*models.Repository, error) {
	return ds.GetRepository(id)
}

// GetRepositoryByName gets a repository by org ID and name
func (ds *DatabaseService) GetRepositoryByName(orgID int64, name string) (*models.Repository, error) {
	query := `SELECT id, org_id, name, github_url, owner, last_commit_time, last_commit_by,
			  is_active, default_branch, env_name, jira_space_key, created_at, updated_at
			  FROM repositories WHERE org_id = ? AND name = ?`

	repo := &models.Repository{}
	err := ds.db.QueryRow(query, orgID, name).Scan(&repo.ID, &repo.OrgID, &repo.Name, &repo.GitHubURL,
		&repo.Owner, &repo.LastCommitTime, &repo.LastCommitBy, &repo.IsActive, &repo.DefaultBranch,
		&repo.EnvName, &repo.JiraSpaceKey, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	// Set aliases
	repo.JiraProjectKey = repo.JiraSpaceKey
	repo.EnvironmentName = repo.EnvName
	return repo, nil
}

// ListRepositories lists all repositories for an organization
func (ds *DatabaseService) ListRepositories(orgID int64) ([]*models.Repository, error) {
	query := `SELECT id, org_id, name, github_url, owner, last_commit_time, last_commit_by,
			  is_active, default_branch, env_name, jira_space_key, created_at, updated_at
			  FROM repositories WHERE org_id = ? ORDER BY name`

	rows, err := ds.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	var repos []*models.Repository
	for rows.Next() {
		repo := &models.Repository{}
		err := rows.Scan(&repo.ID, &repo.OrgID, &repo.Name, &repo.GitHubURL, &repo.Owner,
			&repo.LastCommitTime, &repo.LastCommitBy, &repo.IsActive, &repo.DefaultBranch,
			&repo.EnvName, &repo.JiraSpaceKey, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Set aliases
		repo.JiraProjectKey = repo.JiraSpaceKey
		repo.EnvironmentName = repo.EnvName
		repos = append(repos, repo)
	}
	return repos, nil
}

// UpdateRepository updates a repository
func (ds *DatabaseService) UpdateRepository(repo *models.Repository) error {
	query := `UPDATE repositories SET github_url = ?, owner = ?, last_commit_time = ?, last_commit_by = ?,
			  is_active = ?, default_branch = ?, env_name = ?, jira_space_key = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := ds.db.Exec(query, repo.GitHubURL, repo.Owner, repo.LastCommitTime, repo.LastCommitBy,
		repo.IsActive, repo.DefaultBranch, repo.EnvName, repo.JiraSpaceKey, repo.ID)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════
// GitHub Metrics Operations
// ═══════════════════════════════════════════════════════════════

// SaveGitHubMetrics saves GitHub metrics for a repository
func (ds *DatabaseService) SaveGitHubMetrics(metrics *models.GitHubMetrics) error {
	query := `INSERT INTO github_metrics (repo_id, open_prs, closed_prs, merged_prs, prs_with_conflicts,
			  open_issues, closed_issues, total_commits, commits_last_90_days, contributors, branches,
			  has_readme, score, last_commit_date)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, metrics.RepoID, metrics.OpenPRs, metrics.ClosedPRs, metrics.MergedPRs,
		metrics.PRsWithConflicts, metrics.OpenIssues, metrics.ClosedIssues, metrics.TotalCommits,
		metrics.CommitsLast90Days, metrics.Contributors, metrics.Branches, metrics.HasReadme,
		metrics.Score, metrics.LastCommitDate)
	if err != nil {
		return fmt.Errorf("failed to save GitHub metrics: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	metrics.ID = id
	return nil
}

// GetLatestGitHubMetrics gets the latest GitHub metrics for a repository
func (ds *DatabaseService) GetLatestGitHubMetrics(repoID int64) (*models.GitHubMetrics, error) {
	query := `SELECT id, repo_id, open_prs, closed_prs, merged_prs, prs_with_conflicts,
			  open_issues, closed_issues, total_commits, commits_last_90_days, contributors, branches,
			  has_readme, score, last_commit_date, collected_at
			  FROM github_metrics WHERE repo_id = ? ORDER BY collected_at DESC LIMIT 1`

	metrics := &models.GitHubMetrics{}
	err := ds.db.QueryRow(query, repoID).Scan(&metrics.ID, &metrics.RepoID, &metrics.OpenPRs,
		&metrics.ClosedPRs, &metrics.MergedPRs, &metrics.PRsWithConflicts, &metrics.OpenIssues,
		&metrics.ClosedIssues, &metrics.TotalCommits, &metrics.CommitsLast90Days, &metrics.Contributors,
		&metrics.Branches, &metrics.HasReadme, &metrics.Score, &metrics.LastCommitDate, &metrics.CollectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub metrics: %w", err)
	}
	return metrics, nil
}

// ═══════════════════════════════════════════════════════════════
// SonarCloud Metrics Operations
// ═══════════════════════════════════════════════════════════════

// SaveSonarMetrics saves SonarCloud metrics for a repository
func (ds *DatabaseService) SaveSonarMetrics(metrics *models.SonarMetrics) error {
	query := `INSERT INTO sonar_metrics (repo_id, project_key, quality_gate_status, bugs, vulnerabilities,
			  code_smells, coverage, duplicated_lines_density, lines_of_code, security_rating,
			  reliability_rating, maintainability_rating, technical_debt)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, metrics.RepoID, metrics.ProjectKey, metrics.QualityGateStatus,
		metrics.Bugs, metrics.Vulnerabilities, metrics.CodeSmells, metrics.Coverage,
		metrics.DuplicatedLinesDensity, metrics.LinesOfCode, metrics.SecurityRating,
		metrics.ReliabilityRating, metrics.MaintainabilityRating, metrics.TechnicalDebt)
	if err != nil {
		return fmt.Errorf("failed to save Sonar metrics: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	metrics.ID = id
	return nil
}

// GetLatestSonarMetrics gets the latest SonarCloud metrics for a repository
func (ds *DatabaseService) GetLatestSonarMetrics(repoID int64) (*models.SonarMetrics, error) {
	query := `SELECT id, repo_id, project_key, quality_gate_status, bugs, vulnerabilities,
			  code_smells, coverage, duplicated_lines_density, lines_of_code, security_rating,
			  reliability_rating, maintainability_rating, technical_debt, collected_at
			  FROM sonar_metrics WHERE repo_id = ? ORDER BY collected_at DESC LIMIT 1`

	metrics := &models.SonarMetrics{}
	err := ds.db.QueryRow(query, repoID).Scan(&metrics.ID, &metrics.RepoID, &metrics.ProjectKey,
		&metrics.QualityGateStatus, &metrics.Bugs, &metrics.Vulnerabilities, &metrics.CodeSmells,
		&metrics.Coverage, &metrics.DuplicatedLinesDensity, &metrics.LinesOfCode, &metrics.SecurityRating,
		&metrics.ReliabilityRating, &metrics.MaintainabilityRating, &metrics.TechnicalDebt, &metrics.CollectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get Sonar metrics: %w", err)
	}
	return metrics, nil
}

// ═══════════════════════════════════════════════════════════════
// Jira Metrics Operations
// ═══════════════════════════════════════════════════════════════

// SaveJiraMetrics saves Jira metrics for a repository
func (ds *DatabaseService) SaveJiraMetrics(metrics *models.JiraMetrics) error {
	query := `INSERT INTO jira_metrics (repo_id, project_key, open_bugs, closed_bugs, open_tasks, closed_tasks,
			  open_issues, closed_issues, avg_time_to_resolve, avg_sprint_time, active_sprints, completed_sprints,
			  total_story_points, completed_story_points)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, metrics.RepoID, metrics.ProjectKey, metrics.OpenBugs, metrics.ClosedBugs,
		metrics.OpenTasks, metrics.ClosedTasks, metrics.OpenIssues, metrics.ClosedIssues,
		metrics.AvgTimeToResolve, metrics.AvgSprintTime, metrics.ActiveSprints, metrics.CompletedSprints,
		metrics.TotalStoryPoints, metrics.CompletedStoryPoints)
	if err != nil {
		return fmt.Errorf("failed to save Jira metrics: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	metrics.ID = id
	return nil
}

// GetLatestJiraMetrics gets the latest Jira metrics for a repository
func (ds *DatabaseService) GetLatestJiraMetrics(repoID int64) (*models.JiraMetrics, error) {
	query := `SELECT id, repo_id, project_key, open_bugs, closed_bugs, open_tasks, closed_tasks,
			  open_issues, closed_issues, avg_time_to_resolve, avg_sprint_time, active_sprints, completed_sprints,
			  total_story_points, completed_story_points, collected_at
			  FROM jira_metrics WHERE repo_id = ? ORDER BY collected_at DESC LIMIT 1`

	metrics := &models.JiraMetrics{}
	err := ds.db.QueryRow(query, repoID).Scan(&metrics.ID, &metrics.RepoID, &metrics.ProjectKey,
		&metrics.OpenBugs, &metrics.ClosedBugs, &metrics.OpenTasks, &metrics.ClosedTasks,
		&metrics.OpenIssues, &metrics.ClosedIssues, &metrics.AvgTimeToResolve, &metrics.AvgSprintTime,
		&metrics.ActiveSprints, &metrics.CompletedSprints, &metrics.TotalStoryPoints,
		&metrics.CompletedStoryPoints, &metrics.CollectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get Jira metrics: %w", err)
	}
	return metrics, nil
}

// SaveJiraIssueAssignee saves Jira issue assignee information
func (ds *DatabaseService) SaveJiraIssueAssignee(assignee *models.JiraIssueAssignee) error {
	query := `INSERT INTO jira_issue_assignees (repo_id, issue_key, issue_type, status, priority,
			  assignee, reporter, created_at, updated_at, resolved_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.db.Exec(query, assignee.RepoID, assignee.IssueKey, assignee.IssueType,
		assignee.Status, assignee.Priority, assignee.Assignee, assignee.Reporter,
		assignee.CreatedAt, assignee.UpdatedAt, assignee.ResolvedAt)
	if err != nil {
		return fmt.Errorf("failed to save Jira issue assignee: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	assignee.ID = id
	return nil
}

