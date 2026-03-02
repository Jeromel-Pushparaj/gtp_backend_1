package main

import (
	"log"
	"net/http"

	"github.com/keerthanau/go/config"
	"github.com/keerthanau/go/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration (without sensitive data)
	log.Println("Configuration loaded successfully")

	// Setup HTTP routes
	http.HandleFunc("/api/create-issue", handlers.CreateIssueHandler(cfg))
	http.HandleFunc("/health", handlers.HealthHandler)

	// Start server
	port := "8080"
	log.Printf("Server starting on port %s...", port)
	log.Printf("POST endpoint: http://localhost:%s/api/create-issue", port)
	log.Printf("Health check: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
