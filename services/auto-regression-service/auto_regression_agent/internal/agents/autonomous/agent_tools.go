package autonomous

import (
	"encoding/json"
)

// ToolDefinition represents a function declaration for OpenAI-compatible function calling
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  ToolParameters         `json:"parameters"`
}

// ToolParameters represents the parameters schema for a tool
type ToolParameters struct {
	Type       string                    `json:"type"`
	Properties map[string]ToolProperty   `json:"properties"`
	Required   []string                  `json:"required"`
}

// ToolProperty represents a single parameter property
type ToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Items       *ToolProperty `json:"items,omitempty"` // For array types
}

// ToolCall represents a function call from the AI
type ToolCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// ToolResponse represents the result of a tool execution
type ToolResponse struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// HTTPRequestTool defines the http_request tool for making API calls
var HTTPRequestTool = ToolDefinition{
	Name:        "http_request",
	Description: "Make an HTTP request to the target API. Use this to execute API tests. You can make GET, POST, PUT, DELETE, PATCH requests with headers, query parameters, path parameters, and request body.",
	Parameters: ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"method": {
				Type:        "string",
				Description: "HTTP method to use",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"path": {
				Type:        "string",
				Description: "API endpoint path (e.g., /pet/{petId}). Use actual values, not placeholders.",
			},
			"headers": {
				Type:        "object",
				Description: "HTTP headers to include in the request (key-value pairs)",
			},
			"query_params": {
				Type:        "object",
				Description: "Query parameters to append to the URL (key-value pairs)",
			},
			"path_params": {
				Type:        "object",
				Description: "Path parameters to substitute in the URL (key-value pairs). For path /pet/{petId}, provide {\"petId\": \"123\"}",
			},
			"body": {
				Type:        "object",
				Description: "Request body for POST/PUT/PATCH requests (will be JSON encoded)",
			},
		},
		Required: []string{"method", "path"},
	},
}

// StoreContextTool defines the store_context tool for saving response data
var StoreContextTool = ToolDefinition{
	Name:        "store_context",
	Description: "Store data from an API response for use in subsequent requests. Use this to capture IDs, tokens, or other data that will be needed later.",
	Parameters: ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"context_type": {
				Type:        "string",
				Description: "Type of context being stored (e.g., 'pet', 'user', 'order')",
			},
			"data": {
				Type:        "object",
				Description: "Key-value pairs to store (e.g., {\"id\": 123, \"name\": \"test\"})",
			},
		},
		Required: []string{"context_type", "data"},
	},
}

// GetContextTool defines the get_context tool for retrieving stored data
var GetContextTool = ToolDefinition{
	Name:        "get_context",
	Description: "Retrieve previously stored context data. Use this to get IDs or data captured from earlier API calls.",
	Parameters: ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"context_type": {
				Type:        "string",
				Description: "Type of context to retrieve (e.g., 'pet', 'user', 'order')",
			},
			"field": {
				Type:        "string",
				Description: "Specific field to retrieve (e.g., 'id', 'name'). If omitted, returns all fields.",
			},
		},
		Required: []string{"context_type"},
	},
}

// ReportResultTool defines the report_result tool for reporting test outcomes
var ReportResultTool = ToolDefinition{
	Name:        "report_result",
	Description: "Report the final result of a test case. Call this after executing the test to record pass/fail status.",
	Parameters: ToolParameters{
		Type: "object",
		Properties: map[string]ToolProperty{
			"passed": {
				Type:        "boolean",
				Description: "Whether the test passed (true) or failed (false)",
			},
			"actual_status": {
				Type:        "integer",
				Description: "The HTTP status code received from the API",
			},
			"expected_status": {
				Type:        "integer",
				Description: "The HTTP status code expected by the test",
			},
			"reason": {
				Type:        "string",
				Description: "Explanation of why the test passed or failed",
			},
			"response_body": {
				Type:        "object",
				Description: "The response body from the API (if relevant)",
			},
		},
		Required: []string{"passed", "actual_status", "expected_status", "reason"},
	},
}

// AllAgentTools returns all available tools for the executor agent
func AllAgentTools() []ToolDefinition {
	return []ToolDefinition{
		HTTPRequestTool,
		StoreContextTool,
		GetContextTool,
		ReportResultTool,
	}
}

// ToJSON converts tool definitions to JSON for the OpenAI-compatible API
func ToolsToJSON(tools []ToolDefinition) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		data, err := json.Marshal(tool)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		result[i] = m
	}
	return result, nil
}

