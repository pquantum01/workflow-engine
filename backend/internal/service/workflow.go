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

type CreateWorkflowInput struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	StateMachine model.StateMachine `json:"state_machine"`
	SortOrder    int                `json:"sort_order"`
}

type UpdateWorkflowInput struct {
	Name         *string             `json:"name,omitempty"`
	Description  *string             `json:"description,omitempty"`
	StateMachine *model.StateMachine `json:"state_machine,omitempty"`
	SortOrder    *int                `json:"sort_order,omitempty"`
}

type WorkflowService struct {
	repo        repository.WorkflowRepository
	processRepo repository.ProcessRepository
	logger      *slog.Logger
}

func NewWorkflowService(repo repository.WorkflowRepository, processRepo repository.ProcessRepository, logger *slog.Logger) *WorkflowService {
	return &WorkflowService{repo: repo, processRepo: processRepo, logger: logger}
}

func (s *WorkflowService) Create(ctx context.Context, processID uuid.UUID, input CreateWorkflowInput) (*model.Workflow, error) {
	s.logger.Debug("creating workflow", slog.String("name", input.Name), slog.String("process_id", processID.String()))

	p, err := s.processRepo.GetByID(ctx, processID)
	if err != nil {
		s.logger.Error("failed to create workflow", slog.String("process_id", processID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("process: %w", err)
	}
	if p.Status == "archived" {
		s.logger.Error("failed to create workflow", slog.String("process_id", processID.String()), slog.String("error", "cannot add workflow to archived process"))
		return nil, fmt.Errorf("cannot add workflow to archived process")
	}

	now := time.Now().UTC()
	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    processID,
		Name:         input.Name,
		Description:  input.Description,
		StateMachine: input.StateMachine,
		SortOrder:    input.SortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := w.Validate(); err != nil {
		s.logger.Error("failed to create workflow", slog.String("process_id", processID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Create(ctx, w); err != nil {
		s.logger.Error("failed to create workflow", slog.String("process_id", processID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflow created", slog.String("id", w.ID.String()), slog.String("name", w.Name), slog.String("process_id", processID.String()))
	return w, nil
}

func (s *WorkflowService) GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error) {
	s.logger.Debug("getting workflow", slog.String("id", id.String()))

	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflow retrieved", slog.String("id", id.String()))
	return w, nil
}

func (s *WorkflowService) ListByProcessID(ctx context.Context, processID uuid.UUID) ([]*model.Workflow, error) {
	s.logger.Debug("listing workflows", slog.String("process_id", processID.String()))

	results, err := s.repo.ListByProcessID(ctx, processID)
	if err != nil {
		s.logger.Error("failed to list workflows", slog.String("process_id", processID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflows listed", slog.String("process_id", processID.String()), slog.Int("count", len(results)))
	return results, nil
}

func (s *WorkflowService) Update(ctx context.Context, id uuid.UUID, input UpdateWorkflowInput) (*model.Workflow, error) {
	s.logger.Debug("updating workflow", slog.String("id", id.String()))

	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to update workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	if input.Name != nil {
		w.Name = *input.Name
	}
	if input.Description != nil {
		w.Description = *input.Description
	}
	if input.StateMachine != nil {
		w.StateMachine = *input.StateMachine
	}
	if input.SortOrder != nil {
		w.SortOrder = *input.SortOrder
	}
	w.UpdatedAt = time.Now().UTC()

	if err := w.Validate(); err != nil {
		s.logger.Error("failed to update workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Update(ctx, w); err != nil {
		s.logger.Error("failed to update workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflow updated", slog.String("id", id.String()), slog.String("name", w.Name))
	return w, nil
}

func (s *WorkflowService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting workflow", slog.String("id", id.String()))

	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}
	p, err := s.processRepo.GetByID(ctx, w.ProcessID)
	if err != nil {
		s.logger.Error("failed to delete workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return fmt.Errorf("process: %w", err)
	}
	if p.Status != "draft" {
		s.logger.Error("failed to delete workflow", slog.String("id", id.String()), slog.String("error", "can only delete workflows from draft processes"), slog.String("process_status", p.Status))
		return fmt.Errorf("can only delete workflows from draft processes (process status: %s)", p.Status)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete workflow", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("workflow deleted", slog.String("id", id.String()))
	return nil
}
