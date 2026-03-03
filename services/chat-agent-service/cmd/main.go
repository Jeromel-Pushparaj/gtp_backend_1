package main

import (
	"flag"
	"log"
	"os"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/mcp"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	baseURL := flag.String("base-url", getEnv("API_BASE_URL", "http://localhost:8080"), "Base URL of the API server")
	apiKey := flag.String("api-key", getEnv("API_KEY", ""), "API key for authentication")
	flag.Parse()

	server := mcp.NewServer(*baseURL, *apiKey)

	if err := server.RegisterTools(); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	log.Printf("MCP Server starting with base URL: %s", *baseURL)
	if *apiKey != "" {
		log.Println("API key authentication enabled")
	}

	log.Println("Server ready and waiting for stdio input from MCP client...")

	if err := server.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	select {}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
