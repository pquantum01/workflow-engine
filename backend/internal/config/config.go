package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://workflow:workflow@localhost:5432/workflow_engine?sslmode=disable&search_path=workflow_engine"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
