package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/controller"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/db"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/service"
)

// SetupRoutes sets up all routes and dependencies
func SetupRoutes(router *gin.Engine) {
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize dependencies (Dependency Injection)
	serviceRepo := db.NewServiceRepository()
	onboardingService := service.NewOnboardingService(serviceRepo)
	onboardingController := controller.NewOnboardingController(onboardingService)

	// Health check endpoint
	router.GET("/health", onboardingController.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Onboarding endpoint
		api.POST("/onboard", onboardingController.OnboardService)

		// Service endpoints
		api.GET("/services", onboardingController.GetAllServices)
		api.GET("/services/:id", onboardingController.GetServiceByID)
	}
}
