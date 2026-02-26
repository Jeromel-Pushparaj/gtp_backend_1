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
// Helper Functions
// ═══════════════════════════════════════════════════════════════

// collectGitHubMetrics collects GitHub metrics for a repository
func collectGitHubMetrics(gs *services.GitHubService, org, repo string) (*models.RepositoryMetrics, error) {
	metrics := &models.RepositoryMetrics{
		Repository: repo,
	}

	// Check README
	hasReadme, _ := gs.CheckReadmeExists(org, repo)
	metrics.HasReadme = hasReadme

	// Get default branch
	defaultBranch, err := gs.GetDefaultBranch(org, repo)
	if err == nil {
		metrics.DefaultBranch = defaultBranch
	}

	// Get pull requests
	allPRs, err := gs.ListPullRequests(org, repo, "all")
	if err == nil {
		for _, pr := range allPRs {
			if pr.GetState() == "open" {
				metrics.OpenPRs++
				// Check for conflicts
				if pr.Mergeable != nil && !*pr.Mergeable {
					metrics.PRsWithConflicts++
				}
			} else if pr.GetState() == "closed" {
				if pr.MergedAt != nil {
					metrics.MergedPRs++
				} else {
					metrics.ClosedPRs++
				}
			}
		}
	}

	// Get issues (excluding PRs)
	allIssues, err := gs.ListIssues(org, repo, "all")
	if err == nil {
		for _, issue := range allIssues {
			if !issue.IsPullRequest() {
				if issue.GetState() == "open" {
					metrics.OpenIssues++
				} else {
					metrics.ClosedIssues++
				}
			}
		}
	}

	// Get commit activity
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	allCommits, err := gs.ListCommits(org, repo, nil)
	if err == nil {
		metrics.TotalCommits = len(allCommits)

		if len(allCommits) > 0 {
			lastCommitDate := allCommits[0].GetCommit().GetAuthor().GetDate().Time
			metrics.LastCommitDate = &lastCommitDate
		}

		// Count commits in last 90 days
		for _, commit := range allCommits {
			commitDate := commit.GetCommit().GetAuthor().GetDate().Time
			if commitDate.After(ninetyDaysAgo) {
				metrics.CommitsLast90Days++
			}
		}

		metrics.IsActive = metrics.CommitsLast90Days > 0

		// Count unique contributors
		contributorsMap := make(map[string]bool)
		for _, commit := range allCommits {
			if commit.GetAuthor() != nil {
				contributorsMap[commit.GetAuthor().GetLogin()] = true
			}
		}
		metrics.Contributors = len(contributorsMap)
	}

	// Get branches
	branches, err := gs.ListBranches(org, repo)
	if err == nil {
		metrics.Branches = len(branches)
	}

	// Calculate score
	metrics.Score = calculateRepositoryScore(*metrics)

	return metrics, nil
}

// ═══════════════════════════════════════════════════════════════
// Database-backed Metrics Collection Handlers
// ═══════════════════════════════════════════════════════════════

// collectAndStoreGitHubMetricsHandler handles POST /api/v1/metrics/github/collect
func (s *Server) collectAndStoreGitHubMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	if s.db == nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database not initialized",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	// Get or create organization
	org, err := s.db.GetOrganizationByName(s.config.Organization)
	if err != nil {
		// Create organization if it doesn't exist
		org = &models.Organization{
			Name:        s.config.Organization,
			SonarOrgKey: s.config.SonarOrgKey,
		}
		if err := s.db.CreateOrganization(org); err != nil {
			sendJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   fmt.Sprintf("Failed to create organization: %v", err),
			})
			return
		}
	}

	// Get or create repository in database
	dbRepo, err := s.db.GetRepositoryByName(org.ID, repo)
	if err != nil {
		// Repository doesn't exist, create it
		ghRepo, err := gs.GetRepository(s.config.Organization, repo)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   fmt.Sprintf("Failed to get repository from GitHub: %v", err),
			})
			return
		}

		dbRepo = &models.Repository{
			OrgID:         org.ID,
			Name:          repo,
			GitHubURL:     ghRepo.GetHTMLURL(),
			Owner:         s.config.Organization,
			DefaultBranch: ghRepo.GetDefaultBranch(),
			IsActive:      !ghRepo.GetArchived(),
		}

		if err := s.db.CreateRepository(dbRepo); err != nil {
			sendJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   fmt.Sprintf("Failed to create repository: %v", err),
			})
			return
		}
	}

	// Collect GitHub metrics
	metrics, err := collectGitHubMetrics(gs, s.config.Organization, repo)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to collect GitHub metrics: %v", err),
		})
		return
	}

	// Save metrics to database
	dbMetrics := &models.GitHubMetrics{
		RepoID:            dbRepo.ID,
		OpenPRs:           int64(metrics.OpenPRs),
		ClosedPRs:         int64(metrics.ClosedPRs),
		MergedPRs:         int64(metrics.MergedPRs),
		PRsWithConflicts:  int64(metrics.PRsWithConflicts),
		OpenIssues:        int64(metrics.OpenIssues),
		ClosedIssues:      int64(metrics.ClosedIssues),
		TotalCommits:      int64(metrics.TotalCommits),
		CommitsLast90Days: int64(metrics.CommitsLast90Days),
		Contributors:      int64(metrics.Contributors),
		Branches:          int64(metrics.Branches),
		HasReadme:         metrics.HasReadme,
		Score:             metrics.Score,
		LastCommitDate:    metrics.LastCommitDate,
	}

	if err := s.db.SaveGitHubMetrics(dbMetrics); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save GitHub metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("GitHub metrics collected and stored for repository: %s", repo),
		Data: map[string]interface{}{
			"repository": repo,
			"metrics":    metrics,
			"stored_at":  time.Now(),
		},
	})
}

