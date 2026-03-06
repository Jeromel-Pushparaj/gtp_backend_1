package vectordb

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestEmbeddingClient tests the embedding client with OpenAI API
func TestEmbeddingClient(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping embedding test")
	}

	client := NewEmbeddingClient(EmbeddingConfig{
		APIKey: apiKey,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("single embedding", func(t *testing.T) {
		text := "API authentication using Bearer tokens"
		embedding, err := client.Embed(ctx, text)
		if err != nil {
			t.Fatalf("Embed failed: %v", err)
		}

		if len(embedding) != EmbeddingDimension {
			t.Errorf("Expected embedding dimension %d, got %d", EmbeddingDimension, len(embedding))
		}

		// Verify embedding values are not all zero
		allZero := true
		for _, v := range embedding {
			if v != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Error("Embedding is all zeros")
		}
	})

	t.Run("batch embedding", func(t *testing.T) {
		texts := []string{
			"User authentication failed with 401 error",
			"Database connection timeout after 30 seconds",
			"Rate limiting triggered for API endpoint",
		}

		embeddings, err := client.EmbedBatch(ctx, texts)
		if err != nil {
			t.Fatalf("EmbedBatch failed: %v", err)
		}

		if len(embeddings) != len(texts) {
			t.Errorf("Expected %d embeddings, got %d", len(texts), len(embeddings))
		}

		for i, emb := range embeddings {
			if len(emb) != EmbeddingDimension {
				t.Errorf("Embedding %d: expected dimension %d, got %d", i, EmbeddingDimension, len(emb))
			}
		}
	})
}

// TestVectorToString tests the vector serialization
func TestVectorToString(t *testing.T) {
	v := Vector{0.1, 0.2, 0.3}
	result := vectorToString(v)

	// Should produce something like "[0.100000,0.200000,0.300000]"
	if result[0] != '[' || result[len(result)-1] != ']' {
		t.Errorf("Expected bracketed format, got: %s", result)
	}
}

// TestTypes tests the type structures
func TestTypes(t *testing.T) {
	t.Run("Learning", func(t *testing.T) {
		learning := Learning{
			Category:  "auth_pattern",
			SourceAPI: "petstore-api",
			Content:   "Bearer token authentication with OAuth2",
			Context: map[string]interface{}{
				"endpoint": "/api/pets",
				"method":   "GET",
			},
			Confidence: 0.85,
		}

		if learning.Category != "auth_pattern" {
			t.Error("Learning category mismatch")
		}
	})

	t.Run("FailurePattern", func(t *testing.T) {
		pattern := FailurePattern{
			FailureType:    "auth_failure",
			ErrorSignature: "401 Unauthorized",
			ErrorCode:      "401",
			FixDescription: "Add Bearer token to Authorization header",
		}

		if pattern.FailureType != "auth_failure" {
			t.Error("FailurePattern type mismatch")
		}
	})

	t.Run("SearchOptions", func(t *testing.T) {
		opts := DefaultSearchOptions()
		if opts.Limit != 10 {
			t.Errorf("Expected default limit 10, got %d", opts.Limit)
		}
		if opts.MinScore != 0.5 {
			t.Errorf("Expected default min score 0.5, got %f", opts.MinScore)
		}
	})
}

// TestStoreIntegration tests the store with a real database
// Requires PostgreSQL with pgvector to be running
func TestStoreIntegration(t *testing.T) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		t.Skip("POSTGRES_HOST not set, skipping integration test")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	embClient := NewEmbeddingClient(EmbeddingConfig{APIKey: apiKey})

	store, err := NewStore(StoreConfig{
		Host:     host,
		Port:     5432,
		Database: os.Getenv("POSTGRES_DB"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	}, embClient)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Successfully connected to PostgreSQL with pgvector")
}

