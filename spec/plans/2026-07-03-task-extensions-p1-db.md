# Task Extensions — Plan 1: Migration + Model + Repository

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `priority`, `expected_completion_date`, `expected_completion_time`, `is_adhoc` to `task_instance`; create `task_comment` and `task_attachment` tables; update Go model and repo layer to match.

**Architecture:** One migration file adds columns + two new tables. Go model `TaskInstance` gains new fields with nullable wrappers. New repo interfaces (`TaskCommentRepo`, `TaskAttachmentRepo`) added to `interfaces.go` with Postgres implementations. Mock stubs added to `testutil/mocks.go`.

**Tech Stack:** Go 1.22, `database/sql` + `lib/pq`, `golang-migrate`, `testify`, UUID v4.

## Global Constraints

- No ORM — raw SQL only.
- All IDs: `github.com/google/uuid` UUIDs.
- Logger: `github.com/sirupsen/logrus` — no `fmt.Println`.
- Soft delete: `deleted_at TIMESTAMPTZ`. `task_comment` has **no** `deleted_at` (immutable by design).
- `task_attachment` has exactly one row per `task_instance_id` (UNIQUE constraint).
- Run `make test` and `make test-integration` green before committing.

---

### Task 1: Migration 004

**Files:**
- Create: `backend/migrations/004_task_extensions.up.sql`
- Create: `backend/migrations/004_task_extensions.down.sql`

**Interfaces:**
- Produces: DB schema with `priority`, `expected_completion_date`, `expected_completion_time`, `is_adhoc` on `task_instance`; `task_comment` table; `task_attachment` table.

- [ ] **Step 1: Write up migration**

```sql
-- backend/migrations/004_task_extensions.up.sql
ALTER TABLE task_instance
  ADD COLUMN priority                TEXT NOT NULL DEFAULT 'medium'
                                     CHECK (priority IN ('low','medium','high')),
  ADD COLUMN expected_completion_date DATE,
  ADD COLUMN expected_completion_time TIME,
  ADD COLUMN is_adhoc                BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE task_comment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL REFERENCES task_instance(id),
  author           TEXT NOT NULL,
  content          VARCHAR(2048) NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX task_comment_task_idx ON task_comment (task_instance_id, created_at);

CREATE TABLE task_attachment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL UNIQUE,
  filename         TEXT NOT NULL,
  content_type     TEXT NOT NULL,
  size_bytes       INT  NOT NULL,
  s3_url           TEXT NOT NULL,
  uploaded_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 2: Write down migration**

```sql
-- backend/migrations/004_task_extensions.down.sql
DROP TABLE IF EXISTS task_attachment;
DROP TABLE IF EXISTS task_comment;
ALTER TABLE task_instance
  DROP COLUMN IF EXISTS is_adhoc,
  DROP COLUMN IF EXISTS expected_completion_time,
  DROP COLUMN IF EXISTS expected_completion_date,
  DROP COLUMN IF EXISTS priority;
```

- [ ] **Step 3: Apply migration**

```bash
make migrate-up
```

Expected output: `migrate: applying migration 4/u task_extensions`

- [ ] **Step 4: Verify schema**

```bash
make docker-up
psql postgres://workflow:workflow@localhost:5432/workflow -c "\d task_instance" | grep -E "priority|expected|is_adhoc"
psql postgres://workflow:workflow@localhost:5432/workflow -c "\dt task_comment" 
psql postgres://workflow:workflow@localhost:5432/workflow -c "\dt task_attachment"
```

Expected: columns present, both tables exist.

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/004_task_extensions.up.sql backend/migrations/004_task_extensions.down.sql
git commit -m "feat: migration 004 — task_instance extensions + comment/attachment tables"
```

---

### Task 2: Update Go Model

**Files:**
- Modify: `backend/internal/model/task_instance.go`

**Interfaces:**
- Consumes: migration 004 schema (Task 1)
- Produces: `model.TaskInstance` with `Priority string`, `ExpectedCompletionDate *string`, `ExpectedCompletionTime *string`, `IsAdhoc bool`; `model.TaskComment` struct; `model.TaskAttachment` struct.

- [ ] **Step 1: Write failing test for new fields**

