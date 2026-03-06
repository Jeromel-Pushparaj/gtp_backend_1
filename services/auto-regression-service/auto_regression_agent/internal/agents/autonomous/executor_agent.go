package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// ExecutorAgent is a truly autonomous AI agent that executes API tests
// It receives tasks via Redis, uses AI with tools to execute tests autonomously,
// and publishes results back via Redis
type ExecutorAgent struct {
	*Agent
	httpClient      *HTTPClient
	authMgr         *AuthManager
	testContext     *TestContext
	maxTurns        int           // Maximum AI turns per execution session
	executionMode   ExecutionMode // "autonomous" or "planned"
	plannedExecutor *PlannedExecutor
}

// NewExecutorAgent creates a new autonomous executor agent
func NewExecutorAgent(
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *ExecutorAgent {
	baseAgent := NewAgent(
		"executor_agent",
		AgentTypeExecutor,
		[]string{"test_execution", "http_testing", "autonomous_ai"},
		llmClient,
		eventBus,
		messageBus,
		consensus,
	)

	// Default to planned mode (much faster/cheaper)
	// Can be changed via SetExecutionMode or EXECUTION_MODE env var
	mode := ExecutionMode(os.Getenv("EXECUTION_MODE"))
	if mode == "" {
		mode = ModePlanned // Default to planned mode
	}

	return &ExecutorAgent{
		Agent:         baseAgent,
		httpClient:    NewHTTPClient(),
		authMgr:       NewAuthManager(),
		testContext:   NewTestContext(),
		maxTurns:      50, // Allow up to 50 AI turns per session
		executionMode: mode,
	}
}

// SetExecutionMode sets the execution mode (autonomous or planned)
func (ea *ExecutorAgent) SetExecutionMode(mode ExecutionMode) {
	ea.executionMode = mode
	log.Printf("🔧 Execution mode set to: %s", mode)
}

// Start starts the executor agent - it will listen for events via Redis
func (ea *ExecutorAgent) Start(ctx context.Context) error {
	log.Printf("🤖 Starting Autonomous Executor Agent...")

	// Start base agent (sets up Redis message listeners)
	if err := ea.Agent.Start(ctx); err != nil {
		return err
	}

	// Subscribe to payloads_ready events via Redis
	go ea.listenToPayloadsReady(ctx)

	log.Printf("✅ Executor Agent ready - listening for tasks via Redis")
	return nil
}

// listenToPayloadsReady listens for payloads ready events from Redis
func (ea *ExecutorAgent) listenToPayloadsReady(ctx context.Context) {
	err := ea.EventBus.Subscribe(ctx, events.EventTypePayloadsReady, func(event *events.Event) error {
		log.Printf("🤖 Executor Agent received task via Redis: payloads_ready")

		ea.setState(AgentStateProcessing)
		defer ea.setState(AgentStateIdle)

		specID, ok := event.Payload["spec_id"].(string)
		if !ok {
			return fmt.Errorf("spec_id not found in event payload")
		}

		workflowID, ok := event.Payload["workflow_id"].(string)
		if !ok {
			return fmt.Errorf("workflow_id not found in event payload")
		}

		payloads, ok := event.Payload["payloads"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("payloads not found in event payload")
		}

		baseURL, _ := event.Payload["base_url"].(string)
		if baseURL == "" {
			log.Printf("Warning: base_url not found in event payload")
		}

		// Execute autonomously using AI
		return ea.executeAutonomously(ctx, specID, workflowID, baseURL, payloads)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to payloads_ready: %v", err)
	}
}

// executeAutonomously gives the AI full control over test execution
// The AI receives all test cases and decides how to execute them
func (ea *ExecutorAgent) executeAutonomously(ctx context.Context, specID, workflowID, baseURL string, payloads map[string]interface{}) error {
	// Reset test context
	ea.testContext.Reset()

	// Load spec for auth context
	specPath := fmt.Sprintf("./output/specs/%s.json", specID)
	specData, err := os.ReadFile(specPath)
	if err != nil {
		log.Printf("Warning: failed to read spec file: %v", err)
	} else {
		var spec map[string]interface{}
		if err := json.Unmarshal(specData, &spec); err == nil {
			if err := ea.authMgr.LoadSecuritySchemes(spec); err != nil {
				log.Printf("Warning: failed to load security schemes: %v", err)
			}
			ea.authMgr.LoadCredentialsFromEnv()
		}
	}

	// Dispatch based on execution mode
	if ea.executionMode == ModePlanned {
		return ea.executePlannedMode(ctx, specID, workflowID, baseURL, payloads)
	}

	// Default to autonomous mode
	return ea.executeFullyAutonomous(ctx, specID, workflowID, baseURL, payloads)
}

// executeFullyAutonomous runs in fully autonomous AI mode (original behavior)
func (ea *ExecutorAgent) executeFullyAutonomous(ctx context.Context, specID, workflowID, baseURL string, payloads map[string]interface{}) error {
	log.Printf("🤖 [AUTONOMOUS MODE] Executor Agent starting for spec: %s", specID)

	// Serialize payloads for AI
	payloadsJSON, _ := json.MarshalIndent(payloads, "", "  ")

	// Build the autonomous AI prompt
	systemPrompt := ea.buildAutonomousSystemPrompt(baseURL)
	userPrompt := fmt.Sprintf(`You have received a test execution task. Execute ALL the following test cases autonomously.

## Test Payloads
%s

## Your Mission
1. Execute each test case by making HTTP requests
2. Validate responses against expected outcomes
3. Store any IDs or values needed for subsequent tests (use store_context)
4. Report the final result when ALL tests are complete

Begin execution now. Start with the first test case.`, string(payloadsJSON))

	// Create tool executor with all tools the AI needs
	toolExecutor := NewToolExecutor(ea.httpClient, ea.authMgr, ea.testContext, baseURL)

	// Run the autonomous AI loop
	result := ea.runAutonomousLoop(ctx, systemPrompt, userPrompt, toolExecutor)

	// Extract results from AI execution
	summary := map[string]interface{}{
		"executed_at":    time.Now().Format(time.RFC3339),
		"ai_autonomous":  true,
		"total_turns":    result["turns"],
		"final_status":   result["status"],
		"tests_executed": result["tests_executed"],
		"tests_passed":   result["tests_passed"],
		"tests_failed":   result["tests_failed"],
	}

	// Save results
	outputPath := fmt.Sprintf("./output/results/%s-test-results.json", specID)
	if err := os.MkdirAll("./output/results", 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fullResults := map[string]interface{}{
		"summary":      summary,
		"test_results": result["test_results"],
		"ai_reasoning": result["reasoning"],
	}

	resultsJSON, err := json.MarshalIndent(fullResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(outputPath, resultsJSON, 0644); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	log.Printf("🤖 Autonomous execution complete in %d turns", result["turns"])

	// Build results map for analyzer agent
	results := map[string]interface{}{
		"test_results": result["test_results"],
		"reasoning":    result["reasoning"],
	}

	// Publish completion event via Redis
	return ea.PublishEvent(ctx, events.EventTypeTestsComplete, map[string]interface{}{
		"spec_id":     specID,
		"workflow_id": workflowID,
		"summary":     summary,
		"results":     results,
		"output_path": outputPath,
	})
}

// buildAutonomousSystemPrompt creates the system prompt for autonomous execution
func (ea *ExecutorAgent) buildAutonomousSystemPrompt(baseURL string) string {
	return fmt.Sprintf(`You are an autonomous API test execution agent. You have FULL CONTROL over test execution.

## Your Capabilities
You can use these tools to execute tests:

1. **http_request** - Make HTTP requests to APIs
   - method: GET, POST, PUT, DELETE, PATCH
   - path: API endpoint path (will be appended to base URL: %s)
   - headers: Optional headers as JSON object
   - body: Optional request body as JSON object

2. **store_context** - Store values for use in later tests
   - key: A unique key name
   - value: The value to store (e.g., an ID from a response)

3. **get_context** - Retrieve a previously stored value
   - key: The key to retrieve

4. **report_result** - Report a test result (call this for EACH test)
   - test_name: Name of the test
   - passed: true/false
   - reason: Why the test passed or failed
   - response_data: The actual response received

5. **complete_execution** - Call when ALL tests are done
   - summary: Summary of all test results

## Execution Rules
1. Execute tests in a logical order (create before read, etc.)
2. Store IDs from create operations for use in subsequent tests
3. Validate status codes and response bodies
4. Report each test result as you complete it
5. Call complete_execution when finished

## Base URL: %s

You are fully autonomous. Make decisions, execute tests, handle errors, and complete the task.`, baseURL, baseURL)
}

// runAutonomousLoop runs the AI agent loop until completion
func (ea *ExecutorAgent) runAutonomousLoop(ctx context.Context, systemPrompt, userPrompt string, toolExecutor *ToolExecutor) map[string]interface{} {
	// Combine system prompt and user prompt for first message
	combinedPrompt := systemPrompt + "\n\n---\n\n" + userPrompt

	messages := []llm.AgentMessage{
		{Role: "user", Text: combinedPrompt},
	}

	testResults := make([]map[string]interface{}, 0)
	reasoning := make([]string, 0)
	testsExecuted := 0
	testsPassed := 0
	testsFailed := 0
	completed := false

	// Build tool definitions for the LLM
	tools := ea.buildToolsForLLM()

	for turn := 0; turn < ea.maxTurns && !completed; turn++ {
		log.Printf("🤖 AI Turn %d", turn+1)

		// Call LLM with tools
		opts := llm.AgentCompletionOptions{
			Temperature: 0.2,
			MaxTokens:   4096,
			Tools:       tools,
		}

		response, err := ea.LLMClient.GenerateAgentCompletion(ctx, messages, opts)
		if err != nil {
			log.Printf("❌ LLM error: %v", err)
			reasoning = append(reasoning, fmt.Sprintf("Turn %d: LLM error - %v", turn+1, err))
			break
		}

		// Check for function calls
		if len(response.FunctionCalls) == 0 {
			// No function calls - AI is done or confused
			if response.Text != "" {
				reasoning = append(reasoning, response.Text)
			}
			log.Printf("🤖 AI response (no tools): %s", truncate(response.Text, 100))
			break
		}

		// For thinking models with function calling:
		// 1. Add ALL function calls as a single assistant message (preserving thought signatures)
		// 2. Execute all tools and collect responses
		// 3. Add ALL function responses as tool messages
		// This follows the OpenAI-compatible function calling pattern

		// Step 1: Add all function calls as a single model message
		modelFunctionCalls := make([]llm.FunctionCall, 0, len(response.FunctionCalls))
		for _, fc := range response.FunctionCalls {
			log.Printf("🔧 Tool call: %s", fc.Name)
			modelFunctionCalls = append(modelFunctionCalls, llm.FunctionCall{
				Name:             fc.Name,
				Args:             fc.Args,
				ThoughtSignature: fc.ThoughtSignature, // Only first one will have this
			})
		}
		messages = append(messages, llm.AgentMessage{
			Role:          "model",
			FunctionCalls: modelFunctionCalls,
		})

		// Step 2: Execute all tools and collect responses
		functionResponses := make([]llm.FunctionResponse, 0, len(response.FunctionCalls))
		for _, fc := range response.FunctionCalls {
			if fc.Name == "complete_execution" {
				completed = true
				reasoning = append(reasoning, fmt.Sprintf("Turn %d: Execution completed", turn+1))
				functionResponses = append(functionResponses, llm.FunctionResponse{
					Name:     fc.Name,
					Response: map[string]interface{}{"status": "completed"},
				})
				continue
			}

			if fc.Name == "report_result" {
				testsExecuted++
				result := map[string]interface{}{
					"test_name":     fc.Args["test_name"],
					"passed":        fc.Args["passed"],
					"reason":        fc.Args["reason"],
					"response_data": fc.Args["response_data"],
				}
				testResults = append(testResults, result)

				if passed, ok := fc.Args["passed"].(bool); ok && passed {
					testsPassed++
				} else {
					testsFailed++
				}

				functionResponses = append(functionResponses, llm.FunctionResponse{
					Name:     fc.Name,
					Response: map[string]interface{}{"status": "recorded"},
				})
				continue
			}

			// Execute other tools
			toolCall := ToolCall{Name: fc.Name, Args: fc.Args}
			toolResult := toolExecutor.ExecuteTool(ctx, toolCall)
			functionResponses = append(functionResponses, llm.FunctionResponse{
				Name:     fc.Name,
				Response: toolResult,
			})
			reasoning = append(reasoning, fmt.Sprintf("Turn %d: Called %s", turn+1, fc.Name))
		}

		// Step 3: Add all function responses as a single user message
		messages = append(messages, llm.AgentMessage{
			Role:              "user",
			FunctionResponses: functionResponses,
		})

		// Check if we completed
		if completed {
			break
		}
	}

	return map[string]interface{}{
		"turns":          len(reasoning),
		"status":         map[bool]string{true: "completed", false: "incomplete"}[completed],
		"tests_executed": testsExecuted,
		"tests_passed":   testsPassed,
		"tests_failed":   testsFailed,
		"test_results":   testResults,
		"reasoning":      reasoning,
	}
}

// buildToolsForLLM converts tool definitions to the format expected by LLM
func (ea *ExecutorAgent) buildToolsForLLM() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "http_request",
			"description": "Make an HTTP request to an API endpoint",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"method":  map[string]interface{}{"type": "string", "enum": []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
					"path":    map[string]interface{}{"type": "string", "description": "API endpoint path"},
					"headers": map[string]interface{}{"type": "object", "description": "Request headers"},
					"body":    map[string]interface{}{"type": "object", "description": "Request body"},
				},
				"required": []string{"method", "path"},
			},
		},
		{
			"name":        "store_context",
			"description": "Store a value for use in later tests",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":   map[string]interface{}{"type": "string"},
					"value": map[string]interface{}{"type": "string"},
				},
				"required": []string{"key", "value"},
			},
		},
		{
			"name":        "get_context",
			"description": "Retrieve a previously stored value",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{"type": "string"},
				},
				"required": []string{"key"},
			},
		},
		{
			"name":        "report_result",
			"description": "Report the result of a test case",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"test_name":     map[string]interface{}{"type": "string"},
					"passed":        map[string]interface{}{"type": "boolean"},
					"reason":        map[string]interface{}{"type": "string"},
					"response_data": map[string]interface{}{"type": "object"},
				},
				"required": []string{"test_name", "passed", "reason"},
			},
		},
		{
			"name":        "complete_execution",
			"description": "Signal that all tests have been completed",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary": map[string]interface{}{"type": "string"},
				},
				"required": []string{"summary"},
			},
		},
	}
}

