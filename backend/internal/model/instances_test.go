package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProcessInstance_Validate_Valid(t *testing.T) {
	pi := &ProcessInstance{
		ID:        uuid.New(),
		ProcessID: uuid.New(),
		Status:    "running",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := pi.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessInstance_Validate_AllStatuses(t *testing.T) {
	for _, status := range []string{"running", "completed", "cancelled", "failed"} {
		pi := &ProcessInstance{
			ID:        uuid.New(),
			ProcessID: uuid.New(),
			Status:    status,
			StartedAt: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := pi.Validate(); err != nil {
			t.Errorf("status %q should be valid, got %v", status, err)
		}
	}
}

func TestProcessInstance_Validate_NilProcessID(t *testing.T) {
	pi := &ProcessInstance{
		ID:     uuid.New(),
		Status: "running",
	}
	if err := pi.Validate(); err == nil {
		t.Fatal("expected error for nil process_id")
	}
}

func TestProcessInstance_Validate_InvalidStatus(t *testing.T) {
	pi := &ProcessInstance{
		ID:        uuid.New(),
		ProcessID: uuid.New(),
		Status:    "invalid",
	}
	if err := pi.Validate(); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestWorkflowInstance_Validate_Valid(t *testing.T) {
	wi := &WorkflowInstance{
		ID:                uuid.New(),
		WorkflowID:        uuid.New(),
		ProcessInstanceID: uuid.New(),
		CurrentState:      "open",
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := wi.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWorkflowInstance_Validate_AllStatuses(t *testing.T) {
	for _, status := range []string{"active", "completed", "cancelled"} {
		wi := &WorkflowInstance{
			ID:                uuid.New(),
			WorkflowID:        uuid.New(),
			ProcessInstanceID: uuid.New(),
			CurrentState:      "open",
			Status:            status,
		}
		if err := wi.Validate(); err != nil {
			t.Errorf("status %q should be valid, got %v", status, err)
		}
	}
}

func TestWorkflowInstance_Validate_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		wi   WorkflowInstance
	}{
		{"nil workflow_id", WorkflowInstance{ProcessInstanceID: uuid.New(), CurrentState: "open", Status: "active"}},
		{"nil process_instance_id", WorkflowInstance{WorkflowID: uuid.New(), CurrentState: "open", Status: "active"}},
		{"empty current_state", WorkflowInstance{WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(), Status: "active"}},
		{"invalid status", WorkflowInstance{WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(), CurrentState: "open", Status: "bad"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wi.Validate(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestTaskInstance_Validate_Valid(t *testing.T) {
	ti := &TaskInstance{
		ID:                 uuid.New(),
		TaskID:             uuid.New(),
		WorkflowInstanceID: uuid.New(),
		Status:             "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := ti.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTaskInstance_Validate_AllStatuses(t *testing.T) {
	for _, status := range []string{"pending", "in_progress", "completed", "skipped", "failed"} {
		ti := &TaskInstance{
			ID:                 uuid.New(),
			TaskID:             uuid.New(),
			WorkflowInstanceID: uuid.New(),
			Status:             status,
		}
		if err := ti.Validate(); err != nil {
			t.Errorf("status %q should be valid, got %v", status, err)
		}
	}
}

func TestTaskInstance_Validate_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		ti   TaskInstance
	}{
		{"nil task_id", TaskInstance{WorkflowInstanceID: uuid.New(), Status: "pending"}},
		{"nil workflow_instance_id", TaskInstance{TaskID: uuid.New(), Status: "pending"}},
		{"invalid status", TaskInstance{TaskID: uuid.New(), WorkflowInstanceID: uuid.New(), Status: "bad"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ti.Validate(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestStateTransition_Validate_Valid(t *testing.T) {
	st := &StateTransition{
		ID:                 uuid.New(),
		WorkflowInstanceID: uuid.New(),
		FromState:          "open",
		ToState:            "closed",
		TriggerType:        "human",
		CreatedAt:          time.Now(),
	}
	if err := st.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStateTransition_Validate_AllTriggerTypes(t *testing.T) {
	for _, tt := range []string{"human", "ai_agent", "rule", "process"} {
		st := &StateTransition{
			WorkflowInstanceID: uuid.New(),
			FromState:          "a",
			ToState:            "b",
			TriggerType:        tt,
		}
		if err := st.Validate(); err != nil {
			t.Errorf("trigger_type %q should be valid, got %v", tt, err)
		}
	}
}

func TestStateTransition_Validate_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		st   StateTransition
	}{
		{"nil workflow_instance_id", StateTransition{FromState: "a", ToState: "b", TriggerType: "human"}},
		{"empty from_state", StateTransition{WorkflowInstanceID: uuid.New(), ToState: "b", TriggerType: "human"}},
		{"empty to_state", StateTransition{WorkflowInstanceID: uuid.New(), FromState: "a", TriggerType: "human"}},
		{"invalid trigger_type", StateTransition{WorkflowInstanceID: uuid.New(), FromState: "a", ToState: "b", TriggerType: "bad"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.st.Validate(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