Create `backend/internal/model/task_instance_extensions_test.go`:

```go
package model_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/workflow-engine/backend/internal/model"
)

func TestTaskInstance_DefaultPriority(t *testing.T) {
	task := model.TaskInstance{Priority: "medium"}
	assert.Equal(t, "medium", task.Priority)
	assert.False(t, task.IsAdhoc)
}

func TestTaskComment_Fields(t *testing.T) {
	c := model.TaskComment{Author: "alice", Content: "looks good"}
	assert.Equal(t, "alice", c.Author)
	assert.Equal(t, "looks good", c.Content)
}

func TestTaskAttachment_Fields(t *testing.T) {
	a := model.TaskAttachment{Filename: "doc.pdf", SizeBytes: 1024}
	assert.Equal(t, "doc.pdf", a.Filename)
	assert.Equal(t, 1024, a.SizeBytes)
}
```

- [ ] **Step 2: Run test — expect compile fail**

```bash
cd backend && go test ./internal/model/... 2>&1 | head -20
```

Expected: `model.TaskComment undefined` or similar compile error.

- [ ] **Step 3: Update `task_instance.go` and add new model types**

Replace `backend/internal/model/task_instance.go` with:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusNew        TaskStatus = "new"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusRejected   TaskStatus = "rejected"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

var ValidTaskStatuses = map[TaskStatus]bool{
	TaskStatusNew:        true,
	TaskStatusInProgress: true,
	TaskStatusDone:       true,
	TaskStatusRejected:   true,
	TaskStatusCancelled:  true,
}

type TaskInstance struct {
	ID                     uuid.UUID
	WorkflowInstanceID     uuid.UUID
	StateID                string
	TaskKey                string
	TaskLabel              string
	TaskType               string
	Status                 TaskStatus
	Outcome                *string
	Data                   []byte
	CreatedBy              string
	CreatedAt              time.Time
	CompletedAt            *time.Time
	DeletedAt              *time.Time
	// Phase 7 extensions
	Priority               string
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
	IsAdhoc                bool
}

func (t *TaskInstance) IsCompletable() bool {
	return t.Status == TaskStatusNew || t.Status == TaskStatusInProgress
}

func (t *TaskInstance) IsTerminal() bool {
	return t.Status == TaskStatusDone || t.Status == TaskStatusRejected || t.Status == TaskStatusCancelled
}

type TaskComment struct {
	ID             uuid.UUID
	TaskInstanceID uuid.UUID
	Author         string
	Content        string
	CreatedAt      time.Time
}

type TaskAttachment struct {
	ID             uuid.UUID
	TaskInstanceID uuid.UUID
	Filename       string
	ContentType    string
	SizeBytes      int
	S3URL          string
	UploadedAt     time.Time
	UpdatedAt      time.Time
}
```

- [ ] **Step 4: Run test — expect pass**

```bash
cd backend && go test ./internal/model/... -v
```

Expected: all tests pass including existing `workflow_def_test.go`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/task_instance.go backend/internal/model/task_instance_extensions_test.go
git commit -m "feat: extend TaskInstance model, add TaskComment and TaskAttachment types"
```

---

### Task 3: New Repository Interfaces

**Files:**
- Modify: `backend/internal/repository/interfaces.go`
- Modify: `backend/internal/repository/task_instance.go` (extend `TaskInstanceRepo` methods)

**Interfaces:**
- Consumes: `model.TaskInstance` with new fields (Task 2)
- Produces:
  - `TaskInstanceRepo` gains: `Create(ctx, task) error`, `UpdateFields(ctx, id, patch TaskInstancePatch) error`
  - New interface `TaskCommentRepo` with: `Add(ctx, comment) error`, `ListByTask(ctx, taskID) ([]*model.TaskComment, error)`, `WithTx(tx DB) TaskCommentRepo`
  - New interface `TaskAttachmentRepo` with: `Upsert(ctx, attachment) error`, `GetByTaskID(ctx, taskID) (*model.TaskAttachment, error)`, `DeleteByTaskID(ctx, taskID) error`, `WithTx(tx DB) TaskAttachmentRepo`

