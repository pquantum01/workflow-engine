package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/testutil"
)

// engineFixture builds a fake DB-less engine backed by mock repos for unit tests.
// Since Engine manages its own tx (needs *sql.DB), we test via integration tests.
// This file tests the engine logic helpers directly (transition mode resolution etc.).

func TestTransitionMode_Linear(t *testing.T) {
	def := linearDef()
	assert.Equal(t, model.ModeLinear, def.Definition.TransitionMode("submitted"))
}

func TestTransitionMode_TaskBound(t *testing.T) {
	def := taskBoundDef()
	assert.Equal(t, model.ModeTaskBound, def.Definition.TransitionMode("review"))
}

func TestTransitionMode_UserPicks(t *testing.T) {
	def := userPicksDef()
	assert.Equal(t, model.ModeUserPicks, def.Definition.TransitionMode("triage"))
}

func TestEvent_Constants(t *testing.T) {
	assert.Equal(t, "task.created", engine.EventTaskCreated)
	assert.Equal(t, "instance.completed", engine.EventInstanceCompleted)
	assert.Equal(t, "instance.cancelled", engine.EventInstanceCancelled)
}

// --- helpers ---

func linearDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(),
		Definition: model.WorkflowDefinition{
			Name: "linear", Version: 1, StartState: "submitted", EndState: "approved",
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
	}
}

func taskBoundDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(),
		Definition: model.WorkflowDefinition{
			Name: "task-bound", Version: 1, StartState: "review", EndState: "approved",
			States: []model.StateDef{
				{ID: "review", Label: "Review", Tasks: []model.TaskDef{
					{Key: "approve", Label: "Approve", Type: "approval"},
					{Key: "reject", Label: "Reject", Type: "approval"},
				}},
				{ID: "approved", Label: "Approved", Tasks: []model.TaskDef{}},
				{ID: "denied", Label: "Denied", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{
				{FromState: "review", ToState: "approved", Trigger: &model.TransitionTrigger{TaskKey: "approve"}},
				{FromState: "review", ToState: "denied", Trigger: &model.TransitionTrigger{TaskKey: "reject"}},
			},
		},
	}
}

func userPicksDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(),
		Definition: model.WorkflowDefinition{
			Name: "user-picks", Version: 1, StartState: "triage", EndState: "resolved",
			States: []model.StateDef{
				{ID: "triage", Label: "Triage", Tasks: []model.TaskDef{
					{Key: "decide", Label: "Decision", Type: "manual"},
				}},
				{ID: "in_progress", Label: "In Progress", Tasks: []model.TaskDef{
					{Key: "fix", Label: "Fix", Type: "manual"},
				}},
				{ID: "resolved", Label: "Resolved", Tasks: []model.TaskDef{}},
				{ID: "wont_fix", Label: "Won't Fix", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{
				{FromState: "triage", ToState: "in_progress"},
				{FromState: "triage", ToState: "wont_fix"},
				{FromState: "in_progress", ToState: "resolved"},
			},
		},
	}
}

// taskWith returns a TaskInstance with given state and key.
func taskWith(instanceID uuid.UUID, stateID, key string, status model.TaskStatus) *model.TaskInstance {
	return &model.TaskInstance{
		ID:                 uuid.New(),
		WorkflowInstanceID: instanceID,
		StateID:            stateID,
		TaskKey:            key,
		TaskLabel:          key,
		TaskType:           "manual",
		Status:             status,
		CreatedAt:          time.Now().UTC(),
	}
}

// mockDefRepo returns a defRepo that serves the given def.
func mockDefRepo(def *model.WorkflowDef) repository.WorkflowDefRepo {
	return &testutil.MockWorkflowDefRepo{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
			if id == def.ID {
				return def, nil
			}
			return nil, repository.ErrNotFound
		},
	}
}

func TestMockRepos_Compile(t *testing.T) {
	// Ensure mock repos satisfy interfaces at compile time
	var _ repository.WorkflowDefRepo = &testutil.MockWorkflowDefRepo{}
	var _ repository.WorkflowInstanceRepo = &testutil.MockWorkflowInstanceRepo{}
	var _ repository.TaskInstanceRepo = &testutil.MockTaskInstanceRepo{}
	var _ repository.StateTransitionRepo = &testutil.MockStateTransitionRepo{}
	var _ repository.WebhookOutboxRepo = &testutil.MockWebhookOutboxRepo{}
}

// --- Integration-style engine tests using real DB ---
// These live in tests/integration/engine_test.go to use testDB.
// Unit tests above verify model logic and compile-time interface satisfaction.
