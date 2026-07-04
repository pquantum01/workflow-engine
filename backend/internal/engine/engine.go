package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

const (
	EventTaskCreated       = "task.created"
	EventTaskCompleted     = "task.completed"
	EventTaskCancelled     = "task.cancelled"
	EventStateTransitioned = "state.transitioned"
	EventInstanceStarted   = "instance.started"
	EventInstanceCompleted = "instance.completed"
	EventInstanceCancelled = "instance.cancelled"
)

type Event struct {
	Type               string
	WorkflowInstanceID uuid.UUID
	Data               map[string]any
}

type Engine struct {
	db        *sql.DB
	defRepo   repository.WorkflowDefRepo
	instRepo  repository.WorkflowInstanceRepo
	taskRepo  repository.TaskInstanceRepo
	transRepo repository.StateTransitionRepo
	log       *logrus.Logger
}

func New(
	db *sql.DB,
	defRepo repository.WorkflowDefRepo,
	instRepo repository.WorkflowInstanceRepo,
	taskRepo repository.TaskInstanceRepo,
	transRepo repository.StateTransitionRepo,
	log *logrus.Logger,
) *Engine {
	return &Engine{
		db:        db,
		defRepo:   defRepo,
		instRepo:  instRepo,
		taskRepo:  taskRepo,
		transRepo: transRepo,
		log:       log,
	}
}

