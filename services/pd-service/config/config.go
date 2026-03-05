package config

import (
	"os"
)

type Config struct {
	PagerDutyAPIKey string
	SlackBotToken   string
	GitHubPAT       string
	DefaultOrg      string
	Port            string
	MongoURL        string
	DBName          string
}

func Load() *Config {
	return &Config{
		PagerDutyAPIKey: getEnv("PAGERDUTY_API_KEY", ""),
		SlackBotToken:   getEnv("SLACK_BOT_TOKEN", ""),
		GitHubPAT:       getEnv("GITHUB_PAT", ""),
		DefaultOrg:      getEnv("DEFAULT_ORG", "teknex-poc"),
		Port:            getEnv("PORT", "8080"),
		MongoURL:        getEnv("MONGO_URL", ""),
		DBName:          getEnv("DB_NAME", "pd_service"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

