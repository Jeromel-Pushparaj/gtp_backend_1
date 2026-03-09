package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/client"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/config"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/db"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/handler"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/routes"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database service (local cache)
	log.Println("Initializing local cache database...")
	dbService, err := db.NewDatabaseService(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()

	// Initialize sonar-shell-test API client
	log.Printf("Initializing sonar-shell-test API client (URL: %s)", cfg.SonarShellTestURL)
	sonarClient := client.NewSonarClient(cfg.SonarShellTestURL, cfg.SonarShellTestAPIKey)

	// Initialize service layer
	serviceService := service.NewServiceService(sonarClient, dbService)

	// Initialize handler layer (controller logic merged into handler)
	serviceHandler := handler.NewServiceHandler(serviceService)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Setup routes with dependency injection
	routes.SetupRoutes(router, serviceHandler)

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort)
	log.Printf("🚀 %s starting on %s (Environment: %s)", cfg.ServiceName, address, cfg.Environment)
	log.Printf("📝 Service Catalog API Endpoints:")
	log.Printf("   - POST   /api/v1/service?org_id=<id>  (Fetch/Create services from sonar-shell-test)")
	log.Printf("   - GET    /api/v1/service              (Get all services from cache)")
	log.Printf("   - GET    /api/v1/service/:id          (Get specific service by ID)")
	log.Printf("   - GET    /health                      (Health check)")

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
