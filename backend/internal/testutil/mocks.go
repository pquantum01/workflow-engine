package testutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

// MockWorkflowDefRepo is a hand-written mock for WorkflowDefRepo.
type MockWorkflowDefRepo struct {
	CreateFn           func(ctx context.Context, def *model.WorkflowDef) error
	GetByIDFn          func(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error)
	GetByNameVersionFn func(ctx context.Context, name string, version int) (*model.WorkflowDef, error)
	ListFn             func(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error)
	SoftDeleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *MockWorkflowDefRepo) Create(ctx context.Context, def *model.WorkflowDef) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, def)
	}
	return nil
}
func (m *MockWorkflowDefRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockWorkflowDefRepo) GetByNameVersion(ctx context.Context, name string, version int) (*model.WorkflowDef, error) {
	if m.GetByNameVersionFn != nil {
		return m.GetByNameVersionFn(ctx, name, version)
	}
	return nil, repository.ErrNotFound
}
func (m *MockWorkflowDefRepo) List(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, includeDeleted)
	}
	return nil, nil
}
func (m *MockWorkflowDefRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil {
		return m.SoftDeleteFn(ctx, id)
	}
	return nil
}
func (m *MockWorkflowDefRepo) WithTx(tx repository.DB) repository.WorkflowDefRepo { return m }

// MockWorkflowInstanceRepo is a hand-written mock for WorkflowInstanceRepo.
type MockWorkflowInstanceRepo struct {
	CreateFn           func(ctx context.Context, inst *model.WorkflowInstance) error
	GetByIDFn          func(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
	GetByIDForUpdateFn func(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
	ListFn             func(ctx context.Context, filter repository.InstanceFilter) ([]*model.WorkflowInstance, int, error)
	UpdateStateFn      func(ctx context.Context, id uuid.UUID, newStateID string) error
	UpdateStatusFn     func(ctx context.Context, id uuid.UUID, status model.InstanceStatus, completedAt *time.Time) error
	SoftDeleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *MockWorkflowInstanceRepo) Create(ctx context.Context, inst *model.WorkflowInstance) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, inst)
	}
	return nil
}
func (m *MockWorkflowInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockWorkflowInstanceRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	if m.GetByIDForUpdateFn != nil {
		return m.GetByIDForUpdateFn(ctx, id)
	}
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockWorkflowInstanceRepo) List(ctx context.Context, filter repository.InstanceFilter) ([]*model.WorkflowInstance, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}
func (m *MockWorkflowInstanceRepo) UpdateState(ctx context.Context, id uuid.UUID, newStateID string) error {
	if m.UpdateStateFn != nil {
		return m.UpdateStateFn(ctx, id, newStateID)
	}
	return nil
}
func (m *MockWorkflowInstanceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.InstanceStatus, completedAt *time.Time) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status, completedAt)
	}
	return nil
}
func (m *MockWorkflowInstanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil {
		return m.SoftDeleteFn(ctx, id)
	}
	return nil
}
func (m *MockWorkflowInstanceRepo) WithTx(tx repository.DB) repository.WorkflowInstanceRepo { return m }

// MockTaskInstanceRepo is a hand-written mock for TaskInstanceRepo.
type MockTaskInstanceRepo struct {
	BulkCreateFn              func(ctx context.Context, tasks []*model.TaskInstance) error
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
	ListByInstanceFn          func(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error)
	ListByInstanceAndStateFn  func(ctx context.Context, instanceID uuid.UUID, stateID string) ([]*model.TaskInstance, error)
	UpdateStatusFn            func(ctx context.Context, id uuid.UUID, status model.TaskStatus, completedAt *time.Time) error
	UpdateOutcomeFn           func(ctx context.Context, id uuid.UUID, outcome string, data []byte) error
	SoftDeleteFn              func(ctx context.Context, id uuid.UUID) error
}

