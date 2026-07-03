# Task Extensions — Plan 2: Storage Layer + Config

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `FileStore` interface with S3 implementation (behind `S3_ENABLED` feature flag) and extend `config.Config` with S3 env vars.

**Architecture:** `internal/storage/` package defines `FileStore` interface + S3 implementation using `aws-sdk-go-v2`. Config adds 6 new fields. `main.go` wires `FileStore` at startup based on flag. When `S3_ENABLED=false`, a `NoopFileStore` is wired instead (all methods return `ErrFeatureDisabled`).

**Tech Stack:** Go 1.22, `aws-sdk-go-v2/service/s3`, `logrus`.

## Global Constraints

- Logger: `github.com/sirupsen/logrus` — no `fmt.Println`.
- `S3_ENABLED` defaults to `false`; server boots normally without S3 creds.
- `FileStore` interface injected into service layer — service never imports S3 SDK directly.
- Max upload size 2MB (decoded base64).
- Run `make test` green before committing.

---

### Task 5: Extend Config

**Files:**
- Modify: `backend/internal/config/config.go`

**Interfaces:**
- Produces: `config.Config` with fields `S3Enabled bool`, `S3Bucket string`, `S3Region string`, `S3AccessKey string`, `S3SecretKey string`, `S3Endpoint string`

- [ ] **Step 1: Write failing test**

Create `backend/internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/workflow-engine/backend/internal/config"
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
```

- [ ] **Step 2: Run test — expect fail**

```bash
cd backend && go test ./internal/config/... -v 2>&1 | head -20
```

Expected: `cfg.S3Enabled undefined`.

- [ ] **Step 3: Add S3 fields to config.go**

In `backend/internal/config/config.go`, add to `Config` struct:

```go
	// S3 / file upload
	S3Enabled    bool
	S3Bucket     string
	S3Region     string
	S3AccessKey  string
	S3SecretKey  string
	S3Endpoint   string
```

Add to `Load()` return value:

```go
		S3Enabled:   getBoolEnv("S3_ENABLED", false),
		S3Bucket:    getEnv("S3_BUCKET", ""),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),
		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
```

Add `getBoolEnv` helper:

```go
func getBoolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return fallback
}
```

- [ ] **Step 4: Run test — expect pass**

```bash
cd backend && go test ./internal/config/... -v
```

Expected: both test functions pass.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/config/config.go backend/internal/config/config_test.go
git commit -m "feat: add S3 config fields with defaults"
```

---

### Task 6: FileStore Interface + Implementations

**Files:**
- Create: `backend/internal/storage/storage.go` (interface + error + noop)
- Create: `backend/internal/storage/s3.go` (S3 implementation)
- Create: `backend/internal/storage/storage_test.go` (noop unit tests)

**Interfaces:**
- Consumes: `config.Config.S3*` fields (Task 5)
- Produces:
  - `storage.FileStore` interface with `Upload`, `Download`, `Delete`
  - `storage.ErrFeatureDisabled` sentinel error
  - `storage.NewS3Store(cfg config.Config) (*S3Store, error)` — constructor
  - `storage.NewNoopStore() FileStore` — returns noop that returns `ErrFeatureDisabled` on all calls

- [ ] **Step 1: Write unit test for noop**

Create `backend/internal/storage/storage_test.go`:

```go
package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/workflow-engine/backend/internal/storage"
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
```

- [ ] **Step 2: Run test — expect compile fail**

```bash
cd backend && go test ./internal/storage/... -v 2>&1 | head -15
```

Expected: `package storage not found`.

- [ ] **Step 3: Create `storage.go`**

Create `backend/internal/storage/storage.go`:

```go
package storage

import (
	"context"
	"errors"
	"io"
)

var ErrFeatureDisabled = errors.New("file storage is not enabled")
var ErrNotFound = errors.New("attachment not found")

type FileStore interface {
	// Upload stores content at key; returns the full S3 URL.
	Upload(ctx context.Context, key, contentType string, r io.Reader) (url string, err error)
	// Download fetches the object at url, caller must close the returned ReadCloser.
	Download(ctx context.Context, url string) (io.ReadCloser, error)
	// Delete removes the object at url.
	Delete(ctx context.Context, url string) error
}

// noopStore satisfies FileStore when S3_ENABLED=false.
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
```

- [ ] **Step 4: Run test — expect pass**

```bash
cd backend && go test ./internal/storage/... -v
```

Expected: 3 noop tests pass.

- [ ] **Step 5: Add aws-sdk-go-v2 dependency**

```bash
cd backend && go get github.com/aws/aws-sdk-go-v2/config \
    github.com/aws/aws-sdk-go-v2/credentials \
    github.com/aws/aws-sdk-go-v2/service/s3
```

- [ ] **Step 6: Create `s3.go`**

Create `backend/internal/storage/s3.go`:

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/workflow-engine/backend/internal/config"
)

type S3Store struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Store(cfg config.Config) (*S3Store, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey, cfg.S3SecretKey, "",
		)),
	}
	if cfg.S3Endpoint != "" {
		opts = append(opts, awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: cfg.S3Endpoint, HostnameImmutable: true}, nil
			}),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("storage: aws config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg)
	return &S3Store{client: client, bucket: cfg.S3Bucket, region: cfg.S3Region}, nil
}

func (s *S3Store) Upload(ctx context.Context, key, contentType string, r io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        r,
	})
	if err != nil {
		return "", fmt.Errorf("storage: upload %s: %w", key, err)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

func (s *S3Store) Download(ctx context.Context, rawURL string) (io.ReadCloser, error) {
	// Extract key from S3 URL: https://<bucket>.s3.<region>.amazonaws.com/<key>
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid url: %w", err)
	}
	key := parsed.Path[1:] // strip leading /

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: download %s: %w", key, err)
	}
	return resp.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("storage: invalid url: %w", err)
	}
	key := parsed.Path[1:]

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: delete %s: %w", key, err)
	}
	return nil
}
```

- [ ] **Step 7: Verify compile**

```bash
cd backend && go build ./... 2>&1
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/storage/ backend/go.mod backend/go.sum
git commit -m "feat: FileStore interface, NoopStore, S3Store implementation"
```

---

### Task 7: Wire Storage in main.go

**Files:**
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: `storage.NewS3Store`, `storage.NewNoopStore`, `config.Config.S3Enabled` (Tasks 5, 6)
- Produces: `fileStore storage.FileStore` variable passed into services in next plan

- [ ] **Step 1: Add storage wire-up to main.go**

After `log.Info("database connected")` in `main()`, add:

```go
	// File storage
	var fileStore storage.FileStore
	if cfg.S3Enabled {
		s3Store, err := storage.NewS3Store(cfg)
		if err != nil {
			log.WithField("err", err).Fatal("failed to init S3 storage")
		}
		fileStore = s3Store
		log.Info("S3 file store enabled")
	} else {
		fileStore = storage.NewNoopStore()
		log.Info("file storage disabled (S3_ENABLED=false)")
	}
```

Add import: `"github.com/workflow-engine/backend/internal/storage"`

- [ ] **Step 2: Build**

```bash
cd backend && go build ./cmd/server/... 2>&1
```

Expected: build succeeds (fileStore unused warning is OK for now — it will be used in Plan 3).

- [ ] **Step 3: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat: wire FileStore in main.go based on S3_ENABLED flag"
```
