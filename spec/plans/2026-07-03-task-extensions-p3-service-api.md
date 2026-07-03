# Task Extensions — Plan 3: Service Layer + API Handlers

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `TaskService` extensions for PATCH + ad-hoc creation; add `CommentService` and `AttachmentService`; wire 7 new HTTP routes; update DTOs; update `main.go` and `router.go`.

**Architecture:** `TaskService` gains `UpdateFields` and `CreateAdhoc`. Two new services `CommentService` and `AttachmentService` are thin wrappers over their repos (+ `FileStore` for attachment). New handler file `task_extensions.go` implements 5 new endpoints. Existing `task.go` handler gains PATCH. Router registers all 7 new routes.

**Tech Stack:** Go 1.22, `gorilla/mux`, `logrus`, `testify`.

## Global Constraints

- Logger: `logrus` — no `fmt.Println`.
- All handlers thin — logic in service layer.
- Sentinel errors in `service/errors.go`; handlers map to HTTP codes.
- Ad-hoc `task_key` = slugified label (lowercase, non-alphanum → `_`, max 40 chars) + `_` + 4-char hex random.
- PATCH is partial: only fields present in JSON body are applied. Use pointer fields in request DTO.
- Run `make test` green before committing each task.

---

### Task 8: Service Errors + TaskService Extensions

**Files:**
- Modify: `backend/internal/service/errors.go`
- Modify: `backend/internal/service/task.go`
- Create: `backend/internal/service/task_test.go` (unit tests with mocks)

**Interfaces:**
- Consumes: `repository.TaskInstanceRepo.Create`, `UpdateFields`, `TaskInstancePatch`; `model.TaskInstance.IsTerminal()` (Plan 1)
- Produces:
  - `service.ErrTaskCompleted` sentinel
  - `service.ErrInvalidState` sentinel
  - `service.ErrInstanceNotActive` sentinel
  - `TaskService.UpdateFields(ctx, id, patch) error`
  - `TaskService.CreateAdhoc(ctx, instanceID, req AdhocTaskRequest) (*model.TaskInstance, error)`

- [ ] **Step 1: Add new sentinel errors**

In `backend/internal/service/errors.go`, add:

```go
var ErrTaskCompleted    = errors.New("task is completed, field update not allowed")
var ErrInvalidState     = errors.New("state_id does not exist in workflow definition")
var ErrInstanceNotActive = errors.New("instance is not active")
```

- [ ] **Step 2: Write failing unit tests**

Create `backend/internal/service/task_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/service"
	"github.com/workflow-engine/backend/internal/testutil"
)

func TestTaskService_UpdateFields_TerminalTask(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()

	taskRepo := &testutil.MockTaskInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
			return &model.TaskInstance{ID: id, Status: model.TaskStatusDone, Priority: "medium"}, nil
		},
	}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	priority := "high"
	err := svc.UpdateFields(ctx, taskID, service.TaskFieldPatch{Priority: &priority})
	assert.ErrorIs(t, err, service.ErrTaskCompleted)
}

func TestTaskService_UpdateFields_Success(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()
	var patchReceived repository.TaskInstancePatch

	taskRepo := &testutil.MockTaskInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
			return &model.TaskInstance{ID: id, Status: model.TaskStatusNew, Priority: "medium"}, nil
		},
		UpdateFieldsFn: func(_ context.Context, id uuid.UUID, p repository.TaskInstancePatch) error {
			patchReceived = p
			return nil
		},
	}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	priority := "high"
	err := svc.UpdateFields(ctx, taskID, service.TaskFieldPatch{Priority: &priority})
	require.NoError(t, err)
	require.NotNil(t, patchReceived.Priority)
	assert.Equal(t, "high", *patchReceived.Priority)
}

func TestTaskService_CreateAdhoc_InvalidState(t *testing.T) {
	ctx := context.Background()
	instID := uuid.New()
	defID := uuid.New()

	instRepo := &testutil.MockWorkflowInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
			return &model.WorkflowInstance{
				ID: id, WorkflowDefID: defID, Status: model.InstanceStatusActive,
			}, nil
		},
	}
	defRepo := &testutil.MockWorkflowDefRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
			return &model.WorkflowDef{
				ID:         id,
				Definition: model.WorkflowDefinition{States: []model.StateDef{{ID: "review"}}},
			}, nil
		},
	}
	taskRepo := &testutil.MockTaskInstanceRepo{}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	svc.SetInstRepo(instRepo)
	svc.SetDefRepo(defRepo)

	_, err := svc.CreateAdhoc(ctx, instID, service.AdhocTaskRequest{
		StateID:   "nonexistent",
		Label:     "My Task",
		TaskType:  "manual",
		CreatedBy: "alice",
	})
	assert.ErrorIs(t, err, service.ErrInvalidState)
}
```

