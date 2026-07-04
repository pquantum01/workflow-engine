package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/storage"
)

func TestNoopStore_Upload(t *testing.T) {
	s := storage.NewNoopStore()
	_, err := s.Upload(context.Background(), "key", "text/plain", strings.NewReader("hello"))
	assert.ErrorIs(t, err, storage.ErrFeatureDisabled)
}

func TestNoopStore_Download(t *testing.T) {
	s := storage.NewNoopStore()
	_, err := s.Download(context.Background(), "https://example.com/file")
	assert.ErrorIs(t, err, storage.ErrFeatureDisabled)
}

func TestNoopStore_Delete(t *testing.T) {
	s := storage.NewNoopStore()
	err := s.Delete(context.Background(), "https://example.com/file")
	assert.ErrorIs(t, err, storage.ErrFeatureDisabled)
}
