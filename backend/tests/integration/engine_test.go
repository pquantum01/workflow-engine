package integration_test

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

func newEngine() *engine.Engine {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return engine.New(
		testDB,
		repository.NewWorkflowDefRepo(testDB),
		repository.NewWorkflowInstanceRepo(testDB),
		repository.NewTaskInstanceRepo(testDB),
		repository.NewStateTransitionRepo(testDB),
		log,
	)
}

func seedDef(t *testing.T, def *model.WorkflowDef) {
	t.Helper()
	_, err := def.Definition.Validate()
	require.NoError(t, err, "test def must be valid")
	require.NoError(t, repository.NewWorkflowDefRepo(testDB).Create(context.Background(), def))
}

func linearDefForEngine() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(), Name: "eng-linear", Version: 1,
		Definition: model.WorkflowDefinition{
			Name: "eng-linear", Version: 1, StartState: "s1", EndState: "s2",
			States: []model.StateDef{
				{ID: "s1", Label: "S1", Tasks: []model.TaskDef{{Key: "t1", Label: "T1", Type: "manual"}}},
				{ID: "s2", Label: "S2", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{{FromState: "s1", ToState: "s2"}},
		},
	}
}

func multiTaskLinearDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(), Name: "eng-multitask", Version: 1,
		Definition: model.WorkflowDefinition{
			Name: "eng-multitask", Version: 1, StartState: "s1", EndState: "s2",
			States: []model.StateDef{
				{ID: "s1", Label: "S1", Tasks: []model.TaskDef{
					{Key: "t1", Label: "T1", Type: "manual"},
					{Key: "t2", Label: "T2", Type: "manual"},
				}},
				{ID: "s2", Label: "S2", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{{FromState: "s1", ToState: "s2"}},
		},
	}
}

func taskBoundDefForEngine() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(), Name: "eng-taskbound", Version: 1,
		Definition: model.WorkflowDefinition{
			Name: "eng-taskbound", Version: 1, StartState: "review", EndState: "approved",
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

func userPicksDefForEngine() *model.WorkflowDef {
	return &model.WorkflowDef{
		ID: uuid.New(), Name: "eng-userpicks", Version: 1,
		Definition: model.WorkflowDefinition{
			Name: "eng-userpicks", Version: 1, StartState: "triage", EndState: "resolved",
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

func TestEngine_Start_CreatesInstanceAndTasks(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := linearDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, events, err := eng.Start(ctx, def.ID, "user1")
	require.NoError(t, err)

	assert.Equal(t, model.InstanceStatusActive, inst.Status)
	assert.Equal(t, "s1", inst.CurrentStateID)

	// tasks spawned
	tasks, err := repository.NewTaskInstanceRepo(testDB).ListByInstance(ctx, inst.ID, nil)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "t1", tasks[0].TaskKey)

	// events
	assert.Contains(t, eventsOfType(events, engine.EventInstanceStarted), engine.EventInstanceStarted)
	assert.Contains(t, eventsOfType(events, engine.EventTaskCreated), engine.EventTaskCreated)
}

func TestEngine_Linear_SingleTask_Done_Transitions(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := linearDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, err := eng.Start(ctx, def.ID, "user1")
	require.NoError(t, err)

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)
	require.Len(t, tasks, 1)

	inst, events, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)

	assert.Equal(t, model.InstanceStatusCompleted, inst.Status)
	assert.Equal(t, "s2", inst.CurrentStateID)
	assert.Contains(t, eventsOfType(events, engine.EventInstanceCompleted), engine.EventInstanceCompleted)
}

func TestEngine_Linear_MultiTask_WaitsForAll(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := multiTaskLinearDef()
	seedDef(t, def)

	eng := newEngine()
	inst, _, err := eng.Start(ctx, def.ID, "user1")
	require.NoError(t, err)

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)
	require.Len(t, tasks, 2)

	// Complete first — should still be active
	inst, _, err = eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, model.InstanceStatusActive, inst.Status)
	assert.Equal(t, "s1", inst.CurrentStateID)

	// Complete second — should now complete
	inst, events, err := eng.Complete(ctx, tasks[1].ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, model.InstanceStatusCompleted, inst.Status)
	assert.Contains(t, eventsOfType(events, engine.EventInstanceCompleted), engine.EventInstanceCompleted)
}

func TestEngine_Rejection_CancelsInstance(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := linearDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, err := eng.Start(ctx, def.ID, "user1")
	require.NoError(t, err)

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	inst, events, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusRejected, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, model.InstanceStatusCancelled, inst.Status)
	assert.Contains(t, eventsOfType(events, engine.EventInstanceCancelled), engine.EventInstanceCancelled)
}

func TestEngine_CompleteOnTerminalInstance_Errors(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := linearDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")
	eng.Cancel(ctx, inst.ID)

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	_, _, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "", "user1")
	assert.ErrorIs(t, err, engine.ErrInstanceTerminal)
}

func TestEngine_Cancel_CancelsAllOpenTasks(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := multiTaskLinearDef()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	inst, events, err := eng.Cancel(ctx, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, model.InstanceStatusCancelled, inst.Status)
	assert.Contains(t, eventsOfType(events, engine.EventInstanceCancelled), engine.EventInstanceCancelled)

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)
	for _, task := range tasks {
		assert.Equal(t, model.TaskStatusCancelled, task.Status)
	}
}

