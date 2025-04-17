package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers []string
}

func LoadConfig() (*Config, error) {

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable is not set")
	}

	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")

	return &Config{
		KafkaBrokers: kafkaBrokers,
	}, nil
}

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("ConfigError: %s", e.Message)
}
