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

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Kafka
	KafkaBrokers              string
	KafkaTopicScoreCalculated string
	KafkaTopicScoreRequested  string
	KafkaGroupID              string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

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

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "scorecard_db"),

		// Kafka
		KafkaBrokers:              getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicScoreCalculated: getEnv("KAFKA_TOPIC_SCORE_CALCULATED", "score.calculated"),
		KafkaTopicScoreRequested:  getEnv("KAFKA_TOPIC_SCORE_REQUESTED", "score.requested"),
		KafkaGroupID:              getEnv("KAFKA_GROUP_ID", "score-card-service"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

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
