# Phase 2: Core Engine + Services — Spec

## Scope
State machine engine, completion rule evaluation, service layer for definitions and instances, parent-child enforcement.

---

## 1. State Machine Engine

### Responsibilities
- Evaluate whether a transition is valid from the current state of a workflow instance.
- Check completion rules (are required tasks done?) before allowing guarded transitions.
- Apply transitions atomically (update workflow instance state + log the transition).
- Auto-evaluate: after a task completes, check if any transitions are now unblocked.

### Interface
```go
type Engine interface {
    // Evaluate checks all transitions from the current state.
    // If any transition's completion rule is satisfied, applies it and returns the result.
    // If multiple transitions are valid, applies the first one found (deterministic by definition order).
    // Returns nil result if no transitions are available.
    Evaluate(ctx context.Context, workflowInstanceID uuid.UUID) (*TransitionResult, error)

    // TriggerTransition attempts a specific named transition.
    // Fails if: transition doesn't exist, wrong current state, or completion rule not satisfied.
    TriggerTransition(ctx context.Context, workflowInstanceID uuid.UUID, transitionName string) (*TransitionResult, error)
}

type TransitionResult struct {
    WorkflowInstanceID uuid.UUID
    FromState          string
    ToState            string
    TransitionName     string
    IsTerminal         bool   // true if ToState is a terminal state
}
```

### Completion Rule Evaluation
Given a completion rule and a list of task instances for the workflow instance:

| Rule Type | Logic |
|-----------|-------|
| `single`  | The single referenced task must have status `completed` |
| `all_of`  | All referenced tasks must have status `completed` |
| `any_of`  | At least one referenced task must have status `completed` |
| `n_of_m`  | At least N of the referenced tasks must have status `completed` |

**Task matching**: CompletionRule.TaskRefs contains task definition **names**. The engine looks up task instances by matching their task_definition.name against these refs.

**No completion rule**: If a transition has no completion rule (nil), it can be triggered freely via `TriggerTransition` (manual/unguarded transition).

### Edge Cases
- Workflow instance in terminal state: no transitions allowed, return error.
- Workflow instance status is not `active`: no transitions allowed.
- Transition from wrong state: return error with current state info.
- Task ref in completion rule doesn't match any task definition: return error during evaluation.

---

## 2. Definition Services

### ProcessService
```go
type ProcessService interface {
    Create(ctx context.Context, input CreateProcessInput) (*model.Process, error)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Process, error)
    List(ctx context.Context, offset, limit int) ([]*model.Process, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateProcessInput) (*model.Process, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type CreateProcessInput struct {
    Name        string
    Description string
}

type UpdateProcessInput struct {
    Name        *string  // nil = don't change
    Description *string
    Status      *string
}
```

**Business rules:**
- Create: auto-assigns UUID, version=1, status="draft", timestamps.
- Update: validates new status if provided. Cannot update an archived process.
- Delete: only draft processes can be deleted. Active/archived must be archived first.

### WorkflowService
```go
type WorkflowService interface {
    Create(ctx context.Context, processID uuid.UUID, input CreateWorkflowInput) (*model.Workflow, error)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error)
    ListByProcessID(ctx context.Context, processID uuid.UUID) ([]*model.Workflow, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateWorkflowInput) (*model.Workflow, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type CreateWorkflowInput struct {
    Name         string
    Description  string
    StateMachine model.StateMachine
    SortOrder    int
}

type UpdateWorkflowInput struct {
    Name         *string
    Description  *string
    StateMachine *model.StateMachine
    SortOrder    *int
}
```

**Business rules:**
- Create: validates state machine, verifies process exists and is not archived.
- Update: validates new state machine if provided.
- Delete: only if parent process is in draft status.

### TaskService
```go
type TaskService interface {
    Create(ctx context.Context, workflowID uuid.UUID, input CreateTaskInput) (*model.Task, error)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
    ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*model.Task, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*model.Task, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

type CreateTaskInput struct {
    Name        string
    Description string
    TaskType    string
}

type UpdateTaskInput struct {
    Name        *string
    Description *string
    TaskType    *string
}
```

---

## 3. Instance Services

### ProcessInstanceService
```go
type ProcessInstanceService interface {
    // Start creates a new process instance from a process definition.
    // Also creates workflow instances and task instances for all workflows/tasks in the definition.
    Start(ctx context.Context, input StartProcessInput) (*model.ProcessInstance, error)

    GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error)
    List(ctx context.Context, offset, limit int) ([]*model.ProcessInstance, error)
    ListChildren(ctx context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error)
    Cancel(ctx context.Context, id uuid.UUID) error
}

type StartProcessInput struct {
    ProcessID uuid.UUID
    ParentID  *uuid.UUID  // optional, for child processes
}
```

**Business rules:**
- Start: process definition must be `active`. If ParentID is set, the parent must exist and must NOT itself have a parent (1-level enforcement).
- Start: creates all workflow instances (initial state from state machine) and task instances (status=pending).
- Cancel: sets status to `cancelled`, also cancels all active workflow instances.

### WorkflowInstanceService
```go
type WorkflowInstanceService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
    ListByProcessInstanceID(ctx context.Context, processInstanceID uuid.UUID) ([]*model.WorkflowInstance, error)
    TriggerTransition(ctx context.Context, id uuid.UUID, input TriggerTransitionInput) (*TransitionResult, error)
}

type TriggerTransitionInput struct {
    TransitionName string
    TriggeredBy    string
    TriggerType    string  // human, ai_agent, rule, process
}
```

### TaskInstanceService
```go
type TaskInstanceService interface {
    GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
    ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.TaskInstance, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateTaskStatusInput) (*model.TaskInstance, error)
    Assign(ctx context.Context, id uuid.UUID, assignedTo string) (*model.TaskInstance, error)
}

type UpdateTaskStatusInput struct {
    Status      string  // completed, failed, skipped
    TriggeredBy string
    TriggerType string
}
```

**Business rules:**
- UpdateStatus: after marking a task as `completed`, auto-evaluate the parent workflow instance's engine to check if any transitions are now unblocked.
- Cannot update status of a task whose workflow instance is not `active`.
- Status transitions: pending -> in_progress -> completed/failed/skipped. Cannot go backwards.

---

## 4. Test Plan

### Engine Unit Tests
- Evaluate with no available transitions -> nil result.
- Evaluate with unguarded transition -> applies it.
- Evaluate with `single` rule satisfied -> applies.
- Evaluate with `single` rule not satisfied -> nil.
- Evaluate with `all_of` rule, all done -> applies.
- Evaluate with `all_of` rule, partial -> nil.
- Evaluate with `any_of` rule, one done -> applies.
- Evaluate with `n_of_m` rule, exactly N done -> applies.
- Evaluate with `n_of_m` rule, less than N -> nil.
- TriggerTransition for nonexistent transition -> error.
- TriggerTransition from wrong state -> error.
- TriggerTransition on terminal state -> error.
- TriggerTransition on non-active workflow -> error.

### Service Unit Tests (mocked repos)
- ProcessService: create sets defaults, update validates status, delete only draft.
- WorkflowService: create validates state machine + process exists, delete only if process is draft.
- TaskService: create validates type, update validates type.
- ProcessInstanceService: start creates all instances, parent-child 1-level enforcement, cancel cascades.
- TaskInstanceService: complete triggers engine evaluation, invalid status transitions rejected.
