package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL           string
	LogLevel              string
	LogFormat             string
	HTTPAddr              string
	WorkflowsDir          string
	WebhookConfig         string
	WebhookWorkerInterval time.Duration
	WebhookMaxAttempts    int
	CORSAllowedOrigins    string
	// S3 / file upload
	S3Enabled   bool
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string
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
		S3Enabled:             getBoolEnv("S3_ENABLED", false),
		S3Bucket:              getEnv("S3_BUCKET", ""),
		S3Region:              getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:           getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:           getEnv("S3_SECRET_KEY", ""),
		S3Endpoint:            getEnv("S3_ENDPOINT", ""),
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

func getBoolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
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
