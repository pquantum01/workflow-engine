package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/testutil"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestProcessService_Create(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	p, err := svc.Create(context.Background(), CreateProcessInput{
		Name:        "Test Process",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if p.Name != "Test Process" {
		t.Errorf("Name = %q, want %q", p.Name, "Test Process")
	}
	if p.Status != "draft" {
		t.Errorf("Status = %q, want %q", p.Status, "draft")
	}
	if p.Version != 1 {
		t.Errorf("Version = %d, want 1", p.Version)
	}
	if p.ID == uuid.Nil {
		t.Error("ID should be set")
	}
}

func TestProcessService_Create_EmptyName(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	_, err := svc.Create(context.Background(), CreateProcessInput{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProcessService_Update(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	p, _ := svc.Create(context.Background(), CreateProcessInput{Name: "Original"})

	newName := "Updated"
	updated, err := svc.Update(context.Background(), p.ID, UpdateProcessInput{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Name = %q, want %q", updated.Name, "Updated")
	}
}

func TestProcessService_Update_ArchivedProcess(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	now := time.Now().UTC()
	p := &model.Process{
		ID: uuid.New(), Name: "Archived", Version: 1, Status: "archived",
		CreatedAt: now, UpdatedAt: now,
	}
	repo.Processes[p.ID] = p

	newName := "Changed"
	_, err := svc.Update(context.Background(), p.ID, UpdateProcessInput{Name: &newName})
	if err == nil {
		t.Fatal("expected error for archived process")
	}
}

func TestProcessService_Delete_Draft(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	p, _ := svc.Create(context.Background(), CreateProcessInput{Name: "Draft"})

	if err := svc.Delete(context.Background(), p.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := svc.GetByID(context.Background(), p.ID)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessService_Delete_ActiveProcess(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	now := time.Now().UTC()
	p := &model.Process{
		ID: uuid.New(), Name: "Active", Version: 1, Status: "active",
		CreatedAt: now, UpdatedAt: now,
	}
	repo.Processes[p.ID] = p

	err := svc.Delete(context.Background(), p.ID)
	if err == nil {
		t.Fatal("expected error for active process")
	}
}

func TestProcessService_List_DefaultLimit(t *testing.T) {
	repo := testutil.NewMockProcessRepo()
	svc := NewProcessService(repo, quietLogger())

	for i := 0; i < 5; i++ {
		svc.Create(context.Background(), CreateProcessInput{Name: "P"})
	}

	results, err := svc.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("expected 5, got %d", len(results))
	}
}
