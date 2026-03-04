package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/resources"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/service"
)

// OnboardingController handles HTTP requests for service onboarding
type OnboardingController struct {
	onboardingService *service.OnboardingService
}

// NewOnboardingController creates a new onboarding controller
func NewOnboardingController(onboardingService *service.OnboardingService) *OnboardingController {
	return &OnboardingController{
		onboardingService: onboardingService,
	}
}

// OnboardService handles the service onboarding request
func (ctrl *OnboardingController) OnboardService(c *gin.Context) {
	var req resources.OnboardServiceRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, resources.APIResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Call service layer
	response, validationErrors := ctrl.onboardingService.OnboardService(req)
	if len(validationErrors) > 0 {
		c.JSON(400, resources.APIResponse{
			Status:  "error",
			Message: "Validation failed",
			Errors:  validationErrors,
		})
		return
	}

	// Return success response
	c.JSON(201, resources.APIResponse{
		Status:  "success",
		Message: "Service onboarded successfully",
		Data:    response,
	})
}

// GetAllServices returns all onboarded services
func (ctrl *OnboardingController) GetAllServices(c *gin.Context) {
	services, err := ctrl.onboardingService.GetAllServices()
	if err != nil {
		c.JSON(500, resources.APIResponse{
			Status:  "error",
			Message: "Failed to retrieve services",
		})
		return
	}

	c.JSON(200, resources.APIResponse{
		Status:  "success",
		Message: "Services retrieved successfully",
		Data:    services,
	})
}

// GetServiceByID returns a service by ID
func (ctrl *OnboardingController) GetServiceByID(c *gin.Context) {
	id := c.Param("id")

	service, err := ctrl.onboardingService.GetServiceByID(id)
	if err != nil {
		c.JSON(404, resources.APIResponse{
			Status:  "error",
			Message: "Service not found",
		})
		return
	}

	c.JSON(200, resources.APIResponse{
		Status:  "success",
		Message: "Service retrieved successfully",
		Data:    service,
	})
}

// HealthCheck returns the health status of the service
func (ctrl *OnboardingController) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":         "healthy",
		"service":        "onboarding-service",
		"total_services": ctrl.onboardingService.GetServiceCount(),
	})
}
