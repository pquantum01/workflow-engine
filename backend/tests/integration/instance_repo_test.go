//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/repository/postgres"
)

func TestProcessInstanceRepo_Create_And_GetByID(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: p.ID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.Create(ctx, pi); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, pi.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != "running" {
		t.Errorf("Status = %q, want %q", got.Status, "running")
	}
	if got.ParentID != nil {
		t.Errorf("ParentID should be nil")
	}
}

func TestProcessInstanceRepo_ParentChild(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	parent := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: p.ID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.Create(ctx, parent)

	child := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: p.ID,
		ParentID:  &parent.ID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.Create(ctx, child); err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	children, err := repo.ListByParentID(ctx, parent.ID)
	if err != nil {
		t.Fatalf("ListByParentID failed: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].ID != child.ID {
		t.Errorf("child ID mismatch")
	}
}

func TestProcessInstanceRepo_ForeignKeyViolation(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: uuid.New(), // nonexistent
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.Create(ctx, pi)
	if err != repository.ErrForeignKey {
		t.Fatalf("expected ErrForeignKey, got %v", err)
	}
}

func TestProcessInstanceRepo_Update(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: p.ID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.Create(ctx, pi)

	completedAt := time.Now().UTC().Truncate(time.Microsecond)
	pi.Status = "completed"
	pi.CompletedAt = &completedAt
	pi.UpdatedAt = completedAt

	if err := repo.Update(ctx, pi); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := repo.GetByID(ctx, pi.ID)
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestProcessInstanceRepo_List(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 3; i++ {
		pi := &model.ProcessInstance{
			ID:        uuid.New(),
			ProcessID: p.ID,
			Status:    "running",
			StartedAt: now,
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now,
		}
		repo.Create(ctx, pi)
	}

	results, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
}

func TestWorkflowInstanceRepo_Create_And_GetByID(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	piRepo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: w.ProcessID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	piRepo.Create(ctx, pi)

	wiRepo := postgres.NewWorkflowInstanceRepo(testDB)
	wi := &model.WorkflowInstance{
		ID:                uuid.New(),
		WorkflowID:        w.ID,
		ProcessInstanceID: pi.ID,
		CurrentState:      "open",
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := wiRepo.Create(ctx, wi); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := wiRepo.GetByID(ctx, wi.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.CurrentState != "open" {
		t.Errorf("CurrentState = %q, want %q", got.CurrentState, "open")
	}
}

func TestWorkflowInstanceRepo_ListByProcessInstanceID(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	piRepo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: w.ProcessID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	piRepo.Create(ctx, pi)

	wiRepo := postgres.NewWorkflowInstanceRepo(testDB)
	for i := 0; i < 2; i++ {
		wi := &model.WorkflowInstance{
			ID:                uuid.New(),
			WorkflowID:        w.ID,
			ProcessInstanceID: pi.ID,
			CurrentState:      "open",
			Status:            "active",
			CreatedAt:         now.Add(time.Duration(i) * time.Second),
			UpdatedAt:         now,
		}
		wiRepo.Create(ctx, wi)
	}

	results, err := wiRepo.ListByProcessInstanceID(ctx, pi.ID)
	if err != nil {
		t.Fatalf("ListByProcessInstanceID failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestWorkflowInstanceRepo_Update(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	piRepo := postgres.NewProcessInstanceRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: w.ProcessID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	piRepo.Create(ctx, pi)

	wiRepo := postgres.NewWorkflowInstanceRepo(testDB)
	wi := &model.WorkflowInstance{
		ID:                uuid.New(),
		WorkflowID:        w.ID,
		ProcessInstanceID: pi.ID,
		CurrentState:      "open",
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	wiRepo.Create(ctx, wi)

	wi.CurrentState = "closed"
	wi.Status = "completed"
	wi.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	if err := wiRepo.Update(ctx, wi); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := wiRepo.GetByID(ctx, wi.ID)
	if got.CurrentState != "closed" {
		t.Errorf("CurrentState = %q, want %q", got.CurrentState, "closed")
	}
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
}

func TestTaskInstanceRepo_CRUD(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)

	taskRepo := postgres.NewTaskRepo(testDB)
	piRepo := postgres.NewProcessInstanceRepo(testDB)
	wiRepo := postgres.NewWorkflowInstanceRepo(testDB)
	tiRepo := postgres.NewTaskInstanceRepo(testDB)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create a task definition
	tk := &model.Task{
		ID:         uuid.New(),
		WorkflowID: w.ID,
		Name:       "Review",
		TaskType:   "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	taskRepo.Create(ctx, tk)

	// Create process instance
	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: w.ProcessID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	piRepo.Create(ctx, pi)

	// Create workflow instance
	wi := &model.WorkflowInstance{
		ID:                uuid.New(),
		WorkflowID:        w.ID,
		ProcessInstanceID: pi.ID,
		CurrentState:      "open",
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	wiRepo.Create(ctx, wi)

	// Create task instance
	ti := &model.TaskInstance{
		ID:                 uuid.New(),
		TaskID:             tk.ID,
		WorkflowInstanceID: wi.ID,
		Status:             "pending",
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := tiRepo.Create(ctx, ti); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := tiRepo.GetByID(ctx, ti.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != "pending" {
		t.Errorf("Status = %q, want %q", got.Status, "pending")
	}

	// Update
	completedAt := time.Now().UTC().Truncate(time.Microsecond)
	ti.Status = "completed"
	ti.AssignedTo = "user-123"
	ti.CompletedAt = &completedAt
	ti.UpdatedAt = completedAt
	if err := tiRepo.Update(ctx, ti); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ = tiRepo.GetByID(ctx, ti.ID)
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
	if got.AssignedTo != "user-123" {
		t.Errorf("AssignedTo = %q, want %q", got.AssignedTo, "user-123")
	}

	// List
	results, err := tiRepo.ListByWorkflowInstanceID(ctx, wi.ID)
	if err != nil {
		t.Fatalf("ListByWorkflowInstanceID failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
}

func TestStateTransitionRepo_Create_And_List(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)

	piRepo := postgres.NewProcessInstanceRepo(testDB)
	wiRepo := postgres.NewWorkflowInstanceRepo(testDB)
	stRepo := postgres.NewStateTransitionRepo(testDB)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	pi := &model.ProcessInstance{
		ID:        uuid.New(),
		ProcessID: w.ProcessID,
		Status:    "running",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	piRepo.Create(ctx, pi)

	wi := &model.WorkflowInstance{
		ID:                uuid.New(),
		WorkflowID:        w.ID,
		ProcessInstanceID: pi.ID,
		CurrentState:      "open",
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	wiRepo.Create(ctx, wi)

	st := &model.StateTransition{
		ID:                 uuid.New(),
		WorkflowInstanceID: wi.ID,
		FromState:          "open",
		ToState:            "closed",
		TransitionName:     "close",
		TriggeredBy:        "user-1",
		TriggerType:        "human",
		CreatedAt:          now,
	}

	if err := stRepo.Create(ctx, st); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	results, err := stRepo.ListByWorkflowInstanceID(ctx, wi.ID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
	if results[0].FromState != "open" {
		t.Errorf("FromState = %q, want %q", results[0].FromState, "open")
	}
	if results[0].TriggerType != "human" {
		t.Errorf("TriggerType = %q, want %q", results[0].TriggerType, "human")
	}
}
