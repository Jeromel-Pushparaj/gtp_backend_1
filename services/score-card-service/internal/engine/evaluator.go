package engine

import (
	"fmt"
	"time"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
)

// Evaluator evaluates services against scorecard definitions
type Evaluator struct{}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// EvaluateService evaluates a service against a scorecard definition
func (e *Evaluator) EvaluateService(serviceName string, scorecard models.ScorecardDefinition, serviceData map[string]interface{}) (*models.ScorecardEvaluation, error) {
	evaluation := &models.ScorecardEvaluation{
		ServiceName:   serviceName,
		ScorecardName: scorecard.DisplayName,
		EvaluatedAt:   time.Now(),
	}

	var ruleResults []models.RuleResult
	var achievedLevel *models.Level
	totalRules := 0
	rulesPassed := 0

	// Evaluate levels from highest to lowest to find the best achieved level
	for i := len(scorecard.Levels) - 1; i >= 0; i-- {
		level := scorecard.Levels[i]
		allRulesPassed := true
		levelRulesPassed := 0

		for _, rule := range level.Rules {
			totalRules++
			result := e.evaluateRule(rule, serviceData)
			ruleResults = append(ruleResults, result)

			if result.Passed {
				rulesPassed++
				levelRulesPassed++
			} else {
				allRulesPassed = false
			}
		}

		// If all rules for this level passed, this is the achieved level
		if allRulesPassed && len(level.Rules) > 0 {
			achievedLevel = &level
			break
		}
	}

	evaluation.RulesTotal = totalRules
	evaluation.RulesPassed = rulesPassed
	evaluation.RuleResults = ruleResults

	if totalRules > 0 {
		evaluation.PassPercentage = float64(rulesPassed) / float64(totalRules) * 100
	}

	if achievedLevel != nil {
		evaluation.AchievedLevelName = achievedLevel.DisplayName
		evaluation.AchievedLevelID = &achievedLevel.ID
	} else {
		evaluation.AchievedLevelName = "None"
	}

	return evaluation, nil
}

// evaluateRule evaluates a single rule against service data
func (e *Evaluator) evaluateRule(rule models.Rule, serviceData map[string]interface{}) models.RuleResult {
	result := models.RuleResult{
		RuleName:    rule.Name,
		Passed:      false,
		EvaluatedAt: time.Now(),
	}

	// Get the actual value from service data
	actualValue, exists := serviceData[rule.Property]
	if !exists {
		result.Message = fmt.Sprintf("Property '%s' not found in service data", rule.Property)
		return result
	}

	// Convert to float64 for comparison
	var actualFloat float64
	switch v := actualValue.(type) {
	case float64:
		actualFloat = v
	case int:
		actualFloat = float64(v)
	case int64:
		actualFloat = float64(v)
	default:
		result.Message = fmt.Sprintf("Property '%s' has invalid type", rule.Property)
		return result
	}

	result.ActualValue = actualFloat
	result.ExpectedValue = rule.Threshold

	// Evaluate based on operator
	switch rule.Operator {
	case models.OperatorGreaterThan:
		result.Passed = actualFloat > rule.Threshold
	case models.OperatorGreaterThanOrEqual:
		result.Passed = actualFloat >= rule.Threshold
	case models.OperatorLessThan:
		result.Passed = actualFloat < rule.Threshold
	case models.OperatorLessThanOrEqual:
		result.Passed = actualFloat <= rule.Threshold
	case models.OperatorEqual:
		result.Passed = actualFloat == rule.Threshold
	case models.OperatorNotEqual:
		result.Passed = actualFloat != rule.Threshold
	default:
		result.Message = "Unknown operator"
		return result
	}

	if result.Passed {
		result.Message = "✅ Passed"
	} else {
		result.Message = fmt.Sprintf("❌ Failed: Expected %s %.2f, got %.2f", rule.Operator, rule.Threshold, actualFloat)
	}

	return result
}

// EvaluateAllScorecards evaluates a service against all scorecard definitions
func (e *Evaluator) EvaluateAllScorecards(serviceName string, scorecards []models.ScorecardDefinition, serviceData map[string]interface{}) (*models.ServiceOverallScore, error) {
	overall := &models.ServiceOverallScore{
		ServiceName: serviceName,
		EvaluatedAt: time.Now(),
	}

	var evaluations []models.ScorecardEvaluation
	totalRules := 0
	totalPassed := 0

	for _, scorecard := range scorecards {
		if !scorecard.IsActive {
			continue
		}

		eval, err := e.EvaluateService(serviceName, scorecard, serviceData)
		if err != nil {
			continue
		}

		evaluations = append(evaluations, *eval)
		totalRules += eval.RulesTotal
		totalPassed += eval.RulesPassed
	}

	overall.Scorecards = evaluations
	overall.TotalRules = totalRules
	overall.TotalRulesPassed = totalPassed

	if totalRules > 0 {
		overall.OverallPercentage = float64(totalPassed) / float64(totalRules) * 100
	}

	return overall, nil
}