- [ ] **Step 1: Add `TaskInstancePatch` type to interfaces.go**

Add to `backend/internal/repository/interfaces.go` after the import block:

```go
// TaskInstancePatch carries only the fields that PATCH /task-instances/:id may update.
// nil pointer = field not sent in request (leave unchanged).
type TaskInstancePatch struct {
	Priority               *string
	ExpectedCompletionDate *string // "YYYY-MM-DD" or null
	ExpectedCompletionTime *string // "HH:MM" or null
}
```

- [ ] **Step 2: Extend TaskInstanceRepo interface**

In `interfaces.go`, replace the `TaskInstanceRepo` block:

```go
type TaskInstanceRepo interface {
	BulkCreate(ctx context.Context, tasks []*model.TaskInstance) error
	Create(ctx context.Context, task *model.TaskInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
	ListByInstance(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error)
	ListByInstanceAndState(ctx context.Context, instanceID uuid.UUID, stateID string) ([]*model.TaskInstance, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, completedAt *time.Time) error
	UpdateOutcome(ctx context.Context, id uuid.UUID, outcome string, data []byte) error
	UpdateFields(ctx context.Context, id uuid.UUID, patch TaskInstancePatch) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	WithTx(tx DB) TaskInstanceRepo
}

type TaskCommentRepo interface {
	Add(ctx context.Context, comment *model.TaskComment) error
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error)
	WithTx(tx DB) TaskCommentRepo
}

type TaskAttachmentRepo interface {
	Upsert(ctx context.Context, att *model.TaskAttachment) error
	GetByTaskID(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error)
	DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error
	WithTx(tx DB) TaskAttachmentRepo
}
```

- [ ] **Step 3: Run compile check**

```bash
cd backend && go build ./... 2>&1
```

Expected: errors for missing `Create` and `UpdateFields` on existing `taskInstanceRepo` implementation — that's expected; fix in next task.

---

### Task 4: Postgres Implementations

**Files:**
- Modify: `backend/internal/repository/task_instance.go`
- Create: `backend/internal/repository/task_comment.go`
- Create: `backend/internal/repository/task_attachment.go`
- Create: `backend/tests/integration/task_extensions_repo_test.go`

**Interfaces:**
- Consumes: interfaces from Task 3, schema from Task 1
- Produces: working Postgres implementations for all three repos

- [ ] **Step 1: Write integration test (failing)**

Create `backend/tests/integration/task_extensions_repo_test.go`:

```go
package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

func TestTaskInstanceRepo_CreateAndGetExtended(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, db)

	defID, instID := seedInstanceForTaskTest(t, ctx)

	date := "2026-07-10"
	timeStr := "14:00"
	task := &model.TaskInstance{
		ID:                     uuid.New(),
		WorkflowInstanceID:     instID,
		StateID:                "review",
		TaskKey:                "adhoc_task_a1b2",
		TaskLabel:              "Ad-hoc Review",
		TaskType:               "manual",
		Status:                 model.TaskStatusNew,
		CreatedBy:              "alice",
		CreatedAt:              time.Now(),
		Priority:               "high",
		ExpectedCompletionDate: &date,
		ExpectedCompletionTime: &timeStr,
		IsAdhoc:                true,
	}
	_ = defID

	taskRepo := repository.NewTaskInstanceRepo(db)
	require.NoError(t, taskRepo.Create(ctx, task))

	got, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "high", got.Priority)
	assert.True(t, got.IsAdhoc)
	require.NotNil(t, got.ExpectedCompletionDate)
	assert.Equal(t, "2026-07-10", *got.ExpectedCompletionDate)
	require.NotNil(t, got.ExpectedCompletionTime)
	assert.Equal(t, "14:00", *got.ExpectedCompletionTime)
}

func TestTaskInstanceRepo_UpdateFields(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, db)

	_, instID := seedInstanceForTaskTest(t, ctx)
	taskRepo := repository.NewTaskInstanceRepo(db)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: instID, StateID: "review",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now(),
		Priority: "medium",
	}}))

	newPriority := "low"
	newDate := "2026-08-01"
	require.NoError(t, taskRepo.UpdateFields(ctx, taskID, repository.TaskInstancePatch{
		Priority:               &newPriority,
		ExpectedCompletionDate: &newDate,
	}))

	got, err := taskRepo.GetByID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "low", got.Priority)
	require.NotNil(t, got.ExpectedCompletionDate)
	assert.Equal(t, "2026-08-01", *got.ExpectedCompletionDate)
}

func TestTaskCommentRepo_AddAndList(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, db)

	_, instID := seedInstanceForTaskTest(t, ctx)
	taskRepo := repository.NewTaskInstanceRepo(db)
	commentRepo := repository.NewTaskCommentRepo(db)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: instID, StateID: "review",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now(),
		Priority: "medium",
	}}))

	c1 := &model.TaskComment{ID: uuid.New(), TaskInstanceID: taskID, Author: "alice", Content: "first comment", CreatedAt: time.Now()}
	c2 := &model.TaskComment{ID: uuid.New(), TaskInstanceID: taskID, Author: "bob", Content: "second comment", CreatedAt: time.Now().Add(time.Second)}
	require.NoError(t, commentRepo.Add(ctx, c1))
	require.NoError(t, commentRepo.Add(ctx, c2))

	list, err := commentRepo.ListByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "alice", list[0].Author)
	assert.Equal(t, "bob", list[1].Author)
}

func TestTaskAttachmentRepo_UpsertAndGet(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, db)

	_, instID := seedInstanceForTaskTest(t, ctx)
	taskRepo := repository.NewTaskInstanceRepo(db)
	attRepo := repository.NewTaskAttachmentRepo(db)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: instID, StateID: "review",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now(),
		Priority: "medium",
	}}))

	att := &model.TaskAttachment{
		ID: uuid.New(), TaskInstanceID: taskID,
		Filename: "contract.pdf", ContentType: "application/pdf",
		SizeBytes: 1024, S3URL: "https://bucket.s3.amazonaws.com/attachments/contract.pdf",
		UploadedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, attRepo.Upsert(ctx, att))

	got, err := attRepo.GetByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "contract.pdf", got.Filename)
	assert.Equal(t, 1024, got.SizeBytes)

	// Replace with new file
	att2 := &model.TaskAttachment{
		ID: uuid.New(), TaskInstanceID: taskID,
		Filename: "revised.pdf", ContentType: "application/pdf",
		SizeBytes: 2048, S3URL: "https://bucket.s3.amazonaws.com/attachments/revised.pdf",
		UploadedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, attRepo.Upsert(ctx, att2))

	got2, err := attRepo.GetByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "revised.pdf", got2.Filename)

	require.NoError(t, attRepo.DeleteByTaskID(ctx, taskID))
	_, err = attRepo.GetByTaskID(ctx, taskID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
```

- [ ] **Step 2: Run integration test — expect compile or runtime fail**

```bash
cd backend && go test ./tests/integration/... -run "TestTask.*Repo" -v 2>&1 | head -30
```

Expected: compile errors — missing `Create`, `UpdateFields`, `NewTaskCommentRepo`, `NewTaskAttachmentRepo`.

- [ ] **Step 3: Extend `task_instance.go` — add `Create` and `UpdateFields`**

Add to `backend/internal/repository/task_instance.go`:

```go
func (r *taskInstanceRepo) Create(ctx context.Context, task *model.TaskInstance) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_instance
		  (id, workflow_instance_id, state_id, task_key, task_label, task_type,
		   status, created_by, created_at, priority, is_adhoc,
		   expected_completion_date, expected_completion_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		task.ID, task.WorkflowInstanceID, task.StateID, task.TaskKey, task.TaskLabel,
		task.TaskType, string(task.Status), task.CreatedBy, task.CreatedAt,
		task.Priority, task.IsAdhoc,
		task.ExpectedCompletionDate, task.ExpectedCompletionTime,
	)
	return err
}

func (r *taskInstanceRepo) UpdateFields(ctx context.Context, id uuid.UUID, patch TaskInstancePatch) error {
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if patch.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *patch.Priority)
		argIdx++
	}
	if patch.ExpectedCompletionDate != nil {
		if *patch.ExpectedCompletionDate == "" {
			setClauses = append(setClauses, fmt.Sprintf("expected_completion_date = $%d", argIdx))
			args = append(args, nil)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("expected_completion_date = $%d", argIdx))
			args = append(args, *patch.ExpectedCompletionDate)
		}
		argIdx++
	}
	if patch.ExpectedCompletionTime != nil {
		if *patch.ExpectedCompletionTime == "" {
			setClauses = append(setClauses, fmt.Sprintf("expected_completion_time = $%d", argIdx))
			args = append(args, nil)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("expected_completion_time = $%d", argIdx))
			args = append(args, *patch.ExpectedCompletionTime)
		}
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // nothing to update
	}

	args = append(args, id)
	q := fmt.Sprintf("UPDATE task_instance SET %s WHERE id = $%d AND deleted_at IS NULL",
		strings.Join(setClauses, ", "), argIdx)
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}
```