// executePlannedMode uses AI for planning only, then executes deterministically
func (ea *ExecutorAgent) executePlannedMode(ctx context.Context, specID, workflowID, baseURL string, payloads map[string]interface{}) error {
	log.Printf("🧠 [PLANNED MODE] Executor Agent starting for spec: %s", specID)
	log.Printf("📋 Mode: AI plans once, then deterministic execution (1 LLM call)")

	// Load strategy if available
	strategyPath := fmt.Sprintf("./output/strategy/%s-test-strategy.json", specID)
	var strategy map[string]interface{}
	if strategyData, err := os.ReadFile(strategyPath); err == nil {
		json.Unmarshal(strategyData, &strategy)
	}

	// Load context type mappings from AI analysis
	// LLM returns {canonical: [aliases]}, we need to invert to {alias: canonical}
	var contextMappings map[string]string
	analysisPath := fmt.Sprintf("./output/discovery/%s-ai-analysis.json", specID)
	if analysisData, err := os.ReadFile(analysisPath); err == nil {
		var analysis map[string]interface{}
		if err := json.Unmarshal(analysisData, &analysis); err == nil {
			if mappings, ok := analysis["context_type_mappings"].(map[string]interface{}); ok {
				contextMappings = make(map[string]string)
				for canonical, aliasesRaw := range mappings {
					// Handle array of aliases: {pet: [animal, resource, item]}
					if aliases, ok := aliasesRaw.([]interface{}); ok {
						for _, alias := range aliases {
							if aliasStr, ok := alias.(string); ok {
								contextMappings[aliasStr] = canonical
							}
						}
					}
					// Also handle direct string mapping: {animal: pet}
					if aliasStr, ok := aliasesRaw.(string); ok {
						contextMappings[canonical] = aliasStr
					}
				}
				log.Printf("📋 Loaded %d context type mappings from AI analysis", len(contextMappings))
			}
		}
	}

	// Create planned executor
	plannedExecutor := NewPlannedExecutor(ea.LLMClient, baseURL)
	plannedExecutor.SetAuthManager(ea.authMgr)

	// Set context mappings if available
	if len(contextMappings) > 0 {
		plannedExecutor.SetContextMappings(contextMappings)
	}

	// Phase 1: AI creates execution plan (single LLM call)
	plan, err := plannedExecutor.CreateExecutionPlan(ctx, specID, workflowID, strategy, payloads)
	if err != nil {
		return fmt.Errorf("failed to create execution plan: %w", err)
	}

	// Save execution plan
	planPath := fmt.Sprintf("./output/plans/%s-execution-plan.json", specID)
	if err := os.MkdirAll("./output/plans", 0755); err == nil {
		planJSON, _ := json.MarshalIndent(plan, "", "  ")
		os.WriteFile(planPath, planJSON, 0644)
		log.Printf("📄 Execution plan saved: %s", planPath)
	}

	// Convert and save as reusable test suite
	specName := specID // Use spec ID as name, could be improved with actual spec name
	suite := ConvertFromExecutionPlan(plan, specName, baseURL)
	suitePath, err := SaveTestSuite(suite, "./test_suites")
	if err != nil {
		log.Printf("⚠️  Failed to save test suite: %v", err)
	} else {
		log.Printf("📦 Test suite saved: %s (ID: %s)", suitePath, suite.ID)
	}

	// Phase 2: Execute plan deterministically (no LLM calls)
	result, err := plannedExecutor.ExecutePlan(ctx, plan)
	if err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}

	// Save results
	outputPath := fmt.Sprintf("./output/results/%s-test-results.json", specID)
	if err := os.MkdirAll("./output/results", 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	resultsJSON, _ := json.MarshalIndent(result, "", "  ")
	if err := os.WriteFile(outputPath, resultsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}
	log.Printf("📄 Results saved: %s", outputPath)

	// Build summary for analyzer
	summary := map[string]interface{}{
		"executed_at":    result.ExecutedAt.Format(time.RFC3339),
		"ai_planned":     true,
		"execution_mode": "planned",
		"llm_calls":      result.Summary.LLMCalls,
		"total_tests":    result.Summary.TotalTests,
		"tests_passed":   result.Summary.Passed,
		"tests_failed":   result.Summary.Failed,
		"tests_skipped":  result.Summary.Skipped,
		"pass_rate":      result.Summary.PassRate,
		"duration_ms":    result.Duration.Milliseconds(),
	}

	// Build test results for analyzer
	testResults := make([]map[string]interface{}, len(result.Results))
	for i, r := range result.Results {
		testResults[i] = map[string]interface{}{
			"name":          r.Test.Name,
			"method":        r.Test.Method,
			"path":          r.Test.Path,
			"passed":        r.Passed,
			"skipped":       r.Skipped,
			"status_code":   r.StatusCode,
			"expected":      r.Test.ExpectedStatus,
			"response_time": r.ResponseTime.Milliseconds(),
			"error":         r.Error,
		}
	}

	// Publish tests_complete event
	results := map[string]interface{}{
		"test_results": testResults,
		"reasoning":    plan.AIReasoning,
	}

	return ea.PublishEvent(ctx, events.EventTypeTestsComplete, map[string]interface{}{
		"spec_id":     specID,
		"workflow_id": workflowID,
		"summary":     summary,
		"results":     results,
		"output_path": outputPath,
	})
}

// truncate helper
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
