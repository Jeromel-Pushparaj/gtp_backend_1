package handlers

import (
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/engine"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/fetcher"
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

