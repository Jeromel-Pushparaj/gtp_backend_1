package scorecards

import (
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
)

// GetAllScorecardDefinitions returns all predefined scorecard definitions
func GetAllScorecardDefinitions() []models.ScorecardDefinition {
	return []models.ScorecardDefinition{
		GetCodeQualityScorecard(),
		GetDORAMetricsScorecard(),
		GetSecurityMaturityScorecard(),
		GetProductionReadinessScorecard(),
		GetServiceHealthScorecard(),
		GetPRMetricsScorecard(),
	}
}

// GetCodeQualityScorecard returns the Code Quality scorecard definition
func GetCodeQualityScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternMetal)

	// Bronze Level Rules
	levels[0].Rules = []models.Rule{
		{Name: "Coverage >= 60%", Description: "Test coverage at least 60%", Property: "coverage", Operator: models.OperatorGreaterThanOrEqual, Threshold: 60, RuleType: models.RuleTypeProperty},
		{Name: "Vulnerabilities <= 10", Description: "No more than 10 vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorLessThanOrEqual, Threshold: 10, RuleType: models.RuleTypeProperty},
		{Name: "Duplications <= 5%", Description: "Code duplication under 5%", Property: "duplicated_lines_density", Operator: models.OperatorLessThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "Has README", Description: "Repository has README file", Property: "has_readme", Operator: models.OperatorEqual, Threshold: 1, RuleType: models.RuleTypeProperty},
	}

	// Silver Level Rules
	levels[1].Rules = []models.Rule{
		{Name: "Coverage >= 80%", Description: "Test coverage at least 80%", Property: "coverage", Operator: models.OperatorGreaterThanOrEqual, Threshold: 80, RuleType: models.RuleTypeProperty},
		{Name: "Code Smells <= 50", Description: "No more than 50 code smells", Property: "code_smells", Operator: models.OperatorLessThanOrEqual, Threshold: 50, RuleType: models.RuleTypeProperty},
		{Name: "Vulnerabilities <= 5", Description: "No more than 5 vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorLessThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
	}

	// Gold Level Rules
	levels[2].Rules = []models.Rule{
		{Name: "Coverage >= 90%", Description: "Test coverage at least 90%", Property: "coverage", Operator: models.OperatorGreaterThanOrEqual, Threshold: 90, RuleType: models.RuleTypeProperty},
		{Name: "Code Smells <= 10", Description: "No more than 10 code smells", Property: "code_smells", Operator: models.OperatorLessThanOrEqual, Threshold: 10, RuleType: models.RuleTypeProperty},
		{Name: "Vulnerabilities == 0", Description: "Zero vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorEqual, Threshold: 0, RuleType: models.RuleTypeProperty},
		{Name: "Duplications <= 3%", Description: "Code duplication under 3%", Property: "duplicated_lines_density", Operator: models.OperatorLessThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "CodeQuality",
		DisplayName:  "Code Quality",
		Category:     models.CategoryCodeQuality,
		Description:  "Evaluates code quality based on test coverage, vulnerabilities, code smells, and duplications",
		LevelPattern: models.PatternMetal,
		Levels:       levels,
		IsActive:     true,
	}
}

// GetDORAMetricsScorecard returns the DORA Metrics scorecard definition
func GetDORAMetricsScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternPerformance)

	// Low Level
	levels[0].Rules = []models.Rule{
		{Name: "Deployment Frequency >= 1/week", Description: "At least 1 deployment per week", Property: "deployment_frequency", Operator: models.OperatorGreaterThanOrEqual, Threshold: 1, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 24 hours", Description: "Mean time to resolve under 24 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 24, RuleType: models.RuleTypeProperty},
	}

	// Medium Level
	levels[1].Rules = []models.Rule{
		{Name: "Deployment Frequency >= 3/week", Description: "At least 3 deployments per week", Property: "deployment_frequency", Operator: models.OperatorGreaterThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 12 hours", Description: "Mean time to resolve under 12 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 12, RuleType: models.RuleTypeProperty},
	}

	// High Level
	levels[2].Rules = []models.Rule{
		{Name: "Deployment Frequency >= 7/week", Description: "At least 7 deployments per week (daily)", Property: "deployment_frequency", Operator: models.OperatorGreaterThanOrEqual, Threshold: 7, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 4 hours", Description: "Mean time to resolve under 4 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 4, RuleType: models.RuleTypeProperty},
	}

	// Elite Level
	levels[3].Rules = []models.Rule{
		{Name: "Deployment Frequency >= 15/week", Description: "Multiple deployments per day", Property: "deployment_frequency", Operator: models.OperatorGreaterThanOrEqual, Threshold: 15, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 1 hour", Description: "Mean time to resolve under 1 hour", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 1, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "DORA_Metrics",
		DisplayName:  "DORA Metrics",
		Category:     models.CategoryDevelopmentVelocity,
		Description:  "Evaluates DevOps performance based on DORA metrics (Deployment Frequency, MTTR)",
		LevelPattern: models.PatternPerformance,
		Levels:       levels,
		IsActive:     true,
	}
}

// GetSecurityMaturityScorecard returns the Security Maturity scorecard definition
func GetSecurityMaturityScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternDescriptive)

	// Basic Level
	levels[0].Rules = []models.Rule{
		{Name: "Vulnerabilities <= 20", Description: "No more than 20 vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorLessThanOrEqual, Threshold: 20, RuleType: models.RuleTypeProperty},
		{Name: "Security Hotspots <= 10", Description: "No more than 10 security hotspots", Property: "security_hotspots", Operator: models.OperatorLessThanOrEqual, Threshold: 10, RuleType: models.RuleTypeProperty},
	}

	// Good Level
	levels[1].Rules = []models.Rule{
		{Name: "Vulnerabilities <= 5", Description: "No more than 5 vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorLessThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "Security Hotspots <= 3", Description: "No more than 3 security hotspots", Property: "security_hotspots", Operator: models.OperatorLessThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
	}

	// Great Level
	levels[2].Rules = []models.Rule{
		{Name: "Vulnerabilities == 0", Description: "Zero vulnerabilities", Property: "vulnerabilities", Operator: models.OperatorEqual, Threshold: 0, RuleType: models.RuleTypeProperty},
		{Name: "Security Hotspots == 0", Description: "Zero security hotspots", Property: "security_hotspots", Operator: models.OperatorEqual, Threshold: 0, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "Security_Maturity",
		DisplayName:  "Security Maturity",
		Category:     models.CategorySecurity,
		Description:  "Evaluates security posture based on vulnerabilities and security hotspots",
		LevelPattern: models.PatternDescriptive,
		Levels:       levels,
		IsActive:     true,
	}
}

// GetProductionReadinessScorecard returns the Production Readiness scorecard definition
func GetProductionReadinessScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternTrafficLight)

	// Red Level (Minimum)
	levels[0].Rules = []models.Rule{
		{Name: "Has README", Description: "Repository has README file", Property: "has_readme", Operator: models.OperatorEqual, Threshold: 1, RuleType: models.RuleTypeProperty},
		{Name: "Active in last 90 days", Description: "Commits in last 90 days", Property: "days_since_last_commit", Operator: models.OperatorLessThanOrEqual, Threshold: 90, RuleType: models.RuleTypeProperty},
	}

	// Yellow Level
	levels[1].Rules = []models.Rule{
		{Name: "Active in last 30 days", Description: "Commits in last 30 days", Property: "days_since_last_commit", Operator: models.OperatorLessThanOrEqual, Threshold: 30, RuleType: models.RuleTypeProperty},
		{Name: "Multiple contributors", Description: "At least 2 contributors", Property: "contributors", Operator: models.OperatorGreaterThanOrEqual, Threshold: 2, RuleType: models.RuleTypeProperty},
	}

	// Orange Level
	levels[2].Rules = []models.Rule{
		{Name: "Active in last 14 days", Description: "Commits in last 14 days", Property: "days_since_last_commit", Operator: models.OperatorLessThanOrEqual, Threshold: 14, RuleType: models.RuleTypeProperty},
		{Name: "Team collaboration", Description: "At least 3 contributors", Property: "contributors", Operator: models.OperatorGreaterThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
		{Name: "Quality gate passed", Description: "SonarCloud quality gate passed", Property: "quality_gate_passed", Operator: models.OperatorEqual, Threshold: 1, RuleType: models.RuleTypeProperty},
	}

	// Green Level
	levels[3].Rules = []models.Rule{
		{Name: "Active in last 7 days", Description: "Commits in last 7 days", Property: "days_since_last_commit", Operator: models.OperatorLessThanOrEqual, Threshold: 7, RuleType: models.RuleTypeProperty},
		{Name: "Strong team", Description: "At least 5 contributors", Property: "contributors", Operator: models.OperatorGreaterThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "High coverage", Description: "Test coverage >= 80%", Property: "coverage", Operator: models.OperatorGreaterThanOrEqual, Threshold: 80, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "Production_Readiness",
		DisplayName:  "Production Readiness",
		Category:     models.CategoryProductionReadiness,
		Description:  "Evaluates production readiness based on freshness, documentation, and team collaboration",
		LevelPattern: models.PatternTrafficLight,
		Levels:       levels,
		IsActive:     true,
	}
}

// GetServiceHealthScorecard returns the Service Health scorecard definition
func GetServiceHealthScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternMetal)

	// Bronze Level
	levels[0].Rules = []models.Rule{
		{Name: "Bugs <= 50", Description: "No more than 50 bugs", Property: "bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 50, RuleType: models.RuleTypeProperty},
		{Name: "Open Bugs <= 20", Description: "No more than 20 open bugs", Property: "open_bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 20, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 48 hours", Description: "Mean time to resolve under 48 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 48, RuleType: models.RuleTypeProperty},
	}

	// Silver Level
	levels[1].Rules = []models.Rule{
		{Name: "Bugs <= 20", Description: "No more than 20 bugs", Property: "bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 20, RuleType: models.RuleTypeProperty},
		{Name: "Open Bugs <= 10", Description: "No more than 10 open bugs", Property: "open_bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 10, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 24 hours", Description: "Mean time to resolve under 24 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 24, RuleType: models.RuleTypeProperty},
	}

	// Gold Level
	levels[2].Rules = []models.Rule{
		{Name: "Bugs <= 5", Description: "No more than 5 bugs", Property: "bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "Open Bugs <= 3", Description: "No more than 3 open bugs", Property: "open_bugs", Operator: models.OperatorLessThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
		{Name: "MTTR < 12 hours", Description: "Mean time to resolve under 12 hours", Property: "mttr", Operator: models.OperatorLessThan, Threshold: 12, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "Service_Health",
		DisplayName:  "Service Health",
		Category:     models.CategoryServiceHealth,
		Description:  "Evaluates service health based on bugs and mean time to resolve",
		LevelPattern: models.PatternMetal,
		Levels:       levels,
		IsActive:     true,
	}
}

// GetPRMetricsScorecard returns the PR Metrics scorecard definition
func GetPRMetricsScorecard() models.ScorecardDefinition {
	levels := models.GetLevelsByPattern(models.PatternMetal)

	// Bronze Level
	levels[0].Rules = []models.Rule{
		{Name: "Merged PRs >= 5", Description: "At least 5 merged PRs", Property: "merged_prs", Operator: models.OperatorGreaterThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "PRs with conflicts <= 30%", Description: "Less than 30% PRs with conflicts", Property: "prs_with_conflicts", Operator: models.OperatorLessThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
		{Name: "Open PRs <= 10", Description: "No more than 10 open PRs", Property: "open_prs", Operator: models.OperatorLessThanOrEqual, Threshold: 10, RuleType: models.RuleTypeProperty},
	}

	// Silver Level
	levels[1].Rules = []models.Rule{
		{Name: "Merged PRs >= 20", Description: "At least 20 merged PRs", Property: "merged_prs", Operator: models.OperatorGreaterThanOrEqual, Threshold: 20, RuleType: models.RuleTypeProperty},
		{Name: "PRs with conflicts <= 10%", Description: "Less than 10% PRs with conflicts", Property: "prs_with_conflicts", Operator: models.OperatorLessThanOrEqual, Threshold: 2, RuleType: models.RuleTypeProperty},
		{Name: "Open PRs <= 5", Description: "No more than 5 open PRs", Property: "open_prs", Operator: models.OperatorLessThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
		{Name: "Multiple contributors", Description: "At least 3 contributors", Property: "contributors", Operator: models.OperatorGreaterThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
	}

	// Gold Level
	levels[2].Rules = []models.Rule{
		{Name: "Merged PRs >= 50", Description: "At least 50 merged PRs", Property: "merged_prs", Operator: models.OperatorGreaterThanOrEqual, Threshold: 50, RuleType: models.RuleTypeProperty},
		{Name: "PRs with conflicts <= 5%", Description: "Less than 5% PRs with conflicts", Property: "prs_with_conflicts", Operator: models.OperatorLessThanOrEqual, Threshold: 1, RuleType: models.RuleTypeProperty},
		{Name: "Open PRs <= 3", Description: "No more than 3 open PRs", Property: "open_prs", Operator: models.OperatorLessThanOrEqual, Threshold: 3, RuleType: models.RuleTypeProperty},
		{Name: "Strong team", Description: "At least 5 contributors", Property: "contributors", Operator: models.OperatorGreaterThanOrEqual, Threshold: 5, RuleType: models.RuleTypeProperty},
	}

	return models.ScorecardDefinition{
		Name:         "PR_Metrics",
		DisplayName:  "PR Metrics",
		Category:     models.CategoryDevelopmentVelocity,
		Description:  "Evaluates PR quality and velocity based on merged PRs, conflicts, and collaboration",
		LevelPattern: models.PatternMetal,
		Levels:       levels,
		IsActive:     true,
	}
}
