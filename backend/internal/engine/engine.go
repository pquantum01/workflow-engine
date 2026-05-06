package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type TransitionResult struct {
	WorkflowInstanceID uuid.UUID `json:"workflow_instance_id"`
	FromState          string    `json:"from_state"`
	ToState            string    `json:"to_state"`
	TransitionName     string    `json:"transition_name"`
	IsTerminal         bool      `json:"is_terminal"`
}

type Engine interface {
	Evaluate(ctx context.Context, workflowInstanceID uuid.UUID) (*TransitionResult, error)
	TriggerTransition(ctx context.Context, workflowInstanceID uuid.UUID, transitionName, triggeredBy, triggerType string) (*TransitionResult, error)
}

type engine struct {
	workflowRepo         repository.WorkflowRepository
	workflowInstanceRepo repository.WorkflowInstanceRepository
	taskRepo             repository.TaskRepository
	taskInstanceRepo     repository.TaskInstanceRepository
	stateTransitionRepo  repository.StateTransitionRepository
	logger               *slog.Logger
}

func New(
	workflowRepo repository.WorkflowRepository,
	workflowInstanceRepo repository.WorkflowInstanceRepository,
	taskRepo repository.TaskRepository,
	taskInstanceRepo repository.TaskInstanceRepository,
	stateTransitionRepo repository.StateTransitionRepository,
	logger *slog.Logger,
) Engine {
	return &engine{
		workflowRepo:         workflowRepo,
		workflowInstanceRepo: workflowInstanceRepo,
		taskRepo:             taskRepo,
		taskInstanceRepo:     taskInstanceRepo,
		stateTransitionRepo:  stateTransitionRepo,
		logger:               logger,
	}
}

