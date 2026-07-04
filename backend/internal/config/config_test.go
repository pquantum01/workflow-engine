package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/config"
)

func TestConfig_S3Defaults(t *testing.T) {
	os.Unsetenv("S3_ENABLED")
	cfg := config.Load()
	assert.False(t, cfg.S3Enabled)
	assert.Equal(t, "us-east-1", cfg.S3Region)
}

func TestConfig_S3Enabled(t *testing.T) {
	os.Setenv("S3_ENABLED", "true")
	os.Setenv("S3_BUCKET", "my-bucket")
	defer os.Unsetenv("S3_ENABLED")
	defer os.Unsetenv("S3_BUCKET")

	cfg := config.Load()
	assert.True(t, cfg.S3Enabled)
	assert.Equal(t, "my-bucket", cfg.S3Bucket)
}
