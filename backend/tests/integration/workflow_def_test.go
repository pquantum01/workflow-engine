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

func makeTestDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID:          uuid.New(),
		Name:        "test-flow",
		Version:     1,
		Description: "Integration test workflow",
		Definition: model.WorkflowDefinition{
			Name:       "test-flow",
			Version:    1,
			StartState: "submitted",
			EndState:   "approved",
			States: []model.StateDef{
				{ID: "submitted", Label: "Submitted", Tasks: []model.TaskDef{
					{Key: "review", Label: "Review", Type: "approval"},
				}},
				{ID: "approved", Label: "Approved", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{
				{FromState: "submitted", ToState: "approved"},
			},
		},
		CreatedAt: time.Now().UTC(),
	}
}

func TestWorkflowDefRepo_CreateAndGetByID(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def := makeTestDef()
	require.NoError(t, repo.Create(ctx, def))

	got, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, def.ID, got.ID)
	assert.Equal(t, def.Name, got.Name)
	assert.Equal(t, def.Version, got.Version)
	assert.Equal(t, def.Definition.StartState, got.Definition.StartState)
}

func TestWorkflowDefRepo_GetByNameVersion(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def := makeTestDef()
	require.NoError(t, repo.Create(ctx, def))

	got, err := repo.GetByNameVersion(ctx, def.Name, def.Version)
	require.NoError(t, err)
	assert.Equal(t, def.ID, got.ID)
}

func TestWorkflowDefRepo_GetByID_NotFound(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestWorkflowDefRepo_List(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def1 := makeTestDef()
	def2 := makeTestDef()
	def2.ID = uuid.New()
	def2.Name = "other-flow"

	require.NoError(t, repo.Create(ctx, def1))
	require.NoError(t, repo.Create(ctx, def2))

	defs, err := repo.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, defs, 2)
}

func TestWorkflowDefRepo_SoftDelete(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def := makeTestDef()
	require.NoError(t, repo.Create(ctx, def))
	require.NoError(t, repo.SoftDelete(ctx, def.ID))

	defs, err := repo.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, defs, 0)

	defs, err = repo.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, defs, 1)
	assert.NotNil(t, defs[0].DeletedAt)

	_, err = repo.GetByID(ctx, def.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestWorkflowDefRepo_UniqueNameVersion(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def1 := makeTestDef()
	require.NoError(t, repo.Create(ctx, def1))

	def2 := makeTestDef()
	def2.ID = uuid.New()
	err := repo.Create(ctx, def2)
	assert.Error(t, err)
}

func TestWorkflowDefRepo_SoftDeleteAllowsReuseOfNameVersion(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewWorkflowDefRepo(testDB)

	def1 := makeTestDef()
	require.NoError(t, repo.Create(ctx, def1))
	require.NoError(t, repo.SoftDelete(ctx, def1.ID))

	def2 := makeTestDef()
	def2.ID = uuid.New()
	require.NoError(t, repo.Create(ctx, def2))
}
