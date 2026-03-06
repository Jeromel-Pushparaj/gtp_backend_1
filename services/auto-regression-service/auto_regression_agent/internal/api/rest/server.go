package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/api/rest/handler"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/api/rest/middleware"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/config"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
)

// Server represents the REST API server
type Server struct {
	config       *config.Config
	orchestrator *orchestration.Orchestrator
	router       *gin.Engine
	wsHandler    *handler.WebSocketHandler
}

// NewServer creates a new REST API server
func NewServer(cfg *config.Config, orch *orchestration.Orchestrator) *Server {
	// Set Gin mode based on log level
	if cfg.Observability.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &Server{
		config:       cfg,
		orchestrator: orch,
		router:       router,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(middleware.Logger())

	// CORS middleware
	s.router.Use(middleware.CORS())

	// Request ID middleware
	s.router.Use(middleware.RequestID())
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check endpoints
	s.router.GET("/health", handler.Health())
	s.router.HEAD("/health", handler.Health())
	s.router.GET("/ready", handler.Ready())
	s.router.HEAD("/ready", handler.Ready())

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Spec endpoints
		specHandler := handler.NewSpecHandler(s.orchestrator)
		specs := v1.Group("/specs")
		{
			specs.POST("", specHandler.UploadSpec)
			specs.POST("/validate", specHandler.ValidateSpec)
			specs.GET("/:spec_id/status", specHandler.GetSpecStatus)
		}

		// Queue endpoints
		queueHandler := handler.NewQueueHandler(s.orchestrator)
		queue := v1.Group("/queue")
		{
			queue.GET("/stats", queueHandler.GetStats)
		}

		// Workflow endpoints (placeholder for future implementation)
		workflows := v1.Group("/workflows")
		{
			workflows.GET("/:workflow_id", handler.NotImplemented())
			workflows.GET("/:workflow_id/jobs", handler.NotImplemented())
		}

		// Runs endpoints (for UI)
		runsHandler := handler.NewRunsHandler(s.orchestrator)
		runs := v1.Group("/runs")
		{
			runs.GET("", runsHandler.ListRuns)
			runs.GET("/:runId", runsHandler.GetRun)
			runs.GET("/:runId/report", runsHandler.GetRunReport)
			runs.GET("/:runId/logs", runsHandler.GetRunLogs)
			runs.GET("/:runId/download", runsHandler.DownloadReport)
			runs.GET("/:runId/test-cases", runsHandler.GetTestCases)
			runs.GET("/:runId/collaboration", runsHandler.GetAgentCollaboration)
		}

		// GitHub integration endpoints
		githubHandler := handler.NewGitHubHandler(s.orchestrator)
		github := v1.Group("/github")
		{
			github.POST("/test", githubHandler.TriggerTest)
		}

		// WebSocket endpoint for real-time updates
		s.wsHandler = handler.NewWebSocketHandler()
		v1.GET("/ws/runs/:runId", s.wsHandler.HandleWebSocket)
	}

	// Metrics endpoint (if enabled)
	if s.config.Observability.Metrics.Enabled {
		s.router.GET(s.config.Observability.Metrics.Prometheus.Path, handler.Metrics())
	}
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// WebSocketHandler returns the WebSocket handler
func (s *Server) WebSocketHandler() *handler.WebSocketHandler {
	return s.wsHandler
}
