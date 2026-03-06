package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FunctionCall represents a function call from the Groq API
type FunctionCall struct {
	Name             string                 `json:"name"`
	Args             map[string]interface{} `json:"args"`
	ThoughtSignature string                 `json:"thought_signature,omitempty"` // For thinking models
}

// AgentMessage represents a message in the conversation
type AgentMessage struct {
	Role              string             `json:"role"` // "user", "assistant", "system", or "tool"
	Text              string             `json:"text,omitempty"`
	FunctionCall      *FunctionCall      `json:"function_call,omitempty"`      // Single function call (legacy)
	FunctionCalls     []FunctionCall     `json:"function_calls,omitempty"`     // Multiple function calls (for parallel calls)
	FunctionResponse  *FunctionResponse  `json:"function_response,omitempty"`  // Single function response (legacy)
	FunctionResponses []FunctionResponse `json:"function_responses,omitempty"` // Multiple function responses
	ToolCallID        string             `json:"tool_call_id,omitempty"`       // For tool responses
}

// FunctionResponse represents a function response to send back to the model
type FunctionResponse struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
}

// AgentResponse represents the response from the agent API
type AgentResponse struct {
	Text          string         `json:"text,omitempty"`
	FunctionCalls []FunctionCall `json:"function_calls,omitempty"`
	FinishReason  string         `json:"finish_reason"`
}

// AgentCompletionOptions represents options for agent completion requests
type AgentCompletionOptions struct {
	Temperature float64
	MaxTokens   int
	Tools       []map[string]interface{} // Function declarations
}

// GenerateAgentCompletion generates a completion with function calling support
func (c *Client) GenerateAgentCompletion(ctx context.Context, messages []AgentMessage, opts AgentCompletionOptions) (*AgentResponse, error) {
	if c.provider == "groq" || c.provider == "" || c.provider == "openai" {
		return c.generateGroqAgentCompletion(ctx, messages, opts)
	}
	return nil, fmt.Errorf("unsupported provider: %s", c.provider)
}

// generateGroqAgentCompletion generates a completion with tools using Groq API (OpenAI-compatible)
// Includes retry logic for rate limits (6K TPM, 500K/day for qwen-2.5-32b-instruct)
func (c *Client) generateGroqAgentCompletion(ctx context.Context, messages []AgentMessage, opts AgentCompletionOptions) (*AgentResponse, error) {
	maxRetries := 5
	retryDelay := 15 * time.Second // Wait 15s on rate limit

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("⏳ Rate limit hit (agent), waiting %v before retry %d/%d...", retryDelay, attempt, maxRetries)
			time.Sleep(retryDelay)
		}

		result, err := c.executeGroqAgentRequest(ctx, messages, opts)
		if err != nil {
			// Check if it's a rate limit error
			if strings.Contains(err.Error(), "status 429") || strings.Contains(err.Error(), "rate_limit_exceeded") {
				lastErr = err
				fmt.Printf("⚠️  Rate limit hit (agent, attempt %d/%d): %v", attempt+1, maxRetries+1, err)
				continue // Retry
			}
			// Other errors, don't retry
			return nil, err
		}

		// Success!
		return result, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
	}
	return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

// executeGroqAgentRequest executes a single Groq agent request (extracted for retry logic)
func (c *Client) executeGroqAgentRequest(ctx context.Context, messages []AgentMessage, opts AgentCompletionOptions) (*AgentResponse, error) {
	// Build OpenAI-compatible messages from AgentMessage
	apiMessages := make([]map[string]interface{}, 0, len(messages))

	for _, msg := range messages {
		// Map role: "model" -> "assistant"
		role := msg.Role
		if role == "model" {
			role = "assistant"
		}

		// Handle text messages
		if msg.Text != "" {
			apiMessages = append(apiMessages, map[string]interface{}{
				"role":    role,
				"content": msg.Text,
			})
		}

		// Handle function/tool calls
		if msg.FunctionCall != nil || len(msg.FunctionCalls) > 0 {
			toolCalls := []map[string]interface{}{}

			// Single function call
			if msg.FunctionCall != nil {
				toolCalls = append(toolCalls, map[string]interface{}{
					"id":   fmt.Sprintf("call_%s", msg.FunctionCall.Name),
					"type": "function",
					"function": map[string]interface{}{
						"name":      msg.FunctionCall.Name,
						"arguments": toJSONString(msg.FunctionCall.Args),
					},
				})
			}

			// Multiple function calls
			for _, fc := range msg.FunctionCalls {
				toolCalls = append(toolCalls, map[string]interface{}{
					"id":   fmt.Sprintf("call_%s", fc.Name),
					"type": "function",
					"function": map[string]interface{}{
						"name":      fc.Name,
						"arguments": toJSONString(fc.Args),
					},
				})
			}

			if len(toolCalls) > 0 {
				apiMessages = append(apiMessages, map[string]interface{}{
					"role":       "assistant",
					"tool_calls": toolCalls,
				})
			}
		}

		// Handle function/tool responses
		if msg.FunctionResponse != nil {
			responseJSON := toJSONString(msg.FunctionResponse.Response)
			apiMessages = append(apiMessages, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": fmt.Sprintf("call_%s", msg.FunctionResponse.Name),
				"content":      responseJSON,
			})
		}

		for _, fr := range msg.FunctionResponses {
			responseJSON := toJSONString(fr.Response)
			apiMessages = append(apiMessages, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": fmt.Sprintf("call_%s", fr.Name),
				"content":      responseJSON,
			})
		}
	}

	// Build request body
	reqBody := map[string]interface{}{
		"model":                 c.model,
		"messages":              apiMessages,
		"temperature":           opts.Temperature,
		"max_completion_tokens": opts.MaxTokens,
		"top_p":                 1,
		"stream":                false,
	}

	// Add tools if provided
	if len(opts.Tools) > 0 {
		tools := make([]map[string]interface{}, len(opts.Tools))
		for i, tool := range opts.Tools {
			tools[i] = map[string]interface{}{
				"type":     "function",
				"function": tool,
			}
		}
		reqBody["tools"] = tools
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Execute request
	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	return c.parseGroqAgentResponse(body)
}

// toJSONString converts an interface to JSON string
func toJSONString(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// parseGroqAgentResponse parses the Groq/OpenAI-compatible response with function calls
func (c *Client) parseGroqAgentResponse(body []byte) (*AgentResponse, error) {
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Groq response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in Groq response")
	}

	choice := apiResp.Choices[0]
	result := &AgentResponse{
		Text:         choice.Message.Content,
		FinishReason: choice.FinishReason,
	}

	// Extract tool calls
	for _, toolCall := range choice.Message.ToolCalls {
		if toolCall.Type == "function" {
			// Parse arguments JSON string to map
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				// If parsing fails, use empty map
				args = make(map[string]interface{})
			}

			fc := FunctionCall{
				Name: toolCall.Function.Name,
				Args: args,
			}
			result.FunctionCalls = append(result.FunctionCalls, fc)
		}
	}

	return result, nil
}
