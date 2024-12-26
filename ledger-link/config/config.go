package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	LogLevel    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
