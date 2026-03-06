package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

// AgentTypeFeedback is the feedback agent type
const AgentTypeFeedback AgentType = "feedback"

// feedbackAgentSystemPrompt is the comprehensive system prompt for the Feedback Agent
const feedbackAgentSystemPrompt = `You are the Feedback Agent - an AI assistant that helps users refine API tests.

## Your Capabilities

### Memory & Learning
You have access to a vector database storing:
- **Learned Patterns**: Authentication flows, edge cases, business rules, API patterns
- **Failure Patterns**: Common errors and their fixes
- **Successful Strategies**: Test approaches that have worked well

### Test Suite Operations
You can:
- **Read** test suites, results, and API specs
- **Edit** existing test cases (modify payloads, expected status, headers)
- **Add** new test cases based on user requests
- **Remove** test cases that are redundant or incorrect
- **Update** context capture settings for data flow between tests
- **Execute** tests directly to validate fixes before committing changes

### Available Tools

#### Vector DB Tools (Memory)
- store_learned_pattern: Store patterns for future reference
- search_similar_patterns: Find relevant past learnings
- store_failure_pattern: Store failure patterns with fixes
- search_failure_fixes: Find fixes for similar failures
- store_successful_strategy: Store successful test strategies
- search_strategies: Find proven test approaches
- update_learning_confidence: Boost confidence of useful learnings

#### Read Tools
- read_test_suite: Read the current test suite
- read_test_results: Read test execution results
- read_spec: Read the OpenAPI specification
- list_output_files: List available files for a workflow

#### Write Tools
- edit_test_case: Modify an existing test case
- add_test_case: Add a new test case
- remove_test_case: Remove a test case
- update_test_context: Update context capture settings

#### Analysis Tools
- generate_recommendations: Analyze results and suggest improvements

#### Execute Tools (Run Tests)
- execute_single_test: Run one specific test case by ID, order, or name
- execute_test_subset: Run multiple tests matching filter criteria (methods, paths, tags)
- execute_failed_tests: Re-run only tests that failed in a previous run
- execute_with_context: Run a test with pre-populated context values

## Behavioral Guidelines

1. **Always search your memory first** before suggesting changes
2. **Explain your reasoning** before making modifications
3. **Backups are automatically created** before modifications
4. **Confirm destructive operations** before removing tests
5. **Learn from corrections** - store patterns when users correct you

## Response Format

When making changes:
1. Acknowledge the user's request
2. Search for relevant patterns in memory
3. Explain what you're going to do
4. Execute the tools
5. Summarize what was changed

## Example Interactions

**User**: "The login test is failing with 401"
**You**: I see the login test is failing. Let me search for similar authentication failures...
[search_failure_fixes]
Based on my memory, 401 errors often occur when the Authorization header is missing or malformed.
Let me check the current test configuration...
[read_test_suite]
I notice the test doesn't include the Authorization header. Would you like me to add it?

**User**: "Yes, add the auth header"
**You**: I'll update the login test to include the Authorization header.
[edit_test_case]
Done! I've updated test #1 to include the Bearer token. I've also stored this pattern for future reference.
[store_learned_pattern]

**User**: "What tests should I add for better coverage?"
**You**: Let me analyze your current test suite and results...
[generate_recommendations]
Based on the analysis, I recommend:
1. Adding negative tests for the DELETE endpoint
2. Testing rate limiting scenarios
3. Adding boundary tests for input validation
Would you like me to implement any of these?

**User**: "I fixed the payload, can you run just that test again?"
**You**: Sure! Let me run just that specific test to verify the fix...
[execute_single_test]
The test passed! Status 200 as expected. The fix worked. Would you like me to run the full suite to make sure nothing else broke?

**User**: "Run all the failed tests again"
**You**: I'll re-run all the tests that failed in the previous run...
[execute_failed_tests]
Great news! 3 out of 4 previously failing tests now pass. One test is still failing - the DELETE test returns 403 instead of 204. Would you like me to investigate?

## Important Notes
- Always be helpful and proactive in suggesting improvements
- When you fix an issue, offer to store it as a learned pattern
- If you're unsure about a change, ask for confirmation
- Provide clear explanations of what each tool does
`

