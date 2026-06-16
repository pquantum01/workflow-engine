package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/workflow-engine/backend/internal/engine"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

type TaskService struct {
	eng      *engine.Engine
	taskRepo repository.TaskInstanceRepo
	log      *logrus.Logger
}

func NewTaskService(
	eng *engine.Engine,
	taskRepo repository.TaskInstanceRepo,
	log *logrus.Logger,
) *TaskService {
	return &TaskService{eng: eng, taskRepo: taskRepo, log: log}
}

func (s *TaskService) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		s.log.WithFields(logrus.Fields{"task_id": id, "err": err}).Error("task_service.GetByID failed")
		return nil, err
	}
	return task, nil
}

// Start marks a task in_progress (optional affordance).
func (s *TaskService) Start(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if task.Status != model.TaskStatusNew && task.Status != model.TaskStatusInProgress {
		s.log.WithFields(logrus.Fields{"task_id": id, "status": task.Status}).Warn("task_service.Start: task not startable")
		return nil, ErrTaskNotCompletable
	}
	if task.Status == model.TaskStatusNew {
		if err := s.taskRepo.UpdateStatus(ctx, id, model.TaskStatusInProgress, nil); err != nil {
			return nil, err
		}
		task.Status = model.TaskStatusInProgress
		s.log.WithField("task_id", id).Debug("task marked in_progress")
	}
	return task, nil
}

// Complete delegates to the engine.
func (s *TaskService) Complete(ctx context.Context, id uuid.UUID, status model.TaskStatus, nextState string, completedBy string) (*model.WorkflowInstance, []engine.Event, error) {
	if status != model.TaskStatusDone && status != model.TaskStatusRejected {
		return nil, nil, ErrValidation
	}

	s.log.WithFields(logrus.Fields{
		"task_id":      id,
		"status":       status,
		"next_state":   nextState,
		"completed_by": completedBy,
	}).Debug("task_service.Complete")

	inst, events, err := s.eng.Complete(ctx, id, status, nextState, completedBy)
	if err != nil {
		switch {
		case errors.Is(err, engine.ErrNotFound):
			return nil, nil, ErrNotFound
		case errors.Is(err, engine.ErrInstanceTerminal):
			return nil, nil, ErrInstanceTerminal
		case errors.Is(err, engine.ErrTaskNotCompletable):
			return nil, nil, ErrTaskNotCompletable
		case errors.Is(err, engine.ErrInvalidNextState):
			return nil, nil, ErrInvalidNextState
		case errors.Is(err, engine.ErrNextStateRequired):
			return nil, nil, ErrInvalidNextState
		}
		s.log.WithFields(logrus.Fields{"task_id": id, "err": err}).Error("task_service.Complete engine error")
		return nil, nil, err
	}

	// outcome stored via engine in v2
	if completedBy != "" {
		_ = time.Now()
	}

	s.log.WithFields(logrus.Fields{
		"task_id":   id,
		"status":    status,
		"inst_id":   inst.ID,
		"inst_status": inst.Status,
	}).Info("task completed")
	return inst, events, nil
}
