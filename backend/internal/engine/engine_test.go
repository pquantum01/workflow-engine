package engine

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

// --- Mock Repositories ---

type mockWorkflowRepo struct {
	workflows map[uuid.UUID]*model.Workflow
}

func (m *mockWorkflowRepo) Create(_ context.Context, w *model.Workflow) error {
	m.workflows[w.ID] = w
	return nil
}
func (m *mockWorkflowRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Workflow, error) {
	w, ok := m.workflows[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return w, nil
}
func (m *mockWorkflowRepo) ListByProcessID(_ context.Context, _ uuid.UUID) ([]*model.Workflow, error) {
	return nil, nil
}
func (m *mockWorkflowRepo) Update(_ context.Context, w *model.Workflow) error {
	m.workflows[w.ID] = w
	return nil
}
func (m *mockWorkflowRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type mockWorkflowInstanceRepo struct {
	instances map[uuid.UUID]*model.WorkflowInstance
}

func (m *mockWorkflowInstanceRepo) Create(_ context.Context, wi *model.WorkflowInstance) error {
	m.instances[wi.ID] = wi
	return nil
}
func (m *mockWorkflowInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	wi, ok := m.instances[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return wi, nil
}
func (m *mockWorkflowInstanceRepo) ListByProcessInstanceID(_ context.Context, _ uuid.UUID) ([]*model.WorkflowInstance, error) {
	return nil, nil
}
func (m *mockWorkflowInstanceRepo) Update(_ context.Context, wi *model.WorkflowInstance) error {
	m.instances[wi.ID] = wi
	return nil
}

type mockTaskRepo struct {
	tasks map[uuid.UUID]*model.Task
}

func (m *mockTaskRepo) Create(_ context.Context, t *model.Task) error {
	m.tasks[t.ID] = t
	return nil
}
func (m *mockTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}
func (m *mockTaskRepo) ListByWorkflowID(_ context.Context, wfID uuid.UUID) ([]*model.Task, error) {
	var result []*model.Task
	for _, t := range m.tasks {
		if t.WorkflowID == wfID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (m *mockTaskRepo) Update(_ context.Context, t *model.Task) error {
	m.tasks[t.ID] = t
	return nil
}
func (m *mockTaskRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type mockTaskInstanceRepo struct {
	instances map[uuid.UUID]*model.TaskInstance
}

func (m *mockTaskInstanceRepo) Create(_ context.Context, ti *model.TaskInstance) error {
	m.instances[ti.ID] = ti
	return nil
}
func (m *mockTaskInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	ti, ok := m.instances[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return ti, nil
}
func (m *mockTaskInstanceRepo) ListByWorkflowInstanceID(_ context.Context, wiID uuid.UUID) ([]*model.TaskInstance, error) {
	var result []*model.TaskInstance
	for _, ti := range m.instances {
		if ti.WorkflowInstanceID == wiID {
			result = append(result, ti)
		}
	}
	return result, nil
}
func (m *mockTaskInstanceRepo) Update(_ context.Context, ti *model.TaskInstance) error {
	m.instances[ti.ID] = ti
	return nil
}

type mockStateTransitionRepo struct {
	transitions []*model.StateTransition
}

func (m *mockStateTransitionRepo) Create(_ context.Context, st *model.StateTransition) error {
	m.transitions = append(m.transitions, st)
	return nil
}
func (m *mockStateTransitionRepo) ListByWorkflowInstanceID(_ context.Context, wiID uuid.UUID) ([]*model.StateTransition, error) {
	var result []*model.StateTransition
	for _, st := range m.transitions {
		if st.WorkflowInstanceID == wiID {
			result = append(result, st)
		}
	}
	return result, nil
}

// --- Test helpers ---

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

type testSetup struct {
	engine               Engine
	workflowID           uuid.UUID
	workflowInstanceID   uuid.UUID
	taskIDs              map[string]uuid.UUID // task name -> task ID
	taskInstanceIDs      map[string]uuid.UUID // task name -> task instance ID
	taskInstanceRepo     *mockTaskInstanceRepo
	workflowInstanceRepo *mockWorkflowInstanceRepo
	stRepo               *mockStateTransitionRepo
}

func setupEngine(sm model.StateMachine, taskNames []string, currentState string) *testSetup {
	workflowID := uuid.New()
	workflowInstanceID := uuid.New()
	processInstanceID := uuid.New()
	now := time.Now().UTC()

	wf := &model.Workflow{
		ID:           workflowID,
		ProcessID:    uuid.New(),
		Name:         "Test Workflow",
		StateMachine: sm,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	wi := &model.WorkflowInstance{
		ID:                workflowInstanceID,
		WorkflowID:        workflowID,
		ProcessInstanceID: processInstanceID,
		CurrentState:      currentState,
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	wfRepo := &mockWorkflowRepo{workflows: map[uuid.UUID]*model.Workflow{workflowID: wf}}
	wiRepo := &mockWorkflowInstanceRepo{instances: map[uuid.UUID]*model.WorkflowInstance{workflowInstanceID: wi}}
	taskRepo := &mockTaskRepo{tasks: make(map[uuid.UUID]*model.Task)}
	tiRepo := &mockTaskInstanceRepo{instances: make(map[uuid.UUID]*model.TaskInstance)}
	stRepo := &mockStateTransitionRepo{}

	taskIDs := make(map[string]uuid.UUID)
	taskInstanceIDs := make(map[string]uuid.UUID)

	for _, name := range taskNames {
		tID := uuid.New()
		tiID := uuid.New()
		taskIDs[name] = tID
		taskInstanceIDs[name] = tiID

		taskRepo.tasks[tID] = &model.Task{
			ID:         tID,
			WorkflowID: workflowID,
			Name:       name,
			TaskType:   "manual",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		tiRepo.instances[tiID] = &model.TaskInstance{
			ID:                 tiID,
			TaskID:             tID,
			WorkflowInstanceID: workflowInstanceID,
			Status:             "pending",
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	eng := New(wfRepo, wiRepo, taskRepo, tiRepo, stRepo, testLogger())

	return &testSetup{
		engine:               eng,
		workflowID:           workflowID,
		workflowInstanceID:   workflowInstanceID,
		taskIDs:              taskIDs,
		taskInstanceIDs:      taskInstanceIDs,
		taskInstanceRepo:     tiRepo,
		workflowInstanceRepo: wiRepo,
		stRepo:               stRepo,
	}
}

func (ts *testSetup) completeTask(name string) {
	tiID := ts.taskInstanceIDs[name]
	ti := ts.taskInstanceRepo.instances[tiID]
	ti.Status = "completed"
	now := time.Now().UTC()
	ti.CompletedAt = &now
}

// --- Tests ---

func TestEvaluate_NoTransitionsAvailable(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, nil, "open")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result when no guarded transitions")
	}
}

func TestEvaluate_SingleRule_Satisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions: []model.Transition{
			{Name: "close", FromState: "open", ToState: "closed", CompletionRule: &model.CompletionRule{
				Type: "single", TaskRefs: []string{"review"},
			}},
		},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, []string{"review"}, "open")
	ts.completeTask("review")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected transition result")
	}
	if result.ToState != "closed" {
		t.Errorf("ToState = %q, want %q", result.ToState, "closed")
	}
	if !result.IsTerminal {
		t.Error("expected IsTerminal to be true")
	}
}

func TestEvaluate_SingleRule_NotSatisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions: []model.Transition{
			{Name: "close", FromState: "open", ToState: "closed", CompletionRule: &model.CompletionRule{
				Type: "single", TaskRefs: []string{"review"},
			}},
		},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, []string{"review"}, "open")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result when task not completed")
	}
}

func TestEvaluate_AllOf_Satisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "review"}, {Name: "approved"}},
		Transitions: []model.Transition{
			{Name: "approve", FromState: "review", ToState: "approved", CompletionRule: &model.CompletionRule{
				Type: "all_of", TaskRefs: []string{"code_review", "qa_review"},
			}},
		},
		InitialState:   "review",
		TerminalStates: []string{"approved"},
	}
	ts := setupEngine(sm, []string{"code_review", "qa_review"}, "review")
	ts.completeTask("code_review")
	ts.completeTask("qa_review")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected transition")
	}
	if result.ToState != "approved" {
		t.Errorf("ToState = %q, want %q", result.ToState, "approved")
	}
}

