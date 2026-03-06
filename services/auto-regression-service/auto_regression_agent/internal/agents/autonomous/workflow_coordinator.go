package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

// WorkflowResult contains the results of a complete test workflow
type WorkflowResult struct {
	WorkflowID      string                `json:"workflow_id"`
	SpecID          string                `json:"spec_id"`
	ExecutionResult *ExecutionResult      `json:"execution_result,omitempty"`
	SuiteResult     *SuiteExecutionResult `json:"suite_result,omitempty"`
	Recommendations []Recommendation      `json:"recommendations"`
	LearnedPatterns int                   `json:"learned_patterns"`
	StartTime       time.Time             `json:"start_time"`
	EndTime         time.Time             `json:"end_time"`
	Duration        time.Duration         `json:"duration"`
}

// Recommendation represents a suggested improvement for the test suite
type Recommendation struct {
	ID          string                 `json:"id"`
	Priority    string                 `json:"priority"` // high, medium, low
	Category    string                 `json:"category"` // coverage, reliability, performance, security
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	TestIndex   int                    `json:"test_index,omitempty"` // Index of related test, -1 if general
	Context     map[string]interface{} `json:"context,omitempty"`
}

// WorkflowCoordinator coordinates the complete test workflow including
// execution, analysis, learning, and recommendations
type WorkflowCoordinator struct {
	llmClient     *llm.Client
	eventBus      *events.Bus
	vectorStore   *vectordb.Store
	feedbackAgent *FeedbackAgent
	baseURL       string
}

// NewWorkflowCoordinator creates a new workflow coordinator
func NewWorkflowCoordinator(
	llmClient *llm.Client,
	eventBus *events.Bus,
	vectorStore *vectordb.Store,
	baseURL string,
) *WorkflowCoordinator {
	// Create feedback agent for learning
	var feedbackAgent *FeedbackAgent
	if llmClient != nil {
		feedbackAgent = NewFeedbackAgent(llmClient, eventBus, nil, nil, vectorStore)
	}

	return &WorkflowCoordinator{
		llmClient:     llmClient,
		eventBus:      eventBus,
		vectorStore:   vectorStore,
		feedbackAgent: feedbackAgent,
		baseURL:       baseURL,
	}
}

// RunTestWorkflow runs the complete test workflow with learning and recommendations
func (wc *WorkflowCoordinator) RunTestWorkflow(ctx context.Context, suitePath string) (*WorkflowResult, error) {
	startTime := time.Now()
	log.Printf("🚀 Starting test workflow for suite: %s", suitePath)

	// Load the test suite
	suite, err := LoadTestSuite(suitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load test suite: %w", err)
	}

	// Create planned executor
	executor := NewPlannedExecutor(wc.llmClient, wc.baseURL)
	if suite.BaseURL != "" {
		executor.baseURL = suite.BaseURL
	}

	// Execute the suite
	suiteResult, err := executor.RunSavedSuite(ctx, suitePath)
	if err != nil {
		return nil, fmt.Errorf("test execution failed: %w", err)
	}

	// Save results
	resultsPath := fmt.Sprintf("./output/results/%s-test-results.json", suite.ID)
	if err := os.MkdirAll("./output/results", 0755); err == nil {
		if data, err := json.MarshalIndent(suiteResult, "", "  "); err == nil {
			os.WriteFile(resultsPath, data, 0644)
		}
	}

	// Analyze results and generate recommendations
	recommendations := wc.analyzeResultsAndRecommend(ctx, suite, suiteResult)

	// Store learnings from the execution
	learnedCount := wc.storeLearnedPatterns(ctx, suite, suiteResult)

	endTime := time.Now()
	result := &WorkflowResult{
		WorkflowID:      suite.ID,
		SpecID:          suite.SpecID,
		SuiteResult:     suiteResult,
		Recommendations: recommendations,
		LearnedPatterns: learnedCount,
		StartTime:       startTime,
		EndTime:         endTime,
		Duration:        endTime.Sub(startTime),
	}

	log.Printf("✅ Workflow complete: %d tests, %d recommendations, %d patterns learned",
		len(suite.Tests), len(recommendations), learnedCount)

	return result, nil
}

