package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/service"
	"workflow-engine/backend/internal/testutil"
)

func TestCommentService_Add(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()
	var saved *model.TaskComment

	commentRepo := &testutil.MockTaskCommentRepo{
		AddFn: func(_ context.Context, c *model.TaskComment) error {
			saved = c
			return nil
		},
	}
	svc := service.NewCommentService(commentRepo, logrus.New())
	got, err := svc.Add(ctx, taskID, "alice", "looks good")
	require.NoError(t, err)
	assert.Equal(t, taskID, got.TaskInstanceID)
	assert.Equal(t, "alice", got.Author)
	assert.Equal(t, "looks good", got.Content)
	assert.NotNil(t, saved)
}

func TestCommentService_List(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()

	commentRepo := &testutil.MockTaskCommentRepo{
		ListByTaskFn: func(_ context.Context, id uuid.UUID) ([]*model.TaskComment, error) {
			return []*model.TaskComment{
				{TaskInstanceID: id, Author: "alice", Content: "c1"},
			}, nil
		},
	}
	svc := service.NewCommentService(commentRepo, logrus.New())
	list, err := svc.List(ctx, taskID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "c1", list[0].Content)
}
