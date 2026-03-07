package handlers

import (
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// getScorecardByName returns a scorecard definition by name
func getScorecardByName(name string) (models.ScorecardDefinition, bool) {
	switch name {
	case "CodeQuality":
		return scorecards.GetCodeQualityScorecard(), true
	case "Security_Maturity":
		return scorecards.GetSecurityMaturityScorecard(), true
	case "Production_Readiness":
		return scorecards.GetProductionReadinessScorecard(), true
	case "Service_Health":
		return scorecards.GetServiceHealthScorecard(), true
	case "PR_Metrics":
		return scorecards.GetPRMetricsScorecard(), true
	default:
		return models.ScorecardDefinition{}, false
	}
}

// convertMetricsToServiceData converts CombinedMetrics to service data map
func convertMetricsToServiceData(combinedMetrics *models.CombinedMetrics) map[string]interface{} {
	metricsMap := combinedMetrics.ToMap()
	serviceData := make(map[string]interface{})
	for k, v := range metricsMap {
		serviceData[k] = v
	}
	return serviceData
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

// filterAchievedLevelRules filters rule results to show only the rules from the achieved level
func (h *ScorecardV2Handler) filterAchievedLevelRules(evaluation *models.ScorecardEvaluation, scorecard models.ScorecardDefinition) []models.RuleResult {
	// Find the achieved level in the scorecard definition
	var achievedLevelRules []string
	for _, level := range scorecard.Levels {
		// Compare with DisplayName since evaluation uses DisplayName (e.g., "⚪ Starter")
		if level.DisplayName == evaluation.AchievedLevelName {
			// Get all rule names from this level
			for _, rule := range level.Rules {
				achievedLevelRules = append(achievedLevelRules, rule.Name)
			}
			break
		}
	}

	// Filter the rule results to only include rules from the achieved level
	var filteredResults []models.RuleResult
	for _, result := range evaluation.RuleResults {
		for _, ruleName := range achievedLevelRules {
			if result.RuleName == ruleName {
				filteredResults = append(filteredResults, result)
				break
			}
		}
	}

	return filteredResults
}