// analyzeResultsAndRecommend analyzes test results and generates recommendations
func (wc *WorkflowCoordinator) analyzeResultsAndRecommend(
	ctx context.Context,
	suite *TestSuite,
	result *SuiteExecutionResult,
) []Recommendation {
	var recommendations []Recommendation
	recID := 0

	// Analyze coverage gaps
	recommendations = append(recommendations, wc.analyzeCoverageGaps(suite, result, &recID)...)

	// Analyze failure patterns
	recommendations = append(recommendations, wc.analyzeFailurePatterns(result, &recID)...)

	// Analyze performance issues
	recommendations = append(recommendations, wc.analyzePerformance(result, &recID)...)

	// Search for relevant recommendations from memory
	if wc.vectorStore != nil {
		recommendations = append(recommendations, wc.searchMemoryForRecommendations(ctx, suite, result, &recID)...)
	}

	return recommendations
}

// analyzeCoverageGaps identifies missing test coverage
func (wc *WorkflowCoordinator) analyzeCoverageGaps(suite *TestSuite, result *SuiteExecutionResult, recID *int) []Recommendation {
	var recommendations []Recommendation

	// Check for missing HTTP methods coverage
	methods := make(map[string]int)
	categories := make(map[string]int)

	for _, test := range suite.Tests {
		methods[test.Method]++
		categories[test.Category]++
	}

	// Recommend adding DELETE tests if missing
	if methods["DELETE"] == 0 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "medium",
			Category:    "coverage",
			Title:       "Missing DELETE endpoint tests",
			Description: "No DELETE endpoint tests found. Consider adding tests for resource deletion to ensure proper cleanup functionality.",
			TestIndex:   -1,
		})
	}

	// Check for negative test coverage
	if categories["negative"] == 0 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "high",
			Category:    "coverage",
			Title:       "Missing negative tests",
			Description: "No negative tests found. Add tests for invalid inputs, missing required fields, and error conditions.",
			TestIndex:   -1,
		})
	}

	// Check for boundary test coverage
	if categories["boundary"] == 0 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "medium",
			Category:    "coverage",
			Title:       "Missing boundary tests",
			Description: "No boundary tests found. Add tests for edge cases like empty values, max lengths, and numeric limits.",
			TestIndex:   -1,
		})
	}

	// Check for security tests
	if categories["security"] == 0 && suite.Statistics.TotalTests > 5 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "high",
			Category:    "security",
			Title:       "Missing security tests",
			Description: "No security tests found. Consider adding tests for authentication, authorization, and input validation.",
			TestIndex:   -1,
		})
	}

	return recommendations
}

// analyzeFailurePatterns identifies patterns in test failures
func (wc *WorkflowCoordinator) analyzeFailurePatterns(result *SuiteExecutionResult, recID *int) []Recommendation {
	var recommendations []Recommendation

	failedTests := make([]PersistedTestResult, 0)
	skippedTests := make([]PersistedTestResult, 0)
	timeoutCount := 0

	for _, r := range result.Results {
		if r.Skipped {
			skippedTests = append(skippedTests, r)
		} else if !r.Passed {
			failedTests = append(failedTests, r)
		}
		if r.ResponseTime > 10*time.Second {
			timeoutCount++
		}
	}

	// Analyze authentication failures
	authFailures := 0
	for _, r := range failedTests {
		if r.StatusCode == 401 || r.StatusCode == 403 {
			authFailures++
		}
	}

	if authFailures > 0 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "high",
			Category:    "reliability",
			Title:       fmt.Sprintf("%d authentication failures detected", authFailures),
			Description: "Tests are failing due to authentication issues. Check that API credentials are configured correctly and tokens are not expired.",
			TestIndex:   -1,
			Context: map[string]interface{}{
				"auth_failure_count": authFailures,
			},
		})
	}

	// High skip rate
	if len(skippedTests) > len(result.Results)/3 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "medium",
			Category:    "reliability",
			Title:       fmt.Sprintf("%d tests skipped due to missing context", len(skippedTests)),
			Description: "Many tests are being skipped because context from previous tests is not available. Review test ordering and context capture settings.",
			TestIndex:   -1,
		})
	}

	// Timeout issues
	if timeoutCount > 0 {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "medium",
			Category:    "performance",
			Title:       fmt.Sprintf("%d tests experiencing timeouts", timeoutCount),
			Description: "Some tests are taking too long to complete. Consider checking rate limiting or API performance issues.",
			TestIndex:   -1,
		})
	}

	return recommendations
}

