package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type InstanceService struct {
	eng       *engine.Engine
	instRepo  repository.WorkflowInstanceRepo
	taskRepo  repository.TaskInstanceRepo
	transRepo repository.StateTransitionRepo
	log       *logrus.Logger
}

func NewInstanceService(
	eng *engine.Engine,
	instRepo repository.WorkflowInstanceRepo,
	taskRepo repository.TaskInstanceRepo,
	transRepo repository.StateTransitionRepo,
	log *logrus.Logger,
) *InstanceService {
	return &InstanceService{
		eng:       eng,
		instRepo:  instRepo,
		taskRepo:  taskRepo,
		transRepo: transRepo,
		log:       log,
	}
}

func (s *InstanceService) Start(ctx context.Context, defID uuid.UUID, createdBy string) (*model.WorkflowInstance, []engine.Event, error) {
	inst, events, err := s.eng.Start(ctx, defID, createdBy)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		s.log.WithFields(logrus.Fields{"def_id": defID, "err": err}).Error("instance_service.Start failed")
		return nil, nil, err
	}
	s.log.WithFields(logrus.Fields{
		"inst_id":    inst.ID,
		"def_id":     defID,
		"created_by": createdBy,
	}).Info("workflow instance started")
	return inst, events, nil
}

func (s *InstanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	inst, err := s.instRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.log.WithFields(logrus.Fields{"inst_id": id, "err": err}).Error("instance_service.GetByID failed")
		return nil, err
	}
	return inst, nil
}

func (s *InstanceService) List(ctx context.Context, filter repository.InstanceFilter) ([]*model.WorkflowInstance, int, error) {
	return s.instRepo.List(ctx, filter)
}

func (s *InstanceService) Cancel(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, []engine.Event, error) {
	inst, events, err := s.eng.Cancel(ctx, id)
	if err != nil {
		if errors.Is(err, engine.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		if errors.Is(err, engine.ErrInstanceTerminal) {
			return nil, nil, ErrInstanceTerminal
		}
		s.log.WithFields(logrus.Fields{"inst_id": id, "err": err}).Error("instance_service.Cancel failed")
		return nil, nil, err
	}
	s.log.WithField("inst_id", id).Info("workflow instance cancelled")
	return inst, events, nil
}

func (s *InstanceService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := s.instRepo.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	s.log.WithField("inst_id", id).Info("workflow instance soft-deleted")
	return nil
}

func (s *InstanceService) ListTasks(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error) {
	return s.taskRepo.ListByInstance(ctx, instanceID, statusFilter)
}

func (s *InstanceService) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error) {
	return s.transRepo.ListByInstance(ctx, instanceID)
}
