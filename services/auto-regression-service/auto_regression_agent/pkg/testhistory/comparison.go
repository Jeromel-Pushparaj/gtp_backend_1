package testhistory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// RunComparison represents a comparison between two test runs
type RunComparison struct {
	Run1           *TestRun         `json:"run1"`
	Run2           *TestRun         `json:"run2"`
	PassRateDelta  float64          `json:"pass_rate_delta"`
	TotalDelta     int              `json:"total_delta"`
	PassedDelta    int              `json:"passed_delta"`
	FailedDelta    int              `json:"failed_delta"`
	DurationDelta  int64            `json:"duration_delta_ms"`
	NewFailures    []string         `json:"new_failures,omitempty"`
	FixedTests     []string         `json:"fixed_tests,omitempty"`
	Improvements   []string         `json:"improvements_applied,omitempty"`
	TrendDirection string           `json:"trend_direction"` // "improving", "declining", "stable"
	Summary        string           `json:"summary"`
}

// TestResult represents a single test result for comparison
type TestResult struct {
	TestID   string `json:"test_id"`
	Name     string `json:"name"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
}

// CompareRuns compares two test runs and identifies differences
func (th *TestHistory) CompareRuns(ctx context.Context, runID1, runID2 uuid.UUID) (*RunComparison, error) {
	run1, err := th.GetRun(ctx, runID1)
	if err != nil {
		return nil, fmt.Errorf("failed to get run 1: %w", err)
	}

	run2, err := th.GetRun(ctx, runID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get run 2: %w", err)
	}

	comparison := &RunComparison{
		Run1:          run1,
		Run2:          run2,
		PassRateDelta: run2.PassRate - run1.PassRate,
		TotalDelta:    run2.TotalTests - run1.TotalTests,
		PassedDelta:   run2.Passed - run1.Passed,
		FailedDelta:   run2.Failed - run1.Failed,
		DurationDelta: run2.DurationMs - run1.DurationMs,
		Improvements:  run2.Improvements,
	}

	// Determine trend direction
	if comparison.PassRateDelta > 5 {
		comparison.TrendDirection = "improving"
	} else if comparison.PassRateDelta < -5 {
		comparison.TrendDirection = "declining"
	} else {
		comparison.TrendDirection = "stable"
	}

	// Compare individual test results if available
	comparison.NewFailures, comparison.FixedTests = th.compareTestResults(run1, run2)

	// Generate summary
	comparison.Summary = th.generateComparisonSummary(comparison)

	return comparison, nil
}

// compareTestResults compares individual test results between runs
func (th *TestHistory) compareTestResults(run1, run2 *TestRun) (newFailures, fixedTests []string) {
	// Parse results from both runs
	results1 := th.parseTestResults(run1.Results)
	results2 := th.parseTestResults(run2.Results)

	// Build maps for comparison
	passed1 := make(map[string]bool)
	failed1 := make(map[string]bool)
	for _, r := range results1 {
		if r.Passed {
			passed1[r.TestID] = true
		} else {
			failed1[r.TestID] = true
		}
	}

	// Find new failures and fixed tests
	for _, r := range results2 {
		if !r.Passed && passed1[r.TestID] {
			newFailures = append(newFailures, r.TestID)
		}
		if r.Passed && failed1[r.TestID] {
			fixedTests = append(fixedTests, r.TestID)
		}
	}

	return
}

// parseTestResults parses test results from JSON
func (th *TestHistory) parseTestResults(data json.RawMessage) []TestResult {
	if data == nil {
		return nil
	}

	// Try parsing as array of results
	var wrapper struct {
		Results []TestResult `json:"results"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Results) > 0 {
		return wrapper.Results
	}

	// Try direct array
	var results []TestResult
	json.Unmarshal(data, &results)
	return results
}

// generateComparisonSummary generates a human-readable comparison summary
func (th *TestHistory) generateComparisonSummary(c *RunComparison) string {
	var summary string

	if c.TrendDirection == "improving" {
		summary = fmt.Sprintf("✅ Pass rate improved by %.1f%% (%.1f%% → %.1f%%)",
			c.PassRateDelta, c.Run1.PassRate, c.Run2.PassRate)
	} else if c.TrendDirection == "declining" {
		summary = fmt.Sprintf("⚠️ Pass rate declined by %.1f%% (%.1f%% → %.1f%%)",
			-c.PassRateDelta, c.Run1.PassRate, c.Run2.PassRate)
	} else {
		summary = fmt.Sprintf("➡️ Pass rate stable at %.1f%%", c.Run2.PassRate)
	}

	if len(c.FixedTests) > 0 {
		summary += fmt.Sprintf(" | %d tests fixed", len(c.FixedTests))
	}
	if len(c.NewFailures) > 0 {
		summary += fmt.Sprintf(" | %d new failures", len(c.NewFailures))
	}

	return summary
}

// CompareWithPrevious compares a run with the previous run for the same workflow
func (th *TestHistory) CompareWithPrevious(ctx context.Context, runID uuid.UUID) (*RunComparison, error) {
	// Get the current run
	run, err := th.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Find the previous run
	query := `
		SELECT id FROM test_runs 
		WHERE workflow_id = $1 AND start_time < $2
		ORDER BY start_time DESC
		LIMIT 1
	`

	var prevID uuid.UUID
	err = th.db.QueryRowContext(ctx, query, run.WorkflowID, run.StartTime).Scan(&prevID)
	if err != nil {
		return nil, fmt.Errorf("no previous run found: %w", err)
	}

	return th.CompareRuns(ctx, prevID, runID)
}

