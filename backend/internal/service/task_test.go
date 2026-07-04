package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/service"
	"workflow-engine/backend/internal/testutil"
)

func TestTaskService_UpdateFields_TerminalTask(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()

	taskRepo := &testutil.MockTaskInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
			return &model.TaskInstance{ID: id, Status: model.TaskStatusDone, Priority: "medium"}, nil
		},
	}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	priority := "high"
	err := svc.UpdateFields(ctx, taskID, service.TaskFieldPatch{Priority: &priority})
	assert.ErrorIs(t, err, service.ErrTaskCompleted)
}

func TestTaskService_UpdateFields_Success(t *testing.T) {
	ctx := context.Background()
	taskID := uuid.New()
	var patchReceived repository.TaskInstancePatch

	taskRepo := &testutil.MockTaskInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
			return &model.TaskInstance{ID: id, Status: model.TaskStatusNew, Priority: "medium"}, nil
		},
		UpdateFieldsFn: func(_ context.Context, id uuid.UUID, p repository.TaskInstancePatch) error {
			patchReceived = p
			return nil
		},
	}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	priority := "high"
	err := svc.UpdateFields(ctx, taskID, service.TaskFieldPatch{Priority: &priority})
	require.NoError(t, err)
	require.NotNil(t, patchReceived.Priority)
	assert.Equal(t, "high", *patchReceived.Priority)
}

func TestTaskService_CreateAdhoc_InvalidState(t *testing.T) {
	ctx := context.Background()
	instID := uuid.New()
	defID := uuid.New()

	instRepo := &testutil.MockWorkflowInstanceRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
			return &model.WorkflowInstance{
				ID: id, WorkflowDefID: defID, Status: model.InstanceStatusActive,
			}, nil
		},
	}
	defRepo := &testutil.MockWorkflowDefRepo{
		GetByIDFn: func(_ context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
			return &model.WorkflowDef{
				ID:         id,
				Definition: model.WorkflowDefinition{States: []model.StateDef{{ID: "review"}}},
			}, nil
		},
	}
	taskRepo := &testutil.MockTaskInstanceRepo{}
	svc := service.NewTaskService(nil, taskRepo, logrus.New())
	svc.SetInstRepo(instRepo)
	svc.SetDefRepo(defRepo)

	_, err := svc.CreateAdhoc(ctx, instID, service.AdhocTaskRequest{
		StateID:   "nonexistent",
		Label:     "My Task",
		TaskType:  "manual",
		CreatedBy: "alice",
	})
	assert.ErrorIs(t, err, service.ErrInvalidState)
}
