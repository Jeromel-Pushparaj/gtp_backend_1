package chat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileSessionStore implements SessionStore using file-based storage
type FileSessionStore struct {
	sessionsDir string
}

// NewFileSessionStore creates a new file-based session store
func NewFileSessionStore(sessionsDir string) *FileSessionStore {
	if sessionsDir == "" {
		sessionsDir = DefaultSessionsDir
	}
	return &FileSessionStore{sessionsDir: sessionsDir}
}

// SaveSession saves a chat session to disk
func (fs *FileSessionStore) SaveSession(session *ChatSession) error {
	if err := os.MkdirAll(fs.sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	filename := fmt.Sprintf("%s.json", session.ID)
	filePath := filepath.Join(fs.sessionsDir, filename)

	data, err := json.MarshalIndent(session.ToSessionData(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a chat session from disk
func (fs *FileSessionStore) LoadSession(sessionID string) (*ChatSession, error) {
	filename := fmt.Sprintf("%s.json", sessionID)
	filePath := filepath.Join(fs.sessionsDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Convert to ChatSession (without runtime fields)
	return &ChatSession{
		ID:         sessionData.ID,
		WorkflowID: sessionData.WorkflowID,
		Context:    sessionData.Context,
		History:    sessionData.History,
		CreatedAt:  sessionData.CreatedAt,
		UpdatedAt:  sessionData.UpdatedAt,
	}, nil
}

// ListSessions lists all sessions, optionally filtered by workflow
func (fs *FileSessionStore) ListSessions(workflowID string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(fs.sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []SessionInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(fs.sessionsDir, entry.Name())
		info, err := fs.getSessionInfo(filePath)
		if err != nil {
			continue
		}

		// Filter by workflow if specified
		if workflowID != "" && info.WorkflowID != workflowID {
			continue
		}

		sessions = append(sessions, info)
	}

	// Sort by updated_at descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// getSessionInfo extracts summary info from a session file
func (fs *FileSessionStore) getSessionInfo(filePath string) (SessionInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return SessionInfo{}, err
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return SessionInfo{}, err
	}

	return SessionInfo{
		ID:           session.ID,
		WorkflowID:   session.WorkflowID,
		MessageCount: len(session.History),
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
		FilePath:     filePath,
	}, nil
}

// DeleteSession deletes a session file
func (fs *FileSessionStore) DeleteSession(sessionID string) error {
	filename := fmt.Sprintf("%s.json", sessionID)
	filePath := filepath.Join(fs.sessionsDir, filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetLatestSession returns the most recent session for a workflow
func (fs *FileSessionStore) GetLatestSession(workflowID string) (*SessionInfo, error) {
	sessions, err := fs.ListSessions(workflowID)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, nil
	}

	return &sessions[0], nil
}

// CleanupOldSessions removes sessions older than maxAge
func (fs *FileSessionStore) CleanupOldSessions(maxAge time.Duration) (int, error) {
	sessions, err := fs.ListSessions("")
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	for _, session := range sessions {
		if session.UpdatedAt.Before(cutoff) {
			if err := fs.DeleteSession(session.ID); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