func (e *engine) Evaluate(ctx context.Context, workflowInstanceID uuid.UUID) (*TransitionResult, error) {
	e.logger.Debug("evaluating workflow instance",
		slog.String("workflow_instance_id", workflowInstanceID.String()),
	)

	wi, err := e.workflowInstanceRepo.GetByID(ctx, workflowInstanceID)
	if err != nil {
		e.logger.Error("failed to evaluate workflow instance",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get workflow instance: %w", err)
	}

	if wi.Status != "active" {
		e.logger.Error("workflow instance is not active",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("status", wi.Status),
		)
		return nil, fmt.Errorf("workflow instance is not active (status: %s)", wi.Status)
	}

	wf, err := e.workflowRepo.GetByID(ctx, wi.WorkflowID)
	if err != nil {
		e.logger.Error("failed to get workflow definition",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get workflow definition: %w", err)
	}

	if isTerminal(wf.StateMachine, wi.CurrentState) {
		e.logger.Error("workflow instance is in terminal state",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("current_state", wi.CurrentState),
		)
		return nil, fmt.Errorf("workflow instance is in terminal state %q", wi.CurrentState)
	}

	taskInstances, err := e.taskInstanceRepo.ListByWorkflowInstanceID(ctx, workflowInstanceID)
	if err != nil {
		e.logger.Error("failed to list task instances",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("list task instances: %w", err)
	}

	tasks, err := e.taskRepo.ListByWorkflowID(ctx, wi.WorkflowID)
	if err != nil {
		e.logger.Error("failed to list task definitions",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("list task definitions: %w", err)
	}

	taskNameMap := buildTaskNameMap(tasks, taskInstances)

	for _, t := range wf.StateMachine.Transitions {
		if t.FromState != wi.CurrentState {
			continue
		}
		if t.CompletionRule == nil {
			continue // unguarded transitions need explicit trigger
		}
		if isCompletionRuleSatisfied(t.CompletionRule, taskNameMap) {
			return e.applyTransition(ctx, wi, wf, t, "system", "rule")
		}
	}

	e.logger.Info("evaluate completed with no transition",
		slog.String("workflow_instance_id", workflowInstanceID.String()),
		slog.String("current_state", wi.CurrentState),
	)
	return nil, nil
}

func (e *engine) TriggerTransition(ctx context.Context, workflowInstanceID uuid.UUID, transitionName, triggeredBy, triggerType string) (*TransitionResult, error) {
	e.logger.Debug("triggering transition",
		slog.String("workflow_instance_id", workflowInstanceID.String()),
		slog.String("transition_name", transitionName),
		slog.String("triggered_by", triggeredBy),
		slog.String("trigger_type", triggerType),
	)

	wi, err := e.workflowInstanceRepo.GetByID(ctx, workflowInstanceID)
	if err != nil {
		e.logger.Error("failed to trigger transition",
			slog.String("workflow_instance_id", workflowInstanceID.String()),
			slog.String("transition_name", transitionName),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get workflow instance: %w", err)
	}

	if wi.Status != "active" {
		return nil, fmt.Errorf("workflow instance is not active (status: %s)", wi.Status)
	}

	wf, err := e.workflowRepo.GetByID(ctx, wi.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("get workflow definition: %w", err)
	}

	if isTerminal(wf.StateMachine, wi.CurrentState) {
		return nil, fmt.Errorf("workflow instance is in terminal state %q", wi.CurrentState)
	}

	var transition *model.Transition
	for i := range wf.StateMachine.Transitions {
		if wf.StateMachine.Transitions[i].Name == transitionName {
			transition = &wf.StateMachine.Transitions[i]
			break
		}
	}

	if transition == nil {
		return nil, fmt.Errorf("transition %q not found", transitionName)
	}

	if transition.FromState != wi.CurrentState {
		return nil, fmt.Errorf("transition %q requires state %q but current state is %q", transitionName, transition.FromState, wi.CurrentState)
	}

	if transition.CompletionRule != nil {
		taskInstances, err := e.taskInstanceRepo.ListByWorkflowInstanceID(ctx, workflowInstanceID)
		if err != nil {
			return nil, fmt.Errorf("list task instances: %w", err)
		}
		tasks, err := e.taskRepo.ListByWorkflowID(ctx, wi.WorkflowID)
		if err != nil {
			return nil, fmt.Errorf("list task definitions: %w", err)
		}
		taskNameMap := buildTaskNameMap(tasks, taskInstances)

		if !isCompletionRuleSatisfied(transition.CompletionRule, taskNameMap) {
			return nil, fmt.Errorf("completion rule not satisfied for transition %q", transitionName)
		}
	}

	return e.applyTransition(ctx, wi, wf, *transition, triggeredBy, triggerType)
}

func (e *engine) applyTransition(ctx context.Context, wi *model.WorkflowInstance, wf *model.Workflow, t model.Transition, triggeredBy, triggerType string) (*TransitionResult, error) {
	fromState := wi.CurrentState
	wi.CurrentState = t.ToState
	wi.UpdatedAt = time.Now().UTC()

	terminal := isTerminal(wf.StateMachine, t.ToState)
	if terminal {
		wi.Status = "completed"
	}

	if err := e.workflowInstanceRepo.Update(ctx, wi); err != nil {
		return nil, fmt.Errorf("update workflow instance: %w", err)
	}

	st := &model.StateTransition{
		ID:                 uuid.New(),
		WorkflowInstanceID: wi.ID,
		FromState:          fromState,
		ToState:            t.ToState,
		TransitionName:     t.Name,
		TriggeredBy:        triggeredBy,
		TriggerType:        triggerType,
		CreatedAt:          time.Now().UTC(),
	}
	if err := e.stateTransitionRepo.Create(ctx, st); err != nil {
		return nil, fmt.Errorf("log state transition: %w", err)
	}

	e.logger.Info("state transition applied",
		slog.String("workflow_instance_id", wi.ID.String()),
		slog.String("from", fromState),
		slog.String("to", t.ToState),
		slog.String("transition", t.Name),
	)

	return &TransitionResult{
		WorkflowInstanceID: wi.ID,
		FromState:          fromState,
		ToState:            t.ToState,
		TransitionName:     t.Name,
		IsTerminal:         terminal,
	}, nil
}

func isTerminal(sm model.StateMachine, state string) bool {
	for _, ts := range sm.TerminalStates {
		if ts == state {
			return true
		}
	}
	return false
}

// buildTaskNameMap maps task definition name -> list of task instance statuses
func buildTaskNameMap(tasks []*model.Task, instances []*model.TaskInstance) map[string][]string {
	taskIDToName := make(map[uuid.UUID]string, len(tasks))
	for _, t := range tasks {
		taskIDToName[t.ID] = t.Name
	}

	result := make(map[string][]string)
	for _, ti := range instances {
		name := taskIDToName[ti.TaskID]
		if name != "" {
			result[name] = append(result[name], ti.Status)
		}
	}
	return result
}

func isCompletionRuleSatisfied(rule *model.CompletionRule, taskNameMap map[string][]string) bool {
	switch rule.Type {
	case "single":
		return hasCompleted(taskNameMap, rule.TaskRefs[0])
	case "all_of":
		for _, ref := range rule.TaskRefs {
			if !hasCompleted(taskNameMap, ref) {
				return false
			}
		}
		return true
	case "any_of":
		for _, ref := range rule.TaskRefs {
			if hasCompleted(taskNameMap, ref) {
				return true
			}
		}
		return false
	case "n_of_m":
		count := 0
		for _, ref := range rule.TaskRefs {
			if hasCompleted(taskNameMap, ref) {
				count++
			}
		}
		return count >= rule.N
	default:
		return false
	}
}

func hasCompleted(taskNameMap map[string][]string, taskRef string) bool {
	statuses, ok := taskNameMap[taskRef]
	if !ok {
		return false
	}
	for _, s := range statuses {
		if s == "completed" {
			return true
		}
	}
	return false
}
