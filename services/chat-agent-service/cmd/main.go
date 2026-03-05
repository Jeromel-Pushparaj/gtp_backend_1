package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/agent"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/server"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/session"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Command-line flags
	port := flag.String("port", getEnv("PORT", "8082"), "HTTP server port")
	groqAPIKey := flag.String("groq-api-key", getEnv("GROQ_API_KEY", ""), "Groq API key")
	backendURL := flag.String("backend-url", getEnv("BACKEND_URL", "http://localhost:8080"), "Backend API URL")
	backendAPIKey := flag.String("backend-api-key", getEnv("BACKEND_API_KEY", ""), "Backend API key")
	
	// Redis configuration for conversation memory
	redisHost := flag.String("redis-host", getEnv("REDIS_HOST", "localhost"), "Redis host")
	redisPort := flag.String("redis-port", getEnv("REDIS_PORT", "6379"), "Redis port")
	redisPassword := flag.String("redis-password", getEnv("REDIS_PASSWORD", ""), "Redis password")
	
	flag.Parse()

	// Validate required configuration
	if *groqAPIKey == "" {
		log.Fatal("GROQ_API_KEY is required. Set it via environment variable or -groq-api-key flag")
	}

	log.Println("🤖 Chat Agent Service")
	log.Println("=====================")
	log.Printf("Port: %s", *port)
	log.Printf("Backend URL: %s", *backendURL)
	log.Printf("Groq API Key: %s", maskAPIKey(*groqAPIKey))
	log.Printf("Redis: %s:%s", *redisHost, *redisPort)
	log.Println()

	// Create chat agent
	chatAgent := agent.NewChatAgent(*groqAPIKey, *backendURL, *backendAPIKey)

	// Create Redis session manager for conversation memory
	log.Println("🔄 Initializing Redis session manager...")
	sessionManager, err := session.NewRedisSessionManager(*redisHost, *redisPort, *redisPassword)
	if err != nil {
		log.Fatalf("Failed to initialize Redis session manager: %v", err)
	}
	defer sessionManager.Close()

	// Create and start HTTP server
	httpServer := server.NewHTTPServer(chatAgent, sessionManager, *port)
	if err := httpServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}