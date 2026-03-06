package chat

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/testhistory"
)

// CommandHandler handles chat commands
type CommandHandler struct {
	testHistory *testhistory.TestHistory
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(th *testhistory.TestHistory) *CommandHandler {
	return &CommandHandler{testHistory: th}
}

// HandleCommand processes a chat command and returns the response
func (ch *CommandHandler) HandleCommand(ctx context.Context, session *ChatSession, input string) (string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "/help":
		return ch.handleHelp()
	case "/history":
		return ch.handleHistory(ctx, session, args)
	case "/compare":
		return ch.handleCompare(ctx, args)
	case "/trends":
		return ch.handleTrends(ctx, session, args)
	case "/status":
		return ch.handleStatus(session)
	case "/clear":
		return ch.handleClear(session)
	case "/export":
		return ch.handleExport(session, args)
	default:
		return "", fmt.Errorf("unknown command: %s. Type /help for available commands", command)
	}
}

// IsCommand checks if the input is a command
func IsCommand(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "/")
}

// handleHelp returns help text
func (ch *CommandHandler) handleHelp() (string, error) {
	help := "**Available Commands:**\n\n" +
		"**Test History:**\n" +
		"- `/history [limit]` - Show recent test runs (default: 10)\n" +
		"- `/compare <run1> <run2>` - Compare two test runs\n" +
		"- `/trends [days]` - Show test trends (default: 30 days)\n\n" +
		"**Session:**\n" +
		"- `/status` - Show current session status\n" +
		"- `/clear` - Clear conversation history\n" +
		"- `/export [format]` - Export session (json, markdown)\n\n" +
		"**General:**\n" +
		"- `/help` - Show this help message\n"
	return help, nil
}

// handleHistory shows test run history
func (ch *CommandHandler) handleHistory(ctx context.Context, session *ChatSession, args []string) (string, error) {
	if ch.testHistory == nil {
		return "Test history not available (database not configured)", nil
	}

	limit := 10
	if len(args) > 0 {
		if l, err := strconv.Atoi(args[0]); err == nil && l > 0 {
			limit = l
		}
	}

	runs, err := ch.testHistory.GetRunHistory(ctx, session.WorkflowID, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get history: %w", err)
	}

	if len(runs) == 0 {
		return "No test runs found for this workflow.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Recent Test Runs** (showing %d)\n\n", len(runs)))
	sb.WriteString("| Run ID | Date | Pass Rate | Passed | Failed | Duration |\n")
	sb.WriteString("|--------|------|-----------|--------|--------|----------|\n")

	for _, run := range runs {
		sb.WriteString(fmt.Sprintf("| %s | %s | %.1f%% | %d | %d | %dms |\n",
			run.ID.String()[:8],
			run.StartTime.Format("2006-01-02 15:04"),
			run.PassRate,
			run.Passed,
			run.Failed,
			run.DurationMs,
		))
	}

	return sb.String(), nil
}

// handleCompare compares two test runs
func (ch *CommandHandler) handleCompare(ctx context.Context, args []string) (string, error) {
	if ch.testHistory == nil {
		return "Test history not available (database not configured)", nil
	}

	if len(args) < 2 {
		return "Usage: /compare <run1-id> <run2-id>", nil
	}

	run1ID, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Sprintf("Invalid run ID: %s", args[0]), nil
	}

	run2ID, err := uuid.Parse(args[1])
	if err != nil {
		return fmt.Sprintf("Invalid run ID: %s", args[1]), nil
	}

	comparison, err := ch.testHistory.CompareRuns(ctx, run1ID, run2ID)
	if err != nil {
		return "", fmt.Errorf("failed to compare runs: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("**Test Run Comparison**\n\n")
	sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", comparison.Summary))
	sb.WriteString(fmt.Sprintf("| Metric | Run 1 | Run 2 | Delta |\n"))
	sb.WriteString("|--------|-------|-------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Pass Rate | %.1f%% | %.1f%% | %+.1f%% |\n",
		comparison.Run1.PassRate, comparison.Run2.PassRate, comparison.PassRateDelta))
	sb.WriteString(fmt.Sprintf("| Passed | %d | %d | %+d |\n",
		comparison.Run1.Passed, comparison.Run2.Passed, comparison.PassedDelta))
	sb.WriteString(fmt.Sprintf("| Failed | %d | %d | %+d |\n",
		comparison.Run1.Failed, comparison.Run2.Failed, comparison.FailedDelta))

	if len(comparison.FixedTests) > 0 {
		sb.WriteString(fmt.Sprintf("\n**Fixed Tests:** %d\n", len(comparison.FixedTests)))
	}
	if len(comparison.NewFailures) > 0 {
		sb.WriteString(fmt.Sprintf("\n**New Failures:** %d\n", len(comparison.NewFailures)))
	}

	return sb.String(), nil
}

