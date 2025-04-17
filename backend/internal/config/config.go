package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration
type Config struct {
	KafkaBrokers    []string
	ServerPort      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	KafkaTimeout    time.Duration
	LogLevel        string
	EnableTLS       bool
	CertFile        string
	KeyFile         string
	EnvironmentName string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		ServerPort:      getEnvWithDefault("PORT", "8080"),
		ReadTimeout:     getEnvDurationWithDefault("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnvDurationWithDefault("WRITE_TIMEOUT", 10*time.Second),
		KafkaTimeout:    getEnvDurationWithDefault("KAFKA_TIMEOUT", 5*time.Second),
		LogLevel:        getEnvWithDefault("LOG_LEVEL", "info"),
		EnableTLS:       getEnvBoolWithDefault("ENABLE_TLS", false),
		CertFile:        getEnvWithDefault("CERT_FILE", ""),
		KeyFile:         getEnvWithDefault("KEY_FILE", ""),
		EnvironmentName: getEnvWithDefault("ENVIRONMENT", "development"),
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable is required")
	}
	config.KafkaBrokers = strings.Split(kafkaBrokersStr, ",")

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks configuration for errors
func (c *Config) validate() error {
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("at least one Kafka broker must be specified")
	}

	if c.EnableTLS {
		if c.CertFile == "" {
			return fmt.Errorf("CERT_FILE must be specified when TLS is enabled")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("KEY_FILE must be specified when TLS is enabled")
		}
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
