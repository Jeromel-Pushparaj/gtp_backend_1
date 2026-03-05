package main

import (
	"fmt"
	"log"
	"pd-service/clients"
	"pd-service/config"
	"pd-service/handlers"
	"pd-service/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ No .env file found or error loading it: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Log configuration (without exposing full keys)
	log.Printf("🔧 Configuration loaded:")
	log.Printf("  ✓ PagerDuty API Key: %s", maskKey(cfg.PagerDutyAPIKey))
	log.Printf("  ✓ Slack Bot Token: %s", maskKey(cfg.SlackBotToken))
	log.Printf("  ✓ GitHub PAT: %s", maskKey(cfg.GitHubPAT))
	log.Printf("  ✓ Default Org: %s", cfg.DefaultOrg)
	log.Printf("  ✓ MongoDB URL: %s", maskMongoURL(cfg.MongoURL))
	log.Printf("  ✓ DB Name: %s", cfg.DBName)

	// Initialize storage
	store := storage.NewInMemoryStorage("data/services.json")
	log.Printf("💾 In-memory storage initialized")

	// Initialize MongoDB storage
	var mongoStore *storage.MongoDBStorage
	if cfg.MongoURL != "" {
		var err error
		mongoStore, err = storage.NewMongoDBStorage(cfg.MongoURL, cfg.DBName)
		if err != nil {
			log.Printf("⚠️ Failed to connect to MongoDB: %v", err)
			log.Printf("⚠️ Continuing without MongoDB storage...")
		} else {
			log.Printf("💾 MongoDB storage initialized")
			// Ensure MongoDB connection is closed on exit
			defer func() {
				if err := mongoStore.Close(); err != nil {
					log.Printf("⚠️ Error closing MongoDB connection: %v", err)
				}
			}()
		}
	} else {
		log.Printf("⚠️ MongoDB URL not configured, skipping MongoDB storage")
	}

	// Initialize API clients
	pdClient := clients.NewPagerDutyClient(cfg.PagerDutyAPIKey)
	slackClient := clients.NewSlackClient(cfg.SlackBotToken)
	githubClient := clients.NewGitHubClient(cfg.GitHubPAT)
	log.Printf("🔌 API clients initialized (PagerDuty, Slack, GitHub)")

	// Initialize handlers
	h := handlers.NewHandler(store, mongoStore, pdClient, slackClient, githubClient, cfg.DefaultOrg)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Custom logging middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		statusColor := getStatusColor(param.StatusCode)
		methodColor := getMethodColor(param.Method)

		return fmt.Sprintf("[GIN] %s | %s%3d%s | %13v | %15s | %s%-7s%s %s\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor(),
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor(),
			param.Path,
			param.ErrorMessage,
		)
	}))

	// API routes
	api := r.Group("/api")
	{
		// Organizations
		api.GET("/organizations", h.GetOrganizations)

		// Services
		api.GET("/services", h.ListServices)
		api.POST("/services", h.CreateService)
		api.GET("/services/:id", h.GetService)
		api.DELETE("/services/:id", h.DeleteService)

		// Metrics
		api.GET("/metrics", h.GetAllMetrics)
		api.POST("/metrics/by-name", h.GetMetricsByServiceName)
		api.GET("/services/:id/metrics", h.GetServiceMetrics)

		// PagerDuty
		api.GET("/pagerduty/services", h.ListPDServices)
		api.GET("/pagerduty/escalation-policies", h.ListEscalationPolicies)

		// GitHub
		api.GET("/github/repos", h.ListGitHubRepos)

		// Slack
		api.GET("/slack/users", h.ListSlackUsers)

		// Incidents
		api.POST("/incidents/trigger", h.TriggerIncident)
	}

	// Serve static files (must be after API routes)
	r.StaticFile("/", "./frontend/index.html")
	r.StaticFile("/index.html", "./frontend/index.html")
	r.StaticFile("/styles.css", "./frontend/styles.css")
	r.StaticFile("/app.js", "./frontend/app.js")

	// Start server
	addr := ":" + cfg.Port
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("🌐 Frontend available at http://localhost:%s", cfg.Port)
	log.Printf("📡 API available at http://localhost:%s/api", cfg.Port)
	log.Printf("\n✨ Ready to accept requests!\n")

	if err := r.Run(addr); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

func maskKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func maskMongoURL(url string) string {
	if url == "" {
		return "not configured"
	}
	// Mask password in MongoDB URL
	// Format: mongodb+srv://username:password@host/...
	if len(url) > 20 {
		return url[:14] + "***" + url[len(url)-20:]
	}
	return "***"
}

func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[32m" // Green
	case code >= 300 && code < 400:
		return "\033[36m" // Cyan
	case code >= 400 && code < 500:
		return "\033[33m" // Yellow
	default:
		return "\033[31m" // Red
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // Blue
	case "POST":
		return "\033[32m" // Green
	case "PUT":
		return "\033[33m" // Yellow
	case "DELETE":
		return "\033[31m" // Red
	default:
		return "\033[37m" // White
	}
}

func resetColor() string {
	return "\033[0m"
}

