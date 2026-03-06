package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// VariableCapture uses LLM to intelligently capture variables from API responses
type VariableCapture struct {
	llmClient *llm.Client
	strategy  map[string]interface{} // Strategy from Designer Agent
}

// NewVariableCapture creates a new variable capture service
func NewVariableCapture(llmClient *llm.Client) *VariableCapture {
	return &VariableCapture{
		llmClient: llmClient,
	}
}

// SetStrategy sets the test strategy containing dependency mapping
func (vc *VariableCapture) SetStrategy(strategy map[string]interface{}) {
	vc.strategy = strategy
}

// CaptureRequest represents a request to capture variables from a response
type CaptureRequest struct {
	Endpoint     string                 // e.g., "POST /api/resources"
	Method       string                 // e.g., "POST"
	ResponseBody map[string]interface{} // Parsed JSON response
	StatusCode   int
}

// CaptureResult represents the result of variable capture
type CaptureResult struct {
	Variables map[string]interface{} // Captured variables with their values
	Metadata  map[string]string      // Metadata about each variable (type, purpose, etc.)
}

// CaptureVariables uses LLM to intelligently determine which fields to capture
func (vc *VariableCapture) CaptureVariables(ctx context.Context, req CaptureRequest) (*CaptureResult, error) {
	// First, check if strategy provides explicit mapping
	if vc.strategy != nil {
		if explicitVars := vc.getExplicitVariables(req.Endpoint, req.ResponseBody); len(explicitVars) > 0 {
			log.Printf("📌 Using explicit variable mapping from strategy for %s", req.Endpoint)
			return &CaptureResult{
				Variables: explicitVars,
				Metadata:  map[string]string{},
			}, nil
		}
	}

	// If no explicit mapping, use LLM to intelligently capture
	return vc.captureWithLLM(ctx, req)
}

// getExplicitVariables extracts variables based on strategy's resource_creators mapping
func (vc *VariableCapture) getExplicitVariables(endpoint string, responseBody map[string]interface{}) map[string]interface{} {
	variables := make(map[string]interface{})

	// Navigate to data_strategy.resource_creators
	dataStrategy, ok := vc.strategy["data_strategy"].(map[string]interface{})
	if !ok {
		return variables
	}

	resourceCreators, ok := dataStrategy["resource_creators"].(map[string]interface{})
	if !ok {
		return variables
	}

	// Check if this endpoint is a resource creator
	creatorInfo, ok := resourceCreators[endpoint].(map[string]interface{})
	if !ok {
		return variables
	}

	// Extract the ID field specified in the strategy
	idField, ok := creatorInfo["id_field"].(string)
	if !ok {
		return variables
	}

	// Capture the ID field from response
	if idValue, exists := responseBody[idField]; exists {
		resourceType, _ := creatorInfo["resource_type"].(string)
		variableName := fmt.Sprintf("%sId", resourceType)
		variables[variableName] = idValue
		log.Printf("✅ Captured %s = %v from %s", variableName, idValue, endpoint)
	}

	return variables
}

// captureWithLLM uses LLM to analyze response and determine what to capture
func (vc *VariableCapture) captureWithLLM(ctx context.Context, req CaptureRequest) (*CaptureResult, error) {
	responseJSON, _ := json.MarshalIndent(req.ResponseBody, "", "  ")

	prompt := fmt.Sprintf(`You are an API testing expert. Analyze this API response and determine which fields should be captured as variables for use in subsequent tests.

Endpoint: %s %s
Status Code: %d
Response Body:
%s

Instructions:
1. Identify fields that represent resource identifiers (IDs, UUIDs, keys)
2. Identify fields that might be needed in subsequent API calls
3. Consider common REST patterns (e.g., POST creates resource, returns ID for GET/PUT/DELETE)
4. Provide meaningful variable names based on the resource type (e.g., "resourceId", "entityId", "customerId")

CRITICAL FORMATTING RULES:
1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks
2. Do NOT wrap the JSON in markdown code blocks
3. Start your response directly with the opening brace {
4. End your response with the closing brace }

Required JSON structure:
{
  "variables": {
    "resourceId": 123,
    "name": "example-resource"
  },
  "metadata": {
    "resourceId": "Resource identifier for subsequent operations",
    "name": "Resource name for reference"
  }
}

If no variables should be captured, return:
{
  "variables": {},
  "metadata": {}
}`, req.Method, req.Endpoint, req.StatusCode, string(responseJSON))

	response, err := vc.llmClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.3, // Low temperature for consistent extraction
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture variables with LLM: %w", err)
	}

	// Parse LLM response
	cleanedResponse := stripMarkdownCodeBlocks(response)
	var result CaptureResult
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		log.Printf("Warning: failed to parse LLM capture response: %v", err)
		return &CaptureResult{
			Variables: make(map[string]interface{}),
			Metadata:  make(map[string]string),
		}, nil
	}

	return &result, nil
}
