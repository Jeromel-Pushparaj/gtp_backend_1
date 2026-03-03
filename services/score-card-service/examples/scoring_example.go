package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/engine"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/models"
	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

func main() {
	fmt.Println("=== Port.io-Style Scoring System - Example ===")
	fmt.Println()

	// Step 1: Create sample metrics for a service (simulating data from port 8080)
	sampleMetrics := createSampleMetrics()

	// Step 2: Convert metrics to flat map for rule evaluation
	metricsMap := sampleMetrics.ToMap()

	fmt.Println("📊 Sample Metrics for 'authentication' service:")
	printMetrics(metricsMap)

	// Step 3: Get all scorecard definitions
	allScorecards := scorecards.GetAllScorecardDefinitions()

	// Step 4: Create rule engine
	ruleEngine := engine.NewRuleEngine()

	// Step 5: Evaluate all scorecards
	overallScore := ruleEngine.EvaluateAllScorecards(allScorecards, metricsMap, "authentication")

	// Step 6: Display results
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📈 OVERALL SCORE")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Service: %s\n", overallScore.ServiceName)
	fmt.Printf("Overall Score: %.1f%% (%d/%d rules passing)\n",
		overallScore.OverallPercentage,
		overallScore.TotalRulesPassed,
		overallScore.TotalRules)
	fmt.Printf("Evaluated At: %s\n\n", overallScore.EvaluatedAt.Format(time.RFC3339))

	// Step 7: Display individual scorecard results
	fmt.Println("📋 SCORECARD BREAKDOWN")
	fmt.Println(strings.Repeat("=", 80))
	for _, eval := range overallScore.Scorecards {
		displayScorecardResult(eval)
	}

	// Step 8: Display strengths and improvements
	fmt.Println("\n💪 STRENGTHS")
	fmt.Println(strings.Repeat("=", 80))
	if len(overallScore.Strengths) == 0 {
		fmt.Println("No areas with >90% pass rate yet")
	} else {
		for _, strength := range overallScore.Strengths {
			fmt.Printf("✅ %s\n", strength)
		}
	}

	fmt.Println("\n🎯 IMPROVEMENT AREAS")
	fmt.Println(strings.Repeat("=", 80))
	if len(overallScore.ImprovementAreas) == 0 {
		fmt.Println("No critical areas (<60% pass rate)")
	} else {
		for _, improvement := range overallScore.ImprovementAreas {
			fmt.Printf("⚠️  %s\n", improvement)
		}
	}

	// Step 9: Show detailed rule results for one scorecard
	fmt.Println("\n🔍 DETAILED RULE RESULTS - Code Quality Scorecard")
	fmt.Println(strings.Repeat("=", 80))
	for _, eval := range overallScore.Scorecards {
		if eval.ScorecardName == "Code Quality" {
			displayDetailedRules(eval)
			break
		}
	}

	// Step 10: Export as JSON
	fmt.Println("\n📄 JSON OUTPUT")
	fmt.Println(strings.Repeat("=", 80))
	jsonOutput, _ := json.MarshalIndent(overallScore, "", "  ")
	fmt.Println(string(jsonOutput))
}

func createSampleMetrics() *models.CombinedMetrics {
	lastCommit := time.Now().Add(-5 * 24 * time.Hour) // 5 days ago

	return &models.CombinedMetrics{
		ServiceName: "authentication",
		GitHub: &models.GitHubMetrics{
			Repository:        "authentication",
			HasReadme:         true,
			OpenPRs:           3,
			ClosedPRs:         45,
			MergedPRs:         42,
			PRsWithConflicts:  2,
			OpenIssues:        5,
			ClosedIssues:      38,
			TotalCommits:      156,
			CommitsLast90Days: 45, // ~3.5 per week
			Contributors:      4,
			Branches:          8,
			LastCommitDate:    &lastCommit,
		},
		Sonar: &models.SonarMetrics{
			ProjectKey:             "auth-service",
			QualityGateStatus:      "OK",
			Bugs:                   8,
			Vulnerabilities:        3,
			SecurityHotspots:       2,
			CodeSmells:             35,
			Coverage:               78.5,
			DuplicatedLinesDensity: 4.2,
			LinesOfCode:            5420,
			SecurityRating:         "B",
			ReliabilityRating:      "A",
			MaintainabilityRating:  "A",
		},
		Jira: &models.JiraMetrics{
			ProjectKey:           "AUTH",
			OpenBugs:             6,
			ClosedBugs:           24,
			OpenTasks:            12,
			ClosedTasks:          48,
			AvgTimeToResolve:     18.5, // hours (MTTR)
			ActiveSprints:        1,
			CompletedSprints:     8,
			TotalStoryPoints:     120,
			CompletedStoryPoints: 95,
		},
		CollectedAt: time.Now(),
	}
}

func printMetrics(metrics map[string]float64) {
	fmt.Printf("  Coverage: %.1f%%\n", metrics["coverage"])
	fmt.Printf("  Vulnerabilities: %.0f\n", metrics["vulnerabilities"])
	fmt.Printf("  Code Smells: %.0f\n", metrics["code_smells"])
	fmt.Printf("  Bugs: %.0f\n", metrics["bugs"])
	fmt.Printf("  Merged PRs: %.0f\n", metrics["merged_prs"])
	fmt.Printf("  Contributors: %.0f\n", metrics["contributors"])
	fmt.Printf("  Deployment Frequency: %.1f/week\n", metrics["deployment_frequency"])
	fmt.Printf("  MTTR: %.1f hours\n", metrics["mttr"])
	fmt.Printf("  Days Since Last Commit: %.0f\n", metrics["days_since_last_commit"])
}

func displayScorecardResult(eval models.ScorecardEvaluation) {
	icon := getIcon(eval.PassPercentage)
	fmt.Printf("\n%s %s - %s\n", icon, eval.ScorecardName, eval.AchievedLevelName)
	fmt.Printf("   Rules: %d/%d passing (%.1f%%)\n",
		eval.RulesPassed, eval.RulesTotal, eval.PassPercentage)
}

func displayDetailedRules(eval models.ScorecardEvaluation) {
	for _, result := range eval.RuleResults {
		status := "✅"
		if !result.Passed {
			status = "❌"
		}
		fmt.Printf("%s %s\n", status, result.Message)
	}
}

func getIcon(percentage float64) string {
	if percentage >= 90 {
		return "🟢"
	} else if percentage >= 70 {
		return "🟡"
	} else if percentage >= 50 {
		return "🟠"
	}
	return "🔴"
}
