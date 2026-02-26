package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sonar-automation/controllers"
	"sonar-automation/services"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RepositoryRequest represents a request to process a single repository
type RepositoryRequest struct {
	RepositoryName string `json:"repository_name"`
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// listSecretsHandler handles GET /api/v1/secrets/list
func (s *Server) listSecretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	type SecretInfo struct {
		Repository string   `json:"repository"`
		Secrets    []string `json:"secrets"`
		EnvSecrets []string `json:"env_secrets"`
	}

	var results []SecretInfo

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		info := SecretInfo{Repository: repoName}

		// List repository secrets
		secrets, err := gs.ListSecrets(s.config.Organization, repoName)
		if err == nil {
			for _, secret := range secrets {
				info.Secrets = append(info.Secrets, secret.Name)
			}
		}

		// List environment secrets
		envSecrets, err := gs.ListEnvSecrets(s.config.Organization, repoName, s.config.EnvironmentName)
		if err == nil {
			for _, secret := range envSecrets {
				info.EnvSecrets = append(info.EnvSecrets, secret.Name)
			}
		}

		results = append(results, info)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d repositories", len(results)),
		Data:    results,
	})
}

// addEnvSecretsHandler handles POST /api/v1/secrets/add
func (s *Server) addEnvSecretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, s.config)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	type Result struct {
		Repository string `json:"repository"`
		Status     string `json:"status"`
		Error      string `json:"error,omitempty"`
	}

	var results []Result
	successCount := 0
	errorCount := 0

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		result := Result{Repository: repoName}

		err := rc.SetupEnvironmentSecrets(repoName)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			errorCount++
		} else {
			result.Status = "success"
			successCount++
		}

		results = append(results, result)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Processed %d repositories: %d success, %d errors", len(results), successCount, errorCount),
		Data:    results,
	})
}

// updateWorkflowsHandler handles POST /api/v1/workflows/update
func (s *Server) updateWorkflowsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, s.config)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	type Result struct {
		Repository string `json:"repository"`
		Status     string `json:"status"`
		Error      string `json:"error,omitempty"`
	}

	var results []Result
	successCount := 0
	errorCount := 0

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		result := Result{Repository: repoName}

		defaultBranch, err := gs.GetDefaultBranch(s.config.Organization, repoName)
		if err != nil {
			result.Status = "error"
			result.Error = fmt.Sprintf("Failed to get default branch: %v", err)
			errorCount++
			results = append(results, result)
			continue
		}

		sonarPath := ".github/workflows/sonar.yml"
		exists, err := gs.CheckFileExists(s.config.Organization, repoName, sonarPath, defaultBranch)
		if err != nil || !exists {
			result.Status = "skipped"
			result.Error = "sonar.yml not found"
			results = append(results, result)
			continue
		}

		err = rc.UpdateWorkflowToUseEnvironment(repoName, defaultBranch)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			errorCount++
		} else {
			result.Status = "success"
			successCount++
		}

		results = append(results, result)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Processed %d repositories: %d success, %d errors", len(results), successCount, errorCount),
		Data:    results,
	})
}

// fullSetupHandler handles POST /api/v1/setup/full
func (s *Server) fullSetupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	sc := services.NewSonarCloudService(s.config.SonarToken, s.config.SonarOrgKey)
	rc := controllers.NewRepositoryController(gs, sc, s.config)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	type Result struct {
		Repository string `json:"repository"`
		Status     string `json:"status"`
		Error      string `json:"error,omitempty"`
	}

	var results []Result
	successCount := 0
	errorCount := 0

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		result := Result{Repository: repoName}

		err := rc.ProcessRepositoryFullSetup(repoName)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			errorCount++
		} else {
			result.Status = "success"
			successCount++
		}

		results = append(results, result)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Processed %d repositories: %d success, %d errors", len(results), successCount, errorCount),
		Data:    results,
	})
}

// fetchResultsHandler handles GET /api/v1/results/fetch
func (s *Server) fetchResultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	sc := services.NewSonarCloudService(s.config.SonarToken, s.config.SonarOrgKey)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	type RepositoryResult struct {
		Repository   string                 `json:"repository"`
		ProjectKey   string                 `json:"project_key"`
		QualityGate  string                 `json:"quality_gate,omitempty"`
		Metrics      map[string]string      `json:"metrics,omitempty"`
		IssuesCount  int                    `json:"issues_count,omitempty"`
		Error        string                 `json:"error,omitempty"`
	}

	var results []RepositoryResult

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		projectKey := fmt.Sprintf("%s_%s", s.config.SonarOrgKey, repoName)

		result := RepositoryResult{
			Repository: repoName,
			ProjectKey: projectKey,
		}

		// Get quality gate status
		qgStatus, err := sc.GetQualityGateStatus(projectKey)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.QualityGate = qgStatus.Status

		// Get measures
		measures, err := sc.GetProjectMeasures(projectKey)
		if err == nil {
			result.Metrics = make(map[string]string)
			for _, m := range measures {
				result.Metrics[m.Metric] = m.Value
			}
		}

		// Get issues count
		issues, err := sc.GetIssues(projectKey, 1)
		if err == nil {
			result.IssuesCount = issues.Total
		}

		results = append(results, result)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Fetched results for %d repositories", len(results)),
		Data:    results,
	})
}

// getSonarMetricsHandler handles GET /api/v1/sonar/metrics
func (s *Server) getSonarMetricsHandler(w http.ResponseWriter, r *http.Request) {
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

	if s.config.SonarToken == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "SonarCloud token not configured",
		})
		return
	}

	sc := services.NewSonarCloudService(s.config.SonarToken, s.config.SonarOrgKey)
	projectKey := fmt.Sprintf("%s_%s", s.config.SonarOrgKey, repo)

	type SonarMetricsResponse struct {
		Repository        string            `json:"repository"`
		ProjectKey        string            `json:"project_key"`
		QualityGateStatus string            `json:"quality_gate_status"`
		Metrics           map[string]string `json:"metrics"`
		IssuesCount       int               `json:"issues_count"`
		Issues            []interface{}     `json:"issues,omitempty"`
	}

	response := SonarMetricsResponse{
		Repository: repo,
		ProjectKey: projectKey,
		Metrics:    make(map[string]string),
	}

	// Get quality gate status
	qgStatus, err := sc.GetQualityGateStatus(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get quality gate status: %v", err),
		})
		return
	}
	response.QualityGateStatus = qgStatus.Status

	// Get measures
	measures, err := sc.GetProjectMeasures(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get project measures: %v", err),
		})
		return
	}

	for _, m := range measures {
		response.Metrics[m.Metric] = m.Value
	}

	// Get issues
	includeIssues := r.URL.Query().Get("include_issues") == "true"
	issuesLimit := 10
	if includeIssues {
		issuesLimit = 100
	}

	issues, err := sc.GetIssues(projectKey, issuesLimit)
	if err == nil {
		response.IssuesCount = issues.Total
		if includeIssues {
			for _, issue := range issues.Issues {
				response.Issues = append(response.Issues, issue)
			}
		}
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// processRepositoryHandler handles POST /api/v1/repository/process
func (s *Server) processRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req RepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.RepositoryName == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repository_name is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, s.config)

	err := rc.ProcessRepository(req.RepositoryName)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Successfully processed repository: %s", req.RepositoryName),
	})
}

