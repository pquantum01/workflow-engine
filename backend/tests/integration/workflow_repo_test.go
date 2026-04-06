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

func createTestProcess(t *testing.T) *model.Process {
	t.Helper()
	repo := postgres.NewProcessRepo(testDB)
	now := time.Now().UTC().Truncate(time.Microsecond)
	p := &model.Process{
		ID:        uuid.New(),
		Name:      "Test Process",
		Version:   1,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.Create(context.Background(), p); err != nil {
		t.Fatalf("failed to create test process: %v", err)
	}
	return p
}

func testStateMachine() model.StateMachine {
	return model.StateMachine{
		States: []model.State{
			{Name: "open"},
			{Name: "closed"},
		},
		Transitions: []model.Transition{
			{Name: "close", FromState: "open", ToState: "closed"},
		},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
}

func TestWorkflowRepo_Create_And_GetByID(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    p.ID,
		Name:         "Test Workflow",
		Description:  "desc",
		StateMachine: testStateMachine(),
		SortOrder:    1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Create(ctx, w); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.Name != w.Name {
		t.Errorf("Name = %q, want %q", got.Name, w.Name)
	}
	if got.StateMachine.InitialState != "open" {
		t.Errorf("InitialState = %q, want %q", got.StateMachine.InitialState, "open")
	}
	if len(got.StateMachine.States) != 2 {
		t.Errorf("States count = %d, want 2", len(got.StateMachine.States))
	}
}

func TestWorkflowRepo_Create_ForeignKeyViolation(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    uuid.New(), // nonexistent
		Name:         "Orphan",
		StateMachine: testStateMachine(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, w)
	if err != repository.ErrForeignKey {
		t.Fatalf("expected ErrForeignKey, got %v", err)
	}
}

func TestWorkflowRepo_ListByProcessID(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 3; i++ {
		w := &model.Workflow{
			ID:           uuid.New(),
			ProcessID:    p.ID,
			Name:         "Workflow",
			StateMachine: testStateMachine(),
			SortOrder:    i,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		repo.Create(ctx, w)
	}

	results, err := repo.ListByProcessID(ctx, p.ID)
	if err != nil {
		t.Fatalf("ListByProcessID failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
}

func TestWorkflowRepo_ListByProcessID_Empty(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	results, err := repo.ListByProcessID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("ListByProcessID failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0, got %d", len(results))
	}
}

func TestWorkflowRepo_Update(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    p.ID,
		Name:         "Original",
		StateMachine: testStateMachine(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	repo.Create(ctx, w)

	w.Name = "Updated"
	w.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(ctx, w); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := repo.GetByID(ctx, w.ID)
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated")
	}
}

func TestWorkflowRepo_Delete(t *testing.T) {
	cleanupTables(t)
	p := createTestProcess(t)
	repo := postgres.NewWorkflowRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    p.ID,
		Name:         "To Delete",
		StateMachine: testStateMachine(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	repo.Create(ctx, w)

	if err := repo.Delete(ctx, w.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(ctx, w.ID)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkflowRepo_GetByID_NotFound(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewWorkflowRepo(testDB)
	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
