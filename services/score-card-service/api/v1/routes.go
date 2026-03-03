package v1

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/config"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/repository"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/service"
)

// RegisterRoutes registers all API v1 routes
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Initialize database connection
	db, err := initDB(cfg)
	if err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("⚠️  Service will run without database functionality")
		// Continue without DB for now - you can make this fatal if DB is required
	}

	// Initialize repository and service
	var handler *Handler
	if db != nil {
		repo := repository.NewScoreCardRepository(db)
		svc := service.NewScoreCardService(repo)
		handler = NewHandler(svc)
	}

	// API v1 group - Simple scorecards
	v1 := router.Group("/api/v1")
	{
		// Scorecard routes
		scorecards := v1.Group("/scorecards")
		{
			if handler != nil {
				scorecards.POST("", handler.CreateScoreCard)
				scorecards.GET("/:id", handler.GetScoreCard)
				scorecards.GET("/service/:name", handler.GetScoreCardsByService)
				scorecards.GET("/service/:name/latest", handler.GetLatestScoreCard)
			} else {
				// Placeholder routes when DB is not available
				scorecards.GET("", func(c *gin.Context) {
					c.JSON(503, gin.H{
						"error": "Database not available",
					})
				})
			}
		}
	}

	// API v2 group - Advanced scorecards with levels (Gold/Silver/Bronze, Traffic Lights, etc.)
	// Create v2 handler with metrics fetcher if configured
	var v2Handler *ScorecardV2Handler
	if cfg.MetricsAPIBaseURL != "" {
		v2Handler = NewScorecardV2HandlerWithFetcher(cfg.MetricsAPIBaseURL)
		log.Printf("✅ V2 Handler initialized with metrics fetcher (base URL: %s)", cfg.MetricsAPIBaseURL)
	} else {
		v2Handler = NewScorecardV2Handler()
		log.Println("⚠️  V2 Handler initialized without metrics fetcher (auto-evaluate endpoints will not work)")
	}

	v2 := router.Group("/api/v2")
	{
		// Scorecard definition routes
		v2.GET("/scorecards/definitions", v2Handler.GetAllScorecardDefinitions)
		v2.GET("/scorecards/definitions/:name", v2Handler.GetScorecardDefinition)

		// Manual evaluation routes (requires service_data in request body)
		v2.POST("/scorecards/evaluate", v2Handler.EvaluateService)
		v2.POST("/scorecards/evaluate/:name", v2Handler.EvaluateServiceByScorecardName)

		// Auto-fetch evaluation routes - POST (fetches metrics from external APIs)
		v2.POST("/scorecards/auto-evaluate", v2Handler.AutoEvaluateService)
		v2.POST("/scorecards/auto-evaluate/:name", v2Handler.AutoEvaluateServiceByScorecardName)

		// Auto-fetch evaluation routes - GET (frontend-friendly, uses query parameters)
		v2.GET("/scorecards/auto-evaluate", v2Handler.AutoEvaluateServiceGET)
		v2.GET("/scorecards/auto-evaluate/:name", v2Handler.AutoEvaluateServiceByScorecardNameGET)
	}

	log.Println("✅ API v1 routes registered")
	log.Println("✅ API v2 routes registered (Advanced Scorecards)")
	log.Println("   - Manual evaluation: POST /api/v2/scorecards/evaluate")
	log.Println("   - Auto evaluation (POST): POST /api/v2/scorecards/auto-evaluate")
	log.Println("   - Auto evaluation (GET): GET /api/v2/scorecards/auto-evaluate?service_name=xxx&jira_project_key=yyy")
}

// initDB initializes the database connection
func initDB(cfg *config.Config) (*sql.DB, error) {
	connStr := "host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Database connected successfully")

	// Create table if it doesn't exist
	if err := createTables(db); err != nil {
		log.Printf("⚠️  Failed to create tables: %v", err)
	}

	return db, nil
}

// createTables creates the necessary database tables
func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS scorecards (
		id SERIAL PRIMARY KEY,
		service_name VARCHAR(255) NOT NULL,
		score DECIMAL(5,2) NOT NULL,
		code_quality DECIMAL(5,2) NOT NULL,
		test_coverage DECIMAL(5,2) NOT NULL,
		security_score DECIMAL(5,2) NOT NULL,
		performance_score DECIMAL(5,2) NOT NULL,
		documentation_score DECIMAL(5,2) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_scorecards_service_name ON scorecards(service_name);
	CREATE INDEX IF NOT EXISTS idx_scorecards_created_at ON scorecards(created_at DESC);
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Println("✅ Database tables created/verified")
	return nil
}
