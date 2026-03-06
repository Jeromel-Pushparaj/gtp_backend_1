// Package testhistory provides test run tracking and historical analysis
package testhistory

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TestRun represents a single test execution
type TestRun struct {
	ID             uuid.UUID       `json:"id"`
	WorkflowID     string          `json:"workflow_id"`
	SuiteID        string          `json:"suite_id,omitempty"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time,omitempty"`
	DurationMs     int64           `json:"duration_ms,omitempty"`
	TotalTests     int             `json:"total_tests"`
	Passed         int             `json:"passed"`
	Failed         int             `json:"failed"`
	Skipped        int             `json:"skipped"`
	PassRate       float64         `json:"pass_rate"`
	Results        json.RawMessage `json:"results,omitempty"`
	Improvements   []string        `json:"improvements,omitempty"`
	FailureSummary json.RawMessage `json:"failure_summary,omitempty"`
	Environment    string          `json:"environment,omitempty"`
	BaseURL        string          `json:"base_url,omitempty"`
	TriggeredBy    string          `json:"triggered_by,omitempty"`
	Notes          string          `json:"notes,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// TestHistory provides test run tracking and analysis
type TestHistory struct {
	db *sql.DB
}

// NewTestHistory creates a new test history tracker
func NewTestHistory(db *sql.DB) *TestHistory {
	return &TestHistory{db: db}
}

// RecordRun records a new test run
func (th *TestHistory) RecordRun(ctx context.Context, run *TestRun) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}

	// Calculate pass rate
	if run.TotalTests > 0 {
		run.PassRate = float64(run.Passed) / float64(run.TotalTests) * 100
	}

	// Calculate duration
	if !run.EndTime.IsZero() {
		run.DurationMs = run.EndTime.Sub(run.StartTime).Milliseconds()
	}

	// Marshal JSON fields
	improvementsJSON, _ := json.Marshal(run.Improvements)

	query := `
		INSERT INTO test_runs (
			id, workflow_id, suite_id, start_time, end_time, duration_ms,
			total_tests, passed, failed, skipped, pass_rate,
			results, improvements, failure_summary,
			environment, base_url, triggered_by, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	var endTime interface{}
	if !run.EndTime.IsZero() {
		endTime = run.EndTime
	}

	_, err := th.db.ExecContext(ctx, query,
		run.ID, run.WorkflowID, run.SuiteID, run.StartTime, endTime, run.DurationMs,
		run.TotalTests, run.Passed, run.Failed, run.Skipped, run.PassRate,
		run.Results, improvementsJSON, run.FailureSummary,
		run.Environment, run.BaseURL, run.TriggeredBy, run.Notes,
	)

	return err
}

// GetRun retrieves a specific test run
func (th *TestHistory) GetRun(ctx context.Context, runID uuid.UUID) (*TestRun, error) {
	query := `
		SELECT id, workflow_id, suite_id, start_time, end_time, duration_ms,
			   total_tests, passed, failed, skipped, pass_rate,
			   results, improvements, failure_summary,
			   environment, base_url, triggered_by, notes, created_at
		FROM test_runs WHERE id = $1
	`

	var run TestRun
	var endTime sql.NullTime
	var improvementsJSON []byte

	err := th.db.QueryRowContext(ctx, query, runID).Scan(
		&run.ID, &run.WorkflowID, &run.SuiteID, &run.StartTime, &endTime, &run.DurationMs,
		&run.TotalTests, &run.Passed, &run.Failed, &run.Skipped, &run.PassRate,
		&run.Results, &improvementsJSON, &run.FailureSummary,
		&run.Environment, &run.BaseURL, &run.TriggeredBy, &run.Notes, &run.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		run.EndTime = endTime.Time
	}
	if improvementsJSON != nil {
		json.Unmarshal(improvementsJSON, &run.Improvements)
	}

	return &run, nil
}

// GetRunHistory retrieves test run history for a workflow
func (th *TestHistory) GetRunHistory(ctx context.Context, workflowID string, limit int) ([]TestRun, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, workflow_id, suite_id, start_time, end_time, duration_ms,
			   total_tests, passed, failed, skipped, pass_rate,
			   improvements, environment, triggered_by, created_at
		FROM test_runs 
		WHERE workflow_id = $1
		ORDER BY start_time DESC
		LIMIT $2
	`

	rows, err := th.db.QueryContext(ctx, query, workflowID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []TestRun
	for rows.Next() {
		var run TestRun
		var endTime sql.NullTime
		var improvementsJSON []byte

		if err := rows.Scan(
			&run.ID, &run.WorkflowID, &run.SuiteID, &run.StartTime, &endTime, &run.DurationMs,
			&run.TotalTests, &run.Passed, &run.Failed, &run.Skipped, &run.PassRate,
			&improvementsJSON, &run.Environment, &run.TriggeredBy, &run.CreatedAt,
		); err != nil {
			return nil, err
		}

		if endTime.Valid {
			run.EndTime = endTime.Time
		}
		if improvementsJSON != nil {
			json.Unmarshal(improvementsJSON, &run.Improvements)
		}

		runs = append(runs, run)
	}

	return runs, rows.Err()
}
