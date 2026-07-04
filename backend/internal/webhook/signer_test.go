package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/webhook"
)

func TestSign_ProducesHMACSHA256(t *testing.T) {
	sig := webhook.Sign("secret", []byte(`{"hello":"world"}`))
	assert.True(t, len(sig) > 7)
	assert.Equal(t, "sha256=", sig[:7])
}

func TestSign_DifferentSecrets_DifferentSig(t *testing.T) {
	body := []byte(`{"event":"test"}`)
	sig1 := webhook.Sign("secret1", body)
	sig2 := webhook.Sign("secret2", body)
	assert.NotEqual(t, sig1, sig2)
}

func TestSign_SameInputDeterministic(t *testing.T) {
	body := []byte(`{"event":"test"}`)
	assert.Equal(t, webhook.Sign("key", body), webhook.Sign("key", body))
}
