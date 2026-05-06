package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

var validStatusTransitions = map[string]map[string]bool{
	"pending":     {"in_progress": true},
	"in_progress": {"completed": true, "failed": true, "skipped": true},
}

type UpdateTaskStatusInput struct {
	Status      string `json:"status"`
	TriggeredBy string `json:"triggered_by"`
	TriggerType string `json:"trigger_type"`
}

type TaskInstanceService struct {
	taskInstanceRepo     repository.TaskInstanceRepository
	workflowInstanceRepo repository.WorkflowInstanceRepository
	engine               engine.Engine
	logger               *slog.Logger
}

func NewTaskInstanceService(
	taskInstanceRepo repository.TaskInstanceRepository,
	workflowInstanceRepo repository.WorkflowInstanceRepository,
	eng engine.Engine,
	logger *slog.Logger,
) *TaskInstanceService {
	return &TaskInstanceService{
		taskInstanceRepo:     taskInstanceRepo,
		workflowInstanceRepo: workflowInstanceRepo,
		engine:               eng,
		logger:               logger,
	}
}

func (s *TaskInstanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	s.logger.Debug("getting task instance", slog.String("id", id.String()))

	ti, err := s.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get task instance", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task instance retrieved", slog.String("id", id.String()), slog.String("status", ti.Status))
	return ti, nil
}

func (s *TaskInstanceService) ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.TaskInstance, error) {
	s.logger.Debug("listing task instances", slog.String("workflow_instance_id", workflowInstanceID.String()))

	results, err := s.taskInstanceRepo.ListByWorkflowInstanceID(ctx, workflowInstanceID)
	if err != nil {
		s.logger.Error("failed to list task instances", slog.String("workflow_instance_id", workflowInstanceID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task instances listed", slog.String("workflow_instance_id", workflowInstanceID.String()), slog.Int("count", len(results)))
	return results, nil
}

func (s *TaskInstanceService) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateTaskStatusInput) (*model.TaskInstance, error) {
	s.logger.Debug("updating task instance status",
		slog.String("task_instance_id", id.String()),
		slog.String("new_status", input.Status),
	)

	ti, err := s.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to update task instance status", slog.String("task_instance_id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	// Check workflow instance is active
	wi, err := s.workflowInstanceRepo.GetByID(ctx, ti.WorkflowInstanceID)
	if err != nil {
		s.logger.Error("failed to update task instance status", slog.String("task_instance_id", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("workflow instance: %w", err)
	}
	if wi.Status != "active" {
		s.logger.Error("failed to update task instance status", slog.String("task_instance_id", id.String()), slog.String("error", "workflow instance not active"), slog.String("workflow_instance_status", wi.Status))
		return nil, fmt.Errorf("cannot update task: workflow instance is not active (status: %s)", wi.Status)
	}

	// Validate status transition
	allowed, ok := validStatusTransitions[ti.Status]
	if !ok || !allowed[input.Status] {
		s.logger.Error("failed to update task instance status", slog.String("task_instance_id", id.String()), slog.String("error", "invalid status transition"), slog.String("from_status", ti.Status), slog.String("to_status", input.Status))
		return nil, fmt.Errorf("invalid status transition: %s -> %s", ti.Status, input.Status)
	}

	ti.Status = input.Status
	ti.UpdatedAt = time.Now().UTC()
	if input.Status == "completed" {
		now := ti.UpdatedAt
		ti.CompletedAt = &now
	}

	if err := s.taskInstanceRepo.Update(ctx, ti); err != nil {
		s.logger.Error("failed to update task instance status", slog.String("task_instance_id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task instance status updated",
		slog.String("task_instance_id", id.String()),
		slog.String("status", ti.Status),
		slog.String("workflow_instance_id", ti.WorkflowInstanceID.String()),
	)

	// Auto-evaluate engine after task completion
	if input.Status == "completed" {
		if _, err := s.engine.Evaluate(ctx, ti.WorkflowInstanceID); err != nil {
			s.logger.Error("engine evaluation failed after task completion",
				slog.String("task_instance_id", id.String()),
				slog.String("workflow_instance_id", ti.WorkflowInstanceID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	return ti, nil
}

func (s *TaskInstanceService) Assign(ctx context.Context, id uuid.UUID, assignedTo string) (*model.TaskInstance, error) {
	s.logger.Debug("assigning task instance",
		slog.String("task_instance_id", id.String()),
		slog.String("assigned_to", assignedTo),
	)

	ti, err := s.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to assign task instance", slog.String("task_instance_id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	ti.AssignedTo = assignedTo
	ti.UpdatedAt = time.Now().UTC()

	if err := s.taskInstanceRepo.Update(ctx, ti); err != nil {
		s.logger.Error("failed to assign task instance", slog.String("task_instance_id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task instance assigned",
		slog.String("task_instance_id", id.String()),
		slog.String("assigned_to", assignedTo),
	)
	return ti, nil
}
