package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// evaluateDedicatedScorecard is a helper function to evaluate a specific scorecard
// This reduces code duplication across the 5 dedicated scorecard endpoints
func (h *ScorecardV2Handler) evaluateDedicatedScorecard(c *gin.Context, scorecardName string, scorecardGetter func() models.ScorecardDefinition) {
	serviceName := c.Query("service_name")
	owner := c.Query("owner")
	jiraProjectKey := c.Query("jira_project_key")
	onlyAchieved := c.Query("only_achieved_level")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name query parameter is required",
		})
		return
	}

	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured",
		})
		return
	}

	// Fetch metrics
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)
	if owner != "" {
		combinedMetrics.ServiceName = owner + "/" + serviceName
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics",
			"details": err.Error(),
		})
		return
	}

	// Get the scorecard definition
	scorecard := scorecardGetter()

	// Convert metrics to map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Evaluate
	evaluation, err := h.evaluator.EvaluateService(serviceName, scorecard, serviceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to evaluate service",
			"details": err.Error(),
		})
		return
	}

	// Check if user wants only achieved level rules
	if onlyAchieved == "true" {
		filteredResults := h.filterAchievedLevelRules(evaluation, scorecard)
		evaluation.RuleResults = filteredResults
		evaluation.RulesTotal = len(filteredResults)
	}

	c.JSON(http.StatusOK, gin.H{
		"scorecard":       scorecardName,
		"evaluation":      evaluation,
		"fetched_metrics": combinedMetrics,
	})
}

// GetCodeQualityScorecard handles GET /api/v2/scorecards/code-quality
// Dedicated endpoint for Code Quality scorecard evaluation
func (h *ScorecardV2Handler) GetCodeQualityScorecard(c *gin.Context) {
	h.evaluateDedicatedScorecard(c, "Code Quality", func() models.ScorecardDefinition {
		return scorecards.GetCodeQualityScorecard()
	})
}

// GetServiceHealthScorecard handles GET /api/v2/scorecards/service-health
// Dedicated endpoint for Service Health scorecard evaluation
func (h *ScorecardV2Handler) GetServiceHealthScorecard(c *gin.Context) {
	h.evaluateDedicatedScorecard(c, "Service Health", func() models.ScorecardDefinition {
		return scorecards.GetServiceHealthScorecard()
	})
}

// GetSecurityMaturityScorecard handles GET /api/v2/scorecards/security-maturity
// Dedicated endpoint for Security Maturity scorecard evaluation
func (h *ScorecardV2Handler) GetSecurityMaturityScorecard(c *gin.Context) {
	h.evaluateDedicatedScorecard(c, "Security Maturity", func() models.ScorecardDefinition {
		return scorecards.GetSecurityMaturityScorecard()
	})
}

// GetProductionReadinessScorecard handles GET /api/v2/scorecards/production-readiness
// Dedicated endpoint for Production Readiness scorecard evaluation
func (h *ScorecardV2Handler) GetProductionReadinessScorecard(c *gin.Context) {
	h.evaluateDedicatedScorecard(c, "Production Readiness", func() models.ScorecardDefinition {
		return scorecards.GetProductionReadinessScorecard()
	})
}

// GetPRMetricsScorecard handles GET /api/v2/scorecards/pr-metrics
// Dedicated endpoint for PR Metrics scorecard evaluation
func (h *ScorecardV2Handler) GetPRMetricsScorecard(c *gin.Context) {
	h.evaluateDedicatedScorecard(c, "PR Metrics", func() models.ScorecardDefinition {
		return scorecards.GetPRMetricsScorecard()
	})
}

// ScorecardSummary represents a summary of a single scorecard evaluation
type ScorecardSummary struct {
	Name              string  `json:"name"`
	AchievedLevelName string  `json:"achieved_level_name"`
	RulesPassed       int     `json:"rules_passed"`
	RulesTotal        int     `json:"rules_total"`
	PassPercentage    float64 `json:"pass_percentage"`
}

// GetAllScorecardsResponse represents the response for all scorecards
type GetAllScorecardsResponse struct {
	ServiceName string             `json:"service_name"`
	Scorecards  []ScorecardSummary `json:"scorecards"`
}

// GetAllScorecards handles GET /api/v2/scorecards/all
// Returns all 5 scorecards in a single API call for frontend consumption
func (h *ScorecardV2Handler) GetAllScorecards(c *gin.Context) {
	serviceName := c.Query("service_name")
	owner := c.Query("owner")
	jiraProjectKey := c.Query("jira_project_key")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name query parameter is required",
		})
		return
	}

	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured",
		})
		return
	}

	// Fetch metrics once
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)
	if owner != "" {
		combinedMetrics.ServiceName = owner + "/" + serviceName
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics",
			"details": err.Error(),
		})
		return
	}

	// Convert metrics to map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Evaluate all 5 scorecards
	scorecardGetters := []struct {
		name   string
		getter func() models.ScorecardDefinition
	}{
		{"Code Quality", scorecards.GetCodeQualityScorecard},
		{"Service Health", scorecards.GetServiceHealthScorecard},
		{"Security Maturity", scorecards.GetSecurityMaturityScorecard},
		{"Production Readiness", scorecards.GetProductionReadinessScorecard},
		{"PR Metrics", scorecards.GetPRMetricsScorecard},
	}

	var summaries []ScorecardSummary

	for _, sc := range scorecardGetters {
		scorecard := sc.getter()
		evaluation, err := h.evaluator.EvaluateService(serviceName, scorecard, serviceData)
		if err != nil {
			continue // Skip if evaluation fails
		}

		summaries = append(summaries, ScorecardSummary{
			Name:              sc.name,
			AchievedLevelName: evaluation.AchievedLevelName,
			RulesPassed:       evaluation.RulesPassed,
			RulesTotal:        evaluation.RulesTotal,
			PassPercentage:    evaluation.PassPercentage,
		})
	}

	response := GetAllScorecardsResponse{
		ServiceName: serviceName,
		Scorecards:  summaries,
	}

	c.JSON(http.StatusOK, response)
}
