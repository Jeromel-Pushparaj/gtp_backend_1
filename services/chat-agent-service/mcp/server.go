package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Server represents the MCP HTTP server
type Server struct {
	baseURL string
	apiKey  string
	port    string
}

// NewServer creates a new MCP server instance
func NewServer(baseURL, apiKey, port string) *Server {
	return &Server{
		baseURL: baseURL,
		apiKey:  apiKey,
		port:    port,
	}
}

// ExecuteToolRequest represents the HTTP request body
type ExecuteToolRequest struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ExecuteToolResponse represents the HTTP response body
type ExecuteToolResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// RegisterTools is a no-op for HTTP mode (tools are mapped dynamically)
func (s *Server) RegisterTools() error {
	log.Println("Tools registered (HTTP mode - dynamic mapping)")
	return nil
}

// Serve starts the HTTP server
func (s *Server) Serve() error {
	http.HandleFunc("/execute", s.handleExecuteTool)
	http.HandleFunc("/health", s.handleHealth)

	addr := ":" + s.port
	log.Printf("MCP HTTP Server listening on %s", addr)
	log.Printf("Backend API URL: %s", s.baseURL)
	return http.ListenAndServe(addr, nil)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "mcp-server",
	})
}

// handleExecuteTool handles tool execution requests
func (s *Server) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ExecuteToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to parse request: %v", err)
		respondWithError(w, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	log.Printf("Executing tool: %s with arguments: %v", req.Tool, req.Arguments)

	// Execute the tool
	result, err := s.ExecuteTool(req.Tool, req.Arguments)
	if err != nil {
		log.Printf("Tool execution failed: %v", err)
		respondWithError(w, err.Error())
		return
	}

	// Return success response
	log.Printf("Tool execution successful, result length: %d", len(result))
	respondWithSuccess(w, result)
}

// respondWithSuccess sends a successful response
func respondWithSuccess(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ExecuteToolResponse{
		Result: result,
		Error:  "",
	})
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ExecuteToolResponse{
		Result: "",
		Error:  errorMsg,
	})
}

// ExecuteTool executes a tool by mapping it to a backend API endpoint
func (s *Server) ExecuteTool(toolName string, arguments map[string]interface{}) (string, error) {
	endpoint, method, params, body := s.mapToolToEndpoint(toolName, arguments)
	if endpoint == "" {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
	return s.makeRequest(method, endpoint, params, body)
}

// makeRequest makes an HTTP request to the backend API
func (s *Server) makeRequest(method, endpoint string, queryParams map[string]string, body interface{}) (string, error) {
	fullURL := s.baseURL + endpoint

	if len(queryParams) > 0 {
		params := url.Values{}
		for k, v := range queryParams {
			params.Add(k, v)
		}
		fullURL += "?" + params.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return "", err
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}

// mapToolToEndpoint maps tool names to API endpoints
// mapToolToEndpoint maps tool names to API endpoints by delegating to category-specific mappers
func (s *Server) mapToolToEndpoint(toolName string, args map[string]interface{}) (endpoint, method string, params map[string]string, body interface{}) {
	// Health check (special case)
	if toolName == "health_check" {
		return "/health", "GET", nil, nil
	}

	// Try GitHub tools
	if endpoint, method, params, body, found := mapGitHubToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Try Jira tools
	if endpoint, method, params, body, found := mapJiraToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Try Metrics tools
	if endpoint, method, params, body, found := mapMetricsToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Try SonarCloud tools
	if endpoint, method, params, body, found := mapSonarCloudToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Try Organization tools
	if endpoint, method, params, body, found := mapOrgToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Try Repository tools
	if endpoint, method, params, body, found := mapRepoToolToEndpoint(toolName, args); found {
		return endpoint, method, params, body
	}

	// Tool not found
	return "", "", nil, nil
}
