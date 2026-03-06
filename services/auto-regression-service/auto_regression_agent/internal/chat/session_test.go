package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChatSession(t *testing.T) {
	tests := []struct {
		name       string
		config     SessionConfig
		wantErr    bool
		errContain string
	}{
		{
			name: "valid config",
			config: SessionConfig{
				WorkflowID: "test-workflow-123",
			},
			wantErr: false,
		},
		{
			name: "empty workflow ID",
			config: SessionConfig{
				WorkflowID: "",
			},
			wantErr:    true,
			errContain: "workflow_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewChatSession(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContain)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, session.ID)
			assert.Equal(t, tt.config.WorkflowID, session.WorkflowID)
			assert.NotNil(t, session.Context)
			assert.Empty(t, session.History)
			assert.False(t, session.CreatedAt.IsZero())
		})
	}
}

func TestChatSession_GetConversationHistory(t *testing.T) {
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)

	// Initially empty
	history := session.GetConversationHistory()
	assert.Empty(t, history)

	// Add some messages
	session.History = []ChatMessage{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
	}

	history = session.GetConversationHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, "user", history[0].Role)
	assert.Equal(t, "assistant", history[1].Role)
}

func TestChatSession_GetContextSummary(t *testing.T) {
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow-abc",
	})
	require.NoError(t, err)

	summary := session.GetContextSummary()
	assert.Contains(t, summary, "test-workflow-abc")
}

func TestChatSession_buildContextualPrompt(t *testing.T) {
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)

	// Add some history
	session.History = []ChatMessage{
		{Role: "user", Content: "Previous message", Timestamp: time.Now()},
	}

	prompt := session.buildContextualPrompt("Current question")

	assert.Contains(t, prompt, "test-workflow")
	assert.Contains(t, prompt, "Current question")
	assert.Contains(t, prompt, "User Request")
}

func TestChatSession_extractToolCallsFromResponse(t *testing.T) {
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		response string
		wantLen  int
	}{
		{
			name:     "no tool calls",
			response: "Just a regular response",
			wantLen:  0,
		},
		{
			name:     "with tool results",
			response: "Some text\nTool Results:\nTool read_test_suite result: success\nTool edit_test_case result: done",
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := session.extractToolCallsFromResponse(tt.response)
			assert.Len(t, calls, tt.wantLen)
		})
	}
}

func TestChatSession_SendMessage_NoAgent(t *testing.T) {
	session, err := NewChatSession(SessionConfig{
		WorkflowID: "test-workflow",
	})
	require.NoError(t, err)

	// Without a proper agent, feedbackAgent is nil
	// Verify the session was created but agent is nil
	assert.Nil(t, session.feedbackAgent)
}
