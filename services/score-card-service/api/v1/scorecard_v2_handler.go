package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/engine"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/fetcher"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// ScorecardV2Handler handles advanced scorecard evaluation requests
type ScorecardV2Handler struct {
	evaluator      *engine.Evaluator
	metricsFetcher *fetcher.MetricsFetcher
}

// NewScorecardV2Handler creates a new advanced scorecard handler
func NewScorecardV2Handler() *ScorecardV2Handler {
	return &ScorecardV2Handler{
		evaluator: engine.NewEvaluator(),
	}
}

// NewScorecardV2HandlerWithFetcher creates a new advanced scorecard handler with metrics fetcher
func NewScorecardV2HandlerWithFetcher(metricsBaseURL string) *ScorecardV2Handler {
	return &ScorecardV2Handler{
		evaluator:      engine.NewEvaluator(),
		metricsFetcher: fetcher.NewMetricsFetcher(metricsBaseURL),
	}
}

// EvaluateServiceRequest represents a request to evaluate a service
type EvaluateServiceRequest struct {
	ServiceName string                 `json:"service_name" binding:"required"`
	ServiceData map[string]interface{} `json:"service_data" binding:"required"`
}

// AutoEvaluateServiceRequest represents a request to auto-fetch and evaluate a service
type AutoEvaluateServiceRequest struct {
	ServiceName    string `json:"service_name" binding:"required"`
	JiraProjectKey string `json:"jira_project_key,omitempty"`
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

// createSummaryEvaluation creates a summary view without detailed rule results
func (h *ScorecardV2Handler) createSummaryEvaluation(overall *models.ServiceOverallScore) map[string]interface{} {
	// Create summary scorecards without rule_results
	summaryScorecards := make([]map[string]interface{}, len(overall.Scorecards))
	for i, sc := range overall.Scorecards {
		summaryScorecards[i] = map[string]interface{}{
			"scorecard_name":      sc.ScorecardName,
			"achieved_level_name": sc.AchievedLevelName,
			"rules_passed":        sc.RulesPassed,
			"rules_total":         sc.RulesTotal,
			"pass_percentage":     sc.PassPercentage,
		}
	}

	return map[string]interface{}{
		"service_name":       overall.ServiceName,
		"overall_percentage": overall.OverallPercentage,
		"total_rules_passed": overall.TotalRulesPassed,
		"total_rules":        overall.TotalRules,
		"scorecards":         summaryScorecards,
		"strengths":          overall.Strengths,
		"improvement_areas":  overall.ImprovementAreas,
	}
}

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

	// Convert CombinedMetrics to map for evaluation
	metricsMap := combinedMetrics.ToMap()

	// Convert map[string]float64 to map[string]interface{} for evaluator
	serviceData := make(map[string]interface{})
	for k, v := range metricsMap {
		serviceData[k] = v
	}

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
	var scorecard models.ScorecardDefinition
	var found bool

	switch scorecardName {
	case "CodeQuality":
		scorecard = scorecards.GetCodeQualityScorecard()
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

	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(req.ServiceName, req.JiraProjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch metrics from external APIs",
			"details": err.Error(),
		})
		return
	}

	// Convert CombinedMetrics to map for evaluation
	metricsMap := combinedMetrics.ToMap()

	// Convert map[string]float64 to map[string]interface{} for evaluator
	serviceData := make(map[string]interface{})
	for k, v := range metricsMap {
		serviceData[k] = v
	}

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

	// Note: The shell API expects just the repo name (it uses GITHUB_ORG env var for owner)
	// The owner parameter is for display/documentation purposes only
	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)

	// If owner was provided, update the service name in the response for clarity
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

	// Convert CombinedMetrics to map for evaluation
	metricsMap := combinedMetrics.ToMap()

	// Convert map[string]float64 to map[string]interface{} for evaluator
	serviceData := make(map[string]interface{})
	for k, v := range metricsMap {
		serviceData[k] = v
	}

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
	var scorecard models.ScorecardDefinition
	var found bool

	switch scorecardName {
	case "CodeQuality":
		scorecard = scorecards.GetCodeQualityScorecard()
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

	// Note: The shell API expects just the repo name (it uses GITHUB_ORG env var for owner)
	// Fetch all metrics from external APIs
	combinedMetrics, err := h.metricsFetcher.FetchAllMetrics(serviceName, jiraProjectKey)

	// If owner was provided, update the service name in the response for clarity
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

	// Convert CombinedMetrics to map for evaluation
	metricsMap := combinedMetrics.ToMap()

	// Convert map[string]float64 to map[string]interface{} for evaluator
	serviceData := make(map[string]interface{})
	for k, v := range metricsMap {
		serviceData[k] = v
	}

	// Evaluate the service (use original serviceName for display)
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
