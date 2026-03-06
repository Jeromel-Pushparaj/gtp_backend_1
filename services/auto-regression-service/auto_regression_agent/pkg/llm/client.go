package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a multi-provider LLM client supporting Groq and OpenAI-compatible APIs
type Client struct {
	apiKey   string
	baseURL  string
	model    string
	provider string
	timeout  time.Duration
}

// Config represents LLM client configuration
type Config struct {
	APIKey   string
	BaseURL  string // e.g., https://api.groq.com/openai/v1 (Groq) or https://api.openai.com/v1 (OpenAI)
	Model    string // e.g., openai/gpt-oss-120b (Groq) or gpt-4, claude-3-opus (OpenAI)
	Provider string // groq, openai, anthropic, azure
	Timeout  time.Duration
}

// CompletionOptions represents options for completion requests
type CompletionOptions struct {
	Temperature float64
	MaxTokens   int
}

// NewClient creates a new LLM client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.BaseURL == "" {
		// Default to Groq API (currently active)
		if config.Provider == "groq" || config.Provider == "" {
			config.BaseURL = "https://api.groq.com/openai/v1"
		} else if config.Provider == "openai" {
			config.BaseURL = "https://api.openai.com/v1"
		} else {
			// Default to Groq
			config.BaseURL = "https://api.groq.com/openai/v1"
		}
	}
	return &Client{
		apiKey:   config.APIKey,
		baseURL:  config.BaseURL,
		model:    config.Model,
		provider: config.Provider,
		timeout:  config.Timeout,
	}
}

// Provider returns the LLM provider name
func (c *Client) Provider() string {
	return c.provider
}

// GenerateCompletion generates a completion using the LLM
func (c *Client) GenerateCompletion(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	// Use Groq API (OpenAI-compatible format)
	if c.provider == "groq" || c.provider == "" || c.provider == "openai" {
		return c.generateGroqCompletion(ctx, prompt, opts)
	}

	return "", fmt.Errorf("unsupported provider: %s (currently only 'groq' and 'openai' are supported)", c.provider)
}

// generateGroqCompletion generates a completion using Groq API (OpenAI-compatible format)
// Includes retry logic for rate limits (6K TPM, 500K/day for qwen-2.5-32b-instruct)
func (c *Client) generateGroqCompletion(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	maxRetries := 5
	retryDelay := 15 * time.Second // Wait 15s on rate limit (6K TPM = ~100 tokens/sec)

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {

			fmt.Printf("⏳ Rate limit hit, waiting %v before retry %d/%d...", retryDelay, attempt, maxRetries)
			time.Sleep(retryDelay)
		}

		// Build Groq/OpenAI-compatible API request
		reqBody := map[string]interface{}{
			"model": c.model,
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"temperature":            opts.Temperature,
			"max_completion_tokens":  opts.MaxTokens,
			"top_p":                  1,
			"stream":                 false,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request: %w", err)
		}

		// Create HTTP request - Groq uses Authorization header
		url := fmt.Sprintf("%s/chat/completions", c.baseURL)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

		// Execute request with timeout
		client := &http.Client{Timeout: c.timeout}
		resp, err := client.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("LLM request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		// Check for rate limit (429) and retry
		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("rate limit exceeded (429): %s", string(body))
			fmt.Printf("⚠️  Rate limit hit (attempt %d/%d): %v", attempt+1, maxRetries+1, lastErr)
			continue // Retry
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
		}

		// Parse OpenAI-compatible response format
		var apiResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				TotalTokens int `json:"total_tokens"`
			} `json:"usage"`
			Model string `json:"model"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return "", fmt.Errorf("failed to parse Groq response: %w", err)
		}

		if len(apiResp.Choices) == 0 {
			return "", fmt.Errorf("no choices in Groq response")
		}

		// Log finish reason for debugging truncation issues
		finishReason := apiResp.Choices[0].FinishReason
		if finishReason != "stop" && finishReason != "" {
			fmt.Printf("Warning: Groq finish reason: %s (may indicate truncation)\n", finishReason)
		}

		// Success! Return the response
		return apiResp.Choices[0].Message.Content, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
	}
	return "", fmt.Errorf("failed after %d retries", maxRetries)
}


