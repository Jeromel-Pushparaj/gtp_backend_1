package config

import (
	"fmt"
	"os"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
)

type Config struct {
	ServiceName   string
	ServicePort   string
	ServiceHost   string
	SlackBotToken string
	SlackBotID    string
	SlackBotName  string
	LogLevel      string
	Environment   string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		ServiceName:   getEnv(constants.EnvServiceName, constants.DefaultServiceName),
		ServicePort:   getEnv(constants.EnvServicePort, constants.DefaultServicePort),
		ServiceHost:   getEnv(constants.EnvServiceHost, constants.DefaultServiceHost),
		SlackBotToken: os.Getenv(constants.EnvSlackBotToken),
		SlackBotID:    os.Getenv(constants.EnvSlackBotID),
		SlackBotName:  getEnv(constants.EnvSlackBotName, constants.DefaultBotName),
		LogLevel:      getEnv(constants.EnvLogLevel, constants.DefaultLogLevel),
		Environment:   getEnv(constants.EnvEnvironment, constants.DefaultEnvironment),
	}

	if config.SlackBotToken == "" {
		return nil, fmt.Errorf(constants.ErrorSlackTokenRequired)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
