package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/workflow-engine/backend/internal/model"
)

func TestTaskInstance_IsCompletable(t *testing.T) {
	task := &model.TaskInstance{
		ID:     uuid.New(),
		Status: model.TaskStatusNew,
	}
	assert.True(t, task.IsCompletable())

	task.Status = model.TaskStatusInProgress
	assert.True(t, task.IsCompletable())

	task.Status = model.TaskStatusDone
	assert.False(t, task.IsCompletable())

	task.Status = model.TaskStatusCancelled
	assert.False(t, task.IsCompletable())
}

func TestTaskInstance_IsTerminal(t *testing.T) {
	now := time.Now()
	task := &model.TaskInstance{ID: uuid.New(), CreatedAt: now}

	for _, status := range []model.TaskStatus{model.TaskStatusDone, model.TaskStatusRejected, model.TaskStatusCancelled} {
		task.Status = status
		assert.True(t, task.IsTerminal(), "expected terminal for %s", status)
	}
	for _, status := range []model.TaskStatus{model.TaskStatusNew, model.TaskStatusInProgress} {
		task.Status = status
		assert.False(t, task.IsTerminal(), "expected non-terminal for %s", status)
	}
}
