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

func TestProcessRepo_Create_And_GetByID(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	p := &model.Process{
		ID:          uuid.New(),
		Name:        "Test Process",
		Description: "A test process",
		Version:     1,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if got.Status != p.Status {
		t.Errorf("Status = %q, want %q", got.Status, p.Status)
	}
	if got.Version != p.Version {
		t.Errorf("Version = %d, want %d", got.Version, p.Version)
	}
}

func TestProcessRepo_GetByID_NotFound(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessRepo_List(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 3; i++ {
		p := &model.Process{
			ID:        uuid.New(),
			Name:      "Process " + string(rune('A'+i)),
			Version:   1,
			Status:    "draft",
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now,
		}
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	results, err := repo.List(ctx, 0, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestProcessRepo_List_Pagination(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 5; i++ {
		p := &model.Process{
			ID:        uuid.New(),
			Name:      "Process",
			Version:   1,
			Status:    "draft",
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now,
		}
		repo.Create(ctx, p)
	}

	page1, _ := repo.List(ctx, 0, 2)
	page2, _ := repo.List(ctx, 2, 2)

	if len(page1) != 2 {
		t.Errorf("page1 expected 2, got %d", len(page1))
	}
	if len(page2) != 2 {
		t.Errorf("page2 expected 2, got %d", len(page2))
	}
}

func TestProcessRepo_Update(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	p := &model.Process{
		ID:        uuid.New(),
		Name:      "Original",
		Version:   1,
		Status:    "draft",
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.Create(ctx, p)

	p.Name = "Updated"
	p.Status = "active"
	p.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := repo.GetByID(ctx, p.ID)
	if got.Name != "Updated" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated")
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
}

func TestProcessRepo_Update_NotFound(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	p := &model.Process{
		ID:        uuid.New(),
		Name:      "Ghost",
		Version:   1,
		Status:    "draft",
		UpdatedAt: time.Now(),
	}
	err := repo.Update(ctx, p)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessRepo_Delete(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	p := &model.Process{
		ID:        uuid.New(),
		Name:      "To Delete",
		Version:   1,
		Status:    "draft",
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.Create(ctx, p)

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := repo.GetByID(ctx, p.ID)
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestProcessRepo_Delete_NotFound(t *testing.T) {
	cleanupTables(t)
	repo := postgres.NewProcessRepo(testDB)
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	if err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
