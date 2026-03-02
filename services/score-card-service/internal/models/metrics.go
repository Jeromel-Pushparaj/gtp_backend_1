package models

import "time"

// CombinedMetrics represents all metrics from different sources
type CombinedMetrics struct {
	ServiceName   string         `json:"service_name"`
	GitHub        *GitHubMetrics `json:"github,omitempty"`
	Sonar         *SonarMetrics  `json:"sonar,omitempty"`
	Jira          *JiraMetrics   `json:"jira,omitempty"`
	CollectedAt   time.Time      `json:"collected_at"`
}

// GitHubMetrics represents metrics from GitHub
type GitHubMetrics struct {
	Repository        string    `json:"repository"`
	HasReadme         bool      `json:"has_readme"`
	DefaultBranch     string    `json:"default_branch"`
	OpenPRs           int64     `json:"open_prs"`
	ClosedPRs         int64     `json:"closed_prs"`
	MergedPRs         int64     `json:"merged_prs"`
	PRsWithConflicts  int64     `json:"prs_with_conflicts"`
	OpenIssues        int64     `json:"open_issues"`
	ClosedIssues      int64     `json:"closed_issues"`
	TotalCommits      int64     `json:"total_commits"`
	CommitsLast90Days int64     `json:"commits_last_90_days"`
	IsActive          bool      `json:"is_active"`
	Contributors      int64     `json:"contributors"`
	Branches          int64     `json:"branches"`
	Stars             int64     `json:"stars"`
	Forks             int64     `json:"forks"`
	Watchers          int64     `json:"watchers"`
	Score             float64   `json:"score"`
	LastCommitDate    *time.Time `json:"last_commit_date,omitempty"`
}

// SonarMetrics represents metrics from SonarCloud
type SonarMetrics struct {
	ProjectKey             string  `json:"project_key"`
	QualityGateStatus      string  `json:"quality_gate_status"`
	Bugs                   int     `json:"bugs"`
	Vulnerabilities        int     `json:"vulnerabilities"`
	SecurityHotspots       int     `json:"security_hotspots"`
	CodeSmells             int     `json:"code_smells"`
	Coverage               float64 `json:"coverage"`
	DuplicatedLinesDensity float64 `json:"duplicated_lines_density"`
	LinesOfCode            int     `json:"lines_of_code"`
	SecurityRating         string  `json:"security_rating"`
	ReliabilityRating      string  `json:"reliability_rating"`
	MaintainabilityRating  string  `json:"maintainability_rating"`
	TechnicalDebt          string  `json:"technical_debt"`
}

// JiraMetrics represents metrics from Jira
type JiraMetrics struct {
	ProjectKey           string  `json:"project_key"`
	OpenBugs             int     `json:"open_bugs"`
	ClosedBugs           int     `json:"closed_bugs"`
	OpenTasks            int     `json:"open_tasks"`
	ClosedTasks          int     `json:"closed_tasks"`
	OpenIssues           int     `json:"open_issues"`
	ClosedIssues         int     `json:"closed_issues"`
	AvgTimeToResolve     float64 `json:"avg_time_to_resolve"`     // in hours (MTTR)
	AvgSprintTime        float64 `json:"avg_sprint_time"`         // in days
	ActiveSprints        int     `json:"active_sprints"`
	CompletedSprints     int     `json:"completed_sprints"`
	TotalStoryPoints     int     `json:"total_story_points"`
	CompletedStoryPoints int     `json:"completed_story_points"`
}

// ToMap converts CombinedMetrics to a flat map for rule evaluation
func (cm *CombinedMetrics) ToMap() map[string]float64 {
	metrics := make(map[string]float64)
	
	// GitHub metrics
	if cm.GitHub != nil {
		metrics["open_prs"] = float64(cm.GitHub.OpenPRs)
		metrics["closed_prs"] = float64(cm.GitHub.ClosedPRs)
		metrics["merged_prs"] = float64(cm.GitHub.MergedPRs)
		metrics["prs_with_conflicts"] = float64(cm.GitHub.PRsWithConflicts)
		metrics["open_issues"] = float64(cm.GitHub.OpenIssues)
		metrics["closed_issues"] = float64(cm.GitHub.ClosedIssues)
		metrics["total_commits"] = float64(cm.GitHub.TotalCommits)
		metrics["commits_last_90_days"] = float64(cm.GitHub.CommitsLast90Days)
		metrics["contributors"] = float64(cm.GitHub.Contributors)
		metrics["branches"] = float64(cm.GitHub.Branches)
		metrics["has_readme"] = boolToFloat(cm.GitHub.HasReadme)
		
		// Calculate deployment frequency (commits per week)
		if cm.GitHub.CommitsLast90Days > 0 {
			metrics["deployment_frequency"] = float64(cm.GitHub.CommitsLast90Days) / 13.0 // 90 days ≈ 13 weeks
		}
		
		// Calculate days since last commit (freshness)
		if cm.GitHub.LastCommitDate != nil {
			daysSince := time.Since(*cm.GitHub.LastCommitDate).Hours() / 24
			metrics["days_since_last_commit"] = daysSince
		}
	}
	
	// SonarCloud metrics
	if cm.Sonar != nil {
		metrics["bugs"] = float64(cm.Sonar.Bugs)
		metrics["vulnerabilities"] = float64(cm.Sonar.Vulnerabilities)
		metrics["security_hotspots"] = float64(cm.Sonar.SecurityHotspots)
		metrics["code_smells"] = float64(cm.Sonar.CodeSmells)
		metrics["coverage"] = cm.Sonar.Coverage
		metrics["duplicated_lines_density"] = cm.Sonar.DuplicatedLinesDensity
		metrics["lines_of_code"] = float64(cm.Sonar.LinesOfCode)
		metrics["quality_gate_passed"] = boolToFloat(cm.Sonar.QualityGateStatus == "OK")
	}
	
	// Jira metrics
	if cm.Jira != nil {
		metrics["open_bugs"] = float64(cm.Jira.OpenBugs)
		metrics["closed_bugs"] = float64(cm.Jira.ClosedBugs)
		metrics["open_tasks"] = float64(cm.Jira.OpenTasks)
		metrics["closed_tasks"] = float64(cm.Jira.ClosedTasks)
		metrics["mttr"] = cm.Jira.AvgTimeToResolve // Mean Time To Resolve in hours
		metrics["active_sprints"] = float64(cm.Jira.ActiveSprints)
		metrics["completed_sprints"] = float64(cm.Jira.CompletedSprints)
		
		// Calculate story point completion rate
		if cm.Jira.TotalStoryPoints > 0 {
			metrics["story_point_completion_rate"] = float64(cm.Jira.CompletedStoryPoints) / float64(cm.Jira.TotalStoryPoints) * 100
		}
	}
	
	return metrics
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