// collectAndStoreSonarMetricsHandler handles POST /api/v1/metrics/sonar/collect
func (s *Server) collectAndStoreSonarMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	if s.config.SonarToken == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "SonarCloud token not configured",
		})
		return
	}

	if s.db == nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database not initialized",
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByName(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found in database: %s. Collect GitHub metrics first.", s.config.Organization),
		})
		return
	}

	// Get repository from database
	dbRepo, err := s.db.GetRepositoryByName(org.ID, repo)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found in database: %s. Collect GitHub metrics first.", repo),
		})
		return
	}

	sc := services.NewSonarCloudService(s.config.SonarToken, s.config.SonarOrgKey)
	projectKey := fmt.Sprintf("%s_%s", s.config.SonarOrgKey, repo)

	// Get quality gate status
	qgStatus, err := sc.GetQualityGateStatus(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get quality gate status: %v", err),
		})
		return
	}

	// Get measures
	measures, err := sc.GetProjectMeasures(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get project measures: %v", err),
		})
		return
	}

	// Parse metrics
	metricsMap := make(map[string]string)
	for _, m := range measures {
		metricsMap[m.Metric] = m.Value
	}

	// Create SonarMetrics object
	dbMetrics := &models.SonarMetrics{
		RepoID:            dbRepo.ID,
		ProjectKey:        projectKey,
		QualityGateStatus: qgStatus.Status,
	}

	// Parse numeric values
	if val, ok := metricsMap["bugs"]; ok {
		if bugs, err := strconv.Atoi(val); err == nil {
			dbMetrics.Bugs = bugs
		}
	}
	if val, ok := metricsMap["vulnerabilities"]; ok {
		if vuln, err := strconv.Atoi(val); err == nil {
			dbMetrics.Vulnerabilities = vuln
		}
	}
	if val, ok := metricsMap["code_smells"]; ok {
		if smells, err := strconv.Atoi(val); err == nil {
			dbMetrics.CodeSmells = smells
		}
	}
	if val, ok := metricsMap["coverage"]; ok {
		if cov, err := strconv.ParseFloat(val, 64); err == nil {
			dbMetrics.Coverage = cov
		}
	}
	if val, ok := metricsMap["duplicated_lines_density"]; ok {
		if dup, err := strconv.ParseFloat(val, 64); err == nil {
			dbMetrics.DuplicatedLinesDensity = dup
		}
	}
	if val, ok := metricsMap["ncloc"]; ok {
		if loc, err := strconv.Atoi(val); err == nil {
			dbMetrics.LinesOfCode = loc
		}
	}
	if val, ok := metricsMap["security_rating"]; ok {
		dbMetrics.SecurityRating = val
	}
	if val, ok := metricsMap["reliability_rating"]; ok {
		dbMetrics.ReliabilityRating = val
	}
	if val, ok := metricsMap["sqale_rating"]; ok {
		dbMetrics.MaintainabilityRating = val
	}
	if val, ok := metricsMap["sqale_index"]; ok {
		dbMetrics.TechnicalDebt = val
	}

	// Save to database
	if err := s.db.SaveSonarMetrics(dbMetrics); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to save SonarCloud metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("SonarCloud metrics collected and stored for repository: %s", repo),
		Data: map[string]interface{}{
			"repository":  repo,
			"project_key": projectKey,
			"metrics":     dbMetrics,
			"stored_at":   time.Now(),
		},
	})
}

// getStoredGitHubMetricsHandler handles GET /api/v1/metrics/github/stored
func (s *Server) getStoredGitHubMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	if s.db == nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database not initialized",
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByName(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %s", s.config.Organization),
		})
		return
	}

	// Get repository from database
	dbRepo, err := s.db.GetRepositoryByName(org.ID, repo)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %s", repo),
		})
		return
	}

	// Get latest metrics
	metrics, err := s.db.GetLatestGitHubMetrics(dbRepo.ID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("No metrics found for repository: %s", repo),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"repository": repo,
			"metrics":    metrics,
		},
	})
}

// getStoredSonarMetricsHandler handles GET /api/v1/metrics/sonar/stored
func (s *Server) getStoredSonarMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	if s.db == nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Database not initialized",
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByName(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %s", s.config.Organization),
		})
		return
	}

	// Get repository from database
	dbRepo, err := s.db.GetRepositoryByName(org.ID, repo)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %s", repo),
		})
		return
	}

	// Get latest metrics
	metrics, err := s.db.GetLatestSonarMetrics(dbRepo.ID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("No metrics found for repository: %s", repo),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"repository": repo,
			"metrics":    metrics,
		},
	})
}

