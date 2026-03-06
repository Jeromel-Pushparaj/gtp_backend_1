package migrations

import "database/sql"

// Migration002FailurePatterns creates the failure_patterns table
func Migration002FailurePatterns() Migration {
	return Migration{
		Version:     2,
		Name:        "create_failure_patterns_table",
		Description: "Creates the failure_patterns table for storing failure signatures and fixes",
		Up: func(db *sql.DB) error {
			query := `
				CREATE TABLE IF NOT EXISTS failure_patterns (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					
					-- Failure identification
					failure_type VARCHAR(100) NOT NULL,
					error_signature TEXT NOT NULL,
					error_code VARCHAR(50),
					
					-- Context about where failure occurred
					endpoint_pattern VARCHAR(500),
					http_method VARCHAR(10),
					api_source VARCHAR(255),
					
					-- Solution
					root_cause TEXT,
					fix_description TEXT NOT NULL,
					fix_payload JSONB,
					
					-- Vector embedding for similarity search
					embedding vector(768),
					
					-- Metadata
					times_encountered INT DEFAULT 1,
					times_fixed INT DEFAULT 0,
					fix_success_rate FLOAT DEFAULT 0.0,
					
					-- Timestamps
					first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			if _, err := db.Exec(query); err != nil {
				return err
			}

			// Create indexes
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_failure_patterns_type ON failure_patterns(failure_type)",
				"CREATE INDEX IF NOT EXISTS idx_failure_patterns_error_code ON failure_patterns(error_code)",
				"CREATE INDEX IF NOT EXISTS idx_failure_patterns_embedding ON failure_patterns USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
			}

			for _, idx := range indexes {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}

			// Create trigger for updated_at
			trigger := `
				CREATE TRIGGER update_failure_patterns_updated_at
					BEFORE UPDATE ON failure_patterns
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()
			`
			_, err := db.Exec(trigger)
			return err
		},
		Down: func(db *sql.DB) error {
			if _, err := db.Exec("DROP TRIGGER IF EXISTS update_failure_patterns_updated_at ON failure_patterns"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS failure_patterns CASCADE"); err != nil {
				return err
			}
			return nil
		},
	}
}

