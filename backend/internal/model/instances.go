package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var validProcessInstanceStatuses = map[string]bool{
	"running":   true,
	"completed": true,
	"cancelled": true,
	"failed":    true,
}

var validWorkflowInstanceStatuses = map[string]bool{
	"active":    true,
	"completed": true,
	"cancelled": true,
}

var validTaskInstanceStatuses = map[string]bool{
	"pending":     true,
	"in_progress": true,
	"completed":   true,
	"skipped":     true,
	"failed":      true,
}

var validTriggerTypes = map[string]bool{
	"human":    true,
	"ai_agent": true,
	"rule":     true,
	"process":  true,
}

type ProcessInstance struct {
	ID          uuid.UUID  `json:"id"`
	ProcessID   uuid.UUID  `json:"process_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (pi *ProcessInstance) Validate() error {
	if pi.ProcessID == uuid.Nil {
		return fmt.Errorf("process_id is required")
	}
	if !validProcessInstanceStatuses[pi.Status] {
		return fmt.Errorf("status must be one of: running, completed, cancelled, failed")
	}
	return nil
}

type WorkflowInstance struct {
	ID                uuid.UUID `json:"id"`
	WorkflowID        uuid.UUID `json:"workflow_id"`
	ProcessInstanceID uuid.UUID `json:"process_instance_id"`
	CurrentState      string    `json:"current_state"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (wi *WorkflowInstance) Validate() error {
	if wi.WorkflowID == uuid.Nil {
		return fmt.Errorf("workflow_id is required")
	}
	if wi.ProcessInstanceID == uuid.Nil {
		return fmt.Errorf("process_instance_id is required")
	}
	if wi.CurrentState == "" {
		return fmt.Errorf("current_state is required")
	}
	if !validWorkflowInstanceStatuses[wi.Status] {
		return fmt.Errorf("status must be one of: active, completed, cancelled")
	}
	return nil
}

type TaskInstance struct {
	ID                 uuid.UUID  `json:"id"`
	TaskID             uuid.UUID  `json:"task_id"`
	WorkflowInstanceID uuid.UUID  `json:"workflow_instance_id"`
	Status             string     `json:"status"`
	AssignedTo         string     `json:"assigned_to,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (ti *TaskInstance) Validate() error {
	if ti.TaskID == uuid.Nil {
		return fmt.Errorf("task_id is required")
	}
	if ti.WorkflowInstanceID == uuid.Nil {
		return fmt.Errorf("workflow_instance_id is required")
	}
	if !validTaskInstanceStatuses[ti.Status] {
		return fmt.Errorf("status must be one of: pending, in_progress, completed, skipped, failed")
	}
	return nil
}

type StateTransition struct {
	ID                 uuid.UUID `json:"id"`
	WorkflowInstanceID uuid.UUID `json:"workflow_instance_id"`
	FromState          string    `json:"from_state"`
	ToState            string    `json:"to_state"`
	TransitionName     string    `json:"transition_name,omitempty"`
	TriggeredBy        string    `json:"triggered_by,omitempty"`
	TriggerType        string    `json:"trigger_type"`
	CreatedAt          time.Time `json:"created_at"`
}

func (st *StateTransition) Validate() error {
	if st.WorkflowInstanceID == uuid.Nil {
		return fmt.Errorf("workflow_instance_id is required")
	}
	if st.FromState == "" {
		return fmt.Errorf("from_state is required")
	}
	if st.ToState == "" {
		return fmt.Errorf("to_state is required")
	}
	if !validTriggerTypes[st.TriggerType] {
		return fmt.Errorf("trigger_type must be one of: human, ai_agent, rule, process")
	}
	return nil
}
