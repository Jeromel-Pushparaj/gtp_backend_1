package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/config"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Setup routes with dependency injection
	routes.SetupRoutes(router)

	// Start server
	address := fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort)
	log.Printf("🚀 %s starting on %s (Environment: %s)", cfg.ServiceName, address, cfg.Environment)
	log.Printf("📝 Endpoints:")
	log.Printf("   - POST   /api/onboard")
	log.Printf("   - GET    /api/services")
	log.Printf("   - GET    /api/services/:id")
	log.Printf("   - GET    /health")

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