// Start creates a workflow instance in start_state and spawns its tasks.
func (e *Engine) Start(ctx context.Context, defID uuid.UUID, createdBy string) (*model.WorkflowInstance, []Event, error) {
	e.log.WithFields(logrus.Fields{"def_id": defID, "created_by": createdBy}).Debug("engine.Start")
	def, err := e.defRepo.GetByID(ctx, defID)
	if err != nil {
		e.log.WithFields(logrus.Fields{"def_id": defID, "err": err}).Error("engine.Start: get def failed")
		return nil, nil, fmt.Errorf("get def: %w", err)
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		e.log.WithField("err", err).Error("engine.Start: begin tx failed")
		return nil, nil, err
	}
	defer tx.Rollback()

	txInstRepo := e.instRepo.WithTx(tx)
	txTaskRepo := e.taskRepo.WithTx(tx)

	inst := &model.WorkflowInstance{
		ID:             uuid.New(),
		WorkflowDefID:  def.ID,
		CurrentStateID: def.Definition.StartState,
		Status:         model.InstanceStatusActive,
		CreatedBy:      createdBy,
		Context:        map[string]any{},
		StartedAt:      time.Now().UTC(),
	}
	if err := txInstRepo.Create(ctx, inst); err != nil {
		e.log.WithField("err", err).Error("engine.Start: create instance failed")
		return nil, nil, fmt.Errorf("create instance: %w", err)
	}

	events, err := e.spawnTasks(ctx, txTaskRepo, def, inst, def.Definition.StartState, createdBy)
	if err != nil {
		e.log.WithFields(logrus.Fields{"inst_id": inst.ID, "state": def.Definition.StartState, "err": err}).Error("engine.Start: spawn tasks failed")
		return nil, nil, fmt.Errorf("spawn tasks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		e.log.WithField("err", err).Error("engine.Start: commit failed")
		return nil, nil, err
	}

	e.log.WithFields(logrus.Fields{
		"inst_id":       inst.ID,
		"def_id":        def.ID,
		"def_name":      def.Name,
		"start_state":   inst.CurrentStateID,
		"tasks_spawned": len(events),
	}).Info("engine.Start: instance started")

	events = append([]Event{{
		Type:               EventInstanceStarted,
		WorkflowInstanceID: inst.ID,
		Data:               map[string]any{"workflow_def_id": def.ID.String(), "start_state": inst.CurrentStateID},
	}}, events...)

	return inst, events, nil
}

// Complete processes a task completion. Returns updated instance and events.
func (e *Engine) Complete(ctx context.Context, taskID uuid.UUID, status model.TaskStatus, nextState string, createdBy string) (*model.WorkflowInstance, []Event, error) {
	e.log.WithFields(logrus.Fields{
		"task_id":      taskID,
		"status":       status,
		"next_state":   nextState,
		"completed_by": createdBy,
	}).Debug("engine.Complete")

	if status != model.TaskStatusDone && status != model.TaskStatusRejected {
		return nil, nil, fmt.Errorf("%w: status must be done or rejected", ErrInvalidStatus)
	}

	// Load task outside tx to get instanceID
	task, err := e.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			e.log.WithField("task_id", taskID).Warn("engine.Complete: task not found")
			return nil, nil, ErrNotFound
		}
		e.log.WithFields(logrus.Fields{"task_id": taskID, "err": err}).Error("engine.Complete: get task failed")
		return nil, nil, err
	}

	// Quick non-locking instance check to give correct error before tx
	inst, err := e.instRepo.GetByID(ctx, task.WorkflowInstanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if inst.IsTerminal() {
		e.log.WithFields(logrus.Fields{
			"inst_id":     inst.ID,
			"inst_status": inst.Status,
			"task_id":     taskID,
		}).Warn("engine.Complete: instance already terminal")
		return nil, nil, ErrInstanceTerminal
	}
	if !task.IsCompletable() {
		e.log.WithFields(logrus.Fields{
			"task_id":     taskID,
			"task_status": task.Status,
		}).Warn("engine.Complete: task not completable")
		return nil, nil, ErrTaskNotCompletable
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	txInstRepo := e.instRepo.WithTx(tx)
	txTaskRepo := e.taskRepo.WithTx(tx)
	txTransRepo := e.transRepo.WithTx(tx)

	// Lock instance row — authoritative check under lock
	inst, err = txInstRepo.GetByIDForUpdate(ctx, task.WorkflowInstanceID)
	if err != nil {
		return nil, nil, fmt.Errorf("lock instance: %w", err)
	}
	if inst.IsTerminal() {
		return nil, nil, ErrInstanceTerminal
	}

	// Re-check task under lock (may have changed)
	task, err = txTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if !task.IsCompletable() {
		return nil, nil, ErrTaskNotCompletable
	}

	def, err := e.defRepo.GetByID(ctx, inst.WorkflowDefID)
	if err != nil {
		return nil, nil, fmt.Errorf("get def: %w", err)
	}

	var events []Event
	now := time.Now().UTC()

	if status == model.TaskStatusRejected {
		// Mark task rejected
		if err := txTaskRepo.UpdateStatus(ctx, task.ID, model.TaskStatusRejected, &now); err != nil {
			return nil, nil, err
		}
		events = append(events, Event{Type: EventTaskCompleted, WorkflowInstanceID: inst.ID,
			Data: map[string]any{"task_id": task.ID.String(), "status": "rejected"}})

		// Cancel all other open tasks
		cancelEvents, err := e.cancelOpenTasks(ctx, txTaskRepo, inst.ID, task.ID)
		if err != nil {
			return nil, nil, err
		}
		events = append(events, cancelEvents...)

		// Cancel instance
		if err := txInstRepo.UpdateStatus(ctx, inst.ID, model.InstanceStatusCancelled, &now); err != nil {
			return nil, nil, err
		}
		inst.Status = model.InstanceStatusCancelled
		inst.CompletedAt = &now
		events = append(events, Event{Type: EventInstanceCancelled, WorkflowInstanceID: inst.ID, Data: map[string]any{}})

		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		e.log.WithFields(logrus.Fields{"task_id": taskID, "inst_id": inst.ID}).Info("engine.Complete: task rejected, instance cancelled")
		return inst, events, nil
	}

	// status == done
	if err := txTaskRepo.UpdateStatus(ctx, task.ID, model.TaskStatusDone, &now); err != nil {
		return nil, nil, err
	}
	events = append(events, Event{Type: EventTaskCompleted, WorkflowInstanceID: inst.ID,
		Data: map[string]any{"task_id": task.ID.String(), "status": "done"}})

	// Adhoc task (or any task) in a past state: instance already moved on, no transition.
	if task.StateID != inst.CurrentStateID {
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		e.log.WithFields(logrus.Fields{
			"task_id":       taskID,
			"task_state":    task.StateID,
			"current_state": inst.CurrentStateID,
		}).Debug("engine.Complete: task state != current state, skipping transition")
		return inst, events, nil
	}

	mode := def.Definition.TransitionMode(task.StateID)
	nextStateID, cancelEvents, err := e.resolveTransition(ctx, txTaskRepo, def, inst, task, mode, nextState)
	if err != nil {
		return nil, nil, err
	}
	events = append(events, cancelEvents...)

	if nextStateID == "" {
		// Linear: waiting for other tasks
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		e.log.WithFields(logrus.Fields{"task_id": taskID, "inst_id": inst.ID, "state": task.StateID}).Debug("engine.Complete: task done, waiting for sibling tasks")
		return inst, events, nil
	}

	// Fire transition
	tr := &model.StateTransition{
		ID:                 uuid.New(),
		WorkflowInstanceID: inst.ID,
		FromState:          task.StateID,
		ToState:            nextStateID,
		TransitionedAt:     now,
	}
	if err := txTransRepo.Append(ctx, tr); err != nil {
		return nil, nil, err
	}
	if err := txInstRepo.UpdateState(ctx, inst.ID, nextStateID); err != nil {
		return nil, nil, err
	}
	inst.CurrentStateID = nextStateID
	events = append(events, Event{Type: EventStateTransitioned, WorkflowInstanceID: inst.ID,
		Data: map[string]any{"from_state": task.StateID, "to_state": nextStateID}})

	nextStateDef, _ := def.Definition.FindState(nextStateID)
	isTerminalState := nextStateDef != nil && len(nextStateDef.Tasks) == 0 && len(def.Definition.OutgoingTransitions(nextStateID)) == 0

	if isTerminalState {
		if err := txInstRepo.UpdateStatus(ctx, inst.ID, model.InstanceStatusCompleted, &now); err != nil {
			return nil, nil, err
		}
		inst.Status = model.InstanceStatusCompleted
		inst.CompletedAt = &now
		events = append(events, Event{Type: EventInstanceCompleted, WorkflowInstanceID: inst.ID,
			Data: map[string]any{"final_state": nextStateID}})
	} else {
		// Spawn tasks for next state
		spawnEvents, err := e.spawnTasks(ctx, txTaskRepo, def, inst, nextStateID, createdBy)
		if err != nil {
			return nil, nil, err
		}
		events = append(events, spawnEvents...)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	e.log.WithFields(logrus.Fields{
		"task_id":     taskID,
		"inst_id":     inst.ID,
		"from_state":  task.StateID,
		"to_state":    nextStateID,
		"inst_status": inst.Status,
	}).Info("engine.Complete: transition fired")
	return inst, events, nil
}

// Cancel cancels an active instance and all its open tasks.
func (e *Engine) Cancel(ctx context.Context, instanceID uuid.UUID) (*model.WorkflowInstance, []Event, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	txInstRepo := e.instRepo.WithTx(tx)
	txTaskRepo := e.taskRepo.WithTx(tx)

	inst, err := txInstRepo.GetByIDForUpdate(ctx, instanceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if inst.IsTerminal() {
		return nil, nil, ErrInstanceTerminal
	}

	now := time.Now().UTC()
	cancelEvents, err := e.cancelOpenTasks(ctx, txTaskRepo, instanceID, uuid.Nil)
	if err != nil {
		return nil, nil, err
	}

	if err := txInstRepo.UpdateStatus(ctx, instanceID, model.InstanceStatusCancelled, &now); err != nil {
		return nil, nil, err
	}
	inst.Status = model.InstanceStatusCancelled
	inst.CompletedAt = &now

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	events := append(cancelEvents, Event{
		Type:               EventInstanceCancelled,
		WorkflowInstanceID: instanceID,
		Data:               map[string]any{},
	})
	e.log.WithField("inst_id", instanceID).Info("engine.Cancel: instance cancelled")
	return inst, events, nil
}

// resolveTransition determines the next state ID (empty = wait for more tasks).
func (e *Engine) resolveTransition(
	ctx context.Context,
	taskRepo repository.TaskInstanceRepo,
	def *model.WorkflowDef,
	inst *model.WorkflowInstance,
	task *model.TaskInstance,
	mode model.TransitionMode,
	callerNextState string,
) (nextStateID string, cancelEvents []Event, err error) {
	outs := def.Definition.OutgoingTransitions(task.StateID)

	switch mode {
	case model.ModeLinear:
		// Wait for all non-cancelled tasks in this state to be done
		stateTasks, err := taskRepo.ListByInstanceAndState(ctx, inst.ID, task.StateID)
		if err != nil {
			return "", nil, err
		}
		for _, t := range stateTasks {
			if t.Status == model.TaskStatusCancelled {
				continue
			}
			if t.Status != model.TaskStatusDone {
				return "", nil, nil // still waiting
			}
		}
		return outs[0].ToState, nil, nil

	case model.ModeTaskBound:
		// Find the edge whose trigger matches this task
		for _, tr := range outs {
			if tr.Trigger != nil && tr.Trigger.TaskKey == task.TaskKey {
				// Cancel sibling tasks
				stateTasks, err := taskRepo.ListByInstanceAndState(ctx, inst.ID, task.StateID)
				if err != nil {
					return "", nil, err
				}
				now := time.Now().UTC()
				for _, sibling := range stateTasks {
					if sibling.ID == task.ID || sibling.IsTerminal() {
						continue
					}
					if err := taskRepo.UpdateStatus(ctx, sibling.ID, model.TaskStatusCancelled, &now); err != nil {
						return "", nil, err
					}
					cancelEvents = append(cancelEvents, Event{
						Type:               EventTaskCancelled,
						WorkflowInstanceID: inst.ID,
						Data:               map[string]any{"task_id": sibling.ID.String()},
					})
				}
				return tr.ToState, cancelEvents, nil
			}
		}
		return "", nil, fmt.Errorf("no task-bound edge found for task key %q", task.TaskKey)

	case model.ModeUserPicks:
		if callerNextState == "" {
			return "", nil, ErrNextStateRequired
		}
		// Validate next_state is a valid outgoing edge
		for _, tr := range outs {
			if tr.ToState == callerNextState {
				return callerNextState, nil, nil
			}
		}
		return "", nil, fmt.Errorf("%w: %q", ErrInvalidNextState, callerNextState)
	}

	return "", nil, fmt.Errorf("unknown transition mode %d", mode)
}

func (e *Engine) spawnTasks(ctx context.Context, taskRepo repository.TaskInstanceRepo, def *model.WorkflowDef, inst *model.WorkflowInstance, stateID string, createdBy string) ([]Event, error) {
	stateDef, ok := def.Definition.FindState(stateID)
	if !ok || len(stateDef.Tasks) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	tasks := make([]*model.TaskInstance, 0, len(stateDef.Tasks))
	for _, td := range stateDef.Tasks {
		tasks = append(tasks, &model.TaskInstance{
			ID:                 uuid.New(),
			WorkflowInstanceID: inst.ID,
			StateID:            stateID,
			TaskKey:            td.Key,
			TaskLabel:          td.Label,
			TaskType:           td.Type,
			Status:             model.TaskStatusNew,
			CreatedBy:          createdBy,
			CreatedAt:          now,
		})
	}
	if err := taskRepo.BulkCreate(ctx, tasks); err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(tasks))
	for _, t := range tasks {
		events = append(events, Event{
			Type:               EventTaskCreated,
			WorkflowInstanceID: inst.ID,
			Data:               map[string]any{"task_id": t.ID.String(), "task_key": t.TaskKey, "state_id": stateID},
		})
	}
	return events, nil
}

func (e *Engine) cancelOpenTasks(ctx context.Context, taskRepo repository.TaskInstanceRepo, instanceID uuid.UUID, excludeID uuid.UUID) ([]Event, error) {
	openTasks, err := taskRepo.ListByInstance(ctx, instanceID, []model.TaskStatus{
		model.TaskStatusNew, model.TaskStatusInProgress,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var events []Event
	for _, t := range openTasks {
		if t.ID == excludeID {
			continue
		}
		if err := taskRepo.UpdateStatus(ctx, t.ID, model.TaskStatusCancelled, &now); err != nil {
			return nil, err
		}
		events = append(events, Event{
			Type:               EventTaskCancelled,
			WorkflowInstanceID: instanceID,
			Data:               map[string]any{"task_id": t.ID.String()},
		})
	}
	return events, nil
}

// Sentinel errors
var (
	ErrNotFound        = fmt.Errorf("not found")
	ErrInstanceTerminal = fmt.Errorf("instance is in terminal state")
	ErrTaskNotCompletable = fmt.Errorf("task is not in a completable state")
	ErrInvalidNextState = fmt.Errorf("invalid next_state")
	ErrNextStateRequired = fmt.Errorf("next_state is required for user-picks transition")
	ErrInvalidStatus    = fmt.Errorf("invalid status")
)
