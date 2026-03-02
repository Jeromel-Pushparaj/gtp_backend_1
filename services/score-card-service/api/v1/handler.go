package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/service"
)

// Handler handles HTTP requests for scorecard API
type Handler struct {
	service *service.ScoreCardService
}

// NewHandler creates a new handler
func NewHandler(svc *service.ScoreCardService) *Handler {
	return &Handler{
		service: svc,
	}
}

// CreateScoreCard handles POST /api/v1/scorecards
func (h *Handler) CreateScoreCard(c *gin.Context) {
	var req models.ScoreCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := h.service.CreateScoreCard(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create scorecard",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetScoreCard handles GET /api/v1/scorecards/:id
func (h *Handler) GetScoreCard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scorecard ID",
		})
		return
	}

	response, err := h.service.GetScoreCard(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetScoreCardsByService handles GET /api/v1/scorecards/service/:name
func (h *Handler) GetScoreCardsByService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	responses, err := h.service.GetScoreCardsByService(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get scorecards",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"count":   len(responses),
		"scorecards": responses,
	})
}

// GetLatestScoreCard handles GET /api/v1/scorecards/service/:name/latest
func (h *Handler) GetLatestScoreCard(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	response, err := h.service.GetLatestScoreCard(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No scorecard found for service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

