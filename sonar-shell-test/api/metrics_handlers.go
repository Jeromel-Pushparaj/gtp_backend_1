package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sonar-automation/models"
	"sonar-automation/services"
)

// ═══════════════════════════════════════════════════════════════
// Metrics Fetch Handlers (Fetch and Update DB)
// ═══════════════════════════════════════════════════════════════

// fetchGitHubMetricsByRepoHandler handles GET /api/v1/repos/{repo_id}/metrics/github
func (s *Server) fetchGitHubMetricsByRepoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repoIDStr := r.URL.Query().Get("repo_id")
	if repoIDStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo_id parameter is required",
		})
		return
	}

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid repo_id parameter",
		})
		return
	}

	// Get repository
	repo, err := s.db.GetRepositoryByID(repoID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %v", err),
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByID(repo.OrgID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %v", err),
		})
		return
	}

	// Fetch GitHub metrics
	gs := services.NewGitHubService(org.GitHubPAT)
	
	// Collect metrics
	metrics := &models.GitHubMetrics{
		RepoID:      repo.ID,
		CollectedAt: time.Now(),
	}

	// Get repository details
	ghRepo, err := gs.GetRepository(org.Name, repo.Name)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get repository: %v", err),
		})
		return
	}

	metrics.Stars = int64(ghRepo.GetStargazersCount())
	metrics.Forks = int64(ghRepo.GetForksCount())
	metrics.Watchers = int64(ghRepo.GetWatchersCount())
	metrics.OpenIssuesCount = int64(ghRepo.GetOpenIssuesCount())

	// Get PRs
	prs, err := gs.ListPullRequests(org.Name, repo.Name, "all")
	if err == nil {
		for _, pr := range prs {
			metrics.TotalPRs++
			if pr.GetState() == "open" {
				metrics.OpenPRs++
			} else if pr.GetState() == "closed" {
				if pr.MergedAt != nil {
					metrics.MergedPRs++
				} else {
					metrics.ClosedPRs++
				}
			}
		}
	}

	// Get commits (last 90 days)
	since := time.Now().AddDate(0, 0, -90)
	commits, err := gs.ListCommits(org.Name, repo.Name, &since)
	if err == nil {
		metrics.CommitsLast90Days = int64(len(commits))
	}

	// Get branches
	branches, err := gs.ListBranches(org.Name, repo.Name)
	if err == nil {
		metrics.TotalBranches = int64(len(branches))
	}

	// Get contributors
	contributors, err := gs.ListContributors(org.Name, repo.Name)
	if err == nil {
		metrics.Contributors = int64(len(contributors))
	}

	// Check for README
	hasReadme, err := gs.CheckReadmeExists(org.Name, repo.Name)
	if err == nil {
		metrics.HasReadme = hasReadme
	}

	// Calculate score
	score := calculateGitHubScore(metrics)
	metrics.Score = score

	// Save metrics to database
	if err := s.db.SaveGitHubMetrics(metrics); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "GitHub metrics fetched and saved successfully",
		Data:    metrics,
	})
}

// fetchJiraMetricsByRepoHandler handles GET /api/v1/repos/{repo_id}/metrics/jira
func (s *Server) fetchJiraMetricsByRepoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repoIDStr := r.URL.Query().Get("repo_id")
	if repoIDStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo_id parameter is required",
		})
		return
	}

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid repo_id parameter",
		})
		return
	}

	// Get repository
	repo, err := s.db.GetRepositoryByID(repoID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %v", err),
		})
		return
	}

	if repo.JiraProjectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Repository does not have a Jira project key configured",
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByID(repo.OrgID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %v", err),
		})
		return
	}

	if org.JiraToken == "" || org.JiraDomain == "" || org.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Organization does not have Jira credentials configured",
		})
		return
	}

	// Fetch Jira metrics
	js := services.NewJiraService(org.JiraDomain, org.JiraEmail, org.JiraToken)

	projectMetrics, err := js.GetProjectMetrics(repo.JiraProjectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch Jira metrics: %v", err),
		})
		return
	}

	// Save to database
	metrics := &models.JiraMetrics{
		RepoID:           repo.ID,
		ProjectKey:       repo.JiraProjectKey,
		OpenBugs:         projectMetrics.IssueStats.Bugs,
		ClosedBugs:       0, // We don't have this breakdown yet
		OpenTasks:        projectMetrics.IssueStats.Tasks,
		ClosedTasks:      0, // We don't have this breakdown yet
		OpenIssues:       projectMetrics.IssueStats.OpenIssues,
		ClosedIssues:     projectMetrics.IssueStats.ClosedIssues,
		AvgTimeToResolve: projectMetrics.IssueStats.AvgTimeToResolve,
		AvgSprintTime:    projectMetrics.SprintStats.AvgSprintDuration,
		ActiveSprints:    projectMetrics.SprintStats.ActiveSprints,
		CompletedSprints: projectMetrics.SprintStats.CompletedSprints,
		TotalStoryPoints: 0, // We don't have this yet
		CompletedStoryPoints: 0, // We don't have this yet
		CollectedAt:      time.Now(),
	}

	if err := s.db.SaveJiraMetrics(metrics); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Jira metrics fetched and saved successfully",
		Data:    metrics,
	})
}

