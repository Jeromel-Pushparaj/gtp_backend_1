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

	// HTTP Client Configuration (for proxy connections)
	HTTPTimeout          time.Duration
	HTTPMaxIdleConns     int
	HTTPMaxConnsPerHost  int
	HTTPIdleConnTimeout  time.Duration
	HTTPTLSTimeout       time.Duration

	// Server Timeout Configuration
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration

	// Service URLs - All microservices in the ecosystem
	JiraTriggerServiceURL string // Port 8086 - Jira issue creation service
	ChatAgentServiceURL   string // Port 8082 - AI chat agent service
	ApprovalServiceURL    string // Port 8083 - Slack approval workflow service
	OnboardingServiceURL  string // Port 8084 - Service catalog/onboarding service
	ScoreCardServiceURL   string // Port 8085 - Service scorecard evaluation service
	SonarShellServiceURL  string // Port 8080 - SonarCloud automation service
	PagerDutyServiceURL   string // Port 8091 - PagerDuty service

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

	// Parse HTTP client timeouts
	httpTimeout, err := time.ParseDuration(getEnv("HTTP_TIMEOUT", "30s"))
	if err != nil {
		log.Printf("WARNING: Invalid HTTP_TIMEOUT, using default 30s: %v", err)
		httpTimeout = 30 * time.Second
	}

	httpIdleConnTimeout, err := time.ParseDuration(getEnv("HTTP_IDLE_CONN_TIMEOUT", "90s"))
	if err != nil {
		log.Printf("WARNING: Invalid HTTP_IDLE_CONN_TIMEOUT, using default 90s: %v", err)
		httpIdleConnTimeout = 90 * time.Second
	}

	httpTLSTimeout, err := time.ParseDuration(getEnv("HTTP_TLS_TIMEOUT", "10s"))
	if err != nil {
		log.Printf("WARNING: Invalid HTTP_TLS_TIMEOUT, using default 10s: %v", err)
		httpTLSTimeout = 10 * time.Second
	}

	// Parse server timeouts
	serverReadTimeout, err := time.ParseDuration(getEnv("SERVER_READ_TIMEOUT", "15s"))
	if err != nil {
		log.Printf("WARNING: Invalid SERVER_READ_TIMEOUT, using default 15s: %v", err)
		serverReadTimeout = 15 * time.Second
	}

	serverWriteTimeout, err := time.ParseDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s"))
	if err != nil {
		log.Printf("WARNING: Invalid SERVER_WRITE_TIMEOUT, using default 30s: %v", err)
		serverWriteTimeout = 30 * time.Second
	}

	serverIdleTimeout, err := time.ParseDuration(getEnv("SERVER_IDLE_TIMEOUT", "120s"))
	if err != nil {
		log.Printf("WARNING: Invalid SERVER_IDLE_TIMEOUT, using default 120s: %v", err)
		serverIdleTimeout = 120 * time.Second
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

		// HTTP Client Configuration
		HTTPTimeout:         httpTimeout,
		HTTPMaxIdleConns:    getEnvAsInt("HTTP_MAX_IDLE_CONNS", 100),
		HTTPMaxConnsPerHost: getEnvAsInt("HTTP_MAX_CONNS_PER_HOST", 100),
		HTTPIdleConnTimeout: httpIdleConnTimeout,
		HTTPTLSTimeout:      httpTLSTimeout,

		// Server Timeout Configuration
		ServerReadTimeout:  serverReadTimeout,
		ServerWriteTimeout: serverWriteTimeout,
		ServerIdleTimeout:  serverIdleTimeout,

		// Service URLs
		JiraTriggerServiceURL: getEnv("JIRA_TRIGGER_SERVICE_URL", "http://localhost:8086"),
		ChatAgentServiceURL:   getEnv("CHAT_AGENT_SERVICE_URL", "http://localhost:8082"),
		ApprovalServiceURL:    getEnv("APPROVAL_SERVICE_URL", "http://localhost:8083"),
		OnboardingServiceURL:  getEnv("ONBOARDING_SERVICE_URL", "http://localhost:8084"),
		ScoreCardServiceURL:   getEnv("SCORECARD_SERVICE_URL", "http://localhost:8085"),
		SonarShellServiceURL:  getEnv("SONAR_SHELL_SERVICE_URL", "http://localhost:8080"),
		PagerDutyServiceURL:   getEnv("PAGER_Duty_SERVICE_URL", "http://localhost:8091"),

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
