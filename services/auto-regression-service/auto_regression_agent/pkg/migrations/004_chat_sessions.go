package migrations

import "database/sql"

// Migration004ChatSessions creates the chat_sessions table
func Migration004ChatSessions() Migration {
	return Migration{
		Version:     4,
		Name:        "create_chat_sessions_table",
		Description: "Creates the chat_sessions table for persistent chat session storage",
		Up: func(db *sql.DB) error {
			query := `
				CREATE TABLE IF NOT EXISTS chat_sessions (
					id VARCHAR(255) PRIMARY KEY,
					workflow_id VARCHAR(255) NOT NULL,
					
					-- Session context (spec, test_suite, results, file_paths)
					context JSONB,
					
					-- Conversation history
					history JSONB DEFAULT '[]'::jsonb,
					
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
				"CREATE INDEX IF NOT EXISTS idx_chat_sessions_workflow_id ON chat_sessions(workflow_id)",
				"CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC)",
			}

			for _, idx := range indexes {
				if _, err := db.Exec(idx); err != nil {
					return err
				}
			}

			// Create trigger for updated_at
			trigger := `
				CREATE TRIGGER update_chat_sessions_updated_at
					BEFORE UPDATE ON chat_sessions
					FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()
			`
			_, err := db.Exec(trigger)
			return err
		},
		Down: func(db *sql.DB) error {
			if _, err := db.Exec("DROP TRIGGER IF EXISTS update_chat_sessions_updated_at ON chat_sessions"); err != nil {
				return err
			}
			if _, err := db.Exec("DROP TABLE IF EXISTS chat_sessions CASCADE"); err != nil {
				return err
			}
			return nil
		},
	}
}

