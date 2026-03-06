package chat

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
	ColorDim     = "\033[2m"
)

// Formatter handles output formatting for the chat interface
type Formatter struct {
	useColor bool
}

// NewFormatter creates a new formatter
func NewFormatter(useColor bool) *Formatter {
	return &Formatter{useColor: useColor}
}

// color applies color if enabled
func (f *Formatter) color(text, colorCode string) string {
	if !f.useColor {
		return text
	}
	return colorCode + text + ColorReset
}

// FormatWelcome formats the welcome message
func (f *Formatter) FormatWelcome(workflowID string) string {
	var sb strings.Builder

	sb.WriteString(f.color("\n╔══════════════════════════════════════════════════════════════╗\n", ColorCyan))
	sb.WriteString(f.color("║           🤖 OpenTest Interactive Chat Mode                   ║\n", ColorCyan))
	sb.WriteString(f.color("╚══════════════════════════════════════════════════════════════╝\n", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Workflow: %s\n", f.color(workflowID, ColorYellow)))
	sb.WriteString("\n")
	sb.WriteString(f.color("Available Commands:\n", ColorBold))
	sb.WriteString("  /status   - Show current context status\n")
	sb.WriteString("  /reload   - Reload context from files\n")
	sb.WriteString("  /history  - Show conversation history\n")
	sb.WriteString("  /save     - Save current session\n")
	sb.WriteString("  /clear    - Clear conversation history\n")
	sb.WriteString("  /help     - Show this help message\n")
	sb.WriteString("  /exit     - Exit chat mode\n")
	sb.WriteString("\n")
	sb.WriteString(f.color("Type your message or command to begin...\n", ColorDim))

	return sb.String()
}

// FormatContextLoaded formats the context loaded message
func (f *Formatter) FormatContextLoaded(summary string) string {
	var sb strings.Builder

	sb.WriteString(f.color("\n📂 Context Loaded:\n", ColorGreen))
	sb.WriteString(f.color("─────────────────\n", ColorDim))
	sb.WriteString(summary)
	sb.WriteString("\n")

	return sb.String()
}

// FormatUserPrompt formats the user input prompt
func (f *Formatter) FormatUserPrompt() string {
	return f.color("\n👤 You: ", ColorBlue)
}

// FormatAssistantResponse formats the assistant's response
func (f *Formatter) FormatAssistantResponse(response string) string {
	var sb strings.Builder

	sb.WriteString(f.color("\n🤖 Assistant:\n", ColorMagenta))
	sb.WriteString(f.color("─────────────\n", ColorDim))

	// Format tool results specially
	if strings.Contains(response, "Tool Results:") {
		parts := strings.SplitN(response, "Tool Results:", 2)
		sb.WriteString(parts[0])
		if len(parts) > 1 {
			sb.WriteString(f.color("\n📦 Tool Results:\n", ColorYellow))
			sb.WriteString(f.formatToolResults(parts[1]))
		}
	} else {
		sb.WriteString(response)
	}

	sb.WriteString("\n")
	return sb.String()
}

// formatToolResults formats the tool results section
func (f *Formatter) formatToolResults(results string) string {
	var sb strings.Builder
	lines := strings.Split(results, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Tool ") && strings.Contains(line, " result:") {
			// Success result - extract tool name and format nicely
			toolName := f.extractToolName(line)
			sb.WriteString(f.color("  ✅ ", ColorGreen))
			sb.WriteString(f.color(toolName, ColorBold))
			sb.WriteString(": ")
			// Extract result value
			if idx := strings.Index(line, " result:"); idx != -1 {
				result := strings.TrimSpace(line[idx+8:])
				sb.WriteString(f.formatToolResultValue(result))
			}
			sb.WriteString("\n")
		} else if strings.HasPrefix(line, "Tool ") && strings.Contains(line, " error:") {
			// Error result
			toolName := f.extractToolName(line)
			sb.WriteString(f.color("  ❌ ", ColorRed))
			sb.WriteString(f.color(toolName, ColorBold))
			sb.WriteString(": ")
			if idx := strings.Index(line, " error:"); idx != -1 {
				errMsg := strings.TrimSpace(line[idx+7:])
				sb.WriteString(f.color(errMsg, ColorRed))
			}
			sb.WriteString("\n")
		} else if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "[") {
			// JSON data - format with indentation
			sb.WriteString(f.color("    ", ColorDim))
			sb.WriteString(f.color(line, ColorDim))
			sb.WriteString("\n")
		} else {
			sb.WriteString("  ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// extractToolName extracts the tool name from a "Tool X result:" line
func (f *Formatter) extractToolName(line string) string {
	line = strings.TrimPrefix(line, "Tool ")
	if idx := strings.Index(line, " result:"); idx != -1 {
		return line[:idx]
	}
	if idx := strings.Index(line, " error:"); idx != -1 {
		return line[:idx]
	}
	return line
}

// formatToolResultValue formats a tool result value for display
func (f *Formatter) formatToolResultValue(result string) string {
	// Truncate long results
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}

// FormatError formats an error message
func (f *Formatter) FormatError(err error) string {
	return f.color(fmt.Sprintf("\n❌ Error: %v\n", err), ColorRed)
}

// FormatSuccess formats a success message
func (f *Formatter) FormatSuccess(message string) string {
	return f.color(fmt.Sprintf("\n✅ %s\n", message), ColorGreen)
}

// FormatWarning formats a warning message
func (f *Formatter) FormatWarning(message string) string {
	return f.color(fmt.Sprintf("\n⚠️  %s\n", message), ColorYellow)
}

// FormatInfo formats an info message
func (f *Formatter) FormatInfo(message string) string {
	return f.color(fmt.Sprintf("\nℹ️  %s\n", message), ColorCyan)
}

// FormatToolExecution formats a tool execution in progress
func (f *Formatter) FormatToolExecution(toolName string, args map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString(f.color("  🔧 ", ColorYellow))
	sb.WriteString(f.color(toolName, ColorBold))
	if len(args) > 0 {
		sb.WriteString(f.color(" (", ColorDim))
		first := true
		for k, v := range args {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
		sb.WriteString(f.color(")", ColorDim))
	}
	sb.WriteString("\n")
	return sb.String()
}

// FormatToolResult formats a single tool result
func (f *Formatter) FormatToolResult(toolName string, success bool, result interface{}) string {
	var sb strings.Builder
	if success {
		sb.WriteString(f.color("  ✅ ", ColorGreen))
		sb.WriteString(f.color(toolName, ColorBold))
		sb.WriteString(": ")
		resultStr := fmt.Sprintf("%v", result)
		if len(resultStr) > 100 {
			resultStr = resultStr[:100] + "..."
		}
		sb.WriteString(resultStr)
	} else {
		sb.WriteString(f.color("  ❌ ", ColorRed))
		sb.WriteString(f.color(toolName, ColorBold))
		sb.WriteString(": ")
		sb.WriteString(f.color(fmt.Sprintf("%v", result), ColorRed))
	}
	sb.WriteString("\n")
	return sb.String()
}

// FormatHistory formats conversation history
func (f *Formatter) FormatHistory(messages []ChatMessage) string {
	var sb strings.Builder

	sb.WriteString(f.color("\n📜 Conversation History\n", ColorCyan))
	sb.WriteString(f.color("═══════════════════════\n", ColorDim))

	if len(messages) == 0 {
		sb.WriteString(f.color("(No messages yet)\n", ColorDim))
		return sb.String()
	}

	for i, msg := range messages {
		timestamp := msg.Timestamp.Format("15:04:05")

		if msg.Role == "user" {
			sb.WriteString(f.color(fmt.Sprintf("\n[%s] 👤 User:\n", timestamp), ColorBlue))
		} else {
			sb.WriteString(f.color(fmt.Sprintf("\n[%s] 🤖 Assistant:\n", timestamp), ColorMagenta))
		}

		// Truncate long messages
		content := msg.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		sb.WriteString(content)
		sb.WriteString("\n")

		// Show tool calls if any
		if len(msg.ToolCalls) > 0 {
			sb.WriteString(f.color("  Tools used: ", ColorYellow))
			names := make([]string, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				names[j] = tc.Name
			}
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString("\n")
		}

		if i < len(messages)-1 {
			sb.WriteString(f.color("───────────────\n", ColorDim))
		}
	}

	return sb.String()
}

// FormatTestSuiteChange formats a test suite modification
func (f *Formatter) FormatTestSuiteChange(changeType string, testName string, details map[string]interface{}) string {
	var sb strings.Builder

	switch changeType {
	case "add":
		sb.WriteString(f.color("➕ Test Added: ", ColorGreen))
	case "edit":
		sb.WriteString(f.color("✏️  Test Modified: ", ColorYellow))
	case "remove":
		sb.WriteString(f.color("➖ Test Removed: ", ColorRed))
	default:
		sb.WriteString(f.color("📝 Test Changed: ", ColorCyan))
	}

	sb.WriteString(f.color(testName, ColorBold))
	sb.WriteString("\n")

	if details != nil {
		for key, value := range details {
			sb.WriteString(fmt.Sprintf("  • %s: %v\n", key, value))
		}
	}

	return sb.String()
}

// FormatStatus formats the status display
func (f *Formatter) FormatStatus(session *ChatSession) string {
	var sb strings.Builder

	sb.WriteString(f.color("\n📊 Session Status\n", ColorCyan))
	sb.WriteString(f.color("═════════════════\n", ColorDim))
	sb.WriteString(fmt.Sprintf("Session ID:   %s\n", session.ID))
	sb.WriteString(fmt.Sprintf("Workflow ID:  %s\n", session.WorkflowID))
	sb.WriteString(fmt.Sprintf("Messages:     %d\n", len(session.History)))
	sb.WriteString(fmt.Sprintf("Created:      %s\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Last Updated: %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("\n")
	sb.WriteString(session.GetContextSummary())

	return sb.String()
}

// FormatGoodbye formats the goodbye message
func (f *Formatter) FormatGoodbye() string {
	return f.color("\n👋 Goodbye! Session saved.\n", ColorCyan)
}

// FormatHelp formats the help message
func (f *Formatter) FormatHelp() string {
	var sb strings.Builder

	sb.WriteString(f.color("\n📚 Help - Available Commands\n", ColorCyan))
	sb.WriteString(f.color("════════════════════════════\n", ColorDim))
	sb.WriteString("\n")
	sb.WriteString(f.color("Session Commands:\n", ColorBold))
	sb.WriteString("  /status      - Show current session and context status\n")
	sb.WriteString("  /reload      - Reload context from workflow files\n")
	sb.WriteString("  /history     - Show full conversation history\n")
	sb.WriteString("  /save        - Save session to file\n")
	sb.WriteString("  /clear       - Clear conversation history\n")
	sb.WriteString("  /help        - Show this help message\n")
	sb.WriteString("  /exit        - Exit chat mode (saves automatically)\n")
	sb.WriteString("\n")
	sb.WriteString(f.color("Test Commands:\n", ColorBold))
	sb.WriteString("  /run [url]   - Run the test suite (optional base URL override)\n")
	sb.WriteString("  /refine <feedback> - Refine tests based on your feedback\n")
	sb.WriteString("  /recommend   - Generate improvement recommendations\n")
	sb.WriteString("\n")
	sb.WriteString(f.color("Examples:\n", ColorBold))
	sb.WriteString("  \"The auth test is failing with 401\"\n")
	sb.WriteString("  \"Add a test for the delete endpoint\"\n")
	sb.WriteString("  \"/refine Add more negative tests for validation\"\n")
	sb.WriteString("  \"/run http://localhost:8080\"\n")
	sb.WriteString("  \"/recommend\"\n")

	return sb.String()
}
