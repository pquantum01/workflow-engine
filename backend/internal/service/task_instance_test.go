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

func setupTaskInstanceTest() (*TaskInstanceService, *testutil.MockTaskInstanceRepo, *testutil.MockWorkflowInstanceRepo, *testutil.MockEngine) {
	tiRepo := testutil.NewMockTaskInstanceRepo()
	wiRepo := testutil.NewMockWorkflowInstanceRepo()
	eng := &testutil.MockEngine{}
	svc := NewTaskInstanceService(tiRepo, wiRepo, eng, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))
	return svc, tiRepo, wiRepo, eng
}

func TestTaskInstanceService_UpdateStatus_Complete(t *testing.T) {
	svc, tiRepo, wiRepo, eng := setupTaskInstanceTest()
	now := time.Now().UTC()

	wiID := uuid.New()
	wiRepo.Instances[wiID] = &model.WorkflowInstance{
		ID: wiID, WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(),
		CurrentState: "open", Status: "active", CreatedAt: now, UpdatedAt: now,
	}

	tiID := uuid.New()
	tiRepo.Instances[tiID] = &model.TaskInstance{
		ID: tiID, TaskID: uuid.New(), WorkflowInstanceID: wiID,
		Status: "in_progress", CreatedAt: now, UpdatedAt: now,
	}

	ti, err := svc.UpdateStatus(context.Background(), tiID, UpdateTaskStatusInput{
		Status:      "completed",
		TriggeredBy: "user-1",
		TriggerType: "human",
	})
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
	if ti.Status != "completed" {
		t.Errorf("Status = %q, want %q", ti.Status, "completed")
	}
	if ti.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	// Engine should have been called
	if eng.EvaluateCalls != 1 {
		t.Errorf("EvaluateCalls = %d, want 1", eng.EvaluateCalls)
	}
}

func TestTaskInstanceService_UpdateStatus_InvalidTransition(t *testing.T) {
	svc, tiRepo, wiRepo, _ := setupTaskInstanceTest()
	now := time.Now().UTC()

	wiID := uuid.New()
	wiRepo.Instances[wiID] = &model.WorkflowInstance{
		ID: wiID, WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(),
		CurrentState: "open", Status: "active", CreatedAt: now, UpdatedAt: now,
	}

	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
	}{
		{"pending to completed", "pending", "completed"},
		{"pending to failed", "pending", "failed"},
		{"completed to in_progress", "completed", "in_progress"},
		{"in_progress to pending", "in_progress", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tiID := uuid.New()
			tiRepo.Instances[tiID] = &model.TaskInstance{
				ID: tiID, TaskID: uuid.New(), WorkflowInstanceID: wiID,
				Status: tt.fromStatus, CreatedAt: now, UpdatedAt: now,
			}
			_, err := svc.UpdateStatus(context.Background(), tiID, UpdateTaskStatusInput{
				Status: tt.toStatus,
			})
			if err == nil {
				t.Fatalf("expected error for %s -> %s", tt.fromStatus, tt.toStatus)
			}
		})
	}
}

func TestTaskInstanceService_UpdateStatus_InactiveWorkflow(t *testing.T) {
	svc, tiRepo, wiRepo, _ := setupTaskInstanceTest()
	now := time.Now().UTC()

	wiID := uuid.New()
	wiRepo.Instances[wiID] = &model.WorkflowInstance{
		ID: wiID, WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(),
		CurrentState: "done", Status: "completed", CreatedAt: now, UpdatedAt: now,
	}

	tiID := uuid.New()
	tiRepo.Instances[tiID] = &model.TaskInstance{
		ID: tiID, TaskID: uuid.New(), WorkflowInstanceID: wiID,
		Status: "in_progress", CreatedAt: now, UpdatedAt: now,
	}

	_, err := svc.UpdateStatus(context.Background(), tiID, UpdateTaskStatusInput{
		Status: "completed",
	})
	if err == nil {
		t.Fatal("expected error for inactive workflow")
	}
}

func TestTaskInstanceService_Assign(t *testing.T) {
	svc, tiRepo, _, _ := setupTaskInstanceTest()
	now := time.Now().UTC()

	tiID := uuid.New()
	tiRepo.Instances[tiID] = &model.TaskInstance{
		ID: tiID, TaskID: uuid.New(), WorkflowInstanceID: uuid.New(),
		Status: "pending", CreatedAt: now, UpdatedAt: now,
	}

	ti, err := svc.Assign(context.Background(), tiID, "user-42")
	if err != nil {
		t.Fatalf("Assign failed: %v", err)
	}
	if ti.AssignedTo != "user-42" {
		t.Errorf("AssignedTo = %q, want %q", ti.AssignedTo, "user-42")
	}
}

func TestTaskInstanceService_UpdateStatus_PendingToInProgress(t *testing.T) {
	svc, tiRepo, wiRepo, eng := setupTaskInstanceTest()
	now := time.Now().UTC()

	wiID := uuid.New()
	wiRepo.Instances[wiID] = &model.WorkflowInstance{
		ID: wiID, WorkflowID: uuid.New(), ProcessInstanceID: uuid.New(),
		CurrentState: "open", Status: "active", CreatedAt: now, UpdatedAt: now,
	}

	tiID := uuid.New()
	tiRepo.Instances[tiID] = &model.TaskInstance{
		ID: tiID, TaskID: uuid.New(), WorkflowInstanceID: wiID,
		Status: "pending", CreatedAt: now, UpdatedAt: now,
	}

	ti, err := svc.UpdateStatus(context.Background(), tiID, UpdateTaskStatusInput{
		Status: "in_progress",
	})
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
	if ti.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", ti.Status, "in_progress")
	}
	// Engine should NOT be called for in_progress
	if eng.EvaluateCalls != 0 {
		t.Errorf("EvaluateCalls = %d, want 0", eng.EvaluateCalls)
	}
}
