package main

import (
	"flag"
	"log"
	"os"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/mcp"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Command-line flags
	baseURL := flag.String("base-url", getEnv("API_BASE_URL", "http://localhost:8080"), "Backend API URL")
	apiKey := flag.String("api-key", getEnv("API_KEY", ""), "API key for backend authentication")
	port := flag.String("port", getEnv("MCP_PORT", "8083"), "MCP server HTTP port")
	
	flag.Parse()

	// Create MCP server
	server := mcp.NewServer(*baseURL, *apiKey, *port)
	
	// Register tools (no-op for HTTP mode, but kept for consistency)
	if err := server.RegisterTools(); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Start HTTP server
	log.Println("🚀 MCP Server (HTTP Mode)")
	log.Println("==========================")
	log.Printf("Port: %s", *port)
	log.Printf("Backend API: %s", *baseURL)
	if *apiKey != "" {
		log.Println("Backend API authentication: enabled")
	}
	log.Println()

	if err := server.Serve(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}