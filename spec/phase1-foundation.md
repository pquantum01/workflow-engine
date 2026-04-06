# Phase 1: Backend Foundation — Spec

## Scope
Project scaffolding, database migrations, domain models, repository layer with PostgreSQL, and all associated tests.

---

## 1. Domain Models

### 1.1 Process (Definition)
```go
type Process struct {
    ID          uuid.UUID
    Name        string     // required, max 255
    Description string     // optional
    Version     int        // default 1, positive
    Status      string     // draft | active | archived
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```
**Validation rules:**
- Name must be non-empty, max 255 chars.
- Status must be one of: `draft`, `active`, `archived`.
- Version must be >= 1.

### 1.2 Workflow (Definition)
```go
type Workflow struct {
    ID           uuid.UUID
    ProcessID    uuid.UUID    // required, FK to process
    Name         string       // required, max 255
    Description  string       // optional
    StateMachine StateMachine // required, stored as JSONB
    SortOrder    int          // default 0
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 1.3 StateMachine (embedded JSONB)
```go
type StateMachine struct {
    States         []State      // list of valid states
    Transitions    []Transition // list of valid transitions
    InitialState   string       // must reference a state name
    TerminalStates []string     // must reference state names
}

type State struct {
    Name        string            // unique within the machine
    Description string
    Metadata    map[string]string // arbitrary key-value
}

type Transition struct {
    Name           string          // unique within the machine
    FromState      string          // must reference a state name
    ToState        string          // must reference a state name
    CompletionRule *CompletionRule // optional — guards the transition
}

type CompletionRule struct {
    Type     string   // single | all_of | any_of | n_of_m
    TaskRefs []string // references to task definition names within the workflow
    N        int      // only used for n_of_m
}
```
**Validation rules:**
- InitialState must exist in States.
- TerminalStates must all exist in States.
- Transition FromState/ToState must exist in States.
- State names unique within a machine.
- Transition names unique within a machine.
- CompletionRule.Type must be one of: `single`, `all_of`, `any_of`, `n_of_m`.
- For `n_of_m`: N must be > 0 and <= len(TaskRefs).
- For `single`: TaskRefs must have exactly 1 element.

### 1.4 Task (Definition)
```go
type Task struct {
    ID          uuid.UUID
    WorkflowID  uuid.UUID // required, FK to workflow
    Name        string    // required, max 255
    Description string    // optional
    TaskType    string    // manual | automated | approval
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```
**Validation rules:**
- Name must be non-empty, max 255 chars.
- TaskType must be one of: `manual`, `automated`, `approval`.

### 1.5 ProcessInstance
```go
type ProcessInstance struct {
    ID          uuid.UUID
    ProcessID   uuid.UUID  // FK to process
    ParentID    *uuid.UUID // nullable, FK to process_instances (1-level only)
    Status      string     // running | completed | cancelled | failed
    StartedAt   time.Time
    CompletedAt *time.Time // nullable
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```
**Validation rules:**
- Status must be one of: `running`, `completed`, `cancelled`, `failed`.
- ParentID, if set, must reference a process instance that itself has no parent (1-level rule).

### 1.6 WorkflowInstance
```go
type WorkflowInstance struct {
    ID                uuid.UUID
    WorkflowID        uuid.UUID // FK to workflow
    ProcessInstanceID uuid.UUID // FK to process_instances
    CurrentState      string    // must be a valid state in the workflow's state machine
    Status            string    // active | completed | cancelled
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

### 1.7 TaskInstance
```go
type TaskInstance struct {
    ID                 uuid.UUID
    TaskID             uuid.UUID  // FK to task
    WorkflowInstanceID uuid.UUID  // FK to workflow_instances
    Status             string     // pending | in_progress | completed | skipped | failed
    AssignedTo         string     // optional
    CompletedAt        *time.Time // nullable
    CreatedAt          time.Time
    UpdatedAt          time.Time
}
```

### 1.8 StateTransition (audit log)
```go
type StateTransition struct {
    ID                 uuid.UUID
    WorkflowInstanceID uuid.UUID
    FromState          string
    ToState            string
    TransitionName     string // optional
    TriggeredBy        string // user ID, agent ID, process ID
    TriggerType        string // human | ai_agent | rule | process
    CreatedAt          time.Time
}
```

---

## 2. Repository Interfaces

Each repository follows the same pattern: interface in `internal/repository/`, implementation in `internal/repository/postgres/`.

### 2.1 ProcessRepository
```go
type ProcessRepository interface {
    Create(ctx context.Context, process *model.Process) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.Process, error)
    List(ctx context.Context, offset, limit int) ([]*model.Process, error)
    Update(ctx context.Context, process *model.Process) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### 2.2 WorkflowRepository
```go
type WorkflowRepository interface {
    Create(ctx context.Context, workflow *model.Workflow) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error)
    ListByProcessID(ctx context.Context, processID uuid.UUID) ([]*model.Workflow, error)
    Update(ctx context.Context, workflow *model.Workflow) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### 2.3 TaskRepository
```go
type TaskRepository interface {
    Create(ctx context.Context, task *model.Task) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
    ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*model.Task, error)
    Update(ctx context.Context, task *model.Task) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### 2.4 ProcessInstanceRepository
```go
type ProcessInstanceRepository interface {
    Create(ctx context.Context, instance *model.ProcessInstance) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error)
    List(ctx context.Context, offset, limit int) ([]*model.ProcessInstance, error)
    ListByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error)
    Update(ctx context.Context, instance *model.ProcessInstance) error
}
```

### 2.5 WorkflowInstanceRepository
```go
type WorkflowInstanceRepository interface {
    Create(ctx context.Context, instance *model.WorkflowInstance) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
    ListByProcessInstanceID(ctx context.Context, processInstanceID uuid.UUID) ([]*model.WorkflowInstance, error)
    Update(ctx context.Context, instance *model.WorkflowInstance) error
}
```

### 2.6 TaskInstanceRepository
```go
type TaskInstanceRepository interface {
    Create(ctx context.Context, instance *model.TaskInstance) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
    ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.TaskInstance, error)
    Update(ctx context.Context, instance *model.TaskInstance) error
}
```

### 2.7 StateTransitionRepository
```go
type StateTransitionRepository interface {
    Create(ctx context.Context, transition *model.StateTransition) error
    ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.StateTransition, error)
}
```

**Error conventions:**
- `GetByID` returns a sentinel `ErrNotFound` when no row matches.
- `Delete` returns `ErrNotFound` if the row doesn't exist.
- Foreign key violations return `ErrForeignKey`.
- All methods accept `context.Context` for cancellation/timeouts.

---

## 3. Test Plan

### Unit Tests (no database)
- **Model validation**: Test `Validate()` on each model — valid inputs pass, invalid inputs return descriptive errors.
- **StateMachine validation**: Duplicate states, missing initial state, invalid transitions, completion rule edge cases.

### Integration Tests (real PostgreSQL via Docker)
- **Repository CRUD**: For each repository — create, read, list, update, delete.
- **Foreign key enforcement**: Creating a workflow with nonexistent process_id fails.
- **Parent-child constraint**: Creating a grandchild process instance fails.
- **List filtering**: ListByProcessID, ListByWorkflowID, etc. return correct subsets.
- **ErrNotFound**: GetByID with random UUID returns ErrNotFound.

---

## 4. Infrastructure

### Docker Compose
- PostgreSQL 16 on port 5432.
- Volume for data persistence in local dev.

### Makefile Targets
- `make build` — Build the Go binary.
- `make test` — Run unit tests.
- `make test-integration` — Run integration tests (requires PostgreSQL).
- `make migrate-up` — Apply all migrations.
- `make migrate-down` — Rollback last migration.
- `make docker-up` — Start Docker Compose services.
- `make docker-down` — Stop Docker Compose services.
