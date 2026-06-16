# Phase 2 — Engine + Services

## Goal
Implement workflow loader, service layer, and state machine engine. No HTTP. All business logic lives here.

---

## Package Layout

```
internal/
  engine/
    engine.go
    engine_test.go
  service/
    workflow_def.go
    instance.go
    task.go
    *_test.go
  workflowloader/
    loader.go
    loader_test.go
  testutil/
    mocks.go          hand-written repo mocks
```

---

## Workflow Loader (`internal/workflowloader/loader.go`)

```go
type LoadResult struct {
    Loaded   []LoadedDef
    Warnings []DefWarning
    Errors   []DefError
}

type Loader struct {
    dir     string
    defRepo repository.WorkflowDefRepo
    log     *slog.Logger
}

func (l *Loader) Load(ctx context.Context) (LoadResult, error)
```

**Algorithm:**
1. `os.ReadDir(l.dir)` — collect `*.json` files; skip non-JSON
2. For each file: unmarshal into `model.WorkflowDefinition`
3. Call `def.Validate()` — collect hard errors + warnings
4. Hard error → add to `result.Errors`, continue to next file
5. `repo.GetByNameVersion(name, version)`:
   - Not found → `Create`
   - Found, identical JSONB → skip (no-op, log info)
   - Found, different content → add to errors (version conflict)
6. Return `LoadResult`

**On startup** (`main.go`): `loader.Load(ctx)` runs. If any hard errors → `slog.Error` each + **`os.Exit(1)`** (fail boot — per Q8 decision). Warnings: `slog.Warn` each + continue.

**On `POST /admin/reload-defs`**: same `loader.Load(ctx)`, return result as JSON. Does NOT exit on error.

### Loader Tests
- Valid file loads successfully
- Invalid JSON → error
- Unknown task type → error
- Cycle detected → error
- Unreachable state → error
- Same `(name, version)` identical content → no-op
- Same `(name, version)` different content → error
- Warning cases: W1, W3
- Mixed good + bad files → good ones load, bad ones in errors

---

## Engine (`internal/engine/engine.go`)

```go
type TransitionMode int

const (
    ModeLinear    TransitionMode = iota
    ModeTaskBound
    ModeUserPicks
)

type Event struct {
    Type               string   // event type constants
    WorkflowInstanceID uuid.UUID
    Data               map[string]any
}

// Event type constants
const (
    EventTaskCreated          = "task.created"
    EventTaskCompleted        = "task.completed"
    EventTaskCancelled        = "task.cancelled"
    EventStateTransitioned    = "state.transitioned"
    EventInstanceStarted      = "instance.started"
    EventInstanceCompleted    = "instance.completed"
    EventInstanceCancelled    = "instance.cancelled"
)

type Engine struct {
    defRepo      repository.WorkflowDefRepo
    instanceRepo repository.WorkflowInstanceRepo
    taskRepo     repository.TaskInstanceRepo
    transRepo    repository.StateTransitionRepo
    pool         *pgxpool.Pool
    log          *slog.Logger
}

func (e *Engine) Start(ctx context.Context, defID uuid.UUID, createdBy string) (*model.WorkflowInstance, []Event, error)

func (e *Engine) Complete(ctx context.Context, taskID uuid.UUID, status model.TaskStatus, nextState string, createdBy string) (*model.WorkflowInstance, []Event, error)

func (e *Engine) Cancel(ctx context.Context, instanceID uuid.UUID) (*model.WorkflowInstance, []Event, error)
```

**`Start` flow:**
1. Load def from repo
2. Open tx
3. Create `workflow_instance` (status=active, current_state=start_state)
4. Spawn task_instances for start_state tasks
5. Commit
6. Return instance + `[EventInstanceStarted, ...EventTaskCreated per task]`

**`Complete` flow (see Architecture.md engine semantics):**
1. Open tx
2. `GetByIDForUpdate` on instance (row lock)
3. Validate instance active, task completable
4. Handle rejected → cancel instance
5. Handle done → derive mode → fire transition or wait
6. Commit
7. Return updated instance + events

**`Cancel` flow:**
1. Open tx
2. `GetByIDForUpdate` on instance
3. Validate instance active
4. Set instance status=cancelled, completed_at=now
5. Cancel all open tasks (status=cancelled)
6. Commit
7. Return instance + `[EventInstanceCancelled, ...EventTaskCancelled per cancelled task]`

