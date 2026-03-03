package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/engine"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// ScorecardV2Handler handles advanced scorecard evaluation requests
type ScorecardV2Handler struct {
	evaluator *engine.Evaluator
}

// NewScorecardV2Handler creates a new advanced scorecard handler
func NewScorecardV2Handler() *ScorecardV2Handler {
	return &ScorecardV2Handler{
		evaluator: engine.NewEvaluator(),
	}
}

// EvaluateServiceRequest represents a request to evaluate a service
type EvaluateServiceRequest struct {
	ServiceName string                 `json:"service_name" binding:"required"`
	ServiceData map[string]interface{} `json:"service_data" binding:"required"`
}

// GetAllScorecardDefinitions handles GET /api/v2/scorecards/definitions
func (h *ScorecardV2Handler) GetAllScorecardDefinitions(c *gin.Context) {
	definitions := scorecards.GetAllScorecardDefinitions()

	c.JSON(http.StatusOK, gin.H{
		"count":      len(definitions),
		"scorecards": definitions,
	})
}

// GetScorecardDefinition handles GET /api/v2/scorecards/definitions/:name
func (h *ScorecardV2Handler) GetScorecardDefinition(c *gin.Context) {
	name := c.Param("name")

	var scorecard models.ScorecardDefinition
	var found bool

	switch name {
	case "CodeQuality":
		scorecard = scorecards.GetCodeQualityScorecard()
		found = true
	case "DORA_Metrics":
		scorecard = scorecards.GetDORAMetricsScorecard()
		found = true
	case "Security_Maturity":
		scorecard = scorecards.GetSecurityMaturityScorecard()
		found = true
	case "Production_Readiness":
		scorecard = scorecards.GetProductionReadinessScorecard()
		found = true
	case "Service_Health":
		scorecard = scorecards.GetServiceHealthScorecard()
		found = true
	case "PR_Metrics":
		scorecard = scorecards.GetPRMetricsScorecard()
		found = true
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard definition not found",
		})
		return
	}

	c.JSON(http.StatusOK, scorecard)
}

// EvaluateService handles POST /api/v2/scorecards/evaluate
func (h *ScorecardV2Handler) EvaluateService(c *gin.Context) {
	var req EvaluateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get all scorecard definitions
	definitions := scorecards.GetAllScorecardDefinitions()

	// Evaluate against all scorecards
	overall, err := h.evaluator.EvaluateAllScorecards(req.ServiceName, definitions, req.ServiceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to evaluate service",
			"details": err.Error(),
		})
		return
	}

	// Add strengths and improvement areas
	overall.Strengths = h.identifyStrengths(overall.Scorecards)
	overall.ImprovementAreas = h.identifyImprovements(overall.Scorecards)

	c.JSON(http.StatusOK, overall)
}

// EvaluateServiceByScorecardName handles POST /api/v2/scorecards/evaluate/:name
func (h *ScorecardV2Handler) EvaluateServiceByScorecardName(c *gin.Context) {
	scorecardName := c.Param("name")

	var req EvaluateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the specific scorecard definition
	var scorecard models.ScorecardDefinition
	var found bool

	switch scorecardName {
	case "CodeQuality":
		scorecard = scorecards.GetCodeQualityScorecard()
		found = true
	case "DORA_Metrics":
		scorecard = scorecards.GetDORAMetricsScorecard()
		found = true
	case "Security_Maturity":
		scorecard = scorecards.GetSecurityMaturityScorecard()
		found = true
	case "Production_Readiness":
		scorecard = scorecards.GetProductionReadinessScorecard()
		found = true
	case "Service_Health":
		scorecard = scorecards.GetServiceHealthScorecard()
		found = true
	case "PR_Metrics":
		scorecard = scorecards.GetPRMetricsScorecard()
		found = true
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard definition not found",
		})
		return
	}

	// Evaluate the service
	evaluation, err := h.evaluator.EvaluateService(req.ServiceName, scorecard, req.ServiceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to evaluate service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, evaluation)
}

// identifyStrengths identifies areas where the service is performing well
func (h *ScorecardV2Handler) identifyStrengths(evaluations []models.ScorecardEvaluation) []string {
	var strengths []string

	for _, eval := range evaluations {
		if eval.AchievedLevelName != "None" && eval.PassPercentage >= 80 {
			strengths = append(strengths, eval.ScorecardName+": "+eval.AchievedLevelName)
		}
	}

	return strengths
}

// identifyImprovements identifies areas that need improvement
func (h *ScorecardV2Handler) identifyImprovements(evaluations []models.ScorecardEvaluation) []string {
	var improvements []string

	for _, eval := range evaluations {
		if eval.AchievedLevelName == "None" || eval.PassPercentage < 50 {
			improvements = append(improvements, eval.ScorecardName+": Needs attention")
		}
	}

	return improvements
}
