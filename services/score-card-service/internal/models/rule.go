package models

import "time"

// RuleOperator represents comparison operators for rules
type RuleOperator string

const (
	OperatorGreaterThanOrEqual RuleOperator = ">="
	OperatorLessThanOrEqual    RuleOperator = "<="
	OperatorEqual              RuleOperator = "=="
	OperatorNotEqual           RuleOperator = "!="
	OperatorGreaterThan        RuleOperator = ">"
	OperatorLessThan           RuleOperator = "<"
)

// RuleType represents the type of rule
type RuleType string

const (
	RuleTypeProperty  RuleType = "property"  // Check a metric property
	RuleTypeRelation  RuleType = "relation"  // Compare to other services
	RuleTypeScorecard RuleType = "scorecard" // Depends on another scorecard
)

// Rule represents a single evaluation rule
type Rule struct {
	ID          int64        `json:"id" db:"id"`
	LevelID     int64        `json:"level_id" db:"level_id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Property    string       `json:"property" db:"property"`       // e.g., "coverage", "code_smells"
	Operator    RuleOperator `json:"operator" db:"operator"`       // e.g., ">=", "<="
	Threshold   float64      `json:"threshold" db:"threshold"`     // e.g., 60, 10
	RuleType    RuleType     `json:"rule_type" db:"rule_type"`     // property, relation, scorecard
	Weight      float64      `json:"weight" db:"weight"`           // Optional weight for scoring
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// RuleResult represents the result of evaluating a single rule
type RuleResult struct {
	ID            int64     `json:"id" db:"id"`
	EvaluationID  int64     `json:"evaluation_id" db:"evaluation_id"`
	RuleID        int64     `json:"rule_id" db:"rule_id"`
	RuleName      string    `json:"rule_name" db:"rule_name"`
	Passed        bool      `json:"passed" db:"passed"`
	ActualValue   float64   `json:"actual_value" db:"actual_value"`
	ExpectedValue float64   `json:"expected_value" db:"expected_value"`
	Operator      string    `json:"operator" db:"operator"`
	Message       string    `json:"message" db:"message"`
	EvaluatedAt   time.Time `json:"evaluated_at" db:"evaluated_at"`
}

// Evaluate evaluates a rule against a metric value
func (r *Rule) Evaluate(actualValue float64) bool {
	switch r.Operator {
	case OperatorGreaterThanOrEqual:
		return actualValue >= r.Threshold
	case OperatorLessThanOrEqual:
		return actualValue <= r.Threshold
	case OperatorEqual:
		return actualValue == r.Threshold
	case OperatorNotEqual:
		return actualValue != r.Threshold
	case OperatorGreaterThan:
		return actualValue > r.Threshold
	case OperatorLessThan:
		return actualValue < r.Threshold
	default:
		return false
	}
}

// GetMessage returns a human-readable message about the rule evaluation
func (r *Rule) GetMessage(actualValue float64, passed bool) string {
	if passed {
		return "✅ " + r.Description
	}
	return "❌ " + r.Description + " (expected " + string(r.Operator) + " " + 
		formatFloat(r.Threshold) + ", got " + formatFloat(actualValue) + ")"
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune(int64(f)))
	}
	return string(rune(f))
}

