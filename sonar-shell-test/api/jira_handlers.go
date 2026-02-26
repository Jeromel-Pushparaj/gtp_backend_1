package api

import (
	"fmt"
	"net/http"
	"strconv"

	"sonar-automation/services"
)

// ═══════════════════════════════════════════════════════════════
// Jira Metrics Handlers
// ═══════════════════════════════════════════════════════════════

// getJiraIssueStatsHandler handles GET /api/v1/jira/issues/stats
func (s *Server) getJiraIssueStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	stats, err := js.GetIssueStats(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get issue stats: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// getJiraOpenBugsHandler handles GET /api/v1/jira/bugs/open
func (s *Server) getJiraOpenBugsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	bugs, err := js.GetOpenBugs(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get open bugs: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d open bugs", len(bugs)),
		Data:    bugs,
	})
}

// getJiraOpenTasksHandler handles GET /api/v1/jira/tasks/open
func (s *Server) getJiraOpenTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	tasks, err := js.GetOpenTasks(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get open tasks: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d open tasks", len(tasks)),
		Data:    tasks,
	})
}

// getJiraIssuesByAssigneeHandler handles GET /api/v1/jira/issues/by-assignee
func (s *Server) getJiraIssuesByAssigneeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	issuesByAssignee, err := js.GetIssuesByAssignee(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get issues by assignee: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found issues for %d assignees", len(issuesByAssignee)),
		Data:    issuesByAssignee,
	})
}

// getJiraSprintStatsHandler handles GET /api/v1/jira/sprints/stats
func (s *Server) getJiraSprintStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	stats, err := js.GetSprintStats(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get sprint stats: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// getJiraProjectMetricsHandler handles GET /api/v1/jira/metrics
func (s *Server) getJiraProjectMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	projectKey := r.URL.Query().Get("project")
	if projectKey == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "project parameter is required",
		})
		return
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	metrics, err := js.GetProjectMetrics(projectKey)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get project metrics: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    metrics,
	})
}

// searchJiraIssuesHandler handles GET /api/v1/jira/issues/search
func (s *Server) searchJiraIssuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	jql := r.URL.Query().Get("jql")
	if jql == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "jql parameter is required",
		})
		return
	}

	maxResultsStr := r.URL.Query().Get("max_results")
	maxResults := 50
	if maxResultsStr != "" {
		var err error
		maxResults, err = strconv.Atoi(maxResultsStr)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Error:   "Invalid max_results parameter",
			})
			return
		}
	}

	if s.config.JiraToken == "" || s.config.JiraDomain == "" || s.config.JiraEmail == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Jira credentials not configured",
		})
		return
	}

	js := services.NewJiraService(s.config.JiraDomain, s.config.JiraEmail, s.config.JiraToken)
	issues, err := js.SearchIssues(jql, maxResults)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to search issues: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d issues", len(issues)),
		Data:    issues,
	})
}

