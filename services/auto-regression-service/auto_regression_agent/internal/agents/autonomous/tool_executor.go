package autonomous

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ToolExecutor handles execution of tools called by the AI agent
type ToolExecutor struct {
	httpClient  *HTTPClient
	authMgr     *AuthManager
	testContext *TestContext
	baseURL     string

	// Track the test result reported by the AI
	reportedResult *TestResult
}

// TestResult represents the final test result reported by the AI
type TestResult struct {
	Passed         bool          `json:"passed"`
	ActualStatus   int           `json:"actual_status"`
	ExpectedStatus int           `json:"expected_status"`
	Reason         string        `json:"reason"`
	ResponseBody   interface{}   `json:"response_body,omitempty"`
	ExecutionTime  time.Duration `json:"execution_time"`
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(httpClient *HTTPClient, authMgr *AuthManager, testContext *TestContext, baseURL string) *ToolExecutor {
	return &ToolExecutor{
		httpClient:  httpClient,
		authMgr:     authMgr,
		testContext: testContext,
		baseURL:     baseURL,
	}
}

// ExecuteTool executes a tool call and returns the result
func (te *ToolExecutor) ExecuteTool(ctx context.Context, call ToolCall) ToolResponse {
	log.Printf("🔧 Executing tool: %s", call.Name)

	switch call.Name {
	case "http_request":
		return te.executeHTTPRequest(ctx, call.Args)
	case "store_context":
		return te.executeStoreContext(ctx, call.Args)
	case "get_context":
		return te.executeGetContext(ctx, call.Args)
	case "report_result":
		return te.executeReportResult(ctx, call.Args)
	default:
		return ToolResponse{
			Name:  call.Name,
			Error: fmt.Sprintf("unknown tool: %s", call.Name),
		}
	}
}

// executeHTTPRequest handles the http_request tool
func (te *ToolExecutor) executeHTTPRequest(ctx context.Context, args map[string]interface{}) ToolResponse {
	method, _ := args["method"].(string)
	path, _ := args["path"].(string)

	// Build path params map
	pathParams := make(map[string]string)
	if pp, ok := args["path_params"].(map[string]interface{}); ok {
		for k, v := range pp {
			pathParams[k] = fmt.Sprintf("%v", v)
		}
	}

	// Build query params map
	queryParams := make(map[string]string)
	if qp, ok := args["query_params"].(map[string]interface{}); ok {
		for k, v := range qp {
			queryParams[k] = fmt.Sprintf("%v", v)
		}
	}

	// Build headers map
	headers := make(map[string]string)
	if h, ok := args["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	// Get body
	var body interface{}
	if b, ok := args["body"]; ok {
		body = b
	}

	// Substitute path parameters
	finalPath := path
	for key, value := range pathParams {
		finalPath = strings.ReplaceAll(finalPath, "{"+key+"}", value)
	}

	// Build full URL
	fullURL := te.baseURL + finalPath

	// Add query parameters
	if len(queryParams) > 0 {
		params := make([]string, 0, len(queryParams))
		for k, v := range queryParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		fullURL += "?" + strings.Join(params, "&")
	}

	// Build the HTTP request struct
	httpReq := HTTPRequest{
		BaseURL:     te.baseURL,
		Path:        path,
		Method:      method,
		Headers:     headers,
		QueryParams: queryParams,
		Payload:     body,
		Timeout:     30 * time.Second,
	}

	// Generate curl command for debugging
	curlCmd := te.httpClient.GenerateCurl(httpReq)
	log.Printf("🔧 curl: %s", curlCmd)

	// Make the request
	startTime := time.Now()
	resp, err := te.httpClient.Execute(ctx, httpReq)
	duration := time.Since(startTime)

	if err != nil {
		return ToolResponse{
			Name:  "http_request",
			Error: fmt.Sprintf("HTTP request failed: %v", err),
			Result: map[string]interface{}{
				"error":    err.Error(),
				"duration": duration.Milliseconds(),
			},
		}
	}

	// Parse response body
	var responseBody interface{}
	if resp.Body != nil && len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, &responseBody); err != nil {
			// If not JSON, use raw string
			responseBody = string(resp.Body)
		}
	}

	log.Printf("📡 Response: status=%d, duration=%dms", resp.StatusCode, duration.Milliseconds())

	contentType := ""
	if resp.Headers != nil {
		contentType = resp.Headers.Get("Content-Type")
	}

	return ToolResponse{
		Name: "http_request",
		Result: map[string]interface{}{
			"status_code":  resp.StatusCode,
			"body":         responseBody,
			"duration_ms":  duration.Milliseconds(),
			"content_type": contentType,
		},
	}
}