// fetchSonarMetricsByRepoHandler handles GET /api/v1/repos/{repo_id}/metrics/sonar
func (s *Server) fetchSonarMetricsByRepoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repoIDStr := r.URL.Query().Get("repo_id")
	if repoIDStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo_id parameter is required",
		})
		return
	}

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid repo_id parameter",
		})
		return
	}

	// Get repository
	repo, err := s.db.GetRepositoryByID(repoID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %v", err),
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByID(repo.OrgID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %v", err),
		})
		return
	}

	if org.SonarToken == "" || org.SonarOrgKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Organization does not have SonarCloud credentials configured",
		})
		return
	}

	// Fetch SonarCloud metrics
	ss := services.NewSonarCloudService(org.SonarToken, org.SonarOrgKey)

	// Get project key (usually org:repo format)
	projectKey := fmt.Sprintf("%s_%s", org.SonarOrgKey, repo.Name)

	// Get quality gate status
	qgStatus, err := ss.GetQualityGateStatus(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get quality gate status: %v", err),
		})
		return
	}

	// Get measures
	measures, err := ss.GetProjectMeasures(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get project measures: %v", err),
		})
		return
	}

	// Parse metrics into a map
	metricsMap := make(map[string]string)
	for _, m := range measures {
		metricsMap[m.Metric] = m.Value
	}

	// Helper function to parse int from metrics
	parseInt := func(key string) int {
		if val, ok := metricsMap[key]; ok {
			var result int
			fmt.Sscanf(val, "%d", &result)
			return result
		}
		return 0
	}

	// Helper function to parse float from metrics
	parseFloat := func(key string) float64 {
		if val, ok := metricsMap[key]; ok {
			var result float64
			fmt.Sscanf(val, "%f", &result)
			return result
		}
		return 0.0
	}

	// Save to database
	metrics := &models.SonarMetrics{
		RepoID:                 repo.ID,
		ProjectKey:             projectKey,
		Bugs:                   parseInt("bugs"),
		Vulnerabilities:        parseInt("vulnerabilities"),
		CodeSmells:             parseInt("code_smells"),
		Coverage:               parseFloat("coverage"),
		DuplicatedLinesDensity: parseFloat("duplicated_lines_density"),
		LinesOfCode:            parseInt("ncloc"),
		QualityGateStatus:      qgStatus.Status,
		SecurityRating:         metricsMap["security_rating"],
		ReliabilityRating:      metricsMap["reliability_rating"],
		MaintainabilityRating:  metricsMap["sqale_rating"],
		TechnicalDebt:          metricsMap["sqale_index"],
		CollectedAt:            time.Now(),
	}

	if err := s.db.SaveSonarMetrics(metrics); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "SonarCloud metrics fetched and saved successfully",
		Data:    metrics,
	})
}

// calculateGitHubScore calculates a health score for a repository based on metrics
func calculateGitHubScore(metrics *models.GitHubMetrics) float64 {
	score := 100.0

	// Deduct points for open PRs (max -20 points)
	if metrics.OpenPRs > 10 {
		score -= 20
	} else if metrics.OpenPRs > 5 {
		score -= 10
	} else if metrics.OpenPRs > 0 {
		score -= 5
	}

	// Deduct points for low activity (max -30 points)
	if metrics.CommitsLast90Days == 0 {
		score -= 30
	} else if metrics.CommitsLast90Days < 10 {
		score -= 15
	} else if metrics.CommitsLast90Days < 50 {
		score -= 5
	}

	// Deduct points for no README (max -10 points)
	if !metrics.HasReadme {
		score -= 10
	}

	// Deduct points for low contributors (max -10 points)
	if metrics.Contributors == 0 {
		score -= 10
	} else if metrics.Contributors == 1 {
		score -= 5
	}

	// Bonus points for merged PRs (max +10 points)
	if metrics.MergedPRs > 50 {
		score += 10
	} else if metrics.MergedPRs > 20 {
		score += 5
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

