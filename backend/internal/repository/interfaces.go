package repository

import (
	"context"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
)

type ProcessRepository interface {
	Create(ctx context.Context, process *model.Process) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Process, error)
	List(ctx context.Context, offset, limit int) ([]*model.Process, error)
	Update(ctx context.Context, process *model.Process) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type WorkflowRepository interface {
	Create(ctx context.Context, workflow *model.Workflow) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error)
	ListByProcessID(ctx context.Context, processID uuid.UUID) ([]*model.Workflow, error)
	Update(ctx context.Context, workflow *model.Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProcessInstanceRepository interface {
	Create(ctx context.Context, instance *model.ProcessInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error)
	List(ctx context.Context, offset, limit int) ([]*model.ProcessInstance, error)
	ListByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error)
	Update(ctx context.Context, instance *model.ProcessInstance) error
}

type WorkflowInstanceRepository interface {
	Create(ctx context.Context, instance *model.WorkflowInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
	ListByProcessInstanceID(ctx context.Context, processInstanceID uuid.UUID) ([]*model.WorkflowInstance, error)
	Update(ctx context.Context, instance *model.WorkflowInstance) error
}

type TaskInstanceRepository interface {
	Create(ctx context.Context, instance *model.TaskInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
	ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.TaskInstance, error)
	Update(ctx context.Context, instance *model.TaskInstance) error
}

type StateTransitionRepository interface {
	Create(ctx context.Context, transition *model.StateTransition) error
	ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.StateTransition, error)
}
