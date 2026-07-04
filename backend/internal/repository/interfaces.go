package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
)

// DB is satisfied by both *sql.DB and *sql.Tx.
type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// TaskInstancePatch carries only the fields that PATCH /task-instances/:id may update.
// nil pointer = field not sent in request (leave unchanged).
type TaskInstancePatch struct {
	Priority               *string
	ExpectedCompletionDate *string // "YYYY-MM-DD" or empty string to clear
	ExpectedCompletionTime *string // "HH:MM" or empty string to clear
}

type InstanceFilter struct {
	DefID  *uuid.UUID
	Status *model.InstanceStatus
	Limit  int
	Offset int
}

type WorkflowDefRepo interface {
	Create(ctx context.Context, def *model.WorkflowDef) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error)
	GetByNameVersion(ctx context.Context, name string, version int) (*model.WorkflowDef, error)
	List(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	WithTx(tx DB) WorkflowDefRepo
}

type WorkflowInstanceRepo interface {
	Create(ctx context.Context, inst *model.WorkflowInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
	List(ctx context.Context, filter InstanceFilter) ([]*model.WorkflowInstance, int, error)
	UpdateState(ctx context.Context, id uuid.UUID, newStateID string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.InstanceStatus, completedAt *time.Time) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	WithTx(tx DB) WorkflowInstanceRepo
}

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

type StateTransitionRepo interface {
	Append(ctx context.Context, t *model.StateTransition) error
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error)
	WithTx(tx DB) StateTransitionRepo
}

type WebhookOutboxRepo interface {
	BulkInsert(ctx context.Context, rows []*model.WebhookOutboxRow) error
	ListPending(ctx context.Context, now time.Time, limit int) ([]*model.WebhookOutboxRow, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error
	IncrementAttempt(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time) error
	WithTx(tx DB) WebhookOutboxRepo
}
