package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the API Gateway
type Config struct {
	// Gateway Configuration
	GatewayPort string
	GatewayHost string
	Environment string
	LogLevel    string

	// Security Configuration
	JWTSecret          string
	JWTExpiration      time.Duration
	CORSAllowedOrigins string

	// Rate Limiting
	RateLimitRequests int
	RateLimitDuration time.Duration

	// Service URLs - All microservices in the ecosystem
	JiraTriggerServiceURL string // Port 8086 - Jira issue creation service
	ChatAgentServiceURL   string // Port 8082 - AI chat agent service
	ApprovalServiceURL    string // Port 8083 - Slack approval workflow service
	OnboardingServiceURL  string // Port 8084 - Service catalog/onboarding service
	ScoreCardServiceURL   string // Port 8085 - Service scorecard evaluation service
	SonarShellServiceURL  string // Port 8080 - SonarCloud automation service

	// Kafka Configuration (optional - for event publishing)
	KafkaBrokers string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: No .env file found, using environment variables")
	}

	// Parse JWT expiration duration
	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Printf("WARNING: Invalid JWT_EXPIRATION, using default 24h: %v", err)
		jwtExpiration = 24 * time.Hour
	}

	// Parse rate limit duration
	rateLimitDuration, err := time.ParseDuration(getEnv("RATE_LIMIT_DURATION", "1m"))
	if err != nil {
		log.Printf("WARNING: Invalid RATE_LIMIT_DURATION, using default 1m: %v", err)
		rateLimitDuration = 1 * time.Minute
	}

	config := &Config{
		// Gateway settings
		GatewayPort: getEnv("GATEWAY_PORT", "8089"),
		GatewayHost: getEnv("GATEWAY_HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "debug"),

		// Security
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration:      jwtExpiration,
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),

		// Rate limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: rateLimitDuration,

		// Service URLs
		JiraTriggerServiceURL: getEnv("JIRA_TRIGGER_SERVICE_URL", "http://localhost:8086"),
		ChatAgentServiceURL:   getEnv("CHAT_AGENT_SERVICE_URL", "http://localhost:8082"),
		ApprovalServiceURL:    getEnv("APPROVAL_SERVICE_URL", "http://localhost:8083"),
		OnboardingServiceURL:  getEnv("ONBOARDING_SERVICE_URL", "http://localhost:8084"),
		ScoreCardServiceURL:   getEnv("SCORECARD_SERVICE_URL", "http://localhost:8085"),
		SonarShellServiceURL:  getEnv("SONAR_SHELL_SERVICE_URL", "http://localhost:8080"),

		// Kafka
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
	}

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

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		log.Printf("WARNING: Invalid integer value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}
