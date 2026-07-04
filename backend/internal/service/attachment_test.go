package service_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/service"
	"workflow-engine/backend/internal/storage"
	"workflow-engine/backend/internal/testutil"
)

func TestAttachmentService_Upload_FeatureDisabled(t *testing.T) {
	ctx := context.Background()
	attRepo := &testutil.MockTaskAttachmentRepo{}
	svc := service.NewAttachmentService(storage.NewNoopStore(), attRepo, logrus.New())

	_, err := svc.Upload(ctx, uuid.New(), "doc.pdf", "application/pdf", []byte("hello"))
	assert.ErrorIs(t, err, service.ErrFeatureDisabled)
}

func TestAttachmentService_Upload_TooLarge(t *testing.T) {
	ctx := context.Background()
	attRepo := &testutil.MockTaskAttachmentRepo{}
	store := &fakeStore{url: "https://example.com/file"}
	svc := service.NewAttachmentService(store, attRepo, logrus.New())

	data := make([]byte, 2*1024*1024+1) // 2MB+1
	_, err := svc.Upload(ctx, uuid.New(), "big.bin", "application/octet-stream", data)
	assert.ErrorIs(t, err, service.ErrFileTooLarge)
}

type fakeStore struct{ url string }

func (f *fakeStore) Upload(_ context.Context, _, _ string, _ io.Reader) (string, error) {
	return f.url, nil
}
func (f *fakeStore) Download(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("data"))), nil
}
func (f *fakeStore) Delete(_ context.Context, _ string) error { return nil }