// executeStoreContext handles the store_context tool
func (te *ToolExecutor) executeStoreContext(ctx context.Context, args map[string]interface{}) ToolResponse {
	contextType, _ := args["context_type"].(string)
	data, _ := args["data"].(map[string]interface{})

	if contextType == "" {
		return ToolResponse{
			Name:  "store_context",
			Error: "context_type is required",
		}
	}

	if data == nil {
		return ToolResponse{
			Name:  "store_context",
			Error: "data is required",
		}
	}

	// Store each field in the test context
	te.testContext.StoreContext(contextType, data)

	log.Printf("💾 Stored context: type=%s, data=%v", contextType, data)

	return ToolResponse{
		Name: "store_context",
		Result: map[string]interface{}{
			"success":      true,
			"context_type": contextType,
			"stored_keys":  getMapKeys(data),
		},
	}
}

// executeGetContext handles the get_context tool
func (te *ToolExecutor) executeGetContext(ctx context.Context, args map[string]interface{}) ToolResponse {
	contextType, _ := args["context_type"].(string)
	field, _ := args["field"].(string)

	if contextType == "" {
		return ToolResponse{
			Name:  "get_context",
			Error: "context_type is required",
		}
	}

	// Get context from test context
	data := te.testContext.GetContext(contextType)

	if data == nil {
		return ToolResponse{
			Name: "get_context",
			Result: map[string]interface{}{
				"found":        false,
				"context_type": contextType,
				"message":      fmt.Sprintf("No context stored for type '%s'", contextType),
			},
		}
	}

	// If a specific field is requested
	if field != "" {
		if value, ok := data[field]; ok {
			return ToolResponse{
				Name: "get_context",
				Result: map[string]interface{}{
					"found":        true,
					"context_type": contextType,
					"field":        field,
					"value":        value,
				},
			}
		}
		return ToolResponse{
			Name: "get_context",
			Result: map[string]interface{}{
				"found":          false,
				"context_type":   contextType,
				"field":          field,
				"available_keys": getMapKeys(data),
				"message":        fmt.Sprintf("Field '%s' not found in context '%s'", field, contextType),
			},
		}
	}

	return ToolResponse{
		Name: "get_context",
		Result: map[string]interface{}{
			"found":        true,
			"context_type": contextType,
			"data":         data,
		},
	}
}

// executeReportResult handles the report_result tool
func (te *ToolExecutor) executeReportResult(ctx context.Context, args map[string]interface{}) ToolResponse {
	passed, _ := args["passed"].(bool)
	actualStatus := int(getFloat64(args, "actual_status"))
	expectedStatus := int(getFloat64(args, "expected_status"))
	reason, _ := args["reason"].(string)
	responseBody := args["response_body"]

	te.reportedResult = &TestResult{
		Passed:         passed,
		ActualStatus:   actualStatus,
		ExpectedStatus: expectedStatus,
		Reason:         reason,
		ResponseBody:   responseBody,
	}

	status := "PASSED"
	if !passed {
		status = "FAILED"
	}
	log.Printf("📊 Test Result: %s - %s", status, reason)

	return ToolResponse{
		Name: "report_result",
		Result: map[string]interface{}{
			"acknowledged": true,
			"status":       status,
		},
	}
}

// GetReportedResult returns the test result reported by the AI
func (te *ToolExecutor) GetReportedResult() *TestResult {
	return te.reportedResult
}

// Helper functions
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}