- [ ] **Step 3: Run test — expect compile fail**

```bash
cd backend && go test ./internal/service/... -v 2>&1 | head -20
```

Expected: `service.TaskFieldPatch undefined`, `svc.UpdateFields undefined`, etc.

- [ ] **Step 4: Add `TaskFieldPatch`, `AdhocTaskRequest`, `UpdateFields`, `CreateAdhoc` to task.go**

Add `TaskFieldPatch` and `AdhocTaskRequest` types near top of `task.go`:

```go
type TaskFieldPatch struct {
	Priority               *string
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
}

type AdhocTaskRequest struct {
	StateID                string
	Label                  string
	TaskType               string
	CreatedBy              string
	Priority               string // defaults to "medium" if empty
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
}
```

Add repo fields to `TaskService` struct (inject at construction via setters for testability):

```go
type TaskService struct {
	eng      *engine.Engine
	taskRepo repository.TaskInstanceRepo
	instRepo repository.WorkflowInstanceRepo
	defRepo  repository.WorkflowDefRepo
	log      *logrus.Logger
}

func NewTaskService(
	eng *engine.Engine,
	taskRepo repository.TaskInstanceRepo,
	log *logrus.Logger,
) *TaskService {
	return &TaskService{eng: eng, taskRepo: taskRepo, log: log}
}

func (s *TaskService) SetInstRepo(r repository.WorkflowInstanceRepo) { s.instRepo = r }
func (s *TaskService) SetDefRepo(r repository.WorkflowDefRepo)       { s.defRepo = r }
```

Add `UpdateFields`:

```go
func (s *TaskService) UpdateFields(ctx context.Context, id uuid.UUID, patch TaskFieldPatch) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if task.IsTerminal() {
		return ErrTaskCompleted
	}
	return s.taskRepo.UpdateFields(ctx, id, repository.TaskInstancePatch{
		Priority:               patch.Priority,
		ExpectedCompletionDate: patch.ExpectedCompletionDate,
		ExpectedCompletionTime: patch.ExpectedCompletionTime,
	})
}
```

Add `CreateAdhoc`:

```go
func (s *TaskService) CreateAdhoc(ctx context.Context, instanceID uuid.UUID, req AdhocTaskRequest) (*model.TaskInstance, error) {
	inst, err := s.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if inst.Status != model.InstanceStatusActive {
		return nil, ErrInstanceNotActive
	}

	def, err := s.defRepo.GetByID(ctx, inst.WorkflowDefID)
	if err != nil {
		return nil, err
	}

	stateExists := false
	for _, st := range def.Definition.States {
		if st.ID == req.StateID {
			stateExists = true
			break
		}
	}
	if !stateExists {
		return nil, ErrInvalidState
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	task := &model.TaskInstance{
		ID:                     uuid.New(),
		WorkflowInstanceID:     instanceID,
		StateID:                req.StateID,
		TaskKey:                generateAdhocKey(req.Label),
		TaskLabel:              req.Label,
		TaskType:               req.TaskType,
		Status:                 model.TaskStatusNew,
		CreatedBy:              req.CreatedBy,
		CreatedAt:              time.Now(),
		Priority:               priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
		IsAdhoc:                true,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.log.WithField("err", err).Error("taskService.CreateAdhoc: repo error")
		return nil, err
	}
	return task, nil
}

func generateAdhocKey(label string) string {
	slug := make([]byte, 0, len(label))
	for _, ch := range strings.ToLower(label) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			slug = append(slug, byte(ch))
		} else {
			if len(slug) > 0 && slug[len(slug)-1] != '_' {
				slug = append(slug, '_')
			}
		}
	}
	// trim trailing underscore
	for len(slug) > 0 && slug[len(slug)-1] == '_' {
		slug = slug[:len(slug)-1]
	}
	if len(slug) > 40 {
		slug = slug[:40]
	}
	b := make([]byte, 2)
	rand.Read(b)
	return fmt.Sprintf("%s_%x", string(slug), b)
}
```