### Engine Tests (`engine_test.go`)
Use mock repos from `testutil/mocks.go`. Test:
- Linear: single task done → transition fires
- Linear: two tasks, first done = wait, second done = transition
- Task-bound: complete bound task → correct edge, sibling cancelled
- Task-bound: two edges, different task keys each fire correct edge
- User-picks: no `next_state` → error
- User-picks: invalid `next_state` → error
- User-picks: valid `next_state` → transition fires
- Rejected task → instance cancelled, all open tasks cancelled
- Completing task on cancelled instance → `ErrInstanceTerminal`
- Completing already-done task → `ErrConflict`
- Reaching end_state → instance completed
- Start: tasks spawned, instance in start state
- Cancel: all open tasks cancelled

---

## Service Layer (`internal/service/`)

Services are thin. They:
- Validate inputs
- Call engine or repo
- Translate repo sentinel errors to service errors

```go
// workflow_def.go
type WorkflowDefService struct {
    repo   repository.WorkflowDefRepo
    loader *workflowloader.Loader
    log    *slog.Logger
}

func (s *WorkflowDefService) List(ctx, includeDeleted bool) ([]*model.WorkflowDef, error)
func (s *WorkflowDefService) GetByID(ctx, id uuid.UUID) (*model.WorkflowDef, error)
func (s *WorkflowDefService) Reload(ctx) (workflowloader.LoadResult, error)
func (s *WorkflowDefService) SoftDelete(ctx, id uuid.UUID) error  // blocks if active instances exist

// instance.go
type InstanceService struct {
    engine       *engine.Engine
    instanceRepo repository.WorkflowInstanceRepo
    taskRepo     repository.TaskInstanceRepo
    transRepo    repository.StateTransitionRepo
    log          *slog.Logger
}

func (s *InstanceService) Start(ctx, defID uuid.UUID, createdBy string) (*model.WorkflowInstance, []engine.Event, error)
func (s *InstanceService) GetByID(ctx, id uuid.UUID) (*model.WorkflowInstance, error)
func (s *InstanceService) List(ctx, filter repository.InstanceFilter) ([]*model.WorkflowInstance, int, error)
func (s *InstanceService) Cancel(ctx, id uuid.UUID) (*model.WorkflowInstance, []engine.Event, error)
func (s *InstanceService) SoftDelete(ctx, id uuid.UUID) error

// task.go
type TaskService struct {
    engine   *engine.Engine
    taskRepo repository.TaskInstanceRepo
    log      *slog.Logger
}

func (s *TaskService) GetByID(ctx, id uuid.UUID) (*model.TaskInstance, error)
func (s *TaskService) ListByInstance(ctx, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error)
func (s *TaskService) Start(ctx, id uuid.UUID) (*model.TaskInstance, error)     // in_progress
func (s *TaskService) Complete(ctx, id uuid.UUID, status model.TaskStatus, nextState string, createdBy string) (*model.WorkflowInstance, []engine.Event, error)
```

### Service Tests
Mock repos via `testutil/mocks.go`. Cover:
- Def list (active only, include deleted)
- Def get (not found)
- Reload delegates to loader
- Instance start (delegates to engine)
- Instance list with filters
- Instance cancel (delegates to engine)
- Task start (validates in_progress only from new/in_progress)
- Task complete (delegates to engine)
- All service errors translate correctly

---

## Testutil Mocks (`internal/testutil/mocks.go`)

Hand-written mock for each repo interface. Each mock records calls and returns configurable responses.

```go
type MockWorkflowDefRepo struct {
    CreateFn           func(ctx, def) error
    GetByIDFn          func(ctx, id) (*model.WorkflowDef, error)
    GetByNameVersionFn func(ctx, name, version) (*model.WorkflowDef, error)
    ListFn             func(ctx, includeDeleted) ([]*model.WorkflowDef, error)
    SoftDeleteFn       func(ctx, id) error
    // WithTx returns self by default
}
// ... similarly for all repo interfaces
```

---

## Sentinel Errors (service layer)

```go
var ErrInstanceTerminal = errors.New("instance is in terminal state")
var ErrTaskNotCompletable = errors.New("task is not in a completable state")
var ErrInvalidNextState = errors.New("next_state is not a valid outgoing transition")
var ErrDefHasActiveInstances = errors.New("workflow def has active instances")
```

---

## Exit Criteria
- [ ] `make test` green (all engine + service + loader unit tests)
- [ ] Loader: startup path fails boot on hard errors (manual verify)
- [ ] Loader: reload path returns errors without exit (test)
- [ ] Engine covers all 8 transition scenarios
- [ ] No `fmt.Println`
