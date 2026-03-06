package migrations

import "database/sql"

// Migration003Strategies creates the successful_strategies table
func Migration003Strategies() Migration {
	return Migration{
		Version:     3,
		Name:        "create_strategies_table",
		Description: "Creates the successful_strategies table for storing test strategies that worked well",
		Up: func(db *sql.DB) error {
			query := `
				CREATE TABLE IF NOT EXISTS successful_strategies (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					
					-- Strategy identification
					strategy_type VARCHAR(100) NOT NULL,
					strategy_name VARCHAR(255) NOT NULL,
					description TEXT NOT NULL,
					
					-- Strategy details
					strategy_content JSONB NOT NULL,
					applicable_patterns JSONB,
					
					-- Context
					api_source VARCHAR(255),
					endpoint_patterns TEXT[],
					
					-- Vector embedding for similarity search
					embedding vector(768),
					
					-- Effectiveness metrics
					success_rate FLOAT DEFAULT 0.0,
					times_used INT DEFAULT 0,
					avg_coverage FLOAT DEFAULT 0.0,
					
					-- Timestamps
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			if _, err := db.Exec(query); err != nil {
				return err
			}

			// Create indexes
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_strategies_type ON successful_strategies(strategy_type)",
				"CREATE INDEX IF NOT EXISTS idx_strategies_embedding ON successful_strategies USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
			}

			for _, idx := range indexes {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}

			// Create trigger for updated_at
			trigger := `
				CREATE TRIGGER update_strategies_updated_at
					BEFORE UPDATE ON successful_strategies
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()
			`
			_, err := db.Exec(trigger)
			return err
		},
		Down: func(db *sql.DB) error {
			if _, err := db.Exec("DROP TRIGGER IF EXISTS update_strategies_updated_at ON successful_strategies"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS successful_strategies CASCADE"); err != nil {
				return err
			}
			return nil
		},
	}
}

