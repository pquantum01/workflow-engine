package dto

import (
	"time"
)

type StartInstanceRequest struct {
	WorkflowDefID  string         `json:"workflow_def_id" validate:"required,uuid"`
	CreatedBy      string         `json:"created_by"      validate:"required"`
	InitialContext map[string]any `json:"initial_context"`
}

type WorkflowInstanceResponse struct {
	ID             string                     `json:"id"`
	WorkflowDefID  string                     `json:"workflow_def_id"`
	CurrentStateID string                     `json:"current_state_id"`
	Status         string                     `json:"status"`
	CreatedBy      string                     `json:"created_by"`
	Context        map[string]any             `json:"context"`
	StartedAt      time.Time                  `json:"started_at"`
	CompletedAt    *time.Time                 `json:"completed_at,omitempty"`
	Tasks          []*TaskInstanceResponse    `json:"tasks"`
	History        []*StateTransitionResponse `json:"history"`
}

type InstanceListResponse struct {
	Items  []*WorkflowInstanceResponse `json:"items"`
	Total  int                         `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
}

type StateTransitionResponse struct {
	ID             string    `json:"id"`
	FromState      string    `json:"from_state"`
	ToState        string    `json:"to_state"`
	TransitionedAt time.Time `json:"transitioned_at"`
}
