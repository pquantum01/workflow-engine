package model

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validTask() *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New(),
		WorkflowID:  uuid.New(),
		Name:        "Code Review",
		Description: "Review the pull request",
		TaskType:    "manual",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestTask_Validate_Valid(t *testing.T) {
	tk := validTask()
	if err := tk.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTask_Validate_AllTypes(t *testing.T) {
	for _, tt := range []string{"manual", "automated", "approval"} {
		tk := validTask()
		tk.TaskType = tt
		if err := tk.Validate(); err != nil {
			t.Errorf("task_type %q should be valid, got %v", tt, err)
		}
	}
}

func TestTask_Validate_EmptyName(t *testing.T) {
	tk := validTask()
	tk.Name = ""
	err := tk.Validate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTask_Validate_LongName(t *testing.T) {
	tk := validTask()
	tk.Name = strings.Repeat("a", 256)
	err := tk.Validate()
	if err == nil {
		t.Fatal("expected error for long name")
	}
}

func TestTask_Validate_NilWorkflowID(t *testing.T) {
	tk := validTask()
	tk.WorkflowID = uuid.Nil
	err := tk.Validate()
	if err == nil {
		t.Fatal("expected error for nil workflow_id")
	}
}

func TestTask_Validate_InvalidType(t *testing.T) {
	tk := validTask()
	tk.TaskType = "invalid"
	err := tk.Validate()
	if err == nil {
		t.Fatal("expected error for invalid task_type")
	}
}
