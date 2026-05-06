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

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func setupProcessInstanceTest() (*ProcessInstanceService, *testutil.MockProcessRepo, *testutil.MockWorkflowRepo, *testutil.MockTaskRepo, *testutil.MockProcessInstanceRepo, *testutil.MockWorkflowInstanceRepo, *testutil.MockTaskInstanceRepo) {
	processRepo := testutil.NewMockProcessRepo()
	workflowRepo := testutil.NewMockWorkflowRepo()
	taskRepo := testutil.NewMockTaskRepo()
	piRepo := testutil.NewMockProcessInstanceRepo()
	wiRepo := testutil.NewMockWorkflowInstanceRepo()
	tiRepo := testutil.NewMockTaskInstanceRepo()

	svc := NewProcessInstanceService(processRepo, workflowRepo, taskRepo, piRepo, wiRepo, tiRepo, testLogger())
	return svc, processRepo, workflowRepo, taskRepo, piRepo, wiRepo, tiRepo
}

func createActiveProcess(repo *testutil.MockProcessRepo) uuid.UUID {
	now := time.Now().UTC()
	p := &model.Process{
		ID: uuid.New(), Name: "Active", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
	}
	repo.Processes[p.ID] = p
	return p.ID
}

func TestProcessInstanceService_Start(t *testing.T) {
	svc, processRepo, workflowRepo, taskRepo, _, wiRepo, tiRepo := setupProcessInstanceTest()
	processID := createActiveProcess(processRepo)

	// Add a workflow with 2 tasks
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "done"}},
		Transitions:    []model.Transition{{Name: "finish", FromState: "open", ToState: "done"}},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
	now := time.Now().UTC()
	wfID := uuid.New()
	workflowRepo.Workflows[wfID] = &model.Workflow{
		ID: wfID, ProcessID: processID, Name: "WF", StateMachine: sm,
		CreatedAt: now, UpdatedAt: now,
	}
	t1ID := uuid.New()
	t2ID := uuid.New()
	taskRepo.Tasks[t1ID] = &model.Task{ID: t1ID, WorkflowID: wfID, Name: "Task1", TaskType: "manual", CreatedAt: now, UpdatedAt: now}
	taskRepo.Tasks[t2ID] = &model.Task{ID: t2ID, WorkflowID: wfID, Name: "Task2", TaskType: "manual", CreatedAt: now, UpdatedAt: now}

	pi, err := svc.Start(context.Background(), StartProcessInput{ProcessID: processID})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if pi.Status != "running" {
		t.Errorf("Status = %q, want %q", pi.Status, "running")
	}

	// Should have created 1 workflow instance
	if len(wiRepo.Instances) != 1 {
		t.Fatalf("expected 1 workflow instance, got %d", len(wiRepo.Instances))
	}
	for _, wi := range wiRepo.Instances {
		if wi.CurrentState != "open" {
			t.Errorf("CurrentState = %q, want %q", wi.CurrentState, "open")
		}
	}

	// Should have created 2 task instances
	if len(tiRepo.Instances) != 2 {
		t.Fatalf("expected 2 task instances, got %d", len(tiRepo.Instances))
	}
}

func TestProcessInstanceService_Start_DraftProcess(t *testing.T) {
	svc, processRepo, _, _, _, _, _ := setupProcessInstanceTest()

	now := time.Now().UTC()
	p := &model.Process{
		ID: uuid.New(), Name: "Draft", Version: 1, Status: "draft",
		CreatedAt: now, UpdatedAt: now,
	}
	processRepo.Processes[p.ID] = p

	_, err := svc.Start(context.Background(), StartProcessInput{ProcessID: p.ID})
	if err == nil {
		t.Fatal("expected error for draft process")
	}
}

func TestProcessInstanceService_Start_ParentChild(t *testing.T) {
	svc, processRepo, _, _, piRepo, _, _ := setupProcessInstanceTest()
	processID := createActiveProcess(processRepo)

	parent, _ := svc.Start(context.Background(), StartProcessInput{ProcessID: processID})

	child, err := svc.Start(context.Background(), StartProcessInput{
		ProcessID: processID,
		ParentID:  &parent.ID,
	})
	if err != nil {
		t.Fatalf("Start child failed: %v", err)
	}
	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Error("child should reference parent")
	}

	// Verify parent has child
	children, _ := svc.ListChildren(context.Background(), parent.ID)
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	// Grandchild should fail
	_, err = svc.Start(context.Background(), StartProcessInput{
		ProcessID: processID,
		ParentID:  &child.ID,
	})
	if err == nil {
		t.Fatal("expected error for grandchild")
	}

	_ = piRepo // used implicitly
}

func TestProcessInstanceService_Start_ParentNotRunning(t *testing.T) {
	svc, processRepo, _, _, piRepo, _, _ := setupProcessInstanceTest()
	processID := createActiveProcess(processRepo)

	now := time.Now().UTC()
	parent := &model.ProcessInstance{
		ID: uuid.New(), ProcessID: processID, Status: "completed",
		StartedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	piRepo.Instances[parent.ID] = parent

	_, err := svc.Start(context.Background(), StartProcessInput{
		ProcessID: processID,
		ParentID:  &parent.ID,
	})
	if err == nil {
		t.Fatal("expected error for non-running parent")
	}
}

func TestProcessInstanceService_Cancel(t *testing.T) {
	svc, processRepo, _, _, _, wiRepo, _ := setupProcessInstanceTest()
	processID := createActiveProcess(processRepo)

	pi, _ := svc.Start(context.Background(), StartProcessInput{ProcessID: processID})

	if err := svc.Cancel(context.Background(), pi.ID); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	got, _ := svc.GetByID(context.Background(), pi.ID)
	if got.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", got.Status, "cancelled")
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	// All workflow instances should be cancelled
	for _, wi := range wiRepo.Instances {
		if wi.ProcessInstanceID == pi.ID && wi.Status != "cancelled" {
			t.Errorf("workflow instance status = %q, want %q", wi.Status, "cancelled")
		}
	}
}

func TestProcessInstanceService_Cancel_NotRunning(t *testing.T) {
	svc, processRepo, _, _, _, _, _ := setupProcessInstanceTest()
	processID := createActiveProcess(processRepo)

	pi, _ := svc.Start(context.Background(), StartProcessInput{ProcessID: processID})
	svc.Cancel(context.Background(), pi.ID)

	// Cancel again should fail
	err := svc.Cancel(context.Background(), pi.ID)
	if err == nil {
		t.Fatal("expected error for non-running instance")
	}
}
