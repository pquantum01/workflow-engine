package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/testutil"
)

func validSM() model.StateMachine {
	return model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
}

func setupWorkflowTest() (*WorkflowService, *testutil.MockProcessRepo, *testutil.MockWorkflowRepo, uuid.UUID) {
	processRepo := testutil.NewMockProcessRepo()
	workflowRepo := testutil.NewMockWorkflowRepo()
	svc := NewWorkflowService(workflowRepo, processRepo, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))

	now := time.Now().UTC()
	p := &model.Process{
		ID: uuid.New(), Name: "Test", Version: 1, Status: "draft",
		CreatedAt: now, UpdatedAt: now,
	}
	processRepo.Processes[p.ID] = p
	return svc, processRepo, workflowRepo, p.ID
}

func TestWorkflowService_Create(t *testing.T) {
	svc, _, _, processID := setupWorkflowTest()

	w, err := svc.Create(context.Background(), processID, CreateWorkflowInput{
		Name:         "Workflow",
		StateMachine: validSM(),
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if w.ProcessID != processID {
		t.Errorf("ProcessID mismatch")
	}
	if w.StateMachine.InitialState != "open" {
		t.Errorf("InitialState = %q, want %q", w.StateMachine.InitialState, "open")
	}
}

func TestWorkflowService_Create_ArchivedProcess(t *testing.T) {
	svc, processRepo, _, _ := setupWorkflowTest()

	now := time.Now().UTC()
	archived := &model.Process{
		ID: uuid.New(), Name: "Archived", Version: 1, Status: "archived",
		CreatedAt: now, UpdatedAt: now,
	}
	processRepo.Processes[archived.ID] = archived

	_, err := svc.Create(context.Background(), archived.ID, CreateWorkflowInput{
		Name:         "WF",
		StateMachine: validSM(),
	})
	if err == nil {
		t.Fatal("expected error for archived process")
	}
}

func TestWorkflowService_Create_InvalidStateMachine(t *testing.T) {
	svc, _, _, processID := setupWorkflowTest()

	_, err := svc.Create(context.Background(), processID, CreateWorkflowInput{
		Name:         "Bad SM",
		StateMachine: model.StateMachine{}, // no states
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWorkflowService_Delete_DraftProcess(t *testing.T) {
	svc, _, _, processID := setupWorkflowTest()

	w, _ := svc.Create(context.Background(), processID, CreateWorkflowInput{
		Name:         "WF",
		StateMachine: validSM(),
	})

	if err := svc.Delete(context.Background(), w.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestWorkflowService_Delete_ActiveProcess(t *testing.T) {
	svc, processRepo, _, processID := setupWorkflowTest()
	processRepo.Processes[processID].Status = "active"

	w, _ := svc.Create(context.Background(), processID, CreateWorkflowInput{
		Name:         "WF",
		StateMachine: validSM(),
	})

	err := svc.Delete(context.Background(), w.ID)
	if err == nil {
		t.Fatal("expected error for active process")
	}
}

func TestWorkflowService_Update(t *testing.T) {
	svc, _, _, processID := setupWorkflowTest()

	w, _ := svc.Create(context.Background(), processID, CreateWorkflowInput{
		Name:         "Original",
		StateMachine: validSM(),
	})

	newName := "Updated"
	updated, err := svc.Update(context.Background(), w.ID, UpdateWorkflowInput{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Name = %q, want %q", updated.Name, "Updated")
	}
}
