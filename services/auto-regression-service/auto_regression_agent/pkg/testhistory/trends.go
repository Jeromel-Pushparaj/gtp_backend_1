package testhistory

import (
	"context"
	"fmt"
	"time"
)

// TrendReport contains trend analysis for test runs
type TrendReport struct {
	WorkflowID       string          `json:"workflow_id"`
	Period           string          `json:"period"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	TotalRuns        int             `json:"total_runs"`
	AveragePassRate  float64         `json:"average_pass_rate"`
	MinPassRate      float64         `json:"min_pass_rate"`
	MaxPassRate      float64         `json:"max_pass_rate"`
	PassRateTrend    float64         `json:"pass_rate_trend"` // Positive = improving
	TotalTests       int             `json:"total_tests_executed"`
	TotalPassed      int             `json:"total_passed"`
	TotalFailed      int             `json:"total_failed"`
	AverageDuration  int64           `json:"average_duration_ms"`
	CommonFailures   []FailureCount  `json:"common_failures,omitempty"`
	DailyStats       []DailyStat     `json:"daily_stats,omitempty"`
	TrendDirection   string          `json:"trend_direction"` // "improving", "declining", "stable"
	Summary          string          `json:"summary"`
}

// FailureCount tracks failure frequency
type FailureCount struct {
	TestID     string `json:"test_id"`
	Name       string `json:"name,omitempty"`
	FailCount  int    `json:"fail_count"`
	LastFailed string `json:"last_failed"`
}

// DailyStat contains daily test execution statistics
type DailyStat struct {
	Date      string  `json:"date"`
	RunCount  int     `json:"run_count"`
	PassRate  float64 `json:"pass_rate"`
	Passed    int     `json:"passed"`
	Failed    int     `json:"failed"`
	TotalMs   int64   `json:"total_duration_ms"`
}

// GetTrends analyzes test run trends over a specified period
func (th *TestHistory) GetTrends(ctx context.Context, workflowID string, days int) (*TrendReport, error) {
	if days <= 0 {
		days = 30
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	report := &TrendReport{
		WorkflowID: workflowID,
		Period:     fmt.Sprintf("%d days", days),
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Get aggregate statistics
	aggQuery := `
		SELECT 
			COUNT(*) as run_count,
			COALESCE(AVG(pass_rate), 0) as avg_pass_rate,
			COALESCE(MIN(pass_rate), 0) as min_pass_rate,
			COALESCE(MAX(pass_rate), 0) as max_pass_rate,
			COALESCE(SUM(total_tests), 0) as total_tests,
			COALESCE(SUM(passed), 0) as total_passed,
			COALESCE(SUM(failed), 0) as total_failed,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		FROM test_runs
		WHERE workflow_id = $1 AND start_time >= $2 AND start_time <= $3
	`

	err := th.db.QueryRowContext(ctx, aggQuery, workflowID, startDate, endDate).Scan(
		&report.TotalRuns,
		&report.AveragePassRate,
		&report.MinPassRate,
		&report.MaxPassRate,
		&report.TotalTests,
		&report.TotalPassed,
		&report.TotalFailed,
		&report.AverageDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate stats: %w", err)
	}

	// Get daily statistics
	dailyQuery := `
		SELECT 
			DATE(start_time) as run_date,
			COUNT(*) as run_count,
			AVG(pass_rate) as pass_rate,
			SUM(passed) as passed,
			SUM(failed) as failed,
			SUM(duration_ms) as total_duration
		FROM test_runs
		WHERE workflow_id = $1 AND start_time >= $2 AND start_time <= $3
		GROUP BY DATE(start_time)
		ORDER BY run_date ASC
	`

	rows, err := th.db.QueryContext(ctx, dailyQuery, workflowID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat DailyStat
		var runDate time.Time
		if err := rows.Scan(&runDate, &stat.RunCount, &stat.PassRate, &stat.Passed, &stat.Failed, &stat.TotalMs); err != nil {
			continue
		}
		stat.Date = runDate.Format("2006-01-02")
		report.DailyStats = append(report.DailyStats, stat)
	}

	// Calculate trend direction
	report.calculateTrend()

	return report, nil
}

// calculateTrend calculates the pass rate trend direction
func (r *TrendReport) calculateTrend() {
	if len(r.DailyStats) < 2 {
		r.TrendDirection = "stable"
		r.Summary = "Not enough data to determine trend"
		return
	}

	// Compare first half vs second half average
	mid := len(r.DailyStats) / 2
	var firstHalfSum, secondHalfSum float64
	var firstHalfCount, secondHalfCount int

	for i, stat := range r.DailyStats {
		if i < mid {
			firstHalfSum += stat.PassRate
			firstHalfCount++
		} else {
			secondHalfSum += stat.PassRate
			secondHalfCount++
		}
	}

	firstHalfAvg := firstHalfSum / float64(firstHalfCount)
	secondHalfAvg := secondHalfSum / float64(secondHalfCount)
	r.PassRateTrend = secondHalfAvg - firstHalfAvg

	if r.PassRateTrend > 5 {
		r.TrendDirection = "improving"
		r.Summary = fmt.Sprintf("📈 Pass rate trending up: %.1f%% → %.1f%% (+%.1f%%)",
			firstHalfAvg, secondHalfAvg, r.PassRateTrend)
	} else if r.PassRateTrend < -5 {
		r.TrendDirection = "declining"
		r.Summary = fmt.Sprintf("📉 Pass rate trending down: %.1f%% → %.1f%% (%.1f%%)",
			firstHalfAvg, secondHalfAvg, r.PassRateTrend)
	} else {
		r.TrendDirection = "stable"
		r.Summary = fmt.Sprintf("➡️ Pass rate stable around %.1f%%", r.AveragePassRate)
	}
}

// GetLatestRuns returns the most recent test runs
func (th *TestHistory) GetLatestRuns(ctx context.Context, workflowID string, limit int) ([]TestRun, error) {
	return th.GetRunHistory(ctx, workflowID, limit)
}

