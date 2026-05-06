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

type CreateTaskInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TaskType    string `json:"task_type"`
}

type UpdateTaskInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	TaskType    *string `json:"task_type,omitempty"`
}

type TaskService struct {
	repo         repository.TaskRepository
	workflowRepo repository.WorkflowRepository
	processRepo  repository.ProcessRepository
	logger       *slog.Logger
}

func NewTaskService(repo repository.TaskRepository, workflowRepo repository.WorkflowRepository, processRepo repository.ProcessRepository, logger *slog.Logger) *TaskService {
	return &TaskService{repo: repo, workflowRepo: workflowRepo, processRepo: processRepo, logger: logger}
}

func (s *TaskService) Create(ctx context.Context, workflowID uuid.UUID, input CreateTaskInput) (*model.Task, error) {
	s.logger.Debug("creating task", slog.String("name", input.Name), slog.String("workflow_id", workflowID.String()))

	wf, err := s.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		s.logger.Error("failed to create task", slog.String("workflow_id", workflowID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("workflow: %w", err)
	}
	p, err := s.processRepo.GetByID(ctx, wf.ProcessID)
	if err != nil {
		s.logger.Error("failed to create task", slog.String("workflow_id", workflowID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("process: %w", err)
	}
	if p.Status == "archived" {
		s.logger.Error("failed to create task", slog.String("workflow_id", workflowID.String()), slog.String("error", "cannot add task to workflow in archived process"))
		return nil, fmt.Errorf("cannot add task to workflow in archived process")
	}

	now := time.Now().UTC()
	t := &model.Task{
		ID:          uuid.New(),
		WorkflowID:  workflowID,
		Name:        input.Name,
		Description: input.Description,
		TaskType:    input.TaskType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := t.Validate(); err != nil {
		s.logger.Error("failed to create task", slog.String("workflow_id", workflowID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Create(ctx, t); err != nil {
		s.logger.Error("failed to create task", slog.String("workflow_id", workflowID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task created", slog.String("id", t.ID.String()), slog.String("name", t.Name), slog.String("workflow_id", workflowID.String()))
	return t, nil
}

func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	s.logger.Debug("getting task", slog.String("id", id.String()))

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task retrieved", slog.String("id", id.String()))
	return t, nil
}

func (s *TaskService) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*model.Task, error) {
	s.logger.Debug("listing tasks", slog.String("workflow_id", workflowID.String()))

	results, err := s.repo.ListByWorkflowID(ctx, workflowID)
	if err != nil {
		s.logger.Error("failed to list tasks", slog.String("workflow_id", workflowID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("tasks listed", slog.String("workflow_id", workflowID.String()), slog.Int("count", len(results)))
	return results, nil
}

func (s *TaskService) Update(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*model.Task, error) {
	s.logger.Debug("updating task", slog.String("id", id.String()))

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to update task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	if input.Name != nil {
		t.Name = *input.Name
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	if input.TaskType != nil {
		t.TaskType = *input.TaskType
	}
	t.UpdatedAt = time.Now().UTC()

	if err := t.Validate(); err != nil {
		s.logger.Error("failed to update task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Update(ctx, t); err != nil {
		s.logger.Error("failed to update task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("task updated", slog.String("id", id.String()), slog.String("name", t.Name))
	return t, nil
}

func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting task", slog.String("id", id.String()))

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}
	wf, err := s.workflowRepo.GetByID(ctx, t.WorkflowID)
	if err != nil {
		s.logger.Error("failed to delete task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return fmt.Errorf("workflow: %w", err)
	}
	p, err := s.processRepo.GetByID(ctx, wf.ProcessID)
	if err != nil {
		s.logger.Error("failed to delete task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return fmt.Errorf("process: %w", err)
	}
	if p.Status != "draft" {
		s.logger.Error("failed to delete task", slog.String("id", id.String()), slog.String("error", "can only delete tasks from draft processes"), slog.String("process_status", p.Status))
		return fmt.Errorf("can only delete tasks from draft processes (process status: %s)", p.Status)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete task", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("task deleted", slog.String("id", id.String()))
	return nil
}
