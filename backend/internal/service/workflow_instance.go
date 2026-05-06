package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type TriggerTransitionInput struct {
	TransitionName string `json:"transition_name"`
	TriggeredBy    string `json:"triggered_by"`
	TriggerType    string `json:"trigger_type"`
}

type WorkflowInstanceService struct {
	workflowInstanceRepo  repository.WorkflowInstanceRepository
	stateTransitionRepo   repository.StateTransitionRepository
	engine                engine.Engine
	logger                *slog.Logger
}

func NewWorkflowInstanceService(
	workflowInstanceRepo repository.WorkflowInstanceRepository,
	stateTransitionRepo repository.StateTransitionRepository,
	eng engine.Engine,
	logger *slog.Logger,
) *WorkflowInstanceService {
	return &WorkflowInstanceService{
		workflowInstanceRepo:  workflowInstanceRepo,
		stateTransitionRepo:   stateTransitionRepo,
		engine:                eng,
		logger:                logger,
	}
}

func (s *WorkflowInstanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	s.logger.Debug("getting workflow instance", slog.String("id", id.String()))

	wi, err := s.workflowInstanceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get workflow instance", slog.String("id", id.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflow instance retrieved", slog.String("id", id.String()), slog.String("status", wi.Status))
	return wi, nil
}

func (s *WorkflowInstanceService) ListByProcessInstanceID(ctx context.Context, processInstanceID uuid.UUID) ([]*model.WorkflowInstance, error) {
	s.logger.Debug("listing workflow instances", slog.String("process_instance_id", processInstanceID.String()))

	results, err := s.workflowInstanceRepo.ListByProcessInstanceID(ctx, processInstanceID)
	if err != nil {
		s.logger.Error("failed to list workflow instances", slog.String("process_instance_id", processInstanceID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("workflow instances listed", slog.String("process_instance_id", processInstanceID.String()), slog.Int("count", len(results)))
	return results, nil
}

func (s *WorkflowInstanceService) TriggerTransition(ctx context.Context, id uuid.UUID, input TriggerTransitionInput) (*engine.TransitionResult, error) {
	s.logger.Debug("triggering workflow transition",
		slog.String("workflow_instance_id", id.String()),
		slog.String("transition_name", input.TransitionName),
		slog.String("triggered_by", input.TriggeredBy),
		slog.String("trigger_type", input.TriggerType),
	)

	result, err := s.engine.TriggerTransition(ctx, id, input.TransitionName, input.TriggeredBy, input.TriggerType)
	if err != nil {
		s.logger.Error("failed to trigger workflow transition",
			slog.String("workflow_instance_id", id.String()),
			slog.String("transition_name", input.TransitionName),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("workflow transition triggered",
		slog.String("workflow_instance_id", id.String()),
		slog.String("transition_name", input.TransitionName),
		slog.String("from_state", result.FromState),
		slog.String("to_state", result.ToState),
	)
	return result, nil
}

func (s *WorkflowInstanceService) ListTransitions(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.StateTransition, error) {
	s.logger.Debug("listing state transitions", slog.String("workflow_instance_id", workflowInstanceID.String()))

	results, err := s.stateTransitionRepo.ListByWorkflowInstanceID(ctx, workflowInstanceID)
	if err != nil {
		s.logger.Error("failed to list state transitions", slog.String("workflow_instance_id", workflowInstanceID.String()), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("state transitions listed", slog.String("workflow_instance_id", workflowInstanceID.String()), slog.Int("count", len(results)))
	return results, nil
}
