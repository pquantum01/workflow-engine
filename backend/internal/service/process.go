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

type CreateProcessInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProcessInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type ProcessService struct {
	repo   repository.ProcessRepository
	logger *slog.Logger
}

func NewProcessService(repo repository.ProcessRepository, logger *slog.Logger) *ProcessService {
	return &ProcessService{repo: repo, logger: logger}
}

func (s *ProcessService) Create(ctx context.Context, input CreateProcessInput) (*model.Process, error) {
	s.logger.Debug("creating process", slog.String("name", input.Name))

	now := time.Now().UTC()
	p := &model.Process{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Version:     1,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.Validate(); err != nil {
		s.logger.Error("failed to create process", slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Create(ctx, p); err != nil {
		s.logger.Error("failed to create process", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("process created", slog.String("id", p.ID.String()), slog.String("name", p.Name))
	return p, nil
}

func (s *ProcessService) GetByID(ctx context.Context, id uuid.UUID) (*model.Process, error) {
	s.logger.Debug("getting process", slog.String("id", id.String()))

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("process retrieved", slog.String("id", id.String()))
	return p, nil
}

func (s *ProcessService) List(ctx context.Context, offset, limit int) ([]*model.Process, error) {
	s.logger.Debug("listing processes", slog.Int("offset", offset), slog.Int("limit", limit))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	results, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("failed to list processes", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("processes listed", slog.Int("count", len(results)))
	return results, nil
}

func (s *ProcessService) Update(ctx context.Context, id uuid.UUID, input UpdateProcessInput) (*model.Process, error) {
	s.logger.Debug("updating process", slog.String("id", id.String()))

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to update process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}
	if p.Status == "archived" {
		s.logger.Error("failed to update process", slog.String("id", id.String()), slog.String("error", "cannot update an archived process"))
		return nil, fmt.Errorf("cannot update an archived process")
	}

	if input.Name != nil {
		p.Name = *input.Name
	}
	if input.Description != nil {
		p.Description = *input.Description
	}
	if input.Status != nil {
		p.Status = *input.Status
	}
	p.UpdatedAt = time.Now().UTC()

	if err := p.Validate(); err != nil {
		s.logger.Error("failed to update process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation: %w", err)
	}
	if err := s.repo.Update(ctx, p); err != nil {
		s.logger.Error("failed to update process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("process updated", slog.String("id", id.String()), slog.String("name", p.Name), slog.String("status", p.Status))
	return p, nil
}

func (s *ProcessService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting process", slog.String("id", id.String()))

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}
	if p.Status != "draft" {
		s.logger.Error("failed to delete process", slog.String("id", id.String()), slog.String("error", "can only delete draft processes"), slog.String("status", p.Status))
		return fmt.Errorf("can only delete draft processes (current status: %s)", p.Status)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete process", slog.String("id", id.String()), slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("process deleted", slog.String("id", id.String()))
	return nil
}
