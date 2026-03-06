package api

import (
	"fmt"
	"log"
	"net/http"

	"sonar-automation/models"
	"sonar-automation/services"
)

// Server represents the API server
type Server struct {
	config *models.Config
	router *http.ServeMux
	port   string
	apiKey string
	db     *services.DatabaseService
}

// NewServer creates a new API server
func NewServer(config *models.Config, port, apiKey string, db *services.DatabaseService) *Server {
	server := &Server{
		config: config,
		router: http.NewServeMux(),
		port:   port,
		apiKey: apiKey,
		db:     db,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.healthHandler)

	// SonarCloud API endpoints
	s.router.HandleFunc("/api/v1/secrets/list", s.listSecretsHandler)
	s.router.HandleFunc("/api/v1/secrets/add", s.addEnvSecretsHandler)
	s.router.HandleFunc("/api/v1/workflows/update", s.updateWorkflowsHandler)
	s.router.HandleFunc("/api/v1/setup/full", s.fullSetupHandler)
	s.router.HandleFunc("/api/v1/results/fetch", s.fetchResultsHandler)
	s.router.HandleFunc("/api/v1/repository/process", s.processRepositoryHandler)

	// GitHub Metrics API endpoints
	// Pull Requests
	s.router.HandleFunc("/api/v1/github/pulls", s.listPullRequestsHandler)
	s.router.HandleFunc("/api/v1/github/pulls/get", s.getPullRequestHandler)

	// Commits
	s.router.HandleFunc("/api/v1/github/commits", s.listCommitsHandler)
	s.router.HandleFunc("/api/v1/github/commits/activity", s.getCommitActivityHandler)

	// Issues & Comments
	s.router.HandleFunc("/api/v1/github/issues", s.listIssuesHandler)
	s.router.HandleFunc("/api/v1/github/issues/comments", s.listIssueCommentsHandler)

	// Repository
	s.router.HandleFunc("/api/v1/github/readme", s.checkReadmeHandler)
	s.router.HandleFunc("/api/v1/github/branches", s.listBranchesHandler)

	// Organization
	s.router.HandleFunc("/api/v1/github/org/members", s.listOrgMembersHandler)
	s.router.HandleFunc("/api/v1/github/org/teams", s.listOrgTeamsHandler)

	// Metrics & Scoring
	s.router.HandleFunc("/api/v1/github/metrics", s.getRepositoryMetricsHandler)
	s.router.HandleFunc("/api/v1/github/metrics/all", s.getAllRepositoriesMetricsHandler)

	// SonarCloud repo-specific metrics
	s.router.HandleFunc("/api/v1/sonar/metrics", s.getSonarMetricsHandler)

	// Jira Metrics API endpoints
	s.router.HandleFunc("/api/v1/jira/issues/stats", s.getJiraIssueStatsHandler)
	s.router.HandleFunc("/api/v1/jira/bugs/open", s.getJiraOpenBugsHandler)
	s.router.HandleFunc("/api/v1/jira/tasks/open", s.getJiraOpenTasksHandler)
	s.router.HandleFunc("/api/v1/jira/issues/by-assignee", s.getJiraIssuesByAssigneeHandler)
	s.router.HandleFunc("/api/v1/jira/sprints/stats", s.getJiraSprintStatsHandler)
	s.router.HandleFunc("/api/v1/jira/metrics", s.getJiraProjectMetricsHandler)
	s.router.HandleFunc("/api/v1/jira/issues/search", s.searchJiraIssuesHandler)

	// Database-backed metrics endpoints
	s.router.HandleFunc("/api/v1/metrics/github/collect", s.collectAndStoreGitHubMetricsHandler)
	s.router.HandleFunc("/api/v1/metrics/github/stored", s.getStoredGitHubMetricsHandler)
	s.router.HandleFunc("/api/v1/metrics/sonar/collect", s.collectAndStoreSonarMetricsHandler)
	s.router.HandleFunc("/api/v1/metrics/sonar/stored", s.getStoredSonarMetricsHandler)

	// Organization Management endpoints
	s.router.HandleFunc("/api/v1/orgs", s.fetchOrgsHandler)
	s.router.HandleFunc("/api/v1/orgs/create", s.createOrgHandler)

	// Repository Management endpoints
	s.router.HandleFunc("/api/v1/repos/fetch", s.fetchReposByOrgHandler)
	s.router.HandleFunc("/api/v1/repos/update", s.updateRepoHandler)

	// Metrics Fetch endpoints (fetch and update DB)
	s.router.HandleFunc("/api/v1/repos/metrics/github", s.fetchGitHubMetricsByRepoHandler)
	s.router.HandleFunc("/api/v1/repos/metrics/jira", s.fetchJiraMetricsByRepoHandler)
	s.router.HandleFunc("/api/v1/repos/metrics/sonar", s.fetchSonarMetricsByRepoHandler)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%s", s.port)
	log.Printf("🚀 API Server starting on http://0.0.0.0:%s\n", s.port)
	log.Printf("📚 API Documentation:\n")
	log.Printf("\n")
	log.Printf("   SonarCloud Endpoints:\n")
	log.Printf("   GET  /health                                - Health check\n")
	log.Printf("   GET  /api/v1/secrets/list                   - List all secrets\n")
	log.Printf("   POST /api/v1/secrets/add                    - Add environment secrets\n")
	log.Printf("   POST /api/v1/workflows/update               - Update workflows\n")
	log.Printf("   POST /api/v1/setup/full                     - Full setup\n")
	log.Printf("   GET  /api/v1/results/fetch                  - Fetch SonarCloud results\n")
	log.Printf("   POST /api/v1/repository/process             - Process single repository\n")
	log.Printf("\n")
	log.Printf("   GitHub Metrics Endpoints:\n")
	log.Printf("   GET  /api/v1/github/pulls?repo=<name>&state=<state>           - List pull requests\n")
	log.Printf("   GET  /api/v1/github/pulls/get?repo=<name>&number=<num>        - Get specific PR\n")
	log.Printf("   GET  /api/v1/github/commits?repo=<name>&since=<date>          - List commits\n")
	log.Printf("   GET  /api/v1/github/commits/activity?repo=<name>              - Get commit activity\n")
	log.Printf("   GET  /api/v1/github/issues?repo=<name>&state=<state>          - List issues\n")
	log.Printf("   GET  /api/v1/github/issues/comments?repo=<name>&number=<num>  - List issue comments\n")
	log.Printf("   GET  /api/v1/github/readme?repo=<name>&content=<true|false>   - Check README\n")
	log.Printf("   GET  /api/v1/github/branches?repo=<name>                      - List branches\n")
	log.Printf("   GET  /api/v1/github/org/members                               - List org members\n")
	log.Printf("   GET  /api/v1/github/org/teams                                 - List org teams\n")
	log.Printf("   GET  /api/v1/github/metrics?repo=<name>                       - Get repository metrics & score\n")
	log.Printf("   GET  /api/v1/github/metrics/all                               - Get all repositories metrics\n")
	log.Printf("\n")
	log.Printf("   SonarCloud Repo-Specific Endpoints:\n")
	log.Printf("   GET  /api/v1/sonar/metrics?repo=<name>&include_issues=<true|false> - Get repo SonarCloud metrics\n")
	log.Printf("\n")
	log.Printf("   Jira Metrics Endpoints (Project-based - aggregates all boards):\n")
	log.Printf("   GET  /api/v1/jira/issues/stats?project=<key>                  - Get issue statistics\n")
	log.Printf("   GET  /api/v1/jira/bugs/open?project=<key>                     - Get open bugs\n")
	log.Printf("   GET  /api/v1/jira/tasks/open?project=<key>                    - Get open tasks\n")
	log.Printf("   GET  /api/v1/jira/issues/by-assignee?project=<key>            - Get issues by assignee\n")
	log.Printf("   GET  /api/v1/jira/sprints/stats?project=<key>                 - Get sprint stats (all boards)\n")
	log.Printf("   GET  /api/v1/jira/metrics?project=<key>                       - Get comprehensive project metrics\n")
	log.Printf("   GET  /api/v1/jira/issues/search?jql=<query>&max_results=<num> - Search issues with JQL\n")
	log.Printf("\n")
	log.Printf("   Database-backed Metrics Endpoints:\n")
	log.Printf("   POST /api/v1/metrics/github/collect?repo=<name>               - Collect & store GitHub metrics\n")
	log.Printf("   GET  /api/v1/metrics/github/stored?repo=<name>                - Get stored GitHub metrics\n")
	log.Printf("   POST /api/v1/metrics/sonar/collect?repo=<name>                - Collect & store SonarCloud metrics\n")
	log.Printf("   GET  /api/v1/metrics/sonar/stored?repo=<name>                 - Get stored SonarCloud metrics\n")
	log.Printf("\n")
	log.Printf("   Organization Management Endpoints:\n")
	log.Printf("   GET  /api/v1/orgs                                             - Fetch all organizations\n")
	log.Printf("   POST /api/v1/orgs/create                                      - Create new organization\n")
	log.Printf("\n")
	log.Printf("   Repository Management Endpoints:\n")
	log.Printf("   GET  /api/v1/repos/fetch?org_id=<id>                          - Fetch repos by org (updates DB)\n")
	log.Printf("   PUT  /api/v1/repos/update?repo_id=<id>                        - Update repo (jira_project_key)\n")
	log.Printf("\n")
	log.Printf("   Metrics Fetch Endpoints (Fetch & Update DB):\n")
	log.Printf("   GET  /api/v1/repos/metrics/github?repo_id=<id>                - Fetch GitHub metrics by repo\n")
	log.Printf("   GET  /api/v1/repos/metrics/jira?repo_id=<id>                  - Fetch Jira metrics by repo\n")
	log.Printf("   GET  /api/v1/repos/metrics/sonar?repo_id=<id>                 - Fetch SonarCloud metrics by repo\n")

	if s.apiKey != "" {
		log.Printf("\n")
		log.Printf("🔐 API Key authentication enabled\n")
		log.Printf("   Use: Authorization: Bearer <API_KEY>\n")
	} else {
		log.Printf("\n")
		log.Printf("⚠️  API Key authentication disabled (set API_KEY env var to enable)\n")
	}
	log.Println()

	// Apply middleware chain
	handler := loggingMiddleware(s.enableCORS(authMiddleware(s.apiKey)(s.router)))

	return http.ListenAndServe(addr, handler)
}

// enableCORS adds CORS headers to all responses
func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"sonar-automation","organization":"%s"}`, s.config.Organization)
}
