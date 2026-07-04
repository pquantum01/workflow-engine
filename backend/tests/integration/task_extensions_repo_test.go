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

func TestTaskInstanceRepo_CreateAndGetExtended(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)
	ctx := context.Background()

	date := "2026-07-10"
	timeStr := "14:00"
	task := &model.TaskInstance{
		ID:                     uuid.New(),
		WorkflowInstanceID:     inst.ID,
		StateID:                "submitted",
		TaskKey:                "adhoc_task_a1b2",
		TaskLabel:              "Ad-hoc Review",
		TaskType:               "manual",
		Status:                 model.TaskStatusNew,
		CreatedBy:              "alice",
		CreatedAt:              time.Now().UTC(),
		Priority:               "high",
		ExpectedCompletionDate: &date,
		ExpectedCompletionTime: &timeStr,
		IsAdhoc:                true,
	}

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	require.NoError(t, taskRepo.Create(ctx, task))

	got, err := taskRepo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "high", got.Priority)
	assert.True(t, got.IsAdhoc)
	require.NotNil(t, got.ExpectedCompletionDate)
	assert.Equal(t, "2026-07-10", *got.ExpectedCompletionDate)
	require.NotNil(t, got.ExpectedCompletionTime)
	assert.Equal(t, "14:00:00", *got.ExpectedCompletionTime)
}

func TestTaskInstanceRepo_UpdateFields(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)
	ctx := context.Background()
	taskRepo := repository.NewTaskInstanceRepo(testDB)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: inst.ID, StateID: "submitted",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now().UTC(),
		Priority: "medium",
	}}))

	newPriority := "low"
	newDate := "2026-08-01"
	require.NoError(t, taskRepo.UpdateFields(ctx, taskID, repository.TaskInstancePatch{
		Priority:               &newPriority,
		ExpectedCompletionDate: &newDate,
	}))

	got, err := taskRepo.GetByID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "low", got.Priority)
	require.NotNil(t, got.ExpectedCompletionDate)
	assert.Equal(t, "2026-08-01", *got.ExpectedCompletionDate)
}

func TestTaskCommentRepo_AddAndList(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)
	ctx := context.Background()
	taskRepo := repository.NewTaskInstanceRepo(testDB)
	commentRepo := repository.NewTaskCommentRepo(testDB)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: inst.ID, StateID: "submitted",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now().UTC(),
		Priority: "medium",
	}}))

	c1 := &model.TaskComment{ID: uuid.New(), TaskInstanceID: taskID, Author: "alice", Content: "first comment", CreatedAt: time.Now().UTC()}
	c2 := &model.TaskComment{ID: uuid.New(), TaskInstanceID: taskID, Author: "bob", Content: "second comment", CreatedAt: time.Now().UTC().Add(time.Second)}
	require.NoError(t, commentRepo.Add(ctx, c1))
	require.NoError(t, commentRepo.Add(ctx, c2))

	list, err := commentRepo.ListByTask(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "alice", list[0].Author)
	assert.Equal(t, "bob", list[1].Author)
}

func TestTaskAttachmentRepo_UpsertAndGet(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)
	ctx := context.Background()
	taskRepo := repository.NewTaskInstanceRepo(testDB)
	attRepo := repository.NewTaskAttachmentRepo(testDB)

	taskID := uuid.New()
	require.NoError(t, taskRepo.BulkCreate(ctx, []*model.TaskInstance{{
		ID: taskID, WorkflowInstanceID: inst.ID, StateID: "submitted",
		TaskKey: "t1", TaskLabel: "T1", TaskType: "manual",
		Status: model.TaskStatusNew, CreatedBy: "bob", CreatedAt: time.Now().UTC(),
		Priority: "medium",
	}}))

	att := &model.TaskAttachment{
		ID: uuid.New(), TaskInstanceID: taskID,
		Filename: "contract.pdf", ContentType: "application/pdf",
		SizeBytes: 1024, S3URL: "https://bucket.s3.amazonaws.com/attachments/contract.pdf",
		UploadedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, attRepo.Upsert(ctx, att))

	got, err := attRepo.GetByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "contract.pdf", got.Filename)
	assert.Equal(t, 1024, got.SizeBytes)

	att2 := &model.TaskAttachment{
		ID: uuid.New(), TaskInstanceID: taskID,
		Filename: "revised.pdf", ContentType: "application/pdf",
		SizeBytes: 2048, S3URL: "https://bucket.s3.amazonaws.com/attachments/revised.pdf",
		UploadedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, attRepo.Upsert(ctx, att2))

	got2, err := attRepo.GetByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Equal(t, "revised.pdf", got2.Filename)

	require.NoError(t, attRepo.DeleteByTaskID(ctx, taskID))
	_, err = attRepo.GetByTaskID(ctx, taskID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
