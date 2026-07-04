package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type TaskFieldPatch struct {
	Priority               *string
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
}

type AdhocTaskRequest struct {
	StateID                string
	Label                  string
	TaskType               string
	CreatedBy              string
	Priority               string
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
}

type TaskService struct {
	eng      *engine.Engine
	taskRepo repository.TaskInstanceRepo
	instRepo repository.WorkflowInstanceRepo
	defRepo  repository.WorkflowDefRepo
	log      *logrus.Logger
}

func NewTaskService(
	eng *engine.Engine,
	taskRepo repository.TaskInstanceRepo,
	log *logrus.Logger,
) *TaskService {
	return &TaskService{eng: eng, taskRepo: taskRepo, log: log}
}

func (s *TaskService) SetInstRepo(r repository.WorkflowInstanceRepo) { s.instRepo = r }
func (s *TaskService) SetDefRepo(r repository.WorkflowDefRepo)       { s.defRepo = r }

func (s *TaskService) UpdateFields(ctx context.Context, id uuid.UUID, patch TaskFieldPatch) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if task.IsTerminal() {
		return ErrTaskCompleted
	}
	return s.taskRepo.UpdateFields(ctx, id, repository.TaskInstancePatch{
		Priority:               patch.Priority,
		ExpectedCompletionDate: patch.ExpectedCompletionDate,
		ExpectedCompletionTime: patch.ExpectedCompletionTime,
	})
}

func (s *TaskService) CreateAdhoc(ctx context.Context, instanceID uuid.UUID, req AdhocTaskRequest) (*model.TaskInstance, error) {
	inst, err := s.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if inst.Status != model.InstanceStatusActive {
		return nil, ErrInstanceNotActive
	}

	def, err := s.defRepo.GetByID(ctx, inst.WorkflowDefID)
	if err != nil {
		return nil, err
	}

	stateExists := false
	for _, st := range def.Definition.States {
		if st.ID == req.StateID {
			stateExists = true
			break
		}
	}
	if !stateExists {
		return nil, ErrInvalidState
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	task := &model.TaskInstance{
		ID:                     uuid.New(),
		WorkflowInstanceID:     instanceID,
		StateID:                req.StateID,
		TaskKey:                generateAdhocKey(req.Label),
		TaskLabel:              req.Label,
		TaskType:               req.TaskType,
		Status:                 model.TaskStatusNew,
		CreatedBy:              req.CreatedBy,
		CreatedAt:              time.Now(),
		Priority:               priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
		IsAdhoc:                true,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.log.WithField("err", err).Error("taskService.CreateAdhoc: repo error")
		return nil, err
	}
	return task, nil
}

func generateAdhocKey(label string) string {
	slug := make([]byte, 0, len(label))
	for _, ch := range strings.ToLower(label) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			slug = append(slug, byte(ch))
		} else {
			if len(slug) > 0 && slug[len(slug)-1] != '_' {
				slug = append(slug, '_')
			}
		}
	}
	for len(slug) > 0 && slug[len(slug)-1] == '_' {
		slug = slug[:len(slug)-1]
	}
	if len(slug) > 40 {
		slug = slug[:40]
	}
	b := make([]byte, 2)
	rand.Read(b)
	return fmt.Sprintf("%s_%x", string(slug), b)
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
