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
	SlackAppToken string
	LogLevel      string
	Environment   string
	KafkaBrokers  string
	KafkaGroupID  string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		ServiceName:   getEnv(constants.EnvServiceName, constants.DefaultServiceName),
		ServicePort:   getEnv(constants.EnvServicePort, constants.DefaultServicePort),
		ServiceHost:   getEnv(constants.EnvServiceHost, constants.DefaultServiceHost),
		SlackBotToken: os.Getenv(constants.EnvSlackBotToken),
		SlackAppToken: os.Getenv(constants.EnvSlackAppToken),
		LogLevel:      getEnv(constants.EnvLogLevel, constants.DefaultLogLevel),
		Environment:   getEnv(constants.EnvEnvironment, constants.DefaultEnvironment),
		KafkaBrokers:  getEnv(constants.EnvKafkaBrokers, constants.DefaultKafkaBrokers),
		KafkaGroupID:  getEnv(constants.EnvKafkaGroupID, constants.DefaultKafkaGroupID),
	}

	if config.SlackBotToken == "" {
		return nil, fmt.Errorf(constants.ErrorSlackTokenRequired)
	}

	if config.SlackAppToken == "" {
		return nil, fmt.Errorf(constants.ErrorAppTokenRequired)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
