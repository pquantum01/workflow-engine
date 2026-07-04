package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

func makeTask(instanceID uuid.UUID, stateID, key string) *model.TaskInstance {
	return &model.TaskInstance{
		ID:                 uuid.New(),
		WorkflowInstanceID: instanceID,
		StateID:            stateID,
		TaskKey:            key,
		TaskLabel:          "Test Task",
		TaskType:           "manual",
		Status:             model.TaskStatusNew,
		CreatedBy:          "test",
		CreatedAt:          time.Now().UTC(),
	}
}

func TestTaskInstanceRepo_BulkCreateAndList(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	tasks := []*model.TaskInstance{
		makeTask(inst.ID, "submitted", "review"),
		makeTask(inst.ID, "submitted", "approve"),
	}
	require.NoError(t, repo.BulkCreate(ctx, tasks))

	got, err := repo.ListByInstance(ctx, inst.ID, nil)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestTaskInstanceRepo_GetByID(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	task := makeTask(inst.ID, "submitted", "review")
	require.NoError(t, repo.BulkCreate(ctx, []*model.TaskInstance{task}))

	got, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, got.ID)
	assert.Equal(t, model.TaskStatusNew, got.Status)
}

func TestTaskInstanceRepo_GetByID_NotFound(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestTaskInstanceRepo_UpdateStatus(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	task := makeTask(inst.ID, "submitted", "review")
	require.NoError(t, repo.BulkCreate(ctx, []*model.TaskInstance{task}))

	now := time.Now().UTC()
	require.NoError(t, repo.UpdateStatus(ctx, task.ID, model.TaskStatusDone, &now))

	got, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusDone, got.Status)
	require.NotNil(t, got.CompletedAt)
}

func TestTaskInstanceRepo_ListByInstanceAndState(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	tasks := []*model.TaskInstance{
		makeTask(inst.ID, "submitted", "review"),
		makeTask(inst.ID, "approved", "finalize"),
	}
	require.NoError(t, repo.BulkCreate(ctx, tasks))

	got, err := repo.ListByInstanceAndState(ctx, inst.ID, "submitted")
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "review", got[0].TaskKey)
}

func TestTaskInstanceRepo_ListByInstance_StatusFilter(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	task := makeTask(inst.ID, "submitted", "review")
	require.NoError(t, repo.BulkCreate(ctx, []*model.TaskInstance{task}))

	now := time.Now().UTC()
	require.NoError(t, repo.UpdateStatus(ctx, task.ID, model.TaskStatusDone, &now))

	got, err := repo.ListByInstance(ctx, inst.ID, []model.TaskStatus{model.TaskStatusNew})
	require.NoError(t, err)
	assert.Len(t, got, 0)

	got, err = repo.ListByInstance(ctx, inst.ID, []model.TaskStatus{model.TaskStatusDone})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestTaskInstanceRepo_SoftDelete(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewTaskInstanceRepo(testDB)

	task := makeTask(inst.ID, "submitted", "review")
	require.NoError(t, repo.BulkCreate(ctx, []*model.TaskInstance{task}))
	require.NoError(t, repo.SoftDelete(ctx, task.ID))

	_, err := repo.GetByID(ctx, task.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
