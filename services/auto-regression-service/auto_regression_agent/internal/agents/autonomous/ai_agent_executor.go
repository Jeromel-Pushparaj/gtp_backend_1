package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// AIAgentExecutor executes tests using AI with tool calling
type AIAgentExecutor struct {
	llmClient    *llm.Client
	toolExecutor *ToolExecutor
	maxTurns     int // Maximum number of AI turns to prevent infinite loops
}

// NewAIAgentExecutor creates a new AI agent executor
func NewAIAgentExecutor(llmClient *llm.Client, toolExecutor *ToolExecutor) *AIAgentExecutor {
	return &AIAgentExecutor{
		llmClient:    llmClient,
		toolExecutor: toolExecutor,
		maxTurns:     10, // Default max turns
	}
}

// ExecuteTestCase executes a single test case using the AI agent
func (ae *AIAgentExecutor) ExecuteTestCase(ctx context.Context, testCase map[string]interface{}, endpoint string, availableContext map[string]interface{}) map[string]interface{} {
	startTime := time.Now()

	// Build the system prompt
	systemPrompt := ae.buildSystemPrompt(testCase, endpoint, availableContext)

	// Convert tools to JSON format for the API
	tools, err := ToolsToJSON(AllAgentTools())
	if err != nil {
		return ae.buildErrorResult(testCase, fmt.Sprintf("failed to convert tools: %v", err), startTime)
	}

	// Initialize conversation
	messages := []llm.AgentMessage{
		{Role: "user", Text: systemPrompt},
	}

	// Agent loop
	for turn := 0; turn < ae.maxTurns; turn++ {
		log.Printf("🤖 AI Agent turn %d", turn+1)

		// Call the LLM
		response, err := ae.llmClient.GenerateAgentCompletion(ctx, messages, llm.AgentCompletionOptions{
			Temperature: 0.1,
			MaxTokens:   4096,
			Tools:       tools,
		})

		if err != nil {
			return ae.buildErrorResult(testCase, fmt.Sprintf("LLM call failed: %v", err), startTime)
		}

		// Check if AI returned text (no function call)
		if len(response.FunctionCalls) == 0 {
			log.Printf("🤖 AI Agent completed with text: %s", truncateString(response.Text, 100))

			// Check if we have a reported result
			if result := ae.toolExecutor.GetReportedResult(); result != nil {
				return ae.buildSuccessResult(testCase, result, startTime)
			}

			// No result reported - this is an error
			return ae.buildErrorResult(testCase, "AI completed without reporting a result", startTime)
		}

		// Process each function call
		for _, funcCall := range response.FunctionCalls {
			log.Printf("🔧 AI calling tool: %s", funcCall.Name)

			// Execute the tool
			toolResult := ae.toolExecutor.ExecuteTool(ctx, ToolCall{
				Name: funcCall.Name,
				Args: funcCall.Args,
			})

			// Add function call to messages (from model) - include ThoughtSignature for thinking models
			messages = append(messages, llm.AgentMessage{
				Role: "model",
				FunctionCall: &llm.FunctionCall{
					Name:             funcCall.Name,
					Args:             funcCall.Args,
					ThoughtSignature: funcCall.ThoughtSignature,
				},
			})

			// Add function response to messages (from user/function)
			resultData := map[string]interface{}{
				"result": toolResult.Result,
			}
			if toolResult.Error != "" {
				resultData["error"] = toolResult.Error
			}

			messages = append(messages, llm.AgentMessage{
				Role: "user",
				FunctionResponse: &llm.FunctionResponse{
					Name:     toolResult.Name,
					Response: resultData,
				},
			})

			// If this was report_result, we're done
			if funcCall.Name == "report_result" {
				if result := ae.toolExecutor.GetReportedResult(); result != nil {
					result.ExecutionTime = time.Since(startTime)
					return ae.buildSuccessResult(testCase, result, startTime)
				}
			}
		}
	}

	// Max turns reached without completion
	return ae.buildErrorResult(testCase, fmt.Sprintf("max turns (%d) reached without completion", ae.maxTurns), startTime)
}

// buildSystemPrompt creates the prompt for the AI agent
func (ae *AIAgentExecutor) buildSystemPrompt(testCase map[string]interface{}, endpoint string, availableContext map[string]interface{}) string {
	testCaseJSON, _ := json.MarshalIndent(testCase, "", "  ")
	contextJSON, _ := json.MarshalIndent(availableContext, "", "  ")

	return fmt.Sprintf(`You are an API test execution agent. Your job is to execute the following test case and report the result.

## Test Case
%s

## Endpoint
%s

## Available Context (from previous test executions)
%s

## Instructions
1. First, check if the test case requires any context data (look for {{CONTEXT:type.field}} markers)
2. If context is needed, use the get_context tool to retrieve it
3. Execute the HTTP request using the http_request tool
4. If the test creates resources, use store_context to save important data for future tests
5. Finally, call report_result to report whether the test passed or failed

## Important Rules
- Replace any {{CONTEXT:type.field}} placeholders with actual values from context
- A test PASSES if the actual status code matches the expected status code
- A test FAILS if the status codes don't match
- If you cannot resolve required context, report the test as SKIPPED with an explanation
- Always call report_result at the end

Begin executing the test now.`, string(testCaseJSON), endpoint, string(contextJSON))
}

// buildErrorResult creates an error result for a test case
func (ae *AIAgentExecutor) buildErrorResult(testCase map[string]interface{}, errMsg string, startTime time.Time) map[string]interface{} {
	name, _ := testCase["name"].(string)
	expectedStatus, _ := testCase["expected_status"].(float64)
	method, _ := testCase["method"].(string)
	path, _ := testCase["path"].(string)

	return map[string]interface{}{
		"name":            name,
		"passed":          false,
		"error":           errMsg,
		"expected_status": int(expectedStatus),
		"actual_status":   0,
		"method":          method,
		"path":            path,
		"execution_time":  time.Since(startTime).Milliseconds(),
		"timestamp":       time.Now().Format(time.RFC3339),
		"ai_agent":        true,
	}
}

// buildSuccessResult creates a result from the AI's reported result
func (ae *AIAgentExecutor) buildSuccessResult(testCase map[string]interface{}, result *TestResult, startTime time.Time) map[string]interface{} {
	name, _ := testCase["name"].(string)
	method, _ := testCase["method"].(string)
	path, _ := testCase["path"].(string)

	return map[string]interface{}{
		"name":            name,
		"passed":          result.Passed,
		"expected_status": result.ExpectedStatus,
		"actual_status":   result.ActualStatus,
		"reason":          result.Reason,
		"method":          method,
		"path":            path,
		"execution_time":  time.Since(startTime).Milliseconds(),
		"timestamp":       time.Now().Format(time.RFC3339),
		"ai_agent":        true,
	}
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
