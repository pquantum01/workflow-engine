package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type StartProcessInput struct {
	ProcessID uuid.UUID  `json:"process_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
}

type ProcessInstanceService struct {
	processRepo          repository.ProcessRepository
	workflowRepo         repository.WorkflowRepository
	taskRepo             repository.TaskRepository
	processInstanceRepo  repository.ProcessInstanceRepository
	workflowInstanceRepo repository.WorkflowInstanceRepository
	taskInstanceRepo     repository.TaskInstanceRepository
	logger               *slog.Logger
}

func NewProcessInstanceService(
	processRepo repository.ProcessRepository,
	workflowRepo repository.WorkflowRepository,
	taskRepo repository.TaskRepository,
	processInstanceRepo repository.ProcessInstanceRepository,
	workflowInstanceRepo repository.WorkflowInstanceRepository,
	taskInstanceRepo repository.TaskInstanceRepository,
	logger *slog.Logger,
) *ProcessInstanceService {
	return &ProcessInstanceService{
		processRepo:          processRepo,
		workflowRepo:         workflowRepo,
		taskRepo:             taskRepo,
		processInstanceRepo:  processInstanceRepo,
		workflowInstanceRepo: workflowInstanceRepo,
		taskInstanceRepo:     taskInstanceRepo,
		logger:               logger,
	}
}

func (s *ProcessInstanceService) Start(ctx context.Context, input StartProcessInput) (*model.ProcessInstance, error) {
	s.logger.Debug("starting process instance", slog.String("process_id", input.ProcessID.String()))

	p, err := s.processRepo.GetByID(ctx, input.ProcessID)
	if err != nil {
		s.logger.Error("failed to start process instance", slog.String("process_id", input.ProcessID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("process: %w", err)
	}
	if p.Status != "active" {
		s.logger.Error("failed to start process instance", slog.String("process_id", input.ProcessID.String()), slog.String("error", "process is not active"), slog.String("status", p.Status))
		return nil, fmt.Errorf("can only start instances of active processes (status: %s)", p.Status)
	}

	// Parent-child 1-level enforcement
	if input.ParentID != nil {
		parent, err := s.processInstanceRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			s.logger.Error("failed to start process instance", slog.String("process_id", input.ProcessID.String()), slog.String("parent_id", input.ParentID.String()), slog.String("error", err.Error()))
			return nil, fmt.Errorf("parent process instance: %w", err)
		}
		if parent.ParentID != nil {
			s.logger.Error("failed to start process instance", slog.String("process_id", input.ProcessID.String()), slog.String("error", "cannot create grandchild"))
			return nil, fmt.Errorf("cannot create grandchild: parent already has a parent")
		}
		if parent.Status != "running" {
			s.logger.Error("failed to start process instance", slog.String("process_id", input.ProcessID.String()), slog.String("error", "parent not running"), slog.String("parent_status", parent.Status))
			return nil, fmt.Errorf("parent process instance is not running (status: %s)", parent.Status)
		}
	}

	now := time.Now().UTC()
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: input.ProcessID,
		ParentID:  input.ParentID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.processInstanceRepo.Create(ctx, pi); err != nil {
		s.logger.Error("failed to create process instance", slog.String("process_id", input.ProcessID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("create process instance: %w", err)
	}

	// Create workflow instances and task instances
	workflows, err := s.workflowRepo.ListByProcessID(ctx, input.ProcessID)
	if err != nil {
		s.logger.Error("failed to list workflows for process", slog.String("process_id", input.ProcessID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("list workflows: %w", err)
	}

	for _, wf := range workflows {
		wi := &model.WorkflowInstance{
			ID:                uuid.New(),
			WorkflowID:        wf.ID,
			ProcessInstanceID: pi.ID,
			CurrentState:      wf.StateMachine.InitialState,
			Status:            "active",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := s.workflowInstanceRepo.Create(ctx, wi); err != nil {
			s.logger.Error("failed to create workflow instance", slog.String("workflow_id", wf.ID.String()), slog.String("error", err.Error()))
			return nil, fmt.Errorf("create workflow instance: %w", err)
		}

		tasks, err := s.taskRepo.ListByWorkflowID(ctx, wf.ID)
		if err != nil {
			s.logger.Error("failed to list tasks for workflow", slog.String("workflow_id", wf.ID.String()), slog.String("error", err.Error()))
			return nil, fmt.Errorf("list tasks: %w", err)
		}
		for _, t := range tasks {
			ti := &model.TaskInstance{
				ID:                 uuid.New(),
				TaskID:             t.ID,
				WorkflowInstanceID: wi.ID,
				Status:             "pending",
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			if err := s.taskInstanceRepo.Create(ctx, ti); err != nil {
				s.logger.Error("failed to create task instance", slog.String("task_id", t.ID.String()), slog.String("error", err.Error()))
				return nil, fmt.Errorf("create task instance: %w", err)
			}
		}
	}

	s.logger.Info("process instance started",
		slog.String("process_instance_id", pi.ID.String()),
		slog.String("process_id", input.ProcessID.String()),
	)

	return pi, nil
}

func (s *ProcessInstanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error) {
	s.logger.Debug("getting process instance", slog.String("id", id.String()))

	pi, err := s.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get process instance", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("process instance retrieved", slog.String("id", id.String()), slog.String("status", pi.Status))
	return pi, nil
}

func (s *ProcessInstanceService) List(ctx context.Context, offset, limit int) ([]*model.ProcessInstance, error) {
	s.logger.Debug("listing process instances", slog.Int("offset", offset), slog.Int("limit", limit))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	results, err := s.processInstanceRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("failed to list process instances", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("process instances listed", slog.Int("count", len(results)))
	return results, nil
}

func (s *ProcessInstanceService) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error) {
	s.logger.Debug("listing child process instances", slog.String("parent_id", parentID.String()))

	results, err := s.processInstanceRepo.ListByParentID(ctx, parentID)
	if err != nil {
		s.logger.Error("failed to list child process instances", slog.String("parent_id", parentID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("child process instances listed", slog.String("parent_id", parentID.String()), slog.Int("count", len(results)))
	return results, nil
}

func (s *ProcessInstanceService) Cancel(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("cancelling process instance", slog.String("id", id.String()))

	pi, err := s.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to cancel process instance", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}
	if pi.Status != "running" {
		s.logger.Error("failed to cancel process instance", slog.String("id", id.String()), slog.String("error", "not running"), slog.String("status", pi.Status))
		return fmt.Errorf("can only cancel running process instances (status: %s)", pi.Status)
	}

	now := time.Now().UTC()
	pi.Status = "cancelled"
	pi.CompletedAt = &now
	pi.UpdatedAt = now
	if err := s.processInstanceRepo.Update(ctx, pi); err != nil {
		s.logger.Error("failed to update process instance", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}

	// Cancel all active workflow instances
	wis, err := s.workflowInstanceRepo.ListByProcessInstanceID(ctx, id)
	if err != nil {
		s.logger.Error("failed to list workflow instances for cancel", slog.String("process_instance_id", id.String()), slog.String("error", err.Error()))
		return fmt.Errorf("list workflow instances: %w", err)
	}
	for _, wi := range wis {
		if wi.Status == "active" {
			wi.Status = "cancelled"
			wi.UpdatedAt = now
			if err := s.workflowInstanceRepo.Update(ctx, wi); err != nil {
				s.logger.Error("failed to cancel workflow instance", slog.String("workflow_instance_id", wi.ID.String()), slog.String("error", err.Error()))
				return fmt.Errorf("cancel workflow instance: %w", err)
			}
		}
	}

	s.logger.Info("process instance cancelled", slog.String("process_instance_id", id.String()))
	return nil
}
