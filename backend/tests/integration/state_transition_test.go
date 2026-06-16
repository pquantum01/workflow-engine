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

func TestStateTransitionRepo_AppendAndList(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewStateTransitionRepo(testDB)

	tr := &model.StateTransition{
		ID:                 uuid.New(),
		WorkflowInstanceID: inst.ID,
		FromState:          "submitted",
		ToState:            "approved",
		TransitionedAt:     time.Now().UTC(),
	}
	require.NoError(t, repo.Append(ctx, tr))

	transitions, err := repo.ListByInstance(ctx, inst.ID)
	require.NoError(t, err)
	assert.Len(t, transitions, 1)
	assert.Equal(t, "submitted", transitions[0].FromState)
	assert.Equal(t, "approved", transitions[0].ToState)
}

func TestStateTransitionRepo_ListByInstance_Empty(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewStateTransitionRepo(testDB)

	transitions, err := repo.ListByInstance(ctx, inst.ID)
	require.NoError(t, err)
	assert.Len(t, transitions, 0)
}

func TestStateTransitionRepo_MultipleTransitions_OrderedByTime(t *testing.T) {
	truncateTables(t)
	_, inst := setupDefAndInstance(t)

	ctx := context.Background()
	repo := repository.NewStateTransitionRepo(testDB)

	t1 := &model.StateTransition{
		ID: uuid.New(), WorkflowInstanceID: inst.ID,
		FromState: "submitted", ToState: "review",
		TransitionedAt: time.Now().UTC(),
	}
	t2 := &model.StateTransition{
		ID: uuid.New(), WorkflowInstanceID: inst.ID,
		FromState: "review", ToState: "approved",
		TransitionedAt: time.Now().UTC().Add(time.Minute),
	}
	require.NoError(t, repo.Append(ctx, t1))
	require.NoError(t, repo.Append(ctx, t2))

	transitions, err := repo.ListByInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Len(t, transitions, 2)
	assert.Equal(t, "submitted", transitions[0].FromState)
	assert.Equal(t, "review", transitions[1].FromState)
}
