package v1

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/api/v2/handlers"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/config"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// API v2 group - Advanced scorecards with levels (Gold/Silver/Bronze, Traffic Lights, etc.)
	// Create v2 handler with metrics fetcher if configured
	var v2Handler *handlers.ScorecardV2Handler
	if cfg.MetricsAPIBaseURL != "" {
		v2Handler = handlers.NewScorecardV2HandlerWithFetcher(cfg.MetricsAPIBaseURL)
		log.Printf("✅ V2 Handler initialized with metrics fetcher (base URL: %s)", cfg.MetricsAPIBaseURL)
	} else {
		v2Handler = handlers.NewScorecardV2Handler()
		log.Println("⚠️  V2 Handler initialized without metrics fetcher (auto-evaluate endpoints will not work)")
	}

	v2 := router.Group("/api/v2")
	{
		// Scorecard definition routes
		v2.GET("/scorecards/definitions", v2Handler.GetAllScorecardDefinitions)
		v2.GET("/scorecards/definitions/:name", v2Handler.GetScorecardDefinition)

		// Auto-fetch evaluation routes - POST (fetches metrics from Metrics API)
		v2.POST("/scorecards/auto-evaluate", v2Handler.AutoEvaluateService)
		v2.POST("/scorecards/auto-evaluate/:name", v2Handler.AutoEvaluateServiceByScorecardName)

		// Auto-fetch evaluation routes - GET (frontend-friendly, uses query parameters)
		v2.GET("/scorecards/auto-evaluate", v2Handler.AutoEvaluateServiceGET)
		v2.GET("/scorecards/auto-evaluate/:name", v2Handler.AutoEvaluateServiceByScorecardNameGET)

		// All scorecards in one call (frontend-friendly)
		v2.GET("/scorecards/all", v2Handler.GetAllScorecards)

		// Dedicated scorecard endpoints (GET with query params - fetches from Metrics API)
		v2.GET("/scorecards/code-quality", v2Handler.GetCodeQualityScorecard)
		v2.GET("/scorecards/service-health", v2Handler.GetServiceHealthScorecard)
		v2.GET("/scorecards/security-maturity", v2Handler.GetSecurityMaturityScorecard)
		v2.GET("/scorecards/production-readiness", v2Handler.GetProductionReadinessScorecard)
		v2.GET("/scorecards/pr-metrics", v2Handler.GetPRMetricsScorecard)
	}

	log.Println("✅ API v2 routes registered")
	log.Println("   - All scorecards (single call): GET /api/v2/scorecards/all?service_name=xxx")
	log.Println("   - Dedicated scorecards: GET /api/v2/scorecards/{code-quality|service-health|security-maturity|production-readiness|pr-metrics}?service_name=xxx")
	log.Println("   - Auto evaluation: GET /api/v2/scorecards/auto-evaluate?service_name=xxx&jira_project_key=yyy")
	log.Println("   - Definitions: GET /api/v2/scorecards/definitions")
}
