package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// AutoEvaluateService handles POST /api/v2/scorecards/auto-evaluate
// This endpoint automatically fetches metrics from GitHub, Jira, and SonarCloud APIs
func (h *ScorecardV2Handler) AutoEvaluateService(c *gin.Context) {
	var req AutoEvaluateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if metrics fetcher is configured
	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured. Please use /api/v2/scorecards/evaluate endpoint with manual service_data",
		})
		return
	}

	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(req.ServiceName, req.JiraProjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics from external APIs",
			"details": err.Error(),
		})
		return
	}

	// Convert CombinedMetrics to service data map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Get all scorecard definitions
	definitions := scorecards.GetAllScorecardDefinitions()

	// Evaluate against all scorecards
	overall, err := h.evaluator.EvaluateAllScorecards(req.ServiceName, definitions, serviceData)
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

	c.JSON(http.StatusOK, gin.H{
		"evaluation":      overall,
		"fetched_metrics": combinedMetrics,
	})
}

// AutoEvaluateServiceByScorecardName handles POST /api/v2/scorecards/auto-evaluate/:name
// This endpoint automatically fetches metrics and evaluates against a specific scorecard
func (h *ScorecardV2Handler) AutoEvaluateServiceByScorecardName(c *gin.Context) {
	scorecardName := c.Param("name")

	var req AutoEvaluateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if metrics fetcher is configured
	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured. Please use /api/v2/scorecards/evaluate/:name endpoint with manual service_data",
		})
		return
	}

	// Get the specific scorecard definition
	scorecard, found := getScorecardByName(scorecardName)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard definition not found",
		})
		return
	}

	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(req.ServiceName, req.JiraProjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics from external APIs",
			"details": err.Error(),
		})
		return
	}

	// Convert CombinedMetrics to service data map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Evaluate the service
	evaluation, err := h.evaluator.EvaluateService(req.ServiceName, scorecard, serviceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to evaluate service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"evaluation":      evaluation,
		"fetched_metrics": combinedMetrics,
	})
}

// AutoEvaluateServiceGET handles GET /api/v2/scorecards/auto-evaluate
// This endpoint automatically fetches metrics from GitHub, Jira, and SonarCloud APIs (GET version)
func (h *ScorecardV2Handler) AutoEvaluateServiceGET(c *gin.Context) {
	serviceName := c.Query("service_name")
	owner := c.Query("owner") // Optional: GitHub owner/organization
	jiraProjectKey := c.Query("jira_project_key")
	summary := c.Query("summary") // "true" for summary view (default), "false" for detailed view

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name query parameter is required",
		})
		return
	}

	// Check if metrics fetcher is configured
	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured. Please use /api/v2/scorecards/evaluate endpoint with manual service_data",
		})
		return
	}

	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)
	if owner != "" {
		combinedMetrics.ServiceName = owner + "/" + serviceName
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics from external APIs",
			"details": err.Error(),
		})
		return
	}

	// Convert CombinedMetrics to service data map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Get all scorecard definitions
	definitions := scorecards.GetAllScorecardDefinitions()

	// Evaluate against all scorecards
	overall, err := h.evaluator.EvaluateAllScorecards(serviceName, definitions, serviceData)
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

	// Return summary view by default (summary=true or not specified)
	if summary != "false" {
		// Create summary response without detailed rule results
		summaryEvaluation := h.createSummaryEvaluation(overall)
		c.JSON(http.StatusOK, gin.H{
			"evaluation":      summaryEvaluation,
			"fetched_metrics": combinedMetrics,
		})
		return
	}

	// Return full detailed view
	c.JSON(http.StatusOK, gin.H{
		"evaluation":      overall,
		"fetched_metrics": combinedMetrics,
	})
}

// AutoEvaluateServiceByScorecardNameGET handles GET /api/v2/scorecards/auto-evaluate/:name
// This endpoint automatically fetches metrics and evaluates against a specific scorecard (GET version)
func (h *ScorecardV2Handler) AutoEvaluateServiceByScorecardNameGET(c *gin.Context) {
	scorecardName := c.Param("name")
	serviceName := c.Query("service_name")
	owner := c.Query("owner") // Optional: GitHub owner/organization
	jiraProjectKey := c.Query("jira_project_key")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name query parameter is required",
		})
		return
	}

	// Check if metrics fetcher is configured
	if h.metricsFetcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Metrics fetcher not configured. Please use /api/v2/scorecards/evaluate/:name endpoint with manual service_data",
		})
		return
	}

	// Get the specific scorecard definition
	scorecard, found := getScorecardByName(scorecardName)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard definition not found",
		})
		return
	}

	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)
	if owner != "" {
		combinedMetrics.ServiceName = owner + "/" + serviceName
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics from external APIs",
			"details": err.Error(),
		})
		return
	}

	// Convert CombinedMetrics to service data map
	serviceData := convertMetricsToServiceData(combinedMetrics)

	// Evaluate the service
	evaluation, err := h.evaluator.EvaluateService(serviceName, scorecard, serviceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to evaluate service",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"evaluation":      evaluation,
		"fetched_metrics": combinedMetrics,
	})
}