Add imports needed: `"crypto/rand"`, `"fmt"`, `"strings"`, `"time"`.

- [ ] **Step 5: Run tests — expect pass**

```bash
cd backend && go test ./internal/service/... -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/errors.go backend/internal/service/task.go backend/internal/service/task_test.go
git commit -m "feat: TaskService.UpdateFields and CreateAdhoc with unit tests"
```

---

### Task 9: CommentService + AttachmentService

**Files:**
- Create: `backend/internal/service/comment.go`
- Create: `backend/internal/service/attachment.go`
- Create: `backend/internal/service/comment_test.go`
- Create: `backend/internal/service/attachment_test.go`

**Interfaces:**
- Consumes: `TaskCommentRepo`, `TaskAttachmentRepo`, `storage.FileStore`, `storage.ErrFeatureDisabled` (Plans 1, 2)
- Produces:
  - `CommentService.Add(ctx, taskID, author, content) (*model.TaskComment, error)`
  - `CommentService.List(ctx, taskID) ([]*model.TaskComment, error)`
  - `AttachmentService.Upload(ctx, taskID, filename, contentType string, data []byte) (*model.TaskAttachment, error)`
  - `AttachmentService.Download(ctx, taskID) (io.ReadCloser, *model.TaskAttachment, error)`
  - `AttachmentService.Delete(ctx, taskID) error`
  - `service.ErrFeatureDisabled` (wraps storage.ErrFeatureDisabled)
  - `service.ErrFileTooLarge`

- [ ] **Step 1: Write failing tests**

Create `backend/internal/service/comment_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/service"
	"github.com/workflow-engine/backend/internal/testutil"
)

func TestCommentService_Add(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()
	var saved *model.TaskComment

	commentRepo := &testutil.MockTaskCommentRepo{
		AddFn: func(_ context.Context, c *model.TaskComment) error {
			saved = c
			return nil
		},
	}
	svc := service.NewCommentService(commentRepo, logrus.New())
	got, err := svc.Add(ctx, taskID, "alice", "looks good")
	require.NoError(t, err)
	assert.Equal(t, taskID, got.TaskInstanceID)
	assert.Equal(t, "alice", got.Author)
	assert.Equal(t, "looks good", got.Content)
	assert.NotNil(t, saved)
}

func TestCommentService_List(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()

	commentRepo := &testutil.MockTaskCommentRepo{
		ListByTaskFn: func(_ context.Context, id uuid.UUID) ([]*model.TaskComment, error) {
			return []*model.TaskComment{
				{TaskInstanceID: id, Author: "alice", Content: "c1"},
			}, nil
		},
	}
	svc := service.NewCommentService(commentRepo, logrus.New())
	list, err := svc.List(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "c1", list[0].Content)
}
```

Create `backend/internal/service/attachment_test.go`:

```go
package service_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/workflow-engine/backend/internal/service"
	"github.com/workflow-engine/backend/internal/storage"
	"github.com/workflow-engine/backend/internal/testutil"
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
```

- [ ] **Step 2: Run — expect compile fail**

```bash
cd backend && go test ./internal/service/... -v 2>&1 | head -20
```

Expected: `service.NewCommentService undefined`.

- [ ] **Step 3: Create `comment.go`**

Create `backend/internal/service/comment.go`:

```go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

type CommentService struct {
	commentRepo repository.TaskCommentRepo
	log         *logrus.Logger
}

func NewCommentService(commentRepo repository.TaskCommentRepo, log *logrus.Logger) *CommentService {
	return &CommentService{commentRepo: commentRepo, log: log}
}

func (s *CommentService) Add(ctx context.Context, taskID uuid.UUID, author, content string) (*model.TaskComment, error) {
	c := &model.TaskComment{
		ID:             uuid.New(),
		TaskInstanceID: taskID,
		Author:         author,
		Content:        content,
		CreatedAt:      time.Now(),
	}
	if err := s.commentRepo.Add(ctx, c); err != nil {
		s.log.WithField("err", err).Error("commentService.Add")
		return nil, err
	}
	return c, nil
}

func (s *CommentService) List(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error) {
	return s.commentRepo.ListByTask(ctx, taskID)
}
```

- [ ] **Step 4: Create `attachment.go`**

Create `backend/internal/service/attachment.go`:

```go
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/storage"
)

const maxUploadBytes = 2 * 1024 * 1024 // 2MB

var ErrFeatureDisabled = errors.New("file upload is not enabled")
var ErrFileTooLarge    = errors.New("file exceeds 2MB limit")

type AttachmentService struct {
	store   storage.FileStore
	attRepo repository.TaskAttachmentRepo
	log     *logrus.Logger
}

func NewAttachmentService(store storage.FileStore, attRepo repository.TaskAttachmentRepo, log *logrus.Logger) *AttachmentService {
	return &AttachmentService{store: store, attRepo: attRepo, log: log}
}

func (s *AttachmentService) Upload(ctx context.Context, taskID uuid.UUID, filename, contentType string, data []byte) (*model.TaskAttachment, error) {
	if len(data) > maxUploadBytes {
		return nil, ErrFileTooLarge
	}

	key := fmt.Sprintf("attachments/%s/%s", taskID, filename)
	s3URL, err := s.store.Upload(ctx, key, contentType, bytes.NewReader(data))
	if err != nil {
		if errors.Is(err, storage.ErrFeatureDisabled) {
			return nil, ErrFeatureDisabled
		}
		s.log.WithField("err", err).Error("attachmentService.Upload: store error")
		return nil, err
	}

	now := time.Now()
	att := &model.TaskAttachment{
		ID:             uuid.New(),
		TaskInstanceID: taskID,
		Filename:       filename,
		ContentType:    contentType,
		SizeBytes:      len(data),
		S3URL:          s3URL,
		UploadedAt:     now,
		UpdatedAt:      now,
	}
	if err := s.attRepo.Upsert(ctx, att); err != nil {
		s.log.WithField("err", err).Error("attachmentService.Upload: upsert error")
		return nil, err
	}
	return att, nil
}

func (s *AttachmentService) Download(ctx context.Context, taskID uuid.UUID) (io.ReadCloser, *model.TaskAttachment, error) {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	rc, err := s.store.Download(ctx, att.S3URL)
	if err != nil {
		if errors.Is(err, storage.ErrFeatureDisabled) {
			return nil, nil, ErrFeatureDisabled
		}
		return nil, nil, err
	}
	return rc, att, nil
}

func (s *AttachmentService) Delete(ctx context.Context, taskID uuid.UUID) error {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.store.Delete(ctx, att.S3URL); err != nil && !errors.Is(err, storage.ErrFeatureDisabled) {
		return err
	}
	return s.attRepo.DeleteByTaskID(ctx, taskID)
}

func (s *AttachmentService) GetMeta(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error) {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return att, nil
}
```

- [ ] **Step 5: Run tests — expect pass**

```bash
cd backend && go test ./internal/service/... -v
```

Expected: all tests pass including existing service tests.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/comment.go backend/internal/service/attachment.go \
        backend/internal/service/comment_test.go backend/internal/service/attachment_test.go
