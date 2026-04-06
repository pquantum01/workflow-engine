package model

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validWorkflow() *Workflow {
	now := time.Now()
	return &Workflow{
		ID:           uuid.New(),
		ProcessID:    uuid.New(),
		Name:         "Bug Fix Workflow",
		Description:  "Workflow for fixing bugs",
		StateMachine: validStateMachine(),
		SortOrder:    0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestWorkflow_Validate_Valid(t *testing.T) {
	w := validWorkflow()
	if err := w.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWorkflow_Validate_EmptyName(t *testing.T) {
	w := validWorkflow()
	w.Name = ""
	err := w.Validate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWorkflow_Validate_LongName(t *testing.T) {
	w := validWorkflow()
	w.Name = strings.Repeat("a", 256)
	err := w.Validate()
	if err == nil {
		t.Fatal("expected error for long name")
	}
}

func TestWorkflow_Validate_NilProcessID(t *testing.T) {
	w := validWorkflow()
	w.ProcessID = uuid.Nil
	err := w.Validate()
	if err == nil {
		t.Fatal("expected error for nil process_id")
	}
}

func TestWorkflow_Validate_InvalidStateMachine(t *testing.T) {
	w := validWorkflow()
	w.StateMachine.States = nil
	err := w.Validate()
	if err == nil {
		t.Fatal("expected error for invalid state machine")
	}
}
