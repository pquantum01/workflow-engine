package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/workflowloader"
)

type WorkflowDefService struct {
	repo     repository.WorkflowDefRepo
	instRepo repository.WorkflowInstanceRepo
	loader   *workflowloader.Loader
	log      *logrus.Logger
}

func NewWorkflowDefService(
	repo repository.WorkflowDefRepo,
	instRepo repository.WorkflowInstanceRepo,
	loader *workflowloader.Loader,
	log *logrus.Logger,
) *WorkflowDefService {
	return &WorkflowDefService{repo: repo, instRepo: instRepo, loader: loader, log: log}
}

func (s *WorkflowDefService) List(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error) {
	defs, err := s.repo.List(ctx, includeDeleted)
	if err != nil {
		s.log.WithField("err", err).Error("workflow_def_service.List failed")
		return nil, err
	}
	return defs, nil
}

func (s *WorkflowDefService) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
	def, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.log.WithFields(logrus.Fields{"id": id, "err": err}).Error("workflow_def_service.GetByID failed")
		return nil, err
	}
	return def, nil
}

func (s *WorkflowDefService) Reload(ctx context.Context) (workflowloader.LoadResult, error) {
	result, err := s.loader.Load(ctx)
	if err != nil {
		s.log.WithField("err", err).Error("workflow_def_service.Reload failed")
		return result, err
	}
	s.log.WithFields(logrus.Fields{
		"loaded":   len(result.Loaded),
		"warnings": len(result.Warnings),
		"errors":   len(result.Errors),
	}).Info("workflow defs reloaded")
	return result, nil
}

func (s *WorkflowDefService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	status := model.InstanceStatusActive
	_, total, err := s.instRepo.List(ctx, repository.InstanceFilter{
		DefID:  &id,
		Status: &status,
		Limit:  1,
	})
	if err != nil {
		return err
	}
	if total > 0 {
		s.log.WithField("def_id", id).Warn("workflow_def_service.SoftDelete: active instances exist")
		return ErrDefHasActiveInstances
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	s.log.WithField("def_id", id).Info("workflow def soft-deleted")
	return nil
}