Also update `scanTask` and `scanTaskRows` to scan the new columns. Replace `scanTask`:

```go
func scanTask(row rowScanner) (*model.TaskInstance, error) {
	var t model.TaskInstance
	var status string
	var outcome sql.NullString
	var completedAt, deletedAt sql.NullTime
	var expDate, expTime sql.NullString
	err := row.Scan(
		&t.ID, &t.WorkflowInstanceID, &t.StateID, &t.TaskKey, &t.TaskLabel, &t.TaskType,
		&status, &outcome, &t.Data, &t.CreatedBy, &t.CreatedAt, &completedAt, &deletedAt,
		&t.Priority, &t.IsAdhoc, &expDate, &expTime,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t.Status = model.TaskStatus(status)
	if outcome.Valid {
		t.Outcome = &outcome.String
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if deletedAt.Valid {
		t.DeletedAt = &deletedAt.Time
	}
	if expDate.Valid {
		t.ExpectedCompletionDate = &expDate.String
	}
	if expTime.Valid {
		t.ExpectedCompletionTime = &expTime.String
	}
	return &t, nil
}
```

Update all SELECT statements in `task_instance.go` to include the new columns at the end:

```sql
-- add to every SELECT:
priority, is_adhoc, expected_completion_date::TEXT, expected_completion_time::TEXT
```

Update `scanTaskRows` similarly (same scan order as `scanTask`).

- [ ] **Step 4: Create `task_comment.go`**

Create `backend/internal/repository/task_comment.go`:

```go
package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type taskCommentRepo struct {
	db DB
}

func NewTaskCommentRepo(db DB) TaskCommentRepo {
	return &taskCommentRepo{db: db}
}

func (r *taskCommentRepo) WithTx(tx DB) TaskCommentRepo {
	return &taskCommentRepo{db: tx}
}

func (r *taskCommentRepo) Add(ctx context.Context, comment *model.TaskComment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO task_comment (id, task_instance_id, author, content, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		comment.ID, comment.TaskInstanceID, comment.Author, comment.Content, comment.CreatedAt,
	)
	return err
}

func (r *taskCommentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_instance_id, author, content, created_at
		 FROM task_comment WHERE task_instance_id = $1 ORDER BY created_at`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.TaskComment
	for rows.Next() {
		var c model.TaskComment
		if err := rows.Scan(&c.ID, &c.TaskInstanceID, &c.Author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}
```

- [ ] **Step 5: Create `task_attachment.go`**

Create `backend/internal/repository/task_attachment.go`:

```go
package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type taskAttachmentRepo struct {
	db DB
}

func NewTaskAttachmentRepo(db DB) TaskAttachmentRepo {
	return &taskAttachmentRepo{db: db}
}

func (r *taskAttachmentRepo) WithTx(tx DB) TaskAttachmentRepo {
	return &taskAttachmentRepo{db: tx}
}

func (r *taskAttachmentRepo) Upsert(ctx context.Context, att *model.TaskAttachment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_attachment
		  (id, task_instance_id, filename, content_type, size_bytes, s3_url, uploaded_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (task_instance_id) DO UPDATE SET
		  id          = EXCLUDED.id,
		  filename    = EXCLUDED.filename,
		  content_type = EXCLUDED.content_type,
		  size_bytes  = EXCLUDED.size_bytes,
		  s3_url      = EXCLUDED.s3_url,
		  updated_at  = EXCLUDED.updated_at`,
		att.ID, att.TaskInstanceID, att.Filename, att.ContentType,
		att.SizeBytes, att.S3URL, att.UploadedAt, att.UpdatedAt,
	)
	return err
}

func (r *taskAttachmentRepo) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, task_instance_id, filename, content_type, size_bytes, s3_url, uploaded_at, updated_at
		 FROM task_attachment WHERE task_instance_id = $1`,
		taskID,
	)
	var a model.TaskAttachment
	err := row.Scan(&a.ID, &a.TaskInstanceID, &a.Filename, &a.ContentType,
		&a.SizeBytes, &a.S3URL, &a.UploadedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *taskAttachmentRepo) DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM task_attachment WHERE task_instance_id = $1`, taskID)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}
```

- [ ] **Step 6: Add mock stubs to testutil/mocks.go**

Append to `backend/internal/testutil/mocks.go`:

```go
// MockTaskCommentRepo
type MockTaskCommentRepo struct {
	AddFn        func(ctx context.Context, comment *model.TaskComment) error
	ListByTaskFn func(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error)
}

func (m *MockTaskCommentRepo) Add(ctx context.Context, comment *model.TaskComment) error {
	if m.AddFn != nil {
		return m.AddFn(ctx, comment)
	}
	return nil
}
func (m *MockTaskCommentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error) {
	if m.ListByTaskFn != nil {
		return m.ListByTaskFn(ctx, taskID)
	}
	return nil, nil
}
func (m *MockTaskCommentRepo) WithTx(tx repository.DB) repository.TaskCommentRepo { return m }

// MockTaskAttachmentRepo
type MockTaskAttachmentRepo struct {
	UpsertFn        func(ctx context.Context, att *model.TaskAttachment) error
	GetByTaskIDFn   func(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error)
	DeleteByTaskIDFn func(ctx context.Context, taskID uuid.UUID) error
}

func (m *MockTaskAttachmentRepo) Upsert(ctx context.Context, att *model.TaskAttachment) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, att)
	}
	return nil
}
func (m *MockTaskAttachmentRepo) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error) {
	if m.GetByTaskIDFn != nil {
		return m.GetByTaskIDFn(ctx, taskID)
	}
	return nil, repository.ErrNotFound
}
func (m *MockTaskAttachmentRepo) DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error {
	if m.DeleteByTaskIDFn != nil {
		return m.DeleteByTaskIDFn(ctx, taskID)
	}
	return nil
}
func (m *MockTaskAttachmentRepo) WithTx(tx repository.DB) repository.TaskAttachmentRepo { return m }
```

Also extend `MockTaskInstanceRepo` with the new methods:

```go
// Add to MockTaskInstanceRepo struct:
CreateFn      func(ctx context.Context, task *model.TaskInstance) error
UpdateFieldsFn func(ctx context.Context, id uuid.UUID, patch repository.TaskInstancePatch) error

// Add method implementations:
func (m *MockTaskInstanceRepo) Create(ctx context.Context, task *model.TaskInstance) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, task)
	}
	return nil
}
func (m *MockTaskInstanceRepo) UpdateFields(ctx context.Context, id uuid.UUID, patch repository.TaskInstancePatch) error {
	if m.UpdateFieldsFn != nil {
		return m.UpdateFieldsFn(ctx, id, patch)
	}
	return nil
}
```

- [ ] **Step 7: Run integration tests**

```bash
cd backend && go test ./tests/integration/... -run "TestTask.*Repo" -v
```

Expected: all 4 test functions pass.

- [ ] **Step 8: Run full test suite**

```bash
cd backend && make test && make test-integration
```

Expected: both pass.

- [ ] **Step 9: Commit**

```bash
git add backend/internal/repository/task_instance.go \
        backend/internal/repository/task_comment.go \
        backend/internal/repository/task_attachment.go \
        backend/internal/repository/interfaces.go \
        backend/internal/testutil/mocks.go \
        backend/tests/integration/task_extensions_repo_test.go
git commit -m "feat: repo layer for task extensions — Create/UpdateFields on task, comment repo, attachment repo"
```