git commit -m "feat: CommentService and AttachmentService with unit tests"
```

---

### Task 10: DTOs for New Endpoints

**Files:**
- Modify: `backend/internal/api/dto/task.go`
- Create: `backend/internal/api/dto/comment.go`
- Create: `backend/internal/api/dto/attachment.go`

**Interfaces:**
- Produces:
  - `dto.PatchTaskRequest{Priority *string, ExpectedCompletionDate *string, ExpectedCompletionTime *string}`
  - `dto.CreateAdhocTaskRequest{StateID, Label, TaskType, CreatedBy string; Priority string; ExpectedCompletionDate, ExpectedCompletionTime *string}`
  - `dto.AddCommentRequest{Author, Content string}`
  - `dto.CommentResponse{ID, TaskInstanceID, Author, Content, CreatedAt string}`
  - `dto.UploadAttachmentRequest{Filename, ContentType, Data string}` (Data = base64)
  - `dto.AttachmentMetaResponse{ID, Filename, ContentType string; SizeBytes int; UploadedAt, UpdatedAt string}`
  - `dto.TaskInstanceResponse` extended with `Priority`, `ExpectedCompletionDate`, `ExpectedCompletionTime *string`, `IsAdhoc bool`, `CommentCount int`, `HasAttachment bool`

- [ ] **Step 1: Update `dto/task.go`**

Add `PatchTaskRequest` and `CreateAdhocTaskRequest`:

```go
type PatchTaskRequest struct {
	Priority               *string `json:"priority"`
	ExpectedCompletionDate *string `json:"expected_completion_date"`
	ExpectedCompletionTime *string `json:"expected_completion_time"`
}

type CreateAdhocTaskRequest struct {
	StateID                string  `json:"state_id"`
	Label                  string  `json:"label"`
	TaskType               string  `json:"task_type"`
	CreatedBy              string  `json:"created_by"`
	Priority               string  `json:"priority"`
	ExpectedCompletionDate *string `json:"expected_completion_date"`
	ExpectedCompletionTime *string `json:"expected_completion_time"`
}
```

Extend `TaskInstanceResponse`:

```go
type TaskInstanceResponse struct {
	ID                     string     `json:"id"`
	WorkflowInstanceID     string     `json:"workflow_instance_id"`
	StateID                string     `json:"state_id"`
	TaskKey                string     `json:"task_key"`
	TaskLabel              string     `json:"task_label"`
	TaskType               string     `json:"task_type"`
	Status                 string     `json:"status"`
	Outcome                *string    `json:"outcome,omitempty"`
	CreatedBy              string     `json:"created_by"`
	CreatedAt              time.Time  `json:"created_at"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	// Phase 7
	Priority               string     `json:"priority"`
	ExpectedCompletionDate *string    `json:"expected_completion_date,omitempty"`
	ExpectedCompletionTime *string    `json:"expected_completion_time,omitempty"`
	IsAdhoc                bool       `json:"is_adhoc"`
	CommentCount           int        `json:"comment_count"`
	HasAttachment          bool       `json:"has_attachment"`
}
```

- [ ] **Step 2: Create `dto/comment.go`**

```go
package dto

type AddCommentRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type CommentResponse struct {
	ID             string `json:"id"`
	TaskInstanceID string `json:"task_instance_id"`
	Author         string `json:"author"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}
```

- [ ] **Step 3: Create `dto/attachment.go`**

```go
package dto

type UploadAttachmentRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64-encoded bytes
}

type AttachmentMetaResponse struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
	UploadedAt  string `json:"uploaded_at"`
	UpdatedAt   string `json:"updated_at"`
}
```

- [ ] **Step 4: Update `toTaskResponse` in handler/task.go**

In `backend/internal/api/handler/task.go`, update `toTaskResponse`:

```go
func toTaskResponse(t *model.TaskInstance) *dto.TaskInstanceResponse {
	return &dto.TaskInstanceResponse{
		ID:                     t.ID.String(),
		WorkflowInstanceID:     t.WorkflowInstanceID.String(),
		StateID:                t.StateID,
		TaskKey:                t.TaskKey,
		TaskLabel:              t.TaskLabel,
		TaskType:               t.TaskType,
		Status:                 string(t.Status),
		Outcome:                t.Outcome,
		CreatedBy:              t.CreatedBy,
		CreatedAt:              t.CreatedAt,
		CompletedAt:            t.CompletedAt,
		Priority:               t.Priority,
		ExpectedCompletionDate: t.ExpectedCompletionDate,
		ExpectedCompletionTime: t.ExpectedCompletionTime,
		IsAdhoc:                t.IsAdhoc,
	}
}
```

- [ ] **Step 5: Build**

```bash
cd backend && go build ./... 2>&1
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/api/dto/ backend/internal/api/handler/task.go
git commit -m "feat: DTOs for task extensions (PATCH, ad-hoc, comments, attachments)"
```

---

### Task 11: API Handlers + Routes

**Files:**
- Modify: `backend/internal/api/handler/task.go` (add `Patch` method)
- Create: `backend/internal/api/handler/task_extensions.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: `TaskService.UpdateFields`, `TaskService.CreateAdhoc`, `CommentService.Add`, `CommentService.List`, `AttachmentService.Upload/Download/Delete` (Tasks 8, 9)
- Produces: 7 new HTTP routes all under `/api/v1`

