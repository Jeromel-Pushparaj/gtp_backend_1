package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/handler"
)

// SetupRoutes configures all routes for the service
func SetupRoutes(router *gin.Engine, serviceHandler *handler.ServiceHandler) {
	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Service routes
		service := api.Group("/service")
		{
			// Fetch/Create service from sonar-shell-test API
			service.POST("", serviceHandler.FetchServices)

			// Get all services from cache
			service.GET("", serviceHandler.GetAllServices)

			// Get specific service from cache by ID
			service.GET("/:id", serviceHandler.GetService)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "service-catalog",
		})
	})
}
