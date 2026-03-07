package config

import (
	"os"
)

// Config holds all configuration for the score-card service
type Config struct {
	ServiceName string
	ServicePort string
	ServiceHost string
	Environment string
	LogLevel    string

	// External APIs
	MetricsAPIBaseURL string // Base URL for GitHub/Jira/Sonar metrics API (e.g., http://localhost:8080)
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "score-card-service"),
		ServicePort: getEnv("SERVICE_PORT", "8085"),
		ServiceHost: getEnv("SERVICE_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),

		// External APIs
		MetricsAPIBaseURL: getEnv("METRICS_API_BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
