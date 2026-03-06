package chat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHistoryManager(t *testing.T) {
	// Default directory
	hm := NewHistoryManager("")
	assert.Equal(t, DefaultSessionsDir, hm.sessionsDir)

	// Custom directory
	hm = NewHistoryManager("/custom/path")
	assert.Equal(t, "/custom/path", hm.sessionsDir)
}

func TestHistoryManager_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// Create a session
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow-123",
	})
	require.NoError(t, err)

	// Add some history
	session.History = []ChatMessage{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi!", Timestamp: time.Now()},
	}

	// Save session
	filePath, err := hm.SaveSession(session)
	require.NoError(t, err)
	assert.FileExists(t, filePath)

	// Load session
	loaded, err := hm.LoadSessionData(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.WorkflowID, loaded.WorkflowID)
	assert.Len(t, loaded.History, 2)
}

func TestHistoryManager_ListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// Initially empty
	sessions, err := hm.ListSessions()
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Create and save sessions
	for i := 0; i < 3; i++ {
		session, err := NewChatSession(SessionConfig{
			WorkflowID: "test-workflow",
		})
		require.NoError(t, err)
		_, err = hm.SaveSession(session)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List sessions
	sessions, err = hm.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Should be sorted by UpdatedAt descending
	for i := 0; i < len(sessions)-1; i++ {
		assert.True(t, sessions[i].UpdatedAt.After(sessions[i+1].UpdatedAt) ||
			sessions[i].UpdatedAt.Equal(sessions[i+1].UpdatedAt))
	}
}

func TestHistoryManager_FindSessionsByWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// Create sessions for different workflows
	workflows := []string{"workflow-a", "workflow-b", "workflow-a"}
	for _, wf := range workflows {
		session, err := NewChatSession(SessionConfig{
			WorkflowID: wf,
		})
		require.NoError(t, err)
		_, err = hm.SaveSession(session)
		require.NoError(t, err)
	}

	// Find sessions for workflow-a
	sessions, err := hm.FindSessionsByWorkflow("workflow-a")
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Find sessions for workflow-b
	sessions, err = hm.FindSessionsByWorkflow("workflow-b")
	require.NoError(t, err)
	assert.Len(t, sessions, 1)

	// Find sessions for non-existent workflow
	sessions, err = hm.FindSessionsByWorkflow("workflow-c")
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestHistoryManager_DeleteSession(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// Create and save a session
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)
	filePath, err := hm.SaveSession(session)
	require.NoError(t, err)
	assert.FileExists(t, filePath)

	// Delete session
	err = hm.DeleteSession(session.ID)
	require.NoError(t, err)
	assert.NoFileExists(t, filePath)

	// Delete non-existent session
	err = hm.DeleteSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestHistoryManager_GetLatestSessionForWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// No sessions
	latest, err := hm.GetLatestSessionForWorkflow("test-workflow")
	require.NoError(t, err)
	assert.Nil(t, latest)

	// Create sessions
	var lastSession *ChatSession
	for i := 0; i < 3; i++ {
		session, err := NewChatSession(SessionConfig{
			WorkflowID: "test-workflow",
		})
		require.NoError(t, err)
		_, err = hm.SaveSession(session)
		require.NoError(t, err)
		lastSession = session
		time.Sleep(10 * time.Millisecond)
	}

	// Get latest
	latest, err = hm.GetLatestSessionForWorkflow("test-workflow")
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, lastSession.ID, latest.ID)
}

func TestHistoryManager_ExportHistory(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	// Create and save a session with history
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)
	session.History = []ChatMessage{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi!", Timestamp: time.Now()},
	}
	_, err = hm.SaveSession(session)
	require.NoError(t, err)

	// Export to markdown
	md, err := hm.ExportHistory(session.ID, "markdown")
	require.NoError(t, err)
	assert.Contains(t, md, "# Chat Session")
	assert.Contains(t, md, "Hello")
	assert.Contains(t, md, "Hi!")

	// Export to text
	txt, err := hm.ExportHistory(session.ID, "text")
	require.NoError(t, err)
	assert.Contains(t, txt, "Chat Session")
	assert.Contains(t, txt, "Hello")

	// Invalid format
	_, err = hm.ExportHistory(session.ID, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestHistoryManager_ListSessions_NonExistentDir(t *testing.T) {
	hm := NewHistoryManager("/non/existent/path")
	sessions, err := hm.ListSessions()
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestHistoryManager_LoadSessionData_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	hm := NewHistoryManager(tmpDir)

	_, err := hm.LoadSessionData("non-existent-id")
	assert.Error(t, err)
}

func TestHistoryManager_SaveSession_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "nested", "sessions")
	hm := NewHistoryManager(sessionsDir)

	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)

	filePath, err := hm.SaveSession(session)
	require.NoError(t, err)
	assert.FileExists(t, filePath)

	// Verify directory was created
	_, err = os.Stat(sessionsDir)
	require.NoError(t, err)
}
