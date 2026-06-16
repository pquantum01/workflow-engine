package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
)

func linearDef() model.WorkflowDefinition {
	return model.WorkflowDefinition{
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
	}
}

func TestValidate_LinearDef(t *testing.T) {
	def := linearDef()
	_, err := def.Validate()
	require.NoError(t, err)
}

func TestValidate_InvalidName(t *testing.T) {
	def := linearDef()
	def.Name = "Test Flow"
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidate_VersionZero(t *testing.T) {
	def := linearDef()
	def.Version = 0
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestValidate_MissingStartState(t *testing.T) {
	def := linearDef()
	def.StartState = "nonexistent"
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start_state")
}

func TestValidate_MissingEndState(t *testing.T) {
	def := linearDef()
	def.EndState = "nonexistent"
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_state")
}

func TestValidate_UnknownTaskType(t *testing.T) {
	def := linearDef()
	def.States[0].Tasks[0].Type = "unknown_type"
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestValidate_EndStateHasTasks(t *testing.T) {
	def := linearDef()
	def.States[1].Tasks = []model.TaskDef{{Key: "task1", Label: "Task", Type: "manual"}}
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no outgoing transitions")
}

func TestValidate_NonEndStateNoTasks(t *testing.T) {
	def := linearDef()
	def.States[0].Tasks = []model.TaskDef{}
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tasks")
}

func TestValidate_Cycle(t *testing.T) {
	def := model.WorkflowDefinition{
		Name:       "cyclic-flow",
		Version:    1,
		StartState: "a",
		EndState:   "c",
		States: []model.StateDef{
			{ID: "a", Label: "A", Tasks: []model.TaskDef{{Key: "t1", Label: "T1", Type: "manual"}}},
			{ID: "b", Label: "B", Tasks: []model.TaskDef{{Key: "t2", Label: "T2", Type: "manual"}}},
			{ID: "c", Label: "C", Tasks: []model.TaskDef{}},
		},
		Transitions: []model.TransitionDef{
			{FromState: "a", ToState: "b"},
			{FromState: "b", ToState: "a"}, // cycle
		},
	}
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestValidate_UnreachableState(t *testing.T) {
	def := model.WorkflowDefinition{
		Name:       "unreachable-flow",
		Version:    1,
		StartState: "a",
		EndState:   "b",
		States: []model.StateDef{
			{ID: "a", Label: "A", Tasks: []model.TaskDef{{Key: "t1", Label: "T1", Type: "manual"}}},
			{ID: "b", Label: "B", Tasks: []model.TaskDef{}},
			{ID: "orphan", Label: "Orphan", Tasks: []model.TaskDef{{Key: "t2", Label: "T2", Type: "manual"}}},
		},
		Transitions: []model.TransitionDef{
			{FromState: "a", ToState: "b"},
			{FromState: "orphan", ToState: "b"},
		},
	}
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable")
}

func TestValidate_TaskBoundMultiEdge(t *testing.T) {
	def := model.WorkflowDefinition{
		Name:       "task-bound-flow",
		Version:    1,
		StartState: "review",
		EndState:   "approved",
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
	}
	_, err := def.Validate()
	require.NoError(t, err)
}

func TestValidate_UserPicksMultiEdge(t *testing.T) {
	def := model.WorkflowDefinition{
		Name:       "user-picks-flow",
		Version:    1,
		StartState: "triage",
		EndState:   "done",
		States: []model.StateDef{
			{ID: "triage", Label: "Triage", Tasks: []model.TaskDef{
				{Key: "decide", Label: "Decision", Type: "manual"},
			}},
			{ID: "fast", Label: "Fast Track", Tasks: []model.TaskDef{
				{Key: "process", Label: "Process", Type: "manual"},
			}},
			{ID: "done", Label: "Done", Tasks: []model.TaskDef{}},
		},
		Transitions: []model.TransitionDef{
			{FromState: "triage", ToState: "fast"},
			{FromState: "triage", ToState: "done"},
			{FromState: "fast", ToState: "done"},
		},
	}
	_, err := def.Validate()
	require.NoError(t, err)
}

func TestValidate_UserPicksMustHaveSingleTask(t *testing.T) {
	def := model.WorkflowDefinition{
		Name:       "bad-user-picks",
		Version:    1,
		StartState: "triage",
		EndState:   "done",
		States: []model.StateDef{
			{ID: "triage", Label: "Triage", Tasks: []model.TaskDef{
				{Key: "t1", Label: "T1", Type: "manual"},
				{Key: "t2", Label: "T2", Type: "manual"},
			}},
			{ID: "done", Label: "Done", Tasks: []model.TaskDef{}},
			{ID: "other", Label: "Other", Tasks: []model.TaskDef{}},
		},
		Transitions: []model.TransitionDef{
			{FromState: "triage", ToState: "done"},
			{FromState: "triage", ToState: "other"},
		},
	}
	_, err := def.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 1 task")
}

func TestTransitionMode(t *testing.T) {
	def := linearDef()
	assert.Equal(t, model.ModeLinear, def.TransitionMode("submitted"))
}

func TestFindState(t *testing.T) {
	def := linearDef()
	s, ok := def.FindState("submitted")
	require.True(t, ok)
	assert.Equal(t, "Submitted", s.Label)

	_, ok = def.FindState("nonexistent")
	assert.False(t, ok)
}