func TestEvaluate_AllOf_Partial(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "review"}, {Name: "approved"}},
		Transitions: []model.Transition{
			{Name: "approve", FromState: "review", ToState: "approved", CompletionRule: &model.CompletionRule{
				Type: "all_of", TaskRefs: []string{"code_review", "qa_review"},
			}},
		},
		InitialState:   "review",
		TerminalStates: []string{"approved"},
	}
	ts := setupEngine(sm, []string{"code_review", "qa_review"}, "review")
	ts.completeTask("code_review")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil when not all tasks complete")
	}
}

func TestEvaluate_AnyOf_OneDone(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "waiting"}, {Name: "done"}},
		Transitions: []model.Transition{
			{Name: "finish", FromState: "waiting", ToState: "done", CompletionRule: &model.CompletionRule{
				Type: "any_of", TaskRefs: []string{"task_a", "task_b"},
			}},
		},
		InitialState:   "waiting",
		TerminalStates: []string{"done"},
	}
	ts := setupEngine(sm, []string{"task_a", "task_b"}, "waiting")
	ts.completeTask("task_b")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected transition")
	}
}

func TestEvaluate_NOfM_Satisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "voting"}, {Name: "approved"}},
		Transitions: []model.Transition{
			{Name: "approve", FromState: "voting", ToState: "approved", CompletionRule: &model.CompletionRule{
				Type: "n_of_m", TaskRefs: []string{"vote_a", "vote_b", "vote_c"}, N: 2,
			}},
		},
		InitialState:   "voting",
		TerminalStates: []string{"approved"},
	}
	ts := setupEngine(sm, []string{"vote_a", "vote_b", "vote_c"}, "voting")
	ts.completeTask("vote_a")
	ts.completeTask("vote_c")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected transition")
	}
}

