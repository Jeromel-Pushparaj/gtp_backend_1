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

const (
	// DefaultSessionsDir is the default directory for storing chat sessions
	DefaultSessionsDir = "output/sessions"
)

// HistoryManager manages chat session persistence
type HistoryManager struct {
	sessionsDir string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(sessionsDir string) *HistoryManager {
	if sessionsDir == "" {
		sessionsDir = DefaultSessionsDir
	}
	return &HistoryManager{sessionsDir: sessionsDir}
}

// SaveSession saves a chat session to disk
func (hm *HistoryManager) SaveSession(session *ChatSession) (string, error) {
	if err := os.MkdirAll(hm.sessionsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sessions directory: %w", err)
	}

	filename := fmt.Sprintf("%s.json", session.ID)
	filePath := filepath.Join(hm.sessionsDir, filename)

	// Create a serializable copy without runtime fields
	sessionData := struct {
		ID         string        `json:"id"`
		WorkflowID string        `json:"workflow_id"`
		Context    *ChatContext  `json:"context"`
		History    []ChatMessage `json:"history"`
		CreatedAt  time.Time     `json:"created_at"`
		UpdatedAt  time.Time     `json:"updated_at"`
	}{
		ID:         session.ID,
		WorkflowID: session.WorkflowID,
		Context:    session.Context,
		History:    session.History,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}

	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write session file: %w", err)
	}

	return filePath, nil
}

// LoadSessionData loads session data from a file (without reinitializing agents)
func (hm *HistoryManager) LoadSessionData(sessionID string) (*ChatSession, error) {
	filename := fmt.Sprintf("%s.json", sessionID)
	filePath := filepath.Join(hm.sessionsDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// ListSessions lists all saved sessions
func (hm *HistoryManager) ListSessions() ([]SessionInfo, error) {
	entries, err := os.ReadDir(hm.sessionsDir)
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

		filePath := filepath.Join(hm.sessionsDir, entry.Name())
		info, err := hm.getSessionInfo(filePath)
		if err != nil {
			continue // Skip invalid session files
		}
		sessions = append(sessions, info)
	}

	// Sort by updated_at descending (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// SessionInfo contains summary information about a session
type SessionInfo struct {
	ID           string    `json:"id"`
	WorkflowID   string    `json:"workflow_id"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FilePath     string    `json:"file_path"`
}

// getSessionInfo extracts summary info from a session file
func (hm *HistoryManager) getSessionInfo(filePath string) (SessionInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return SessionInfo{}, err
	}

	var session struct {
		ID         string        `json:"id"`
		WorkflowID string        `json:"workflow_id"`
		History    []ChatMessage `json:"history"`
		CreatedAt  time.Time     `json:"created_at"`
		UpdatedAt  time.Time     `json:"updated_at"`
	}

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

// FindSessionsByWorkflow finds all sessions for a specific workflow
func (hm *HistoryManager) FindSessionsByWorkflow(workflowID string) ([]SessionInfo, error) {
	allSessions, err := hm.ListSessions()
	if err != nil {
		return nil, err
	}

	var matching []SessionInfo
	for _, session := range allSessions {
		if session.WorkflowID == workflowID {
			matching = append(matching, session)
		}
	}

	return matching, nil
}

// DeleteSession deletes a session file
func (hm *HistoryManager) DeleteSession(sessionID string) error {
	filename := fmt.Sprintf("%s.json", sessionID)
	filePath := filepath.Join(hm.sessionsDir, filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetLatestSessionForWorkflow returns the most recent session for a workflow
func (hm *HistoryManager) GetLatestSessionForWorkflow(workflowID string) (*SessionInfo, error) {
	sessions, err := hm.FindSessionsByWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, nil
	}

	// Sessions are already sorted by UpdatedAt descending
	return &sessions[0], nil
}

// ExportHistory exports conversation history to a readable format
func (hm *HistoryManager) ExportHistory(sessionID string, format string) (string, error) {
	session, err := hm.LoadSessionData(sessionID)
	if err != nil {
		return "", err
	}

	switch format {
	case "markdown":
		return hm.exportToMarkdown(session), nil
	case "text":
		return hm.exportToText(session), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// exportToMarkdown exports session to markdown format
func (hm *HistoryManager) exportToMarkdown(session *ChatSession) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Chat Session: %s\n\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Workflow:** %s\n\n", session.WorkflowID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n\n", session.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("---\n\n")

	for _, msg := range session.History {
		if msg.Role == "user" {
			sb.WriteString("## 👤 User\n\n")
		} else {
			sb.WriteString("## 🤖 Assistant\n\n")
		}
		sb.WriteString(fmt.Sprintf("*%s*\n\n", msg.Timestamp.Format("15:04:05")))
		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")

		if len(msg.ToolCalls) > 0 {
			sb.WriteString("**Tools Used:**\n")
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("- `%s`\n", tc.Name))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// exportToText exports session to plain text format
func (hm *HistoryManager) exportToText(session *ChatSession) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Chat Session: %s\n", session.ID))
	sb.WriteString(fmt.Sprintf("Workflow: %s\n", session.WorkflowID))
	sb.WriteString(fmt.Sprintf("Created: %s\n", session.CreatedAt.Format(time.RFC3339)))
	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n\n")

	for _, msg := range session.History {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		sb.WriteString(fmt.Sprintf("[%s] %s:\n", msg.Timestamp.Format("15:04:05"), role))
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("-", 40))
		sb.WriteString("\n\n")
	}

	return sb.String()
}
