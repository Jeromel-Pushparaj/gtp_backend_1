package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the service-catalog
type Config struct {
	// Service Configuration
	ServiceName string
	ServicePort string
	ServiceHost string
	Environment string
	LogLevel    string

	// Sonar Shell Test API Configuration
	SonarShellTestURL    string
	SonarShellTestAPIKey string

	// Local Database Configuration
	DBPath string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		// Service Configuration
		ServiceName: getEnv("SERVICE_NAME", "service-catalog"),
		ServicePort: getEnv("SERVICE_PORT", "8084"),
		ServiceHost: getEnv("SERVICE_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),

		// Sonar Shell Test API Configuration
		SonarShellTestURL:    getEnv("SONAR_SHELL_TEST_URL", "http://localhost:8080"),
		SonarShellTestAPIKey: getEnv("SONAR_SHELL_TEST_API_KEY", ""),

		// Local Database Configuration
		DBPath: getEnv("DB_PATH", "./data/cache.db"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
