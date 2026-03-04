package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/resources"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/service"
)

// ServiceHandler handles HTTP requests for service operations
type ServiceHandler struct {
	serviceService *service.ServiceService
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(serviceService *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

// FetchServices handles POST /api/v1/service
func (h *ServiceHandler) FetchServices(c *gin.Context) {
	// Bind query parameters to DTO
	var req resources.CreateServiceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.APIResponse{
			Status:  "error",
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	log.Printf("Handler: Fetching services for org_id=%d", req.OrgID)

	// Call service
	repos, err := h.serviceService.FetchServices(req.OrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.APIResponse{
			Status:  "error",
			Message: "Failed to fetch services",
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resources.APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Successfully fetched %d services", len(repos)),
		Data: resources.ServicesResponse{
			Total:    len(repos),
			Services: repos,
		},
	})
}

// GetService handles GET /api/v1/service/:id
func (h *ServiceHandler) GetService(c *gin.Context) {
	// Bind URI parameter to DTO
	var req resources.GetServiceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, resources.APIResponse{
			Status:  "error",
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	log.Printf("Handler: Getting service %s", req.ID)

	// Call service
	repo, err := h.serviceService.GetService(req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, resources.APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Service '%s' not found", req.ID),
			Error:   err.Error(),
		})
		return
	}

	if repo == nil {
		c.JSON(http.StatusNotFound, resources.APIResponse{
			Status:  "error",
			Message: "Service not found",
			Error:   "service not found in cache",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resources.APIResponse{
		Status:  "success",
		Message: "Service retrieved successfully",
		Data:    repo,
	})
}

// GetAllServices handles GET /api/v1/service
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	log.Printf("Handler: Getting all services")

	// Call service
	reposResponse, err := h.serviceService.GetAllServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, resources.APIResponse{
			Status:  "error",
			Message: "Failed to get services",
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resources.APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Successfully retrieved %d services", reposResponse.Total),
		Data:    reposResponse,
	})
}
