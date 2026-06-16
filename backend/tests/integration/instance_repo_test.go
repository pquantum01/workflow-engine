package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

func setupDefAndInstance(t *testing.T) (*model.WorkflowDef, *model.WorkflowInstance) {
	t.Helper()
	ctx := context.Background()
	defRepo := repository.NewWorkflowDefRepo(testDB)
	instRepo := repository.NewWorkflowInstanceRepo(testDB)

	def := makeTestDef()
	require.NoError(t, defRepo.Create(ctx, def))

	inst := &model.WorkflowInstance{
		ID:             uuid.New(),
		WorkflowDefID:  def.ID,
		CurrentStateID: "submitted",
		Status:         model.InstanceStatusActive,
		Context:        map[string]any{},
		StartedAt:      time.Now().UTC(),
	}
	require.NoError(t, instRepo.Create(ctx, inst))
	return def, inst
}

func TestWorkflowInstanceRepo_CreateAndGetByID(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	got, err := repo.GetByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, inst.ID, got.ID)
	assert.Equal(t, inst.CurrentStateID, got.CurrentStateID)
	assert.Equal(t, model.InstanceStatusActive, got.Status)
}

func TestWorkflowInstanceRepo_GetByID_NotFound(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestWorkflowInstanceRepo_UpdateState(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	require.NoError(t, repo.UpdateState(ctx, inst.ID, "approved"))

	got, err := repo.GetByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, "approved", got.CurrentStateID)
}

func TestWorkflowInstanceRepo_UpdateStatus(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	now := time.Now().UTC()
	require.NoError(t, repo.UpdateStatus(ctx, inst.ID, model.InstanceStatusCompleted, &now))

	got, err := repo.GetByID(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, model.InstanceStatusCompleted, got.Status)
	require.NotNil(t, got.CompletedAt)
}

func TestWorkflowInstanceRepo_List_WithFilters(t *testing.T) {
	truncateTables(t)
	_, inst1 := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	now := time.Now().UTC()
	require.NoError(t, repo.UpdateStatus(ctx, inst1.ID, model.InstanceStatusCompleted, &now))

	status := model.InstanceStatusActive
	insts, total, err := repo.List(ctx, repository.InstanceFilter{Status: &status, Limit: 50})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, insts, 0)

	status = model.InstanceStatusCompleted
	insts, total, err = repo.List(ctx, repository.InstanceFilter{Status: &status, Limit: 50})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, insts, 1)
}

func TestWorkflowInstanceRepo_SoftDelete(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewWorkflowInstanceRepo(testDB)

	require.NoError(t, repo.SoftDelete(ctx, inst.ID))

	_, err := repo.GetByID(ctx, inst.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestWorkflowInstanceRepo_GetByIDForUpdate(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()

	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Use tx-bound repo so FOR UPDATE runs inside the transaction
	repo := repository.NewWorkflowInstanceRepo(tx)
	got, err := repo.GetByIDForUpdate(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, inst.ID, got.ID)

	require.NoError(t, tx.Commit())
}