func TestEngine_TaskBound_CorrectEdgeFires(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := taskBoundDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	// Find "approve" task
	var approveTask *model.TaskInstance
	for _, t := range tasks {
		if t.TaskKey == "approve" {
			approveTask = t
			break
		}
	}
	require.NotNil(t, approveTask)

	inst, _, err := eng.Complete(ctx, approveTask.ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, "approved", inst.CurrentStateID)
	assert.Equal(t, model.InstanceStatusCompleted, inst.Status)

	// "reject" task should be cancelled
	allTasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)
	for _, task := range allTasks {
		if task.TaskKey == "reject" {
			assert.Equal(t, model.TaskStatusCancelled, task.Status)
		}
	}
}

func TestEngine_TaskBound_RejectEdgeFires(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := taskBoundDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	var rejectTask *model.TaskInstance
	for _, t := range tasks {
		if t.TaskKey == "reject" {
			rejectTask = t
			break
		}
	}
	require.NotNil(t, rejectTask)

	inst, _, err := eng.Complete(ctx, rejectTask.ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, "denied", inst.CurrentStateID)
	// denied is a terminal state but not end_state → instance completed
	assert.Equal(t, model.InstanceStatusCompleted, inst.Status)
}

func TestEngine_UserPicks_RequiresNextState(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := userPicksDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	_, _, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "", "user1")
	assert.ErrorIs(t, err, engine.ErrNextStateRequired)
}

func TestEngine_UserPicks_InvalidNextState(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := userPicksDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	_, _, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "nonexistent", "user1")
	assert.ErrorIs(t, err, engine.ErrInvalidNextState)
}

func TestEngine_UserPicks_ValidNextState(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := userPicksDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)

	inst, _, err := eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "in_progress", "user1")
	require.NoError(t, err)
	assert.Equal(t, "in_progress", inst.CurrentStateID)
}

func TestEngine_StateTransition_Recorded(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()
	def := linearDefForEngine()
	seedDef(t, def)

	eng := newEngine()
	inst, _, _ := eng.Start(ctx, def.ID, "user1")

	taskRepo := repository.NewTaskInstanceRepo(testDB)
	tasks, _ := taskRepo.ListByInstance(ctx, inst.ID, nil)
	eng.Complete(ctx, tasks[0].ID, model.TaskStatusDone, "", "user1")

	transitions, err := repository.NewStateTransitionRepo(testDB).ListByInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.Len(t, transitions, 1)
	assert.Equal(t, "s1", transitions[0].FromState)
	assert.Equal(t, "s2", transitions[0].ToState)
}

func TestEngine_AdhocTask_InPastState_NoTransition(t *testing.T) {
	truncateTables(t)
	ctx := context.Background()

	// 3-state workflow: s1(T1) → s2(T2) → s3(terminal)
	// s2 is non-terminal so instance stays active after s1→s2
	def := &model.WorkflowDef{
		ID: uuid.New(), Name: "eng-adhoc-paststate", Version: 1,
		Definition: model.WorkflowDefinition{
			Name: "eng-adhoc-paststate", Version: 1, StartState: "s1", EndState: "s3",
			States: []model.StateDef{
				{ID: "s1", Label: "S1", Tasks: []model.TaskDef{{Key: "t1", Label: "T1", Type: "manual"}}},
				{ID: "s2", Label: "S2", Tasks: []model.TaskDef{{Key: "t2", Label: "T2", Type: "manual"}}},
				{ID: "s3", Label: "S3", Tasks: []model.TaskDef{}},
			},
			Transitions: []model.TransitionDef{
				{FromState: "s1", ToState: "s2"},
				{FromState: "s2", ToState: "s3"},
			},
		},
	}
	seedDef(t, def)

	eng := newEngine()
	taskRepo := repository.NewTaskInstanceRepo(testDB)

	// Start instance — T1 created in s1
	inst, _, err := eng.Start(ctx, def.ID, "user1")
	require.NoError(t, err)
	require.Equal(t, "s1", inst.CurrentStateID)

	// Complete T1 → transitions to s2, T2 spawned
	s1Tasks, _ := taskRepo.ListByInstanceAndState(ctx, inst.ID, "s1")
	inst, _, err = eng.Complete(ctx, s1Tasks[0].ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	require.Equal(t, "s2", inst.CurrentStateID)

	// Confirm T2 exists in s2
	s2Tasks, _ := taskRepo.ListByInstanceAndState(ctx, inst.ID, "s2")
	require.Len(t, s2Tasks, 1)

	// Create adhoc task for past state s1 (instance is now at s2)
	adhoc := &model.TaskInstance{
		ID:                 uuid.New(),
		WorkflowInstanceID: inst.ID,
		StateID:            "s1",
		TaskKey:            "adhoc_followup_aa00",
		TaskLabel:          "Adhoc Follow-up",
		TaskType:           "manual",
		Status:             model.TaskStatusNew,
		CreatedBy:          "user1",
		Priority:           "medium",
		IsAdhoc:            true,
	}
	require.NoError(t, taskRepo.Create(ctx, adhoc))

	// Complete adhoc — must NOT re-fire s1→s2 transition (would duplicate T2)
	inst, _, err = eng.Complete(ctx, adhoc.ID, model.TaskStatusDone, "", "user1")
	require.NoError(t, err)
	assert.Equal(t, "s2", inst.CurrentStateID) // still at s2

	// T2 must still exist exactly once — no duplicate
	s2TasksAfter, _ := taskRepo.ListByInstanceAndState(ctx, inst.ID, "s2")
	assert.Len(t, s2TasksAfter, 1)
}

func eventsOfType(events []engine.Event, eventType string) []string {
	var found []string
	for _, e := range events {
		if e.Type == eventType {
			found = append(found, e.Type)
		}
	}
	return found
}
