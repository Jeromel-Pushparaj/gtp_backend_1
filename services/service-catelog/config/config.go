package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServiceName string
	ServicePort string
	ServiceHost string
	Environment string
	LogLevel    string
}

// GlobalConfig is the global configuration instance
var GlobalConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		ServiceName: getEnv("SERVICE_NAME", "onboarding-service"),
		ServicePort: getEnv("SERVICE_PORT", "8084"),
		ServiceHost: getEnv("SERVICE_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
	}

	GlobalConfig = config
	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

