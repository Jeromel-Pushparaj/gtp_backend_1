package migrations

import "database/sql"

// Migration001Learnings creates the learnings table with vector support
func Migration001Learnings() Migration {
	return Migration{
		Version:     1,
		Name:        "create_learnings_table",
		Description: "Creates the learnings table with vector embedding support for similarity search",
		Up: func(db *sql.DB) error {
			// Enable pgvector extension
			if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
				return err
			}

			// Create learnings table
			query := `
				CREATE TABLE IF NOT EXISTS learnings (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					
					-- Content and context
					category VARCHAR(100) NOT NULL,
					source_api VARCHAR(255),
					content TEXT NOT NULL,
					context JSONB,
					
					-- Vector embedding (all-MiniLM-L6-v2 = 384 dimensions by default)
					-- Change to vector(1536) for OpenAI text-embedding-3-small
					-- Change to vector(768) for all-mpnet-base-v2
					embedding vector(384),
					
					-- Metadata
					confidence FLOAT DEFAULT 0.5,
					times_applied INT DEFAULT 0,
					last_applied_at TIMESTAMP,
					
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
				"CREATE INDEX IF NOT EXISTS idx_learnings_category ON learnings(category)",
				"CREATE INDEX IF NOT EXISTS idx_learnings_source_api ON learnings(source_api)",
				"CREATE INDEX IF NOT EXISTS idx_learnings_embedding ON learnings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
			}

			for _, idx := range indexes {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}

			// Create trigger for updated_at
			triggerFunc := `
				CREATE OR REPLACE FUNCTION update_updated_at_column()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.updated_at = CURRENT_TIMESTAMP;
					RETURN NEW;
				END;
				$$ language 'plpgsql'
			`
			if _, err := db.Exec(triggerFunc); err != nil {
				return err
			}

			trigger := `
				CREATE TRIGGER update_learnings_updated_at
					BEFORE UPDATE ON learnings
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()
			`
			_, err := db.Exec(trigger)
			return err
		},
		Down: func(db *sql.DB) error {
			if _, err := db.Exec("DROP TRIGGER IF EXISTS update_learnings_updated_at ON learnings"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS learnings CASCADE"); err != nil {
				return err
			}
			return nil
		},
	}
}

