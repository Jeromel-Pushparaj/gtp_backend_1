package chat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseSessionStore implements SessionStore using PostgreSQL
type DatabaseSessionStore struct {
	db *sql.DB
}

// DatabaseStoreConfig holds database configuration
type DatabaseStoreConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// NewDatabaseSessionStore creates a new database-backed session store
func NewDatabaseSessionStore(db *sql.DB) *DatabaseSessionStore {
	return &DatabaseSessionStore{db: db}
}

// NewDatabaseSessionStoreWithConfig creates a store with its own connection
func NewDatabaseSessionStoreWithConfig(cfg DatabaseStoreConfig) (*DatabaseSessionStore, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseSessionStore{db: db}, nil
}

// SaveSession saves a chat session to the database
func (ds *DatabaseSessionStore) SaveSession(session *ChatSession) error {
	data := session.ToSessionData()

	contextJSON, err := json.Marshal(data.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	historyJSON, err := json.Marshal(data.History)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	query := `
		INSERT INTO chat_sessions (id, workflow_id, context, history, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			context = EXCLUDED.context,
			history = EXCLUDED.history,
			updated_at = EXCLUDED.updated_at
	`

	_, err = ds.db.Exec(query,
		data.ID,
		data.WorkflowID,
		contextJSON,
		historyJSON,
		data.CreatedAt,
		time.Now(),
	)

	return err
}

// LoadSession loads a chat session from the database
func (ds *DatabaseSessionStore) LoadSession(sessionID string) (*ChatSession, error) {
	query := `
		SELECT id, workflow_id, context, history, created_at, updated_at
		FROM chat_sessions WHERE id = $1
	`

	var (
		id, workflowID        string
		contextJSON, histJSON []byte
		createdAt, updatedAt  time.Time
	)

	err := ds.db.QueryRow(query, sessionID).Scan(
		&id, &workflowID, &contextJSON, &histJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	var context ChatContext
	if contextJSON != nil {
		json.Unmarshal(contextJSON, &context)
	}

	var history []ChatMessage
	if histJSON != nil {
		json.Unmarshal(histJSON, &history)
	}

	return &ChatSession{
		ID:         id,
		WorkflowID: workflowID,
		Context:    &context,
		History:    history,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// ListSessions lists sessions, optionally filtered by workflow
func (ds *DatabaseSessionStore) ListSessions(workflowID string) ([]SessionInfo, error) {
	var query string
	var args []interface{}

	if workflowID != "" {
		query = `
			SELECT id, workflow_id, jsonb_array_length(COALESCE(history, '[]'::jsonb)) as msg_count,
				   created_at, updated_at
			FROM chat_sessions WHERE workflow_id = $1
			ORDER BY updated_at DESC
		`
		args = []interface{}{workflowID}
	} else {
		query = `
			SELECT id, workflow_id, jsonb_array_length(COALESCE(history, '[]'::jsonb)) as msg_count,
				   created_at, updated_at
			FROM chat_sessions ORDER BY updated_at DESC
		`
	}

	rows, err := ds.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var info SessionInfo
		if err := rows.Scan(&info.ID, &info.WorkflowID, &info.MessageCount,
			&info.CreatedAt, &info.UpdatedAt); err != nil {
			continue
		}
		sessions = append(sessions, info)
	}

	return sessions, rows.Err()
}

// DeleteSession deletes a session from the database
func (ds *DatabaseSessionStore) DeleteSession(sessionID string) error {
	result, err := ds.db.Exec("DELETE FROM chat_sessions WHERE id = $1", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// GetLatestSession returns the most recent session for a workflow
func (ds *DatabaseSessionStore) GetLatestSession(workflowID string) (*SessionInfo, error) {
	query := `
		SELECT id, workflow_id, jsonb_array_length(COALESCE(history, '[]'::jsonb)) as msg_count,
			   created_at, updated_at
		FROM chat_sessions
		WHERE workflow_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var info SessionInfo
	err := ds.db.QueryRow(query, workflowID).Scan(
		&info.ID, &info.WorkflowID, &info.MessageCount, &info.CreatedAt, &info.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest session: %w", err)
	}

	return &info, nil
}

// CleanupOldSessions removes sessions older than maxAge
func (ds *DatabaseSessionStore) CleanupOldSessions(maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)

	result, err := ds.db.Exec(
		"DELETE FROM chat_sessions WHERE updated_at < $1",
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// Close closes the database connection
func (ds *DatabaseSessionStore) Close() error {
	return ds.db.Close()
}
