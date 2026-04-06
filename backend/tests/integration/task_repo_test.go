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

func createTestWorkflow(t *testing.T) (*model.Process, *model.Workflow) {
	t.Helper()
	p := createTestProcess(t)
	repo := postgres.NewWorkflowRepo(testDB)
	now := time.Now().UTC().Truncate(time.Microsecond)
	w := &model.Workflow{
		ID:           uuid.New(),
		ProcessID:    p.ID,
		Name:         "Test Workflow",
		StateMachine: testStateMachine(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(context.Background(), w); err != nil {
		t.Fatalf("failed to create test workflow: %v", err)
	}
	return p, w
}

func TestTaskRepo_Create_And_GetByID(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	repo := postgres.NewTaskRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tk := &model.Task{
		ID:         uuid.New(),
		WorkflowID: w.ID,
		Name:       "Code Review",
		TaskType:   "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repo.Create(ctx, tk); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, tk.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != tk.Name {
		t.Errorf("Name = %q, want %q", got.Name, tk.Name)
	}
	if got.TaskType != "manual" {
		t.Errorf("TaskType = %q, want %q", got.TaskType, "manual")
	}
}

func TestTaskRepo_Create_ForeignKeyViolation(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewTaskRepo(testDB)
	ctx := context.Background()

	tk := &model.Task{
		ID:         uuid.New(),
		WorkflowID: uuid.New(),
		Name:       "Orphan",
		TaskType:   "manual",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, tk)
	if err != repository.ErrForeignKey {
		t.Fatalf("expected ErrForeignKey, got %v", err)
	}
}

func TestTaskRepo_ListByWorkflowID(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	repo := postgres.NewTaskRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 3; i++ {
		tk := &model.Task{
			ID:         uuid.New(),
			WorkflowID: w.ID,
			Name:       "Task",
			TaskType:   "manual",
			CreatedAt:  now.Add(time.Duration(i) * time.Second),
			UpdatedAt:  now,
		}
		repo.Create(ctx, tk)
	}

	results, err := repo.ListByWorkflowID(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListByWorkflowID failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
}

func TestTaskRepo_Update(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	repo := postgres.NewTaskRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tk := &model.Task{
		ID:         uuid.New(),
		WorkflowID: w.ID,
		Name:       "Original",
		TaskType:   "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	repo.Create(ctx, tk)

	tk.Name = "Updated"
	tk.TaskType = "approval"
	tk.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(ctx, tk); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := repo.GetByID(ctx, tk.ID)
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated")
	}
	if got.TaskType != "approval" {
		t.Errorf("TaskType = %q, want %q", got.TaskType, "approval")
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	cleanupTables(t)
	_, w := createTestWorkflow(t)
	repo := postgres.NewTaskRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tk := &model.Task{
		ID:         uuid.New(),
		WorkflowID: w.ID,
		Name:       "To Delete",
		TaskType:   "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	repo.Create(ctx, tk)

	if err := repo.Delete(ctx, tk.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err := repo.GetByID(ctx, tk.ID)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
