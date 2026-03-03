package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	BaseURL    string
	Email      string
	APIToken   string
	ProjectKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get API token
	apiToken := os.Getenv("JIRA_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("JIRA_API_TOKEN environment variable is required")
	}

	// Get other required variables
	baseURL := os.Getenv("JIRA_BASE_URL")
	email := os.Getenv("JIRA_EMAIL")
	projectKey := os.Getenv("JIRA_PROJECT_KEY")

	// Validate required fields
	if baseURL == "" || email == "" || projectKey == "" {
		return nil, fmt.Errorf("missing required environment variables: JIRA_BASE_URL, JIRA_EMAIL, JIRA_PROJECT_KEY")
	}

	return &Config{
		BaseURL:    baseURL,
		Email:      email,
		APIToken:   apiToken,
		ProjectKey: projectKey,
	}, nil
}