// handleTrends shows test trends
func (ch *CommandHandler) handleTrends(ctx context.Context, session *ChatSession, args []string) (string, error) {
	if ch.testHistory == nil {
		return "Test history not available (database not configured)", nil
	}

	days := 30
	if len(args) > 0 {
		if d, err := strconv.Atoi(args[0]); err == nil && d > 0 {
			days = d
		}
	}

	report, err := ch.testHistory.GetTrends(ctx, session.WorkflowID, days)
	if err != nil {
		return "", fmt.Errorf("failed to get trends: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Test Trends** (last %d days)\n\n", days))
	sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", report.Summary))
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Runs | %d |\n", report.TotalRuns))
	sb.WriteString(fmt.Sprintf("| Avg Pass Rate | %.1f%% |\n", report.AveragePassRate))
	sb.WriteString(fmt.Sprintf("| Min Pass Rate | %.1f%% |\n", report.MinPassRate))
	sb.WriteString(fmt.Sprintf("| Max Pass Rate | %.1f%% |\n", report.MaxPassRate))
	sb.WriteString(fmt.Sprintf("| Total Tests | %d |\n", report.TotalTests))
	sb.WriteString(fmt.Sprintf("| Avg Duration | %dms |\n", report.AverageDuration))

	return sb.String(), nil
}

// handleStatus shows session status
func (ch *CommandHandler) handleStatus(session *ChatSession) (string, error) {
	var sb strings.Builder
	sb.WriteString("**Session Status**\n\n")
	sb.WriteString(fmt.Sprintf("- **Session ID:** %s\n", session.ID))
	sb.WriteString(fmt.Sprintf("- **Workflow ID:** %s\n", session.WorkflowID))
	sb.WriteString(fmt.Sprintf("- **Messages:** %d\n", len(session.History)))
	sb.WriteString(fmt.Sprintf("- **Created:** %s\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **Updated:** %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05")))

	if session.Context != nil && session.Context.TestSuite != nil {
		sb.WriteString(fmt.Sprintf("- **Test Suite:** %s\n", session.Context.TestSuite.Name))
	}

	return sb.String(), nil
}

// handleClear clears conversation history
func (ch *CommandHandler) handleClear(session *ChatSession) (string, error) {
	session.History = []ChatMessage{}
	return "Conversation history cleared.", nil
}

// handleExport exports the session
func (ch *CommandHandler) handleExport(session *ChatSession, args []string) (string, error) {
	format := "markdown"
	if len(args) > 0 {
		format = strings.ToLower(args[0])
	}

	switch format {
	case "json":
		return "Export to JSON: Use the session file at ~/.opentest/sessions/" + session.ID + ".json", nil
	case "markdown", "md":
		return ch.exportMarkdown(session), nil
	default:
		return "Supported formats: json, markdown", nil
	}
}

// exportMarkdown exports session as markdown
func (ch *CommandHandler) exportMarkdown(session *ChatSession) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Chat Session: %s\n\n", session.ID))
	sb.WriteString(fmt.Sprintf("**Workflow:** %s\n", session.WorkflowID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range session.History {
		role := "**User:**"
		if msg.Role == "assistant" {
			role = "**Assistant:**"
		}
		sb.WriteString(fmt.Sprintf("%s\n%s\n\n", role, msg.Content))
	}

	return sb.String()
}