func (m *MockTaskInstanceRepo) BulkCreate(ctx context.Context, tasks []*model.TaskInstance) error {
	if m.BulkCreateFn != nil {
		return m.BulkCreateFn(ctx, tasks)
	}
	return nil
}
func (m *MockTaskInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockTaskInstanceRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error) {
	if m.ListByInstanceFn != nil {
		return m.ListByInstanceFn(ctx, instanceID, statusFilter)
	}
	return nil, nil
}
func (m *MockTaskInstanceRepo) ListByInstanceAndState(ctx context.Context, instanceID uuid.UUID, stateID string) ([]*model.TaskInstance, error) {
	if m.ListByInstanceAndStateFn != nil {
		return m.ListByInstanceAndStateFn(ctx, instanceID, stateID)
	}
	return nil, nil
}
func (m *MockTaskInstanceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, completedAt *time.Time) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status, completedAt)
	}
	return nil
}
func (m *MockTaskInstanceRepo) UpdateOutcome(ctx context.Context, id uuid.UUID, outcome string, data []byte) error {
	if m.UpdateOutcomeFn != nil {
		return m.UpdateOutcomeFn(ctx, id, outcome, data)
	}
	return nil
}
func (m *MockTaskInstanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil {
		return m.SoftDeleteFn(ctx, id)
	}
	return nil
}
func (m *MockTaskInstanceRepo) WithTx(tx repository.DB) repository.TaskInstanceRepo { return m }

// MockStateTransitionRepo is a hand-written mock for StateTransitionRepo.
type MockStateTransitionRepo struct {
	AppendFn          func(ctx context.Context, t *model.StateTransition) error
	ListByInstanceFn  func(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error)
}

func (m *MockStateTransitionRepo) Append(ctx context.Context, t *model.StateTransition) error {
	if m.AppendFn != nil {
		return m.AppendFn(ctx, t)
	}
	return nil
}
func (m *MockStateTransitionRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error) {
	if m.ListByInstanceFn != nil {
		return m.ListByInstanceFn(ctx, instanceID)
	}
	return nil, nil
}
func (m *MockStateTransitionRepo) WithTx(tx repository.DB) repository.StateTransitionRepo { return m }

// MockWebhookOutboxRepo is a hand-written mock for WebhookOutboxRepo.
type MockWebhookOutboxRepo struct {
	BulkInsertFn       func(ctx context.Context, rows []*model.WebhookOutboxRow) error
	ListPendingFn      func(ctx context.Context, now time.Time, limit int) ([]*model.WebhookOutboxRow, error)
	MarkDeliveredFn    func(ctx context.Context, id uuid.UUID) error
	MarkFailedFn       func(ctx context.Context, id uuid.UUID, lastError string) error
	IncrementAttemptFn func(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time) error
}

func (m *MockWebhookOutboxRepo) BulkInsert(ctx context.Context, rows []*model.WebhookOutboxRow) error {
	if m.BulkInsertFn != nil {
		return m.BulkInsertFn(ctx, rows)
	}
	return nil
}
func (m *MockWebhookOutboxRepo) ListPending(ctx context.Context, now time.Time, limit int) ([]*model.WebhookOutboxRow, error) {
	if m.ListPendingFn != nil {
		return m.ListPendingFn(ctx, now, limit)
	}
	return nil, nil
}
func (m *MockWebhookOutboxRepo) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	if m.MarkDeliveredFn != nil {
		return m.MarkDeliveredFn(ctx, id)
	}
	return nil
}
func (m *MockWebhookOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	if m.MarkFailedFn != nil {
		return m.MarkFailedFn(ctx, id, lastError)
	}
	return nil
}
func (m *MockWebhookOutboxRepo) IncrementAttempt(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time) error {
	if m.IncrementAttemptFn != nil {
		return m.IncrementAttemptFn(ctx, id, nextAttemptAt)
	}
	return nil
}
func (m *MockWebhookOutboxRepo) WithTx(tx repository.DB) repository.WebhookOutboxRepo { return m }
