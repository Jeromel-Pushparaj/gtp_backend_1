package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/agents/autonomous"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

// ChatContext holds all relevant context for a chat session
type ChatContext struct {
	Spec      map[string]interface{} `json:"spec,omitempty"`
	TestSuite *autonomous.TestSuite  `json:"test_suite,omitempty"`
	Results   map[string]interface{} `json:"results,omitempty"`
	FilePaths map[string]string      `json:"file_paths"`
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role      string           `json:"role"` // "user" or "assistant"
	Content   string           `json:"content"`
	Timestamp time.Time        `json:"timestamp"`
	ToolCalls []ToolCallRecord `json:"tool_calls,omitempty"`
}

// ToolCallRecord records a tool invocation
type ToolCallRecord struct {
	Name      string                 `json:"name"`
	Args      map[string]interface{} `json:"args"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChatSession manages an interactive chat session with the Feedback Agent
type ChatSession struct {
	ID         string        `json:"id"`
	WorkflowID string        `json:"workflow_id"`
	Context    *ChatContext  `json:"context"`
	History    []ChatMessage `json:"history"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`

	feedbackAgent *autonomous.FeedbackAgent
	llmClient     *llm.Client
	vectorStore   *vectordb.Store
	mu            sync.RWMutex
}

// SessionConfig holds configuration for creating a chat session
type SessionConfig struct {
	WorkflowID  string
	LLMClient   *llm.Client
	VectorStore *vectordb.Store
}

// NewChatSession creates a new chat session
func NewChatSession(cfg SessionConfig) (*ChatSession, error) {
	if cfg.WorkflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	session := &ChatSession{
		ID:          fmt.Sprintf("chat-%s-%d", cfg.WorkflowID, time.Now().UnixNano()),
		WorkflowID:  cfg.WorkflowID,
		Context:     &ChatContext{FilePaths: make(map[string]string)},
		History:     make([]ChatMessage, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		llmClient:   cfg.LLMClient,
		vectorStore: cfg.VectorStore,
	}

	// Create FeedbackAgent only if LLM client is provided
	if cfg.LLMClient != nil {
		session.feedbackAgent = autonomous.NewFeedbackAgent(
			cfg.LLMClient,
			nil, // eventBus - not needed for interactive mode
			nil, // messageBus
			nil, // consensus
			cfg.VectorStore,
		)
	}

	return session, nil
}

// LoadContext loads all relevant files for the workflow
func (s *ChatSession) LoadContext() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use the list_output_files tool logic to find files
	dirs := map[string]string{
		"specs":   "output/specs",
		"suites":  "output/suites",
		"results": "output/results",
		"plans":   "output/plans",
	}

	for category, dir := range dirs {
		if files, err := s.findWorkflowFiles(dir); err == nil && len(files) > 0 {
			s.Context.FilePaths[category] = files[0] // Use first match
		}
	}

	// Load test suite if found
	if suitePath, ok := s.Context.FilePaths["suites"]; ok {
		if suite, err := autonomous.LoadTestSuite(suitePath); err == nil {
			s.Context.TestSuite = suite
		}
	}

	// Load spec if found
	if specPath, ok := s.Context.FilePaths["specs"]; ok {
		if data, err := os.ReadFile(specPath); err == nil {
			var spec map[string]interface{}
			if json.Unmarshal(data, &spec) == nil {
				s.Context.Spec = spec
			}
		}
	}

	// Load results if found
	if resultsPath, ok := s.Context.FilePaths["results"]; ok {
		if data, err := os.ReadFile(resultsPath); err == nil {
			var results map[string]interface{}
			if json.Unmarshal(data, &results) == nil {
				s.Context.Results = results
			}
		}
	}

	return nil
}

// findWorkflowFiles finds files matching the workflow ID in a directory
func (s *ChatSession) findWorkflowFiles(dir string) ([]string, error) {
	var matches []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, s.WorkflowID) && strings.HasSuffix(name, ".json") {
			matches = append(matches, filepath.Join(dir, name))
		}
	}

	return matches, nil
}

// SendMessage processes a user message and returns the agent's response
func (s *ChatSession) SendMessage(ctx context.Context, input string) (string, []ToolCallRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedbackAgent == nil {
		return "", nil, fmt.Errorf("feedback agent not initialized - LLM client required")
	}

	// Record user message
	userMsg := ChatMessage{
		Role:      "user",
		Content:   input,
		Timestamp: time.Now(),
	}
	s.History = append(s.History, userMsg)

	// Build enhanced prompt with context
	enhancedPrompt := s.buildContextualPrompt(input)

	// Process through FeedbackAgent
	response, err := s.feedbackAgent.ProcessInteractiveFeedback(ctx, enhancedPrompt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to process feedback: %w", err)
	}

	// Parse tool calls from response (simplified - FeedbackAgent handles internally)
	toolCalls := s.extractToolCallsFromResponse(response)

	// Record assistant message
	assistantMsg := ChatMessage{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
		ToolCalls: toolCalls,
	}
	s.History = append(s.History, assistantMsg)

	s.UpdatedAt = time.Now()

	return response, toolCalls, nil
}

// buildContextualPrompt builds a prompt with workflow context
func (s *ChatSession) buildContextualPrompt(userMessage string) string {
	var sb strings.Builder

	sb.WriteString("## Current Workflow Context\n")
	sb.WriteString(fmt.Sprintf("Workflow ID: %s\n\n", s.WorkflowID))

	// Add test suite summary
	if s.Context.TestSuite != nil {
		suite := s.Context.TestSuite
		sb.WriteString("### Test Suite Summary\n")
		sb.WriteString(fmt.Sprintf("- Name: %s\n", suite.Name))
		sb.WriteString(fmt.Sprintf("- Total Tests: %d\n", len(suite.Tests)))
		sb.WriteString(fmt.Sprintf("- Positive: %d, Negative: %d, Boundary: %d, Security: %d\n",
			suite.Statistics.PositiveTests, suite.Statistics.NegativeTests,
			suite.Statistics.BoundaryTests, suite.Statistics.SecurityTests))
		sb.WriteString("\nTest Cases:\n")
		for i, test := range suite.Tests {
			status := "📋"
			sb.WriteString(fmt.Sprintf("%d. %s [%s] %s %s - Expected: %d\n",
				i+1, status, test.Category, test.Method, test.Path, test.ExpectedStatus))
		}
		sb.WriteString("\n")
	}

	// Add results summary if available
	if s.Context.Results != nil {
		sb.WriteString("### Test Results Summary\n")
		if summary, ok := s.Context.Results["summary"].(map[string]interface{}); ok {
			if total, ok := summary["total"].(float64); ok {
				sb.WriteString(fmt.Sprintf("- Total Executed: %.0f\n", total))
			}
			if passed, ok := summary["passed"].(float64); ok {
				sb.WriteString(fmt.Sprintf("- Passed: %.0f\n", passed))
			}
			if failed, ok := summary["failed"].(float64); ok {
				sb.WriteString(fmt.Sprintf("- Failed: %.0f\n", failed))
			}
		}
		sb.WriteString("\n")
	}

	// Add conversation history (last 5 messages for context)
	if len(s.History) > 0 {
		sb.WriteString("### Recent Conversation\n")
		start := len(s.History) - 5
		if start < 0 {
			start = 0
		}
		for _, msg := range s.History[start:] {
			role := "User"
			if msg.Role == "assistant" {
				role = "Assistant"
			}
			// Truncate long messages
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", role, content))
		}
		sb.WriteString("\n")
	}

	// Add available file paths
	sb.WriteString("### Available Files\n")
	for category, path := range s.Context.FilePaths {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", category, path))
	}
	sb.WriteString("\n")

	sb.WriteString("## User Request\n")
	sb.WriteString(userMessage)

	return sb.String()
}

// extractToolCallsFromResponse extracts tool call records from response
func (s *ChatSession) extractToolCallsFromResponse(response string) []ToolCallRecord {
	var records []ToolCallRecord

	// Look for "Tool Results:" section in response
	if idx := strings.Index(response, "Tool Results:"); idx != -1 {
		resultsSection := response[idx:]
		lines := strings.Split(resultsSection, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Tool ") && strings.Contains(line, " result:") {
				// Parse "Tool name result: ..."
				parts := strings.SplitN(line, " result:", 2)
				if len(parts) == 2 {
					name := strings.TrimPrefix(parts[0], "Tool ")
					records = append(records, ToolCallRecord{
						Name:      name,
						Timestamp: time.Now(),
					})
				}
			}
		}
	}

	return records
}

// GetConversationHistory returns formatted conversation history
func (s *ChatSession) GetConversationHistory() []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.History
}

// GetContextSummary returns a summary of the loaded context
func (s *ChatSession) GetContextSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Workflow: %s\n", s.WorkflowID))

	if s.Context.TestSuite != nil {
		sb.WriteString(fmt.Sprintf("Test Suite: %s (%d tests)\n",
			s.Context.TestSuite.Name, len(s.Context.TestSuite.Tests)))
	}

	if s.Context.Spec != nil {
		if info, ok := s.Context.Spec["info"].(map[string]interface{}); ok {
			if title, ok := info["title"].(string); ok {
				sb.WriteString(fmt.Sprintf("API Spec: %s\n", title))
			}
		}
	}

	if s.Context.Results != nil {
		sb.WriteString("Results: Loaded\n")
	}

	return sb.String()
}

// SaveSession saves the session to a file
func (s *ChatSession) SaveSession(outputDir string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("%s-session.json", s.ID)
	filePath := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write session file: %w", err)
	}

	return filePath, nil
}

// LoadSession loads a session from a file
func LoadSession(filePath string, llmClient *llm.Client, vectorStore *vectordb.Store) (*ChatSession, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Reinitialize runtime fields
	session.llmClient = llmClient
	session.vectorStore = vectorStore
	session.feedbackAgent = autonomous.NewFeedbackAgent(
		llmClient, nil, nil, nil, vectorStore,
	)

	return &session, nil
}

// ReloadContext reloads the workflow context from files
func (s *ChatSession) ReloadContext() error {
	return s.LoadContext()
}

// RefineTestSuite uses the Feedback Agent to refine the test suite based on user feedback
func (s *ChatSession) RefineTestSuite(ctx context.Context, feedback string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedbackAgent == nil {
		return "", fmt.Errorf("feedback agent not initialized - LLM client required")
	}

	if s.Context.TestSuite == nil {
		return "", fmt.Errorf("no test suite loaded - use /load first")
	}

	// Build a refinement prompt
	prompt := fmt.Sprintf(`Please refine the test suite based on this feedback:

%s

Current test suite has %d tests. Use the edit_test_case, add_test_case, or remove_test_case tools to make improvements.

After making changes, summarize what was modified.`, feedback, len(s.Context.TestSuite.Tests))

	// Process through feedback agent
	response, err := s.feedbackAgent.ProcessInteractiveFeedback(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("refinement failed: %w", err)
	}

	// Reload context to pick up changes
	s.mu.Unlock()
	if err := s.LoadContext(); err != nil {
		s.mu.Lock()
		return response, fmt.Errorf("refinement applied but failed to reload: %w", err)
	}
	s.mu.Lock()

	return response, nil
}

// RunTests executes the test suite and returns results
func (s *ChatSession) RunTests(ctx context.Context, baseURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	suitePath := s.Context.FilePaths["suite"]
	if suitePath == "" {
		return "", fmt.Errorf("no test suite path found - use /load first")
	}

	// Create executor
	executor := autonomous.NewPlannedExecutor(s.llmClient, baseURL)

	// Run the suite
	result, err := executor.RunSavedSuite(ctx, suitePath)
	if err != nil {
		return "", fmt.Errorf("test execution failed: %w", err)
	}

	// Save results
	resultsDir := "./output/results"
	if err := os.MkdirAll(resultsDir, 0755); err == nil {
		resultsPath := filepath.Join(resultsDir, fmt.Sprintf("%s-test-results.json", s.WorkflowID))
		if data, err := json.MarshalIndent(result, "", "  "); err == nil {
			os.WriteFile(resultsPath, data, 0644)
			s.Context.FilePaths["results"] = resultsPath
		}
	}

	// Build summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Test Execution Complete\n\n"))
	sb.WriteString(fmt.Sprintf("**Total Tests:** %d\n", result.Summary.TotalTests))
	sb.WriteString(fmt.Sprintf("**Passed:** %d\n", result.Summary.Passed))
	sb.WriteString(fmt.Sprintf("**Failed:** %d\n", result.Summary.Failed))
	sb.WriteString(fmt.Sprintf("**Skipped:** %d\n", result.Summary.Skipped))
	sb.WriteString(fmt.Sprintf("**Duration:** %v\n\n", result.Summary.TotalTime))

	passRate := 0.0
	if result.Summary.TotalTests > 0 {
		passRate = float64(result.Summary.Passed) / float64(result.Summary.TotalTests) * 100
	}
	sb.WriteString(fmt.Sprintf("**Pass Rate:** %.1f%%\n\n", passRate))

	// List failures
	if result.Summary.Failed > 0 {
		sb.WriteString("### Failed Tests\n")
		for _, r := range result.Results {
			if !r.Passed && !r.Skipped {
				sb.WriteString(fmt.Sprintf("- **%s %s**: %s (Status: %d)\n",
					r.Test.Method, r.Test.Path, r.Error, r.StatusCode))
			}
		}
	}

	return sb.String(), nil
}

// GenerateRecommendations generates improvement recommendations for the test suite
func (s *ChatSession) GenerateRecommendations(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedbackAgent == nil {
		return "", fmt.Errorf("feedback agent not initialized - LLM client required")
	}

	prompt := `Analyze the current test suite and results, then use the generate_recommendations tool to provide improvement suggestions.

Focus on:
1. Coverage gaps
2. Reliability issues
3. Performance concerns
4. Security testing

Provide actionable recommendations.`

	response, err := s.feedbackAgent.ProcessInteractiveFeedback(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("recommendation generation failed: %w", err)
	}

	return response, nil
}
