package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	v1 "github.com/jeromelp/gtp_backen_1/service/score-card-service/api/v1"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/config"
)

func main() {
	// Load .env file - try current directory first, then parent directory
	if err := godotenv.Load(); err != nil {
		// Try loading from parent directory (when running from cmd/)
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("⚠️  No .env file found, using system environment variables")
		}
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": cfg.ServiceName,
		})
	})

	// Register API v1 routes
	v1.RegisterRoutes(router, cfg)

	// Start server
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("🚀 %s starting on port %s\n", cfg.ServiceName, port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v\n", err)
	}
}
