package migrations

import "database/sql"

// Migration005TestRuns creates the test_runs table
func Migration005TestRuns() Migration {
	return Migration{
		Version:     5,
		Name:        "create_test_runs_table",
		Description: "Creates the test_runs table for tracking test execution history",
		Up: func(db *sql.DB) error {
			query := `
				CREATE TABLE IF NOT EXISTS test_runs (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					workflow_id VARCHAR(255) NOT NULL,
					suite_id VARCHAR(255),
					
					-- Execution timing
					start_time TIMESTAMP NOT NULL,
					end_time TIMESTAMP,
					duration_ms BIGINT,
					
					-- Test counts
					total_tests INT DEFAULT 0,
					passed INT DEFAULT 0,
					failed INT DEFAULT 0,
					skipped INT DEFAULT 0,
					
					-- Pass rate (calculated)
					pass_rate FLOAT DEFAULT 0.0,
					
					-- Full results JSON
					results JSONB,
					
					-- Improvements applied since last run
					improvements JSONB DEFAULT '[]'::jsonb,
					
					-- Failure summary for quick access
					failure_summary JSONB,
					
					-- Environment info
					environment VARCHAR(100),
					base_url VARCHAR(500),
					
					-- Metadata
					triggered_by VARCHAR(100),
					notes TEXT,
					
					-- Timestamps
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			if _, err := db.Exec(query); err != nil {
				return err
			}

			// Create indexes
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_test_runs_workflow_id ON test_runs(workflow_id)",
				"CREATE INDEX IF NOT EXISTS idx_test_runs_suite_id ON test_runs(suite_id)",
				"CREATE INDEX IF NOT EXISTS idx_test_runs_start_time ON test_runs(start_time DESC)",
				"CREATE INDEX IF NOT EXISTS idx_test_runs_workflow_time ON test_runs(workflow_id, start_time DESC)",
			}

			for _, idx := range indexes {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}

			return nil
		},
		Down: func(db *sql.DB) error {
			_, err := db.Exec("DROP TABLE IF EXISTS test_runs CASCADE")
			return err
		},
	}
}