func TestEvaluate_NOfM_NotEnough(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "voting"}, {Name: "approved"}},
		Transitions: []model.Transition{
			{Name: "approve", FromState: "voting", ToState: "approved", CompletionRule: &model.CompletionRule{
				Type: "n_of_m", TaskRefs: []string{"vote_a", "vote_b", "vote_c"}, N: 2,
			}},
		},
		InitialState:   "voting",
		TerminalStates: []string{"approved"},
	}
	ts := setupEngine(sm, []string{"vote_a", "vote_b", "vote_c"}, "voting")
	ts.completeTask("vote_a")

	result, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil when less than N tasks done")
	}
}

func TestEvaluate_TerminalState_Error(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, nil, "closed")

	_, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err == nil {
		t.Fatal("expected error for terminal state")
	}
}

func TestEvaluate_NonActiveWorkflow_Error(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, nil, "open")
	ts.workflowInstanceRepo.instances[ts.workflowInstanceID].Status = "cancelled"

	_, err := ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)
	if err == nil {
		t.Fatal("expected error for non-active workflow")
	}
}

func TestTriggerTransition_Unguarded_Success(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "in_progress"}, {Name: "done"}},
		Transitions:    []model.Transition{{Name: "start", FromState: "open", ToState: "in_progress"}},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
	ts := setupEngine(sm, nil, "open")

	result, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "start", "user-1", "human")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ToState != "in_progress" {
		t.Errorf("ToState = %q, want %q", result.ToState, "in_progress")
	}
	if result.IsTerminal {
		t.Error("should not be terminal")
	}

	// Verify state was updated
	wi := ts.workflowInstanceRepo.instances[ts.workflowInstanceID]
	if wi.CurrentState != "in_progress" {
		t.Errorf("workflow instance state = %q, want %q", wi.CurrentState, "in_progress")
	}

	// Verify transition was logged
	if len(ts.stRepo.transitions) != 1 {
		t.Fatalf("expected 1 logged transition, got %d", len(ts.stRepo.transitions))
	}
	if ts.stRepo.transitions[0].TriggeredBy != "user-1" {
		t.Errorf("TriggeredBy = %q, want %q", ts.stRepo.transitions[0].TriggeredBy, "user-1")
	}
}

func TestTriggerTransition_NotFound(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, nil, "open")

	_, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "nonexistent", "user", "human")
	if err == nil {
		t.Fatal("expected error for nonexistent transition")
	}
}

func TestTriggerTransition_WrongState(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "in_progress"}, {Name: "done"}},
		Transitions: []model.Transition{
			{Name: "start", FromState: "open", ToState: "in_progress"},
			{Name: "finish", FromState: "in_progress", ToState: "done"},
		},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
	ts := setupEngine(sm, nil, "open")

	_, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "finish", "user", "human")
	if err == nil {
		t.Fatal("expected error for wrong state")
	}
}

func TestTriggerTransition_TerminalState_Error(t *testing.T) {
	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, nil, "closed")

	_, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "close", "user", "human")
	if err == nil {
		t.Fatal("expected error for terminal state")
	}
}

func TestTriggerTransition_GuardedNotSatisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions: []model.Transition{
			{Name: "close", FromState: "open", ToState: "closed", CompletionRule: &model.CompletionRule{
				Type: "single", TaskRefs: []string{"review"},
			}},
		},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, []string{"review"}, "open")

	_, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "close", "user", "human")
	if err == nil {
		t.Fatal("expected error when completion rule not satisfied")
	}
}

func TestTriggerTransition_GuardedSatisfied(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions: []model.Transition{
			{Name: "close", FromState: "open", ToState: "closed", CompletionRule: &model.CompletionRule{
				Type: "single", TaskRefs: []string{"review"},
			}},
		},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	ts := setupEngine(sm, []string{"review"}, "open")
	ts.completeTask("review")

	result, err := ts.engine.TriggerTransition(context.Background(), ts.workflowInstanceID, "close", "user", "human")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ToState != "closed" {
		t.Errorf("ToState = %q, want %q", result.ToState, "closed")
	}
	if !result.IsTerminal {
		t.Error("expected terminal")
	}
}

func TestEvaluate_TransitionSetsWorkflowCompleted(t *testing.T) {
	sm := model.StateMachine{
		States: []model.State{{Name: "open"}, {Name: "done"}},
		Transitions: []model.Transition{
			{Name: "finish", FromState: "open", ToState: "done", CompletionRule: &model.CompletionRule{
				Type: "single", TaskRefs: []string{"task"},
			}},
		},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
	ts := setupEngine(sm, []string{"task"}, "open")
	ts.completeTask("task")

	ts.engine.Evaluate(context.Background(), ts.workflowInstanceID)

	wi := ts.workflowInstanceRepo.instances[ts.workflowInstanceID]
	if wi.Status != "completed" {
		t.Errorf("workflow instance status = %q, want %q", wi.Status, "completed")
	}
}
