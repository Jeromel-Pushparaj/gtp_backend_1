package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient executes HTTP requests
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	BaseURL     string
	Endpoint    string
	Method      string
	Headers     map[string]string
	PathParams  map[string]string
	QueryParams map[string]string
	Payload     map[string]interface{}
	Timeout     time.Duration
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode   int
	Headers      map[string]string
	Body         map[string]interface{}
	RawBody      []byte
	ResponseTime int64 // milliseconds
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Execute executes an HTTP request
func (c *HTTPClient) Execute(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	startTime := time.Now()

	// Build URL
	fullURL := req.BaseURL + req.Endpoint

	// Replace path parameters
	for key, value := range req.PathParams {
		fullURL = replacePathParam(fullURL, key, value)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		query := parsedURL.Query()
		for key, value := range req.QueryParams {
			query.Set(key, value)
		}
		parsedURL.RawQuery = query.Encode()
		fullURL = parsedURL.String()
	}

	// Marshal payload
	var body io.Reader
	if req.Payload != nil {
		jsonData, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// Parse JSON response
	var bodyMap map[string]interface{}
	if len(rawBody) > 0 {
		// Try to parse as JSON
		if err := json.Unmarshal(rawBody, &bodyMap); err != nil {
			// Not JSON, store as raw string
			bodyMap = map[string]interface{}{
				"_raw": string(rawBody),
			}
		}
	}

	// Build response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &HTTPResponse{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		Body:         bodyMap,
		RawBody:      rawBody,
		ResponseTime: responseTime,
	}, nil
}

// replacePathParam replaces a path parameter in the URL
func replacePathParam(urlStr, key, value string) string {
	// Replace {key} with value
	return strings.ReplaceAll(urlStr, "{"+key+"}", value)
}

// GenerateCurl generates a curl command string for the given HTTP request
// This is useful for debugging and manual testing
func (c *HTTPClient) GenerateCurl(req HTTPRequest) string {
	// Build URL
	fullURL := req.BaseURL + req.Endpoint

	// Replace path parameters
	for key, value := range req.PathParams {
		fullURL = replacePathParam(fullURL, key, value)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err == nil {
			query := parsedURL.Query()
			for key, value := range req.QueryParams {
				query.Set(key, value)
			}
			parsedURL.RawQuery = query.Encode()
			fullURL = parsedURL.String()
		}
	}

	// Start building curl command
	var curlParts []string
	curlParts = append(curlParts, "curl")

	// Add method (if not GET)
	if req.Method != "" && req.Method != "GET" {
		curlParts = append(curlParts, "-X", req.Method)
	}

	// Add Content-Type header
	curlParts = append(curlParts, "-H", "'Content-Type: application/json'")

	// Add custom headers
	for key, value := range req.Headers {
		if strings.EqualFold(key, "Content-Type") {
			continue
		}
		curlParts = append(curlParts, "-H", fmt.Sprintf("'%s: %s'", key, value))
	}

	// Add payload if present
	if req.Payload != nil {
		jsonData, err := json.Marshal(req.Payload)
		if err == nil {
			// Escape single quotes in JSON for shell safety
			jsonStr := strings.ReplaceAll(string(jsonData), "'", "'\\''")
			curlParts = append(curlParts, "-d", fmt.Sprintf("'%s'", jsonStr))
		}
	}

	// Add URL (quoted for safety)
	curlParts = append(curlParts, fmt.Sprintf("'%s'", fullURL))

	return strings.Join(curlParts, " ")
}
