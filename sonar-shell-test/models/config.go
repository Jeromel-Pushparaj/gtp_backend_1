package models

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	GitHubPAT       string
	SonarToken      string
	Organization    string
	SonarOrgKey     string
	DefaultBranch   string
	EnvironmentName string
	JiraToken       string
	JiraDomain      string
	JiraEmail       string
	DatabasePath    string
}

// LoadConfig loads configuration from environment variables or uses defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		GitHubPAT:       os.Getenv("GITHUB_PAT"),
		SonarToken:      os.Getenv("SONAR_TOKEN"),
		Organization:    os.Getenv("GITHUB_ORG"),
		SonarOrgKey:     os.Getenv("SONAR_ORG_KEY"),
		DefaultBranch:   getEnvOrDefault("DEFAULT_BRANCH", "main"),
		EnvironmentName: getEnvOrDefault("ENVIRONMENT_NAME", "production"),
		JiraToken:       os.Getenv("JIRA_TOKEN"),
		JiraDomain:      os.Getenv("JIRA_DOMAIN"),
		JiraEmail:       os.Getenv("JIRA_EMAIL"),
		DatabasePath:    getEnvOrDefault("DATABASE_PATH", "./data/metrics.db"),
	}

	// Validate required fields
	if config.GitHubPAT == "" {
		return nil, fmt.Errorf("GITHUB_PAT environment variable is required")
	}
	if config.Organization == "" {
		return nil, fmt.Errorf("GITHUB_ORG environment variable is required")
	}
	if config.SonarOrgKey == "" {
		config.SonarOrgKey = config.Organization // Default to same as GitHub org
	}

	// SonarToken and Jira credentials are optional
	// They will be validated when those specific services are used

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

