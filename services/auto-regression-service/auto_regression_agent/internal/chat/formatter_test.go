package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	// With color
	f := NewFormatter(true)
	assert.NotNil(t, f)
	assert.True(t, f.useColor)

	// Without color
	f = NewFormatter(false)
	assert.NotNil(t, f)
	assert.False(t, f.useColor)
}

func TestFormatter_color(t *testing.T) {
	// With color enabled
	f := NewFormatter(true)
	result := f.color("test", ColorRed)
	assert.Contains(t, result, ColorRed)
	assert.Contains(t, result, ColorReset)
	assert.Contains(t, result, "test")

	// With color disabled
	f = NewFormatter(false)
	result = f.color("test", ColorRed)
	assert.Equal(t, "test", result)
	assert.NotContains(t, result, ColorRed)
}

func TestFormatter_FormatWelcome(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatWelcome("test-workflow-123")

	assert.Contains(t, result, "test-workflow-123")
	assert.Contains(t, result, "Interactive Chat Mode")
	assert.Contains(t, result, "/status")
	assert.Contains(t, result, "/exit")
}

func TestFormatter_FormatContextLoaded(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatContextLoaded("Test Suite: API Tests (10 tests)")

	assert.Contains(t, result, "Context Loaded")
	assert.Contains(t, result, "Test Suite: API Tests")
}

func TestFormatter_FormatUserPrompt(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatUserPrompt()

	assert.Contains(t, result, "You:")
}

func TestFormatter_FormatAssistantResponse(t *testing.T) {
	f := NewFormatter(false)

	// Regular response
	result := f.FormatAssistantResponse("This is a test response")
	assert.Contains(t, result, "Assistant")
	assert.Contains(t, result, "This is a test response")

	// Response with tool results
	result = f.FormatAssistantResponse("Analysis complete\nTool Results:\nTool read_test_suite result: success")
	assert.Contains(t, result, "Tool Results")
}

func TestFormatter_FormatError(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatError(assert.AnError)

	assert.Contains(t, result, "Error")
	assert.Contains(t, result, assert.AnError.Error())
}

func TestFormatter_FormatSuccess(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatSuccess("Operation completed")

	assert.Contains(t, result, "Operation completed")
}

func TestFormatter_FormatWarning(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatWarning("Be careful")

	assert.Contains(t, result, "Be careful")
}

func TestFormatter_FormatInfo(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatInfo("Some information")

	assert.Contains(t, result, "Some information")
}

func TestFormatter_FormatHistory(t *testing.T) {
	f := NewFormatter(false)

	// Empty history
	result := f.FormatHistory(nil)
	assert.Contains(t, result, "Conversation History")
	assert.Contains(t, result, "No messages yet")

	// With messages
	messages := []ChatMessage{
		{Role: "user", Content: "Hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
	}
	result = f.FormatHistory(messages)
	assert.Contains(t, result, "User")
	assert.Contains(t, result, "Assistant")
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "Hi there!")
}

func TestFormatter_FormatTestSuiteChange(t *testing.T) {
	f := NewFormatter(false)

	// Add
	result := f.FormatTestSuiteChange("add", "New Test", nil)
	assert.Contains(t, result, "Test Added")
	assert.Contains(t, result, "New Test")

	// Edit
	result = f.FormatTestSuiteChange("edit", "Modified Test", map[string]interface{}{
		"status": 201,
	})
	assert.Contains(t, result, "Test Modified")
	assert.Contains(t, result, "status")

	// Remove
	result = f.FormatTestSuiteChange("remove", "Deleted Test", nil)
	assert.Contains(t, result, "Test Removed")
}

func TestFormatter_FormatHelp(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatHelp()

	assert.Contains(t, result, "Help")
	assert.Contains(t, result, "/status")
	assert.Contains(t, result, "/reload")
	assert.Contains(t, result, "/history")
	assert.Contains(t, result, "/save")
	assert.Contains(t, result, "/clear")
	assert.Contains(t, result, "/exit")
}

func TestFormatter_FormatGoodbye(t *testing.T) {
	f := NewFormatter(false)
	result := f.FormatGoodbye()

	assert.Contains(t, result, "Goodbye")
}

