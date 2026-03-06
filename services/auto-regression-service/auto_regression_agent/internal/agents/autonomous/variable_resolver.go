package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// VariableResolver uses LLM to intelligently resolve variables in test cases
type VariableResolver struct {
	llmClient *llm.Client
	strategy  map[string]interface{} // Strategy from Designer Agent
}

// NewVariableResolver creates a new variable resolver service
func NewVariableResolver(llmClient *llm.Client) *VariableResolver {
	return &VariableResolver{
		llmClient: llmClient,
	}
}

// SetStrategy sets the test strategy containing dependency mapping
func (vr *VariableResolver) SetStrategy(strategy map[string]interface{}) {
	vr.strategy = strategy
}

// ResolveRequest represents a request to resolve variables in a test case
type ResolveRequest struct {
	Endpoint           string                 // e.g., "GET /api/resources/{id}"
	Method             string                 // e.g., "GET"
	PathParams         map[string]string      // Path parameters that may contain placeholders
	QueryParams        map[string]string      // Query parameters that may contain placeholders
	Payload            map[string]interface{} // Request body that may contain placeholders
	AvailableVariables map[string]interface{} // Variables available in context
}

// ResolveResult represents the result of variable resolution
type ResolveResult struct {
	PathParams  map[string]string      // Resolved path parameters
	QueryParams map[string]string      // Resolved query parameters
	Payload     map[string]interface{} // Resolved request body
	Resolved    bool                   // Whether any variables were resolved
}

// ResolveVariables intelligently resolves variables in test cases
func (vr *VariableResolver) ResolveVariables(ctx context.Context, req ResolveRequest) (*ResolveResult, error) {
	// First, try explicit resolution from strategy
	if vr.strategy != nil {
		if explicitResult := vr.resolveExplicit(req); explicitResult.Resolved {
			log.Printf("📌 Using explicit variable resolution from strategy for %s", req.Endpoint)
			return explicitResult, nil
		}
	}

	// If no explicit mapping or not resolved, use LLM
	return vr.resolveWithLLM(ctx, req)
}

// resolveExplicit uses strategy's resource_consumers mapping to resolve variables
func (vr *VariableResolver) resolveExplicit(req ResolveRequest) *ResolveResult {
	result := &ResolveResult{
		PathParams:  make(map[string]string),
		QueryParams: make(map[string]string),
		Payload:     req.Payload,
		Resolved:    false,
	}

	// Navigate to data_strategy.resource_consumers
	dataStrategy, ok := vr.strategy["data_strategy"].(map[string]interface{})
	if !ok {
		return result
	}

	resourceConsumers, ok := dataStrategy["resource_consumers"].(map[string]interface{})
	if !ok {
		return result
	}

	// Check if this endpoint is a resource consumer
	consumerInfo, ok := resourceConsumers[req.Endpoint].(map[string]interface{})
	if !ok {
		return result
	}

	// Get required fields
	requiredFields, ok := consumerInfo["required_fields"].([]interface{})
	if !ok {
		return result
	}

	// Resolve each required field
	for _, fieldInterface := range requiredFields {
		field, ok := fieldInterface.(string)
		if !ok {
			continue
		}

		// Check if we have this variable
		if value, exists := req.AvailableVariables[field]; exists {
			result.PathParams[field] = fmt.Sprintf("%v", value)
			result.Resolved = true
			log.Printf("✅ Resolved %s = %v for %s", field, value, req.Endpoint)
		}
	}

	// Copy over any non-resolved path params
	for k, v := range req.PathParams {
		if _, exists := result.PathParams[k]; !exists {
			result.PathParams[k] = v
		}
	}

	// Copy query params as-is
	result.QueryParams = req.QueryParams

	return result
}

// resolveWithLLM uses LLM to intelligently resolve variables
func (vr *VariableResolver) resolveWithLLM(ctx context.Context, req ResolveRequest) (*ResolveResult, error) {
	availableVarsJSON, _ := json.MarshalIndent(req.AvailableVariables, "", "  ")
	pathParamsJSON, _ := json.MarshalIndent(req.PathParams, "", "  ")
	queryParamsJSON, _ := json.MarshalIndent(req.QueryParams, "", "  ")
	payloadJSON, _ := json.MarshalIndent(req.Payload, "", "  ")

	prompt := fmt.Sprintf(`You are an API testing expert. Resolve variables in this test case using available context.

Endpoint: %s %s
Available Variables:
%s

Current Path Parameters:
%s

Current Query Parameters:
%s

Current Payload:
%s

Instructions:
1. Replace any placeholder values (like "{id}", "{{resourceId}}", or "PLACEHOLDER") with actual values from available variables
2. Match variable names intelligently (e.g., "resourceId" can match "resource_id" or "id" if context suggests it's the right resource)
3. Keep non-placeholder values as-is
4. If a variable is not available, keep the placeholder

CRITICAL FORMATTING RULES:
1. Your response MUST be ONLY valid JSON - no explanations, no markdown, no code blocks
2. Do NOT wrap the JSON in markdown code blocks
3. Start your response directly with the opening brace {
4. End your response with the closing brace }

Required JSON structure:
{
  "path_params": {
    "id": "123"
  },
  "query_params": {
    "status": "active"
  },
  "payload": {
    "name": "Updated Resource"
  },
  "resolved": true
}`, req.Method, req.Endpoint, string(availableVarsJSON), string(pathParamsJSON), string(queryParamsJSON), string(payloadJSON))

	response, err := vr.llmClient.GenerateCompletion(ctx, prompt, llm.CompletionOptions{
		Temperature: 0.2, // Very low temperature for consistent resolution
		MaxTokens:   2048,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve variables with LLM: %w", err)
	}

	// Parse LLM response
	cleanedResponse := stripMarkdownCodeBlocks(response)

	var llmResult struct {
		PathParams  map[string]string      `json:"path_params"`
		QueryParams map[string]string      `json:"query_params"`
		Payload     map[string]interface{} `json:"payload"`
		Resolved    bool                   `json:"resolved"`
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &llmResult); err != nil {
		log.Printf("Warning: failed to parse LLM resolution response: %v", err)
		return &ResolveResult{
			PathParams:  req.PathParams,
			QueryParams: req.QueryParams,
			Payload:     req.Payload,
			Resolved:    false,
		}, nil
	}

	return &ResolveResult{
		PathParams:  llmResult.PathParams,
		QueryParams: llmResult.QueryParams,
		Payload:     llmResult.Payload,
		Resolved:    llmResult.Resolved,
	}, nil
}
