package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL            string
	LogLevel               string
	LogFormat              string
	HTTPAddr               string
	WorkflowsDir           string
	WebhookConfig          string
	WebhookWorkerInterval  time.Duration
	WebhookMaxAttempts     int
	CORSAllowedOrigins     string
}

func Load() Config {
	return Config{
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://workflow:workflow@localhost:5432/workflow?sslmode=disable"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		LogFormat:             getEnv("LOG_FORMAT", "json"),
		HTTPAddr:              getEnv("HTTP_ADDR", ":8000"),
		WorkflowsDir:          getEnv("WORKFLOWS_DIR", "./workflows"),
		WebhookConfig:         getEnv("WEBHOOK_CONFIG", "./webhooks.json"),
		WebhookWorkerInterval: getDurationEnv("WEBHOOK_WORKER_INTERVAL", 5*time.Second),
		WebhookMaxAttempts:    getIntEnv("WEBHOOK_MAX_ATTEMPTS", 6),
		CORSAllowedOrigins:    getEnv("CORS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
