package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	fmt.Println("=== Testing Groq API Integration ===\n")
	testGroqAPI()

	fmt.Println("\n=== Testing Local Embedding Service ===\n")
	testLocalEmbeddings()
}

func testGroqAPI() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ GROQ_API_KEY not set")
		return
	}

	fmt.Printf("✓ GROQ_API_KEY found: %s...%s\n", apiKey[:10], apiKey[len(apiKey)-10:])

	// Create LLM client
	client := llm.NewClient(llm.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.groq.com/openai/v1",
		Model:   "openai/gpt-oss-120b",
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	// Test 1: Simple completion
	fmt.Println("\n📝 Test 1: Simple Completion")
	fmt.Println("Prompt: What is 2+2?")
	
	response, err := client.GenerateCompletion(ctx, "What is 2+2? Answer in one sentence.", llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	})

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Response: %s\n", response.Text)
	fmt.Printf("   Tokens used: %d\n", response.TokensUsed)

	// Test 2: More complex prompt
	fmt.Println("\n📝 Test 2: Complex Prompt")
	fmt.Println("Prompt: Explain what API testing is in 2 sentences.")
	
	response2, err := client.GenerateCompletion(ctx, "Explain what API testing is in 2 sentences.", llm.CompletionOptions{
		Temperature: 0.7,
		MaxTokens:   200,
	})

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Response: %s\n", response2.Text)
	fmt.Printf("   Tokens used: %d\n", response2.TokensUsed)
}

func testLocalEmbeddings() {
	// Test with local embedding service
	embeddingProvider := os.Getenv("EMBEDDING_PROVIDER")
	if embeddingProvider == "" {
		embeddingProvider = "local"
	}

	fmt.Printf("✓ Using embedding provider: %s\n", embeddingProvider)

	// Create embedding client
	client := vectordb.NewEmbeddingClient(vectordb.EmbeddingConfig{
		Provider: embeddingProvider,
		BaseURL:  "http://localhost:8000",
		Model:    "all-MiniLM-L6-v2",
		Timeout:  30 * time.Second,
	})

	ctx := context.Background()

	// Test 1: Single text embedding
	fmt.Println("\n📝 Test 1: Single Text Embedding")
	text1 := "This is a test sentence for embedding generation."
	fmt.Printf("Text: %s\n", text1)

	embeddings1, err := client.GenerateEmbeddings(ctx, []string{text1})
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("⚠️  Make sure the embedding service is running:")
		fmt.Println("   cd services/embedding-service && python main.py")
		return
	}

	fmt.Printf("✅ Generated embedding with %d dimensions\n", len(embeddings1[0]))
	fmt.Printf("   First 5 values: [%.4f, %.4f, %.4f, %.4f, %.4f]\n", 
		embeddings1[0][0], embeddings1[0][1], embeddings1[0][2], embeddings1[0][3], embeddings1[0][4])

	// Test 2: Batch embeddings
	fmt.Println("\n📝 Test 2: Batch Embeddings")
	texts := []string{
		"API testing is important for software quality.",
		"Machine learning models need good data.",
		"Docker containers make deployment easier.",
	}
	fmt.Printf("Generating embeddings for %d texts...\n", len(texts))

	embeddings2, err := client.GenerateEmbeddings(ctx, texts)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated %d embeddings\n", len(embeddings2))
	for i, emb := range embeddings2 {
		fmt.Printf("   Text %d: %d dimensions, first value: %.4f\n", i+1, len(emb), emb[0])
	}

	// Test 3: Similarity check
	fmt.Println("\n📝 Test 3: Semantic Similarity")
	similar1 := "The cat sat on the mat."
	similar2 := "A feline rested on the rug."
	different := "Python is a programming language."

	testTexts := []string{similar1, similar2, different}
	embeddings3, err := client.GenerateEmbeddings(ctx, testTexts)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Calculate cosine similarity
	sim12 := cosineSimilarity(embeddings3[0], embeddings3[1])
	sim13 := cosineSimilarity(embeddings3[0], embeddings3[2])

	fmt.Printf("Text 1: %s\n", similar1)
	fmt.Printf("Text 2: %s\n", similar2)
	fmt.Printf("Text 3: %s\n", different)
	fmt.Printf("\n✅ Similarity (Text 1 vs Text 2): %.4f (should be high)\n", sim12)
	fmt.Printf("✅ Similarity (Text 1 vs Text 3): %.4f (should be low)\n", sim13)
}

// Helper function to calculate cosine similarity
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