- [ ] **Step 1: Add `Patch` to `handler/task.go`**

Add `TaskExtHandler` struct and `Patch` handler:

```go
type TaskExtHandler struct {
	taskSvc   *service.TaskService
	commentSvc *service.CommentService
	attSvc    *service.AttachmentService
}

func NewTaskExtHandler(taskSvc *service.TaskService, commentSvc *service.CommentService, attSvc *service.AttachmentService) *TaskExtHandler {
	return &TaskExtHandler{taskSvc: taskSvc, commentSvc: commentSvc, attSvc: attSvc}
}
```

- [ ] **Step 2: Create `task_extensions.go`**

Create `backend/internal/api/handler/task_extensions.go`:

```go
package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/workflow-engine/backend/internal/api/dto"
	"github.com/workflow-engine/backend/internal/service"
)

func (h *TaskExtHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.PatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	err = h.taskSvc.UpdateFields(r.Context(), id, service.TaskFieldPatch{
		Priority:               req.Priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrTaskCompleted):
			writeError(w, http.StatusConflict, "TASK_COMPLETED", "task is terminal, fields locked", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskExtHandler) CreateAdhoc(w http.ResponseWriter, r *http.Request) {
	instID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.CreateAdhocTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.StateID == "" || req.Label == "" || req.TaskType == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "state_id, label, task_type are required", nil)
		return
	}

	task, err := h.taskSvc.CreateAdhoc(r.Context(), instID, service.AdhocTaskRequest{
		StateID:                req.StateID,
		Label:                  req.Label,
		TaskType:               req.TaskType,
		CreatedBy:              req.CreatedBy,
		Priority:               req.Priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrInstanceNotActive):
			writeError(w, http.StatusConflict, "INSTANCE_TERMINAL", "instance is not active", nil)
		case errors.Is(err, service.ErrInvalidState):
			writeError(w, http.StatusUnprocessableEntity, "INVALID_STATE", "state_id not in workflow definition", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	writeJSON(w, http.StatusCreated, toTaskResponse(task))
}

func (h *TaskExtHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.Author == "" || req.Content == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "author and content are required", nil)
		return
	}
	if len(req.Content) > 2048 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "content exceeds 2048 characters", nil)
		return
	}

	comment, err := h.commentSvc.Add(r.Context(), taskID, req.Author, req.Content)
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusCreated, dto.CommentResponse{
		ID:             comment.ID.String(),
		TaskInstanceID: comment.TaskInstanceID.String(),
		Author:         comment.Author,
		Content:        comment.Content,
		CreatedAt:      comment.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *TaskExtHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	comments, err := h.commentSvc.List(r.Context(), taskID)
	if err != nil {
		writeInternalError(w)
		return
	}
	out := make([]dto.CommentResponse, 0, len(comments))
	for _, c := range comments {
		out = append(out, dto.CommentResponse{
			ID:             c.ID.String(),
			TaskInstanceID: c.TaskInstanceID.String(),
			Author:         c.Author,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *TaskExtHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.UploadAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.Filename == "" || req.ContentType == "" || req.Data == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "filename, content_type, data are required", nil)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "data is not valid base64", nil)
		return
	}

	att, err := h.attSvc.Upload(r.Context(), taskID, req.Filename, req.ContentType, data)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file upload is not enabled", nil)
		case errors.Is(err, service.ErrFileTooLarge):
			writeError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds 2MB limit", nil)
		default:
			writeInternalError(w)
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.AttachmentMetaResponse{
		ID:          att.ID.String(),
		Filename:    att.Filename,
		ContentType: att.ContentType,
		SizeBytes:   att.SizeBytes,
		UploadedAt:  att.UploadedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   att.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *TaskExtHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}

	rc, att, err := h.attSvc.Download(r.Context(), taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file download is not enabled", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", att.SizeBytes))
	http.ServeContent(w, r, att.Filename, att.UploadedAt, nil)
	// stream body
	if _, err := copyTo(w, rc); err != nil {
		// headers already sent, can't write error
		return
	}
}

func (h *TaskExtHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	err = h.attSvc.Delete(r.Context(), taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file storage is not enabled", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

Add `copyTo` helper to `response.go` (or inline in the handler file):

```go
import "io"
func copyTo(w http.ResponseWriter, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}
```

- [ ] **Step 3: Register routes in router.go**

In `backend/internal/api/router.go`, update signature and add new routes:

```go
func NewRouter(
	db *sql.DB,
	defSvc *service.WorkflowDefService,
	instSvc *service.InstanceService,
	taskSvc *service.TaskService,
	commentSvc *service.CommentService,
	attSvc *service.AttachmentService,
	log *logrus.Logger,
	corsOrigins string,
) http.Handler {
	// ...existing code...
	taskExtHandler := handler.NewTaskExtHandler(taskSvc, commentSvc, attSvc)

	// existing task routes...
	api.HandleFunc("/task-instances/{id}", taskExtHandler.Patch).Methods(http.MethodPatch)
	api.HandleFunc("/task-instances/{id}/comments", taskExtHandler.ListComments).Methods(http.MethodGet)
	api.HandleFunc("/task-instances/{id}/comments", taskExtHandler.AddComment).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.UploadAttachment).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.DownloadAttachment).Methods(http.MethodGet)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.DeleteAttachment).Methods(http.MethodDelete)
	api.HandleFunc("/workflow-instances/{id}/tasks", taskExtHandler.CreateAdhoc).Methods(http.MethodPost)
	// ...rest of routes
}
```

- [ ] **Step 4: Wire new services in main.go**

In `backend/cmd/server/main.go`, after existing service construction:

```go
	commentRepo := repository.NewTaskCommentRepo(db)
	attRepo := repository.NewTaskAttachmentRepo(db)
	commentSvc := service.NewCommentService(commentRepo, log)
	attSvc := service.NewAttachmentService(fileStore, attRepo, log)
	taskSvc.SetInstRepo(instRepo)
	taskSvc.SetDefRepo(defRepo)

	handler := api.NewRouter(db, defSvc, instSvc, taskSvc, commentSvc, attSvc, log, cfg.CORSAllowedOrigins)
