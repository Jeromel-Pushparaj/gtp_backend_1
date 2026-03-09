package engine

import (
	"fmt"
	"time"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
)

// RuleEngine handles rule evaluation logic
type RuleEngine struct{}

// NewRuleEngine creates a new rule engine
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// EvaluateRule evaluates a single rule against metrics
func (re *RuleEngine) EvaluateRule(rule *models.Rule, metrics map[string]float64) *models.RuleResult {
	// Get the actual value from metrics
	actualValue, exists := metrics[rule.Property]
	if !exists {
		// Property not found in metrics
		return &models.RuleResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Passed:        false,
			ActualValue:   0,
			ExpectedValue: rule.Threshold,
			Operator:      string(rule.Operator),
			Message:       fmt.Sprintf("❌ Metric '%s' not found in data", rule.Property),
			EvaluatedAt:   time.Now(),
		}
	}

	// Evaluate the rule
	passed := rule.Evaluate(actualValue)
	message := rule.GetMessage(actualValue, passed)

	return &models.RuleResult{
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		Passed:        passed,
		ActualValue:   actualValue,
		ExpectedValue: rule.Threshold,
		Operator:      string(rule.Operator),
		Message:       message,
		EvaluatedAt:   time.Now(),
	}
}

// EvaluateLevel evaluates all rules in a level
func (re *RuleEngine) EvaluateLevel(level *models.Level, metrics map[string]float64) (bool, []models.RuleResult) {
	if len(level.Rules) == 0 {
		return true, []models.RuleResult{} // No rules means level passes
	}

	allPassed := true
	results := make([]models.RuleResult, 0, len(level.Rules))

	for _, rule := range level.Rules {
		result := re.EvaluateRule(&rule, metrics)
		results = append(results, *result)
		
		if !result.Passed {
			allPassed = false
		}
	}

	return allPassed, results
}

// EvaluateScorecard evaluates a scorecard and determines the achieved level
func (re *RuleEngine) EvaluateScorecard(
	scorecard *models.ScorecardDefinition,
	metrics map[string]float64,
	serviceName string,
) *models.ScorecardEvaluation {
	
	var achievedLevel *models.Level
	var achievedLevelID *int64
	allRuleResults := make([]models.RuleResult, 0)
	totalRules := 0
	passedRules := 0

	// Evaluate levels from lowest to highest
	for i := range scorecard.Levels {
		level := &scorecard.Levels[i]
		totalRules += len(level.Rules)
		
		levelPassed, results := re.EvaluateLevel(level, metrics)
		allRuleResults = append(allRuleResults, results...)
		
		// Count passed rules
		for _, result := range results {
			if result.Passed {
				passedRules++
			}
		}
		
		if levelPassed {
			achievedLevel = level
			achievedLevelID = &level.ID
		} else {
			// Can't achieve higher levels if this one failed
			break
		}
	}

	achievedLevelName := "None"
	if achievedLevel != nil {
		achievedLevelName = achievedLevel.DisplayName
	}

	passPercentage := 0.0
	if totalRules > 0 {
		passPercentage = float64(passedRules) / float64(totalRules) * 100
	}

	return &models.ScorecardEvaluation{
		ServiceName:       serviceName,
		ScorecardID:       scorecard.ID,
		ScorecardName:     scorecard.DisplayName,
		AchievedLevelID:   achievedLevelID,
		AchievedLevelName: achievedLevelName,
		RulesPassed:       passedRules,
		RulesTotal:        totalRules,
		PassPercentage:    passPercentage,
		RuleResults:       allRuleResults,
		EvaluatedAt:       time.Now(),
	}
}

// EvaluateAllScorecards evaluates all scorecards for a service
func (re *RuleEngine) EvaluateAllScorecards(
	scorecards []models.ScorecardDefinition,
	metrics map[string]float64,
	serviceName string,
) *models.ServiceOverallScore {
	
	evaluations := make([]models.ScorecardEvaluation, 0, len(scorecards))
	totalRules := 0
	totalPassed := 0

	for i := range scorecards {
		evaluation := re.EvaluateScorecard(&scorecards[i], metrics, serviceName)
		evaluations = append(evaluations, *evaluation)
		totalRules += evaluation.RulesTotal
		totalPassed += evaluation.RulesPassed
	}

	overallPercentage := 0.0
	if totalRules > 0 {
		overallPercentage = float64(totalPassed) / float64(totalRules) * 100
	}

	strengths, improvements := re.analyzeResults(evaluations)

	return &models.ServiceOverallScore{
		ServiceName:       serviceName,
		OverallPercentage: overallPercentage,
		TotalRulesPassed:  totalPassed,
		TotalRules:        totalRules,
		Scorecards:        evaluations,
		Strengths:         strengths,
		ImprovementAreas:  improvements,
		EvaluatedAt:       time.Now(),
	}
}

// analyzeResults identifies strengths and improvement areas
func (re *RuleEngine) analyzeResults(evaluations []models.ScorecardEvaluation) ([]string, []string) {
	strengths := make([]string, 0)
	improvements := make([]string, 0)

	for _, eval := range evaluations {
		if eval.PassPercentage >= 90 {
			strengths = append(strengths, fmt.Sprintf("%s: %s level achieved", eval.ScorecardName, eval.AchievedLevelName))
		} else if eval.PassPercentage < 60 {
			improvements = append(improvements, fmt.Sprintf("%s: Only %d/%d rules passing", eval.ScorecardName, eval.RulesPassed, eval.RulesTotal))
		}
	}

	return strengths, improvements
}