// FeedbackAgent is an autonomous agent that learns from user feedback and
// refines test generation using vector-based memory
type FeedbackAgent struct {
	*Agent
	vectorStore  *vectordb.Store
	toolExecutor *FeedbackToolExecutor
}

// NewFeedbackAgent creates a new feedback agent
func NewFeedbackAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
	vectorStore *vectordb.Store,
) *FeedbackAgent {
	baseAgent := NewAgent(
		"feedback_agent",
		AgentTypeFeedback,
		[]string{"learning", "pattern_recognition", "feedback_processing", "memory"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	toolExecutor := NewFeedbackToolExecutor(vectorStore)

	return &FeedbackAgent{
		Agent:        baseAgent,
		vectorStore:  vectorStore,
		toolExecutor: toolExecutor,
	}
}

// Start starts the feedback agent
func (fa *FeedbackAgent) Start(ctx context.Context) error {
	log.Printf("🧠 Starting Feedback Agent...")

	if err := fa.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to test results for learning
	go fa.listenToTestResults(ctx)

	// Subscribe to user feedback events
	go fa.listenToUserFeedback(ctx)

	// Subscribe to consensus requests
	go fa.listenToConsensusRequests(ctx)

	log.Printf("✅ Feedback Agent ready with %d tools", len(fa.toolExecutor.GetToolDefinitions()))
	return nil
}

// listenToTestResults listens for test completion events to learn from
func (fa *FeedbackAgent) listenToTestResults(ctx context.Context) {
	err := fa.EventBus.Subscribe(ctx, events.EventTypeTestsComplete, func(event *events.Event) error {
		log.Printf("🧠 Feedback Agent received test results for learning")
		fa.setState(AgentStateProcessing)
		defer fa.setState(AgentStateIdle)

		return fa.learnFromTestResults(ctx, event.Payload)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to test results: %v", err)
	}
}

// listenToUserFeedback listens for user feedback events
func (fa *FeedbackAgent) listenToUserFeedback(ctx context.Context) {
	err := fa.EventBus.Subscribe(ctx, events.EventTypeUserFeedback, func(event *events.Event) error {
		log.Printf("🧠 Feedback Agent received user feedback")
		fa.setState(AgentStateProcessing)
		defer fa.setState(AgentStateIdle)

		return fa.processFeedback(ctx, event.Payload)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to user feedback: %v", err)
	}
}

// listenToConsensusRequests listens for consensus requests via event bus
func (fa *FeedbackAgent) listenToConsensusRequests(ctx context.Context) {
	err := fa.EventBus.Subscribe(ctx, events.EventTypeConsensusRequest, func(event *events.Event) error {
		// Extract consensus request from event payload
		req := &events.ConsensusRequest{
			ID:           getStringFromPayload(event.Payload, "request_id"),
			DecisionType: getStringFromPayload(event.Payload, "decision_type"),
			Question:     getStringFromPayload(event.Payload, "question"),
		}

		if options, ok := event.Payload["options"].([]interface{}); ok {
			for _, opt := range options {
				if s, ok := opt.(string); ok {
					req.Options = append(req.Options, s)
				}
			}
		}

		if !fa.shouldParticipate(req) {
			return nil
		}

		vote, confidence, reasoning := fa.evaluateConsensusRequest(ctx, req)
		return fa.SubmitVote(ctx, req.ID, vote, confidence, reasoning)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to consensus: %v", err)
	}
}

// getStringFromPayload safely extracts a string from payload
func getStringFromPayload(payload map[string]interface{}, key string) string {
	if v, ok := payload[key].(string); ok {
		return v
	}
	return ""
}

// ProcessInteractiveFeedback processes feedback from interactive user session
func (fa *FeedbackAgent) ProcessInteractiveFeedback(ctx context.Context, userMessage string) (string, error) {
	log.Printf("🧠 Processing interactive feedback: %s", truncateString(userMessage, 100))

	// First, search for relevant existing learnings
	relevantLearnings, err := fa.searchRelevantContext(ctx, userMessage)
	if err != nil {
		log.Printf("Warning: failed to search relevant context: %v", err)
	}

	// Build prompt with context
	prompt := fa.buildFeedbackPrompt(userMessage, relevantLearnings)

	// Call LLM with tool definitions
	response, err := fa.callLLMWithTools(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	return response, nil
}

// searchRelevantContext searches for relevant learnings in vector store
func (fa *FeedbackAgent) searchRelevantContext(ctx context.Context, query string) ([]vectordb.LearningResult, error) {
	if fa.vectorStore == nil {
		return nil, nil
	}

	opts := vectordb.SearchOptions{Limit: 5, MinScore: 0.5}
	return fa.vectorStore.SearchSimilarLearnings(ctx, query, opts)
}

// buildFeedbackPrompt builds the prompt for processing feedback
func (fa *FeedbackAgent) buildFeedbackPrompt(userMessage string, learnings []vectordb.LearningResult) string {
	var sb strings.Builder

	// Use the comprehensive system prompt
	sb.WriteString(feedbackAgentSystemPrompt)
	sb.WriteString("\n---\n\n")

	// Add relevant prior learnings if any
	if len(learnings) > 0 {
		sb.WriteString("## Relevant Prior Learnings from Memory\n")
		for i, l := range learnings {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s (confidence: %.2f, relevance: %.2f)\n",
				i+1, l.Category, l.Content, l.Confidence, l.Score))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Current User Request\n")
	sb.WriteString(userMessage)
	sb.WriteString("\n\n")

	sb.WriteString("Please respond to the user's request. Use the appropriate tools to fulfill their needs.")

	return sb.String()
}

// callLLMWithTools calls the LLM with function calling for tools
func (fa *FeedbackAgent) callLLMWithTools(ctx context.Context, prompt string) (string, error) {
	// Get tool definitions
	tools := fa.toolExecutor.GetToolDefinitions()
	toolsJSON, _ := ToolsToJSON(tools)

	// Build the request with function calling
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"tools": []map[string]interface{}{
			{"function_declarations": toolsJSON},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.3,
			"maxOutputTokens": 4096,
		},
	}

	// Call LLM
	response, err := fa.executeLLMRequest(ctx, reqBody)
	if err != nil {
		return "", err
	}

	// Process response - handle tool calls if any
	return fa.processLLMResponse(ctx, response)
}

// executeLLMRequest executes the LLM request
func (fa *FeedbackAgent) executeLLMRequest(ctx context.Context, reqBody map[string]interface{}) (map[string]interface{}, error) {
	// For now, use direct completion and parse tool calls from response
	// This is a simplified version - in production you'd use the full OpenAI-compatible function calling API

	prompt := ""
	if contents, ok := reqBody["contents"].([]map[string]interface{}); ok && len(contents) > 0 {
		if parts, ok := contents[0]["parts"].([]map[string]string); ok && len(parts) > 0 {
			prompt = parts[0]["text"]
		}
	}

	// Add tool descriptions to prompt
	prompt += "\n\nAvailable tools (call by responding with JSON in format {\"tool\": \"name\", \"args\": {...}}):\n"
	for _, tool := range fa.toolExecutor.GetToolDefinitions() {
		prompt += fmt.Sprintf("- %s: %s\n", tool.Name, tool.Description)
	}

	response, err := fa.LLMClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   4096,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"text": response,
	}, nil
}

// processLLMResponse processes the LLM response and executes any tool calls
func (fa *FeedbackAgent) processLLMResponse(ctx context.Context, response map[string]interface{}) (string, error) {
	text, _ := response["text"].(string)

	// Try to parse tool calls from response
	toolCalls := fa.parseToolCalls(text)

	var results []string
	for _, call := range toolCalls {
		log.Printf("🔧 Executing feedback tool: %s", call.Name)
		result := fa.toolExecutor.ExecuteTool(ctx, call)
		if result.Error != "" {
			results = append(results, fmt.Sprintf("Tool %s error: %s", call.Name, result.Error))
		} else {
			resultJSON, _ := json.Marshal(result.Result)
			results = append(results, fmt.Sprintf("Tool %s result: %s", call.Name, string(resultJSON)))
		}
	}

	// Return combined response
	if len(results) > 0 {
		return text + "\n\nTool Results:\n" + strings.Join(results, "\n"), nil
	}

	return text, nil
}

// parseToolCalls attempts to parse tool calls from LLM response
func (fa *FeedbackAgent) parseToolCalls(text string) []ToolCall {
	var calls []ToolCall

	// Look for JSON objects in the response
	start := strings.Index(text, "{\"tool\"")
	if start == -1 {
		return calls
	}

	// Find matching brace
	depth := 0
	end := start
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end > start {
		var parsed struct {
			Tool string                 `json:"tool"`
			Args map[string]interface{} `json:"args"`
		}
		if err := json.Unmarshal([]byte(text[start:end]), &parsed); err == nil {
			calls = append(calls, ToolCall{Name: parsed.Tool, Args: parsed.Args})
		}
	}

	return calls
}

// learnFromTestResults analyzes test results and extracts learnings
func (fa *FeedbackAgent) learnFromTestResults(ctx context.Context, payload map[string]interface{}) error {
	log.Printf("🧠 Learning from test results...")

	// Extract summary and results
	summary, _ := payload["summary"].(map[string]interface{})
	results, _ := payload["results"].(map[string]interface{})

	// Build learning prompt
	payloadJSON, _ := json.MarshalIndent(map[string]interface{}{
		"summary": summary,
		"results": results,
	}, "", "  ")

	prompt := fmt.Sprintf(`Analyze these test results and extract learnings:

%s

For each interesting pattern, call the appropriate tool:
- store_learned_pattern for general patterns
- store_failure_pattern for failures with fixes
- store_successful_strategy for strategies that worked well

Focus on:
1. Authentication patterns that worked or failed
2. Common error patterns and their solutions
3. Effective test strategies
`, string(payloadJSON))

	_, err := fa.callLLMWithTools(ctx, prompt)
	return err
}

// processFeedback processes user feedback
func (fa *FeedbackAgent) processFeedback(ctx context.Context, payload map[string]interface{}) error {
	feedback, _ := payload["feedback"].(string)
	if feedback == "" {
		return nil
	}

	_, err := fa.ProcessInteractiveFeedback(ctx, feedback)
	return err
}

// shouldParticipate determines if the agent should participate in a consensus request
func (fa *FeedbackAgent) shouldParticipate(req *events.ConsensusRequest) bool {
	// Participate in strategy and learning-related decisions
	return req.DecisionType == "strategy" || req.DecisionType == "learning"
}

// evaluateConsensusRequest evaluates a consensus request and returns a vote
func (fa *FeedbackAgent) evaluateConsensusRequest(ctx context.Context, req *events.ConsensusRequest) (vote string, confidence float64, reasoning string) {
	// Search for relevant learnings
	learnings, err := fa.searchRelevantContext(ctx, req.Question)
	if err != nil || len(learnings) == 0 {
		// No relevant learnings, abstain or vote neutral
		if len(req.Options) > 0 {
			return req.Options[0], 0.3, "No relevant prior learnings found"
		}
		return "", 0.3, "No relevant prior learnings"
	}

	// Use the highest confidence learning to inform vote
	bestLearning := learnings[0]
	return req.Options[0], bestLearning.Confidence * bestLearning.Score,
		fmt.Sprintf("Based on prior learning: %s", bestLearning.Content)
}