```

- [ ] **Step 5: Build**

```bash
cd backend && go build ./... 2>&1
```

Expected: clean build.

- [ ] **Step 6: Run all tests**

```bash
cd backend && make test
```

Expected: pass.

- [ ] **Step 7: Smoke test endpoints manually**

```bash
# Start stack
make docker-up && make run &
sleep 2

# Patch task (needs an existing task ID from DB)
TASK_ID=$(psql postgres://workflow:workflow@localhost:5432/workflow -t -c "SELECT id FROM task_instance LIMIT 1" | tr -d ' ')
curl -s -X PATCH http://localhost:8080/api/v1/task-instances/$TASK_ID \
  -H 'Content-Type: application/json' \
  -d '{"priority":"high"}' | jq .

# Add comment
curl -s -X POST http://localhost:8080/api/v1/task-instances/$TASK_ID/comments \
  -H 'Content-Type: application/json' \
  -d '{"author":"alice","content":"looks good"}' | jq .

# List comments
curl -s http://localhost:8080/api/v1/task-instances/$TASK_ID/comments | jq .
```

- [ ] **Step 8: Commit**

```bash
git add backend/internal/api/handler/task_extensions.go \
        backend/internal/api/handler/task.go \
        backend/internal/api/dto/ \
        backend/internal/api/router.go \
        backend/cmd/server/main.go
git commit -m "feat: 7 new API routes for task extensions (PATCH, ad-hoc, comments, attachment)"
```