// analyzePerformance identifies performance-related issues
func (wc *WorkflowCoordinator) analyzePerformance(result *SuiteExecutionResult, recID *int) []Recommendation {
	var recommendations []Recommendation

	if result.Summary.TotalTime > 5*time.Minute {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "low",
			Category:    "performance",
			Title:       "Long test suite execution time",
			Description: fmt.Sprintf("The test suite took %v to complete. Consider parallelizing independent tests.", result.Summary.TotalTime),
			TestIndex:   -1,
		})
	}

	return recommendations
}

// searchMemoryForRecommendations searches vector DB for relevant past learnings
func (wc *WorkflowCoordinator) searchMemoryForRecommendations(
	ctx context.Context,
	suite *TestSuite,
	result *SuiteExecutionResult,
	recID *int,
) []Recommendation {
	var recommendations []Recommendation

	// Build a query based on failures
	failureDescriptions := make([]string, 0)
	for _, r := range result.Results {
		if !r.Passed && !r.Skipped && r.Error != "" {
			failureDescriptions = append(failureDescriptions, r.Error)
		}
	}

	if len(failureDescriptions) == 0 {
		return recommendations
	}

	// Search for similar failures
	query := fmt.Sprintf("API test failures: %v", failureDescriptions[:min(3, len(failureDescriptions))])
	opts := vectordb.SearchOptions{Limit: 3, MinScore: 0.5}
	results, err := wc.vectorStore.SearchSimilarFailures(ctx, query, opts)
	if err != nil {
		log.Printf("Warning: failed to search failure patterns: %v", err)
		return recommendations
	}

	for _, r := range results {
		*recID++
		recommendations = append(recommendations, Recommendation{
			ID:          fmt.Sprintf("rec-%d", *recID),
			Priority:    "medium",
			Category:    "reliability",
			Title:       fmt.Sprintf("Known fix for %s", r.FailureType),
			Description: r.FixDescription,
			TestIndex:   -1,
			Context: map[string]interface{}{
				"failure_pattern_id": r.ID.String(),
				"success_rate":       r.FixSuccessRate,
			},
		})
	}

	return recommendations
}

// storeLearnedPatterns stores patterns from successful and failed tests
func (wc *WorkflowCoordinator) storeLearnedPatterns(
	ctx context.Context,
	suite *TestSuite,
	result *SuiteExecutionResult,
) int {
	if wc.vectorStore == nil {
		return 0
	}

	learnedCount := 0

	// Store patterns from successful tests
	for _, r := range result.Results {
		if r.Passed && r.Test.Category == "positive" {
			// Store successful pattern
			learning := &vectordb.Learning{
				Category:   "api_pattern",
				Content:    fmt.Sprintf("Successful %s %s returned %d", r.Test.Method, r.Test.Path, r.StatusCode),
				SourceAPI:  suite.SpecID,
				Confidence: 0.8,
				Context: map[string]interface{}{
					"method":   r.Test.Method,
					"path":     r.Test.Path,
					"category": r.Test.Category,
				},
			}
			if err := wc.vectorStore.StoreLearning(ctx, learning); err == nil {
				learnedCount++
			}
		}

		// Store failure patterns
		if !r.Passed && !r.Skipped && r.Error != "" {
			failureType := "validation_error"
			if r.StatusCode == 401 || r.StatusCode == 403 {
				failureType = "auth_failure"
			} else if r.StatusCode >= 500 {
				failureType = "server_error"
			} else if r.StatusCode == 404 {
				failureType = "not_found"
			}

			pattern := &vectordb.FailurePattern{
				FailureType:     failureType,
				ErrorSignature:  r.Error,
				ErrorCode:       fmt.Sprintf("%d", r.StatusCode),
				EndpointPattern: r.Test.Path,
				HTTPMethod:      r.Test.Method,
			}
			if err := wc.vectorStore.StoreFailurePattern(ctx, pattern); err == nil {
				learnedCount++
			}
		}
	}

	return learnedCount
}

// RunFromSpec runs a complete workflow from an OpenAPI spec path
func (wc *WorkflowCoordinator) RunFromSpec(ctx context.Context, specPath string) (*WorkflowResult, error) {
	// This would coordinate with discovery and payload agents
	// For now, just return an error directing to use RunTestWorkflow
	return nil, fmt.Errorf("RunFromSpec not yet implemented - use RunTestWorkflow with an existing suite")
}
