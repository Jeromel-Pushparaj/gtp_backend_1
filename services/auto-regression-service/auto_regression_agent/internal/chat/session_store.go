package chat

import (
	"time"
)

// SessionStore defines the interface for chat session persistence
type SessionStore interface {
	// SaveSession saves a chat session
	SaveSession(session *ChatSession) error

	// LoadSession loads a chat session by ID
	LoadSession(sessionID string) (*ChatSession, error)

	// ListSessions lists all sessions for a workflow
	ListSessions(workflowID string) ([]SessionInfo, error)

	// DeleteSession deletes a session by ID
	DeleteSession(sessionID string) error

	// GetLatestSession gets the most recent session for a workflow
	GetLatestSession(workflowID string) (*SessionInfo, error)

	// CleanupOldSessions removes sessions older than the specified duration
	CleanupOldSessions(maxAge time.Duration) (int, error)
}

// SessionData represents the serializable portion of a chat session
type SessionData struct {
	ID         string        `json:"id"`
	WorkflowID string        `json:"workflow_id"`
	Context    *ChatContext  `json:"context"`
	History    []ChatMessage `json:"history"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// ToSessionData converts a ChatSession to serializable SessionData
func (s *ChatSession) ToSessionData() *SessionData {
	return &SessionData{
		ID:         s.ID,
		WorkflowID: s.WorkflowID,
		Context:    s.Context,
		History:    s.History,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

