package autonomous

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedbackToolExecutor_GetToolDefinitions(t *testing.T) {
	// Create executor without vector store (nil is allowed for testing)
	executor := NewFeedbackToolExecutor(nil)

	tools := executor.GetToolDefinitions()

	// Should have 20 tools (7 vector DB + 4 read + 4 write + 1 analysis + 4 execute)
	assert.Len(t, tools, 20)

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		// Vector DB tools
		"store_learned_pattern",
		"search_similar_patterns",
		"store_failure_pattern",
		"search_failure_fixes",
		"store_successful_strategy",
		"search_strategies",
		"update_learning_confidence",
		// Read tools
		"read_test_suite",
		"read_test_results",
		"read_spec",
		"list_output_files",
		// Write tools
		"edit_test_case",
		"add_test_case",
		"remove_test_case",
		"update_test_context",
		// Analysis tools
		"generate_recommendations",
		// Execute tools
		"execute_single_test",
		"execute_test_subset",
		"execute_failed_tests",
		"execute_with_context",
	}

	for _, name := range expectedTools {
		assert.True(t, toolNames[name], "Expected tool %s to be registered", name)
	}
}

func TestFeedbackToolExecutor_ExecuteTool_UnknownTool(t *testing.T) {
	executor := NewFeedbackToolExecutor(nil)

	result := executor.ExecuteTool(context.Background(), ToolCall{
		Name: "unknown_tool",
		Args: map[string]interface{}{},
	})

	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "unknown")
}

func TestToolsToJSON(t *testing.T) {
	tools := []ToolDefinition{
		{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"param1": {
						Type:        "string",
						Description: "First parameter",
					},
				},
				Required: []string{"param1"},
			},
		},
	}

	jsonTools, err := ToolsToJSON(tools)
	require.NoError(t, err)
	require.Len(t, jsonTools, 1)

	// Verify structure
	tool := jsonTools[0]
	assert.Equal(t, "test_tool", tool["name"])
	assert.Equal(t, "A test tool", tool["description"])

	params, ok := tool["parameters"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", params["type"])
}

func TestParseToolCalls(t *testing.T) {
	fa := &FeedbackAgent{}

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "no tool call",
			text:     "This is just regular text without any tool calls.",
			expected: 0,
		},
		{
			name:     "single tool call",
			text:     `I'll store this pattern. {"tool": "store_learned_pattern", "args": {"category": "auth_pattern", "content": "test"}}`,
			expected: 1,
		},
		{
			name:     "malformed json",
			text:     `{"tool": "incomplete`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := fa.parseToolCalls(tt.text)
			assert.Len(t, calls, tt.expected)
		})
	}
}

func TestGetStringFromPayload(t *testing.T) {
	payload := map[string]interface{}{
		"string_key": "value",
		"int_key":    123,
		"nil_key":    nil,
	}

	assert.Equal(t, "value", getStringFromPayload(payload, "string_key"))
	assert.Equal(t, "", getStringFromPayload(payload, "int_key"))
	assert.Equal(t, "", getStringFromPayload(payload, "nil_key"))
	assert.Equal(t, "", getStringFromPayload(payload, "missing_key"))
}
