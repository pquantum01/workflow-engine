# Phase 1 — Foundation

## Goal
Establish the complete data layer: repo scaffold, DB schema + migrations, domain models, repository interfaces + Postgres implementations, integration tests.

No business logic in this phase. No HTTP. No engine.

---

## Deliverables

### 1. Repo Scaffold
- `go.work` (Go workspace, single module `./backend`)
- `backend/go.mod` with all Phase 1 deps
- `docker-compose.yml` (Postgres 16 only, named volumes, healthcheck)
- `Makefile` with targets: `build`, `run`, `test`, `test-integration`, `docker-up`, `docker-down`, `migrate-up`, `migrate-up-test`, `migrate-down`
- `.gitignore`

### 2. Migrations
`backend/migrations/001_create_base_tables.up.sql`:
- `workflow_def`
- `workflow_instance`
- `task_instance`
- `state_transition`

`backend/migrations/002_webhook_outbox.up.sql`:
- `webhook_outbox`

All with corresponding `.down.sql` reversal.

See `Architecture.md` for full DDL.

### 3. Domain Models (`backend/internal/model/`)

**`workflow_def.go`**
```go
type WorkflowDef struct {
    ID          uuid.UUID
    Name        string
    Version     int
    Description string
    Definition  WorkflowDefinition  // parsed from JSONB
    CreatedAt   time.Time
    DeletedAt   *time.Time
}

type WorkflowDefinition struct {
    Name        string
    Version     int
    Description string
    StartState  string
    EndState    string
    States      []StateDef
    Transitions []TransitionDef
}

type StateDef struct {
    ID    string
    Label string
    Tasks []TaskDef
}

type TaskDef struct {
    Key   string
    Label string
    Type  string
}

type TransitionDef struct {
    FromState string
    ToState   string
    Trigger   *TransitionTrigger  // nil = no trigger
}

type TransitionTrigger struct {
    TaskKey string
}

func (d *WorkflowDefinition) Validate() (warnings []string, err error)
func (d *WorkflowDefinition) TransitionMode(stateID string) TransitionMode
func (d *WorkflowDefinition) OutgoingTransitions(stateID string) []TransitionDef
func (d *WorkflowDefinition) FindState(stateID string) (*StateDef, bool)
```

**`workflow_instance.go`**
```go
type WorkflowInstance struct {
    ID             uuid.UUID
    WorkflowDefID  uuid.UUID
    CurrentStateID string
    Status         InstanceStatus  // active | completed | cancelled
    Context        map[string]any
    StartedAt      time.Time
    CompletedAt    *time.Time
    DeletedAt      *time.Time
}
```

**`task_instance.go`**
```go
type TaskInstance struct {
    ID                  uuid.UUID
    WorkflowInstanceID  uuid.UUID
    StateID             string
    TaskKey             string
    TaskLabel           string
    TaskType            string
    Status              TaskStatus  // new | in_progress | done | rejected | cancelled
    Outcome             *string
    Data                []byte      // JSONB
    CreatedBy           string
    CreatedAt           time.Time
    CompletedAt         *time.Time
    DeletedAt           *time.Time
}
```

**`state_transition.go`**
```go
type StateTransition struct {
    ID                 uuid.UUID
    WorkflowInstanceID uuid.UUID
    FromState          string
    ToState            string
    TransitionedAt     time.Time
}
```

All models have unit tests for constructor helpers and any validation methods.

### 4. Repository Interfaces (`backend/internal/repository/interfaces.go`)

```go
type DB interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type WorkflowDefRepo interface {
    Create(ctx context.Context, def *model.WorkflowDef) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error)
    GetByNameVersion(ctx context.Context, name string, version int) (*model.WorkflowDef, error)
    List(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error)
    SoftDelete(ctx context.Context, id uuid.UUID) error
    WithTx(tx DB) WorkflowDefRepo
}

type WorkflowInstanceRepo interface {
    Create(ctx context.Context, inst *model.WorkflowInstance) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error)
    GetByIDForUpdate(ctx context.Context, id uuid.UUID, tx DB) (*model.WorkflowInstance, error)
    List(ctx context.Context, filter InstanceFilter) ([]*model.WorkflowInstance, int, error)
    UpdateState(ctx context.Context, id uuid.UUID, newStateID string) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status model.InstanceStatus, completedAt *time.Time) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    WithTx(tx DB) WorkflowInstanceRepo
}

type TaskInstanceRepo interface {
    BulkCreate(ctx context.Context, tasks []*model.TaskInstance) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error)
    ListByInstance(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error)
    ListByInstanceAndState(ctx context.Context, instanceID uuid.UUID, stateID string) ([]*model.TaskInstance, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, completedAt *time.Time) error
    UpdateOutcome(ctx context.Context, id uuid.UUID, outcome string, data []byte) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    WithTx(tx DB) TaskInstanceRepo
}

type StateTransitionRepo interface {
    Append(ctx context.Context, t *model.StateTransition) error
    ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error)
    WithTx(tx DB) StateTransitionRepo
}

type WebhookOutboxRepo interface {
    BulkInsert(ctx context.Context, rows []*model.WebhookOutboxRow) error
    ListPending(ctx context.Context, now time.Time, limit int) ([]*model.WebhookOutboxRow, error)
    MarkDelivered(ctx context.Context, id uuid.UUID) error
    MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error
    IncrementAttempt(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time) error
    WithTx(tx DB) WebhookOutboxRepo
}
```

Sentinel errors in `repository/errors.go`:
```go
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
```

### 5. Postgres Implementations (`backend/internal/repository/`)
One file per entity (flat, no subfolder). Each struct holds `db DB`, implements interface, `WithTx` returns new instance with tx.

### 6. Integration Tests (`backend/tests/integration/`)
- `setup_test.go`: DB connection, truncate helper, test main
- `workflow_def_test.go`: CRUD + soft delete + unique constraint
- `workflow_instance_test.go`: CRUD + state/status updates + filters + FOR UPDATE
- `task_instance_test.go`: bulk create + list by instance/state + status updates
- `state_transition_test.go`: append + list

---

## Dependency List (`go.mod`)
```
github.com/lib/pq
github.com/google/uuid
github.com/golang-migrate/migrate/v4
github.com/gorilla/mux                  (Phase 3 but declare now)
github.com/go-playground/validator/v10  (Phase 3 but declare now)
github.com/stretchr/testify
```

---

## Makefile Targets (Phase 1 subset)
```makefile
BINARY=./bin/server
DB_URL=postgres://workflow:workflow@localhost:5432/workflow?sslmode=disable
TEST_DB_URL=postgres://workflow:workflow@localhost:5432/workflow_test?sslmode=disable

build:
	go build -o $(BINARY) ./backend/cmd/server

test:
	go test ./backend/...

test-integration:
	DATABASE_URL=$(TEST_DB_URL) go test ./backend/tests/integration/... -v -count=1

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

migrate-up:
	migrate -path backend/migrations -database "$(DB_URL)" up

migrate-up-test:
	migrate -path backend/migrations -database "$(TEST_DB_URL)" up

migrate-down:
	migrate -path backend/migrations -database "$(DB_URL)" down
```

---

## Exit Criteria
- [ ] `make docker-up && make migrate-up && make migrate-up-test` succeeds
- [ ] `make test` green (model unit tests)
- [ ] `make test-integration` green (all repo tests)
- [ ] No `fmt.Println` in any Go file
