package models

import "time"

// ScorecardCategory represents the category of a scorecard
type ScorecardCategory string

const (
	CategoryCodeQuality        ScorecardCategory = "code_quality"
	CategorySecurity           ScorecardCategory = "security"
	CategoryReliability        ScorecardCategory = "reliability"
	CategoryDevelopmentVelocity ScorecardCategory = "development_velocity"
	CategoryProductionReadiness ScorecardCategory = "production_readiness"
	CategoryServiceHealth      ScorecardCategory = "service_health"
)

// ScorecardDefinition represents a scorecard template (e.g., "Code Quality", "DORA Metrics")
type ScorecardDefinition struct {
	ID           int64             `json:"id" db:"id"`
	Name         string            `json:"name" db:"name"`                   // e.g., "CodeQuality"
	DisplayName  string            `json:"display_name" db:"display_name"`   // e.g., "Code Quality"
	Category     ScorecardCategory `json:"category" db:"category"`
	Description  string            `json:"description" db:"description"`
	LevelPattern LevelPattern      `json:"level_pattern" db:"level_pattern"` // metal, performance, etc.
	Levels       []Level           `json:"levels,omitempty"`                 // Levels with rules
	IsActive     bool              `json:"is_active" db:"is_active"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
}

// ScorecardEvaluation represents the result of evaluating a service against a scorecard
type ScorecardEvaluation struct {
	ID                int64     `json:"id" db:"id"`
	ServiceName       string    `json:"service_name" db:"service_name"`
	ScorecardID       int64     `json:"scorecard_id" db:"scorecard_id"`
	ScorecardName     string    `json:"scorecard_name" db:"scorecard_name"`
	AchievedLevelID   *int64    `json:"achieved_level_id,omitempty" db:"achieved_level_id"`
	AchievedLevelName string    `json:"achieved_level_name" db:"achieved_level_name"`
	RulesPassed       int       `json:"rules_passed" db:"rules_passed"`
	RulesTotal        int       `json:"rules_total" db:"rules_total"`
	PassPercentage    float64   `json:"pass_percentage" db:"pass_percentage"`
	RuleResults       []RuleResult `json:"rule_results,omitempty"`
	EvaluatedAt       time.Time `json:"evaluated_at" db:"evaluated_at"`
}

// ServiceOverallScore represents the overall score across all scorecards
type ServiceOverallScore struct {
	ServiceName       string                `json:"service_name"`
	OverallPercentage float64               `json:"overall_percentage"`
	TotalRulesPassed  int                   `json:"total_rules_passed"`
	TotalRules        int                   `json:"total_rules"`
	Scorecards        []ScorecardEvaluation `json:"scorecards"`
	Strengths         []string              `json:"strengths"`
	ImprovementAreas  []string              `json:"improvement_areas"`
	EvaluatedAt       time.Time             `json:"evaluated_at"`
}

// ScorecardSummary provides a quick overview of a scorecard evaluation
type ScorecardSummary struct {
	ScorecardName     string  `json:"scorecard_name"`
	CurrentLevel      string  `json:"current_level"`
	RulesPassed       int     `json:"rules_passed"`
	RulesTotal        int     `json:"rules_total"`
	PassPercentage    float64 `json:"pass_percentage"`
	Icon              string  `json:"icon"`
	Color             string  `json:"color"`
}

