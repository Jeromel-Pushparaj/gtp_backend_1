package vectordb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient generates embeddings using OpenAI-compatible embedding APIs
// Supports both local sentence transformers and cloud providers (OpenAI, etc.)
type EmbeddingClient struct {
	apiKey   string
	baseURL  string
	model    string
	provider string // "local", "openai", etc.
	timeout  time.Duration
	client   *http.Client
}

// EmbeddingConfig configures the embedding client
type EmbeddingConfig struct {
	APIKey   string
	BaseURL  string        // defaults to local service (http://localhost:8000) or OpenAI API
	Model    string        // defaults to all-MiniLM-L6-v2 (local) or text-embedding-3-small (OpenAI)
	Provider string        // "local" (default) or "openai"
	Timeout  time.Duration // defaults to 30s
}

// NewEmbeddingClient creates a new embedding client
// Supports local sentence transformers (default) or cloud providers like OpenAI
func NewEmbeddingClient(config EmbeddingConfig) *EmbeddingClient {
	// Default to local provider
	if config.Provider == "" {
		config.Provider = "local"
	}

	// Set defaults based on provider
	if config.BaseURL == "" {
		if config.Provider == "local" {
			config.BaseURL = "http://localhost:8000"
		} else if config.Provider == "openai" {
			config.BaseURL = "https://api.openai.com/v1"
		}
	}

	if config.Model == "" {
		if config.Provider == "local" {
			config.Model = "all-MiniLM-L6-v2"
		} else if config.Provider == "openai" {
			config.Model = "text-embedding-3-small"
		}
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &EmbeddingClient{
		apiKey:   config.APIKey,
		baseURL:  config.BaseURL,
		model:    config.Model,
		provider: config.Provider,
		timeout:  config.Timeout,
		client:   &http.Client{Timeout: config.Timeout},
	}
}

// Embed generates an embedding for the given text
func (c *EmbeddingClient) Embed(ctx context.Context, text string) (Vector, error) {
	vectors, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return vectors[0], nil
}

// EmbedBatch generates embeddings for multiple texts in a single request
// Uses OpenAI-compatible embedding API format
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Build request body for OpenAI embedding API
	reqBody := map[string]interface{}{
		"model": c.model,
		"input": texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/embeddings", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Only set Authorization header if API key is provided (not needed for local service)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI-compatible response
	var apiResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to vectors
	vectors := make([]Vector, len(apiResp.Data))
	for i, data := range apiResp.Data {
		// Note: OpenAI's text-embedding-3-small has 1536 dimensions by default
		// You may need to adjust EmbeddingDimension constant if using different models
		vectors[i] = data.Embedding
	}

	return vectors, nil
}

