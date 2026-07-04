package storage

import (
	"context"
	"errors"
	"io"
)

var ErrFeatureDisabled = errors.New("file storage is not enabled")
var ErrNotFound = errors.New("attachment not found")

type FileStore interface {
	Upload(ctx context.Context, key, contentType string, r io.Reader) (url string, err error)
	Download(ctx context.Context, url string) (io.ReadCloser, error)
	Delete(ctx context.Context, url string) error
}

type noopStore struct{}

func NewNoopStore() FileStore { return &noopStore{} }

func (n *noopStore) Upload(_ context.Context, _, _ string, _ io.Reader) (string, error) {
	return "", ErrFeatureDisabled
}
func (n *noopStore) Download(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, ErrFeatureDisabled
}
func (n *noopStore) Delete(_ context.Context, _ string) error {
	return ErrFeatureDisabled
}
