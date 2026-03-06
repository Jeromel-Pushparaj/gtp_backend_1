package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// MCPToolExecutor executes tools via HTTP calls to the MCP server
type MCPToolExecutor struct {
	mcpServerURL string
	httpClient   *http.Client
}

// NewMCPToolExecutor creates a new MCP tool executor
func NewMCPToolExecutor(mcpServerURL string) *MCPToolExecutor {
	return &MCPToolExecutor{
		mcpServerURL: mcpServerURL,
		httpClient:   &http.Client{},
	}
}

// ExecuteToolRequest represents the request to MCP server
type ExecuteToolRequest struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ExecuteToolResponse represents the response from MCP server
type ExecuteToolResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// ExecuteTool executes a tool by making an HTTP POST request to the MCP server
func (m *MCPToolExecutor) ExecuteTool(toolName string, arguments map[string]interface{}) (string, error) {
	log.Printf("Executing tool via MCP HTTP: %s with arguments: %v", toolName, arguments)

	// Prepare request body
	reqBody := ExecuteToolRequest{
		Tool:      toolName,
		Arguments: arguments,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP POST request to MCP server
	url := m.mcpServerURL + "/execute"
	log.Printf("Calling MCP server: POST %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Printf("MCP server call failed: %v", err)
		return "", fmt.Errorf("failed to call MCP server: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var mcpResp ExecuteToolResponse
	if err := json.Unmarshal(responseBody, &mcpResp); err != nil {
		return "", fmt.Errorf("failed to parse MCP response: %w", err)
	}

	// Check for errors in response
	if mcpResp.Error != "" {
		log.Printf("MCP tool execution failed: %s", mcpResp.Error)
		return "", fmt.Errorf("MCP error: %s", mcpResp.Error)
	}

	log.Printf("MCP tool execution successful, result length: %d", len(mcpResp.Result))
	return mcpResp.Result, nil
}