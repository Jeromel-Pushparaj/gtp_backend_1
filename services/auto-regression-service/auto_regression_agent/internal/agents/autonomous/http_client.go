package autonomous

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

// HTTPClient handles HTTP requests for autonomous agents
type HTTPClient struct {
	client *http.Client
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	BaseURL     string
	Path        string
	Method      string
	Headers     map[string]string
	PathParams  map[string]string // Path parameters like {petId}
	QueryParams map[string]string // Query parameters like ?status=active
	Payload     interface{}
	Timeout     time.Duration
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute executes an HTTP request
func (c *HTTPClient) Execute(ctx context.Context, req HTTPRequest) (*HTTPResponse, error) {
	startTime := time.Now()

	// Build full URL
	fullURL := req.BaseURL + req.Path

	// Replace path parameters (e.g., {petId} -> 123)
	for key, value := range req.PathParams {
		placeholder := "{" + key + "}"
		fullURL = strings.ReplaceAll(fullURL, placeholder, value)
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

	// Marshal payload if present
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

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
		httpReq = httpReq.WithContext(ctx)
	}

	// Execute request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(startTime)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}

// GenerateCurl generates a curl command string for the given HTTP request
// This is useful for debugging and manual testing
func (c *HTTPClient) GenerateCurl(req HTTPRequest) string {
	// Build full URL
	fullURL := req.BaseURL + req.Path

	// Replace path parameters (e.g., {petId} -> 123)
	for key, value := range req.PathParams {
		placeholder := "{" + key + "}"
		fullURL = strings.ReplaceAll(fullURL, placeholder, value)
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

	// Add headers
	// Always add Content-Type and Accept as defaults
	curlParts = append(curlParts, "-H", "'Content-Type: application/json'")
	curlParts = append(curlParts, "-H", "'Accept: application/json'")

	// Add custom headers
	for key, value := range req.Headers {
		// Skip Content-Type and Accept since we already added them
		if strings.EqualFold(key, "Content-Type") || strings.EqualFold(key, "Accept") {
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
