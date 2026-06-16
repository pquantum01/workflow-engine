# Phase 3 — REST API Layer

## Goal
Wire HTTP handlers, middleware, webhook dispatcher+worker, and integration tests. After this phase, all backend functionality is usable via curl.

---

## Package Layout

```
internal/
  api/
    handler/
      workflow_def.go
      instance.go
      task.go
      admin.go
      health.go
    middleware/
      cors.go
      logging.go
      recovery.go
      requestid.go
    dto/
      workflow_def.go
      instance.go
      task.go
    static.go           serves embedded frontend SPA
    router.go           all route registrations
  webhook/
    dispatcher.go       engine.Event → outbox rows (called inside tx)
    worker.go           background goroutine, retry loop
    registry.go         webhooks.json loader
    signer.go           HMAC-SHA256
    worker_test.go
    signer_test.go
```

---

## Router (`internal/api/router.go`)

```go
func NewRouter(
    defSvc    *service.WorkflowDefService,
    instSvc   *service.InstanceService,
    taskSvc   *service.TaskService,
    loader    *workflowloader.Loader,
    log       *slog.Logger,
) *mux.Router
```

Route table:
```
GET    /api/v1/workflow-defs
GET    /api/v1/workflow-defs/{id}
POST   /api/v1/admin/reload-defs
POST   /api/v1/workflow-instances
GET    /api/v1/workflow-instances
GET    /api/v1/workflow-instances/{id}
POST   /api/v1/workflow-instances/{id}/cancel
DELETE /api/v1/workflow-instances/{id}
GET    /api/v1/workflow-instances/{id}/tasks
GET    /api/v1/workflow-instances/{id}/history
GET    /api/v1/task-instances/{id}
POST   /api/v1/task-instances/{id}/start
POST   /api/v1/task-instances/{id}/complete
GET    /api/v1/task-types
GET    /healthz
GET    /readyz
/*                                        SPA fallback (static.go)
```

Middleware chain (applied globally): requestID → slog logging → panic recovery → CORS.

---

## Middleware

**`requestid.go`**: generates UUID per request, injects into context + `X-Request-ID` response header. Logger reads it from context.

**`logging.go`**: logs `method`, `path`, `status`, `duration_ms`, `request_id`. Reads status from response writer wrapper.

**`recovery.go`**: catches panic, logs with `slog.Error`, returns `500 INTERNAL`.

**`cors.go`**: reads `CORS_ALLOWED_ORIGINS` env. Handles preflight + CORS headers.

---

## Request / Response DTOs (`internal/api/dto/`)

### Workflow Def
```go
// GET /workflow-defs response item
type WorkflowDefResponse struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Version     int             `json:"version"`
    Description string          `json:"description"`
    Definition  json.RawMessage `json:"definition"`   // full JSONB on detail; nil on list
    CreatedAt   time.Time       `json:"created_at"`
    DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

// GET /workflow-defs list response
type WorkflowDefListResponse struct {
    Items  []*WorkflowDefResponse `json:"items"`
    Total  int                    `json:"total"`
}
```

### Instance
```go
type StartInstanceRequest struct {
    WorkflowDefID  string         `json:"workflow_def_id" validate:"required,uuid"`
    CreatedBy      string         `json:"created_by"      validate:"required"`
    InitialContext map[string]any `json:"initial_context"`
}

type WorkflowInstanceResponse struct {
    ID             string          `json:"id"`
    WorkflowDefID  string          `json:"workflow_def_id"`
    CurrentStateID string          `json:"current_state_id"`
    Status         string          `json:"status"`
    Context        map[string]any  `json:"context"`
    StartedAt      time.Time       `json:"started_at"`
    CompletedAt    *time.Time      `json:"completed_at,omitempty"`
    // on detail endpoint only:
    Tasks          []*TaskInstanceResponse    `json:"tasks,omitempty"`
    History        []*StateTransitionResponse `json:"history,omitempty"`
}

type InstanceListResponse struct {
    Items  []*WorkflowInstanceResponse `json:"items"`
    Total  int                         `json:"total"`
    Limit  int                         `json:"limit"`
    Offset int                         `json:"offset"`
}
```

### Task
```go
type CompleteTaskRequest struct {
    Status      string `json:"status"     validate:"required,oneof=done rejected"`
    NextState   string `json:"next_state"`
    CompletedBy string `json:"completed_by"`
    Outcome     string `json:"outcome"`      // v2 use; persisted if present
    Data        any    `json:"data"`         // v2 use; persisted if present
}

type TaskInstanceResponse struct {
    ID                 string     `json:"id"`
    WorkflowInstanceID string     `json:"workflow_instance_id"`
    StateID            string     `json:"state_id"`
    TaskKey            string     `json:"task_key"`
    TaskLabel          string     `json:"task_label"`
    TaskType           string     `json:"task_type"`
    Status             string     `json:"status"`
    Outcome            *string    `json:"outcome,omitempty"`
    CreatedBy          string     `json:"created_by"`
    CreatedAt          time.Time  `json:"created_at"`
    CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

type StateTransitionResponse struct {
    ID             string    `json:"id"`
    FromState      string    `json:"from_state"`
    ToState        string    `json:"to_state"`
    TransitionedAt time.Time `json:"transitioned_at"`
}
```

---

## Error Handling (`internal/api/handler/response.go`)

```go
type ErrorResponse struct {
    Error ErrorBody `json:"error"`
}

type ErrorBody struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, status int, code, message string, details any)
func writeJSON(w http.ResponseWriter, status int, v any)
```

Service error → HTTP status mapping:
```
service.ErrNotFound          → 404 NOT_FOUND
service.ErrConflict          → 409 CONFLICT
service.ErrInstanceTerminal  → 409 INSTANCE_TERMINAL
service.ErrTaskNotCompletable→ 409 TASK_NOT_COMPLETABLE
service.ErrInvalidNextState  → 422 INVALID_NEXT_STATE
validation errors            → 422 VALIDATION_FAILED
panic (via middleware)       → 500 INTERNAL
```

Error messages must NOT expose internal stack traces or raw SQL.

---

## Webhook Dispatcher

### Registry (`webhook/registry.go`)
```go
type Subscriber struct {
    URL              string   `json:"url"`
    Secret           string   `json:"secret"`
    EventTypes       []string `json:"event_types"`
    WorkflowDefNames []string `json:"workflow_def_names"` // empty = all
}

func LoadRegistry(path string) ([]Subscriber, error)
```

### Dispatcher (`webhook/dispatcher.go`)
Called inside engine transaction (receives `tx DB`):
```go
func Dispatch(ctx context.Context, tx repository.DB, outboxRepo repository.WebhookOutboxRepo,
    subscribers []Subscriber, events []engine.Event, defName string) error
```
Filters subscribers by event type + def name. Creates one outbox row per (subscriber, event) pair.

### Signer (`webhook/signer.go`)
```go
func Sign(secret string, body []byte) string  // returns "sha256=<hex>"
```

### Worker (`webhook/worker.go`)
```go
type Worker struct {
    outboxRepo  repository.WebhookOutboxRepo
    subscribers []Subscriber
    client      *http.Client
    interval    time.Duration
    maxAttempts int
    log         *slog.Logger
}

func (w *Worker) Run(ctx context.Context)   // ticker loop; stops on ctx cancel
func (w *Worker) Stop(ctx context.Context)  // graceful drain
```

**Tick logic:**
1. `SELECT … FOR UPDATE SKIP LOCKED WHERE next_attempt_at <= now() AND delivered_at IS NULL AND failed_at IS NULL LIMIT 20`
2. For each row: POST to `target_url`, timeout 10s
3. 2xx → `MarkDelivered`
4. Non-2xx / error → `IncrementAttempt(nextAttemptAt)` using backoff schedule; if `attempts >= maxAttempts` → `MarkFailed`

**Backoff schedule (attempt → delay before next):**
1 → 1m, 2 → 5m, 3 → 30m, 4 → 2h, 5 → 12h, 6 → dead

### Worker Tests
- Successful delivery marks delivered
- Non-2xx increments attempt + schedules next
- After maxAttempts marks failed
- Signer produces correct HMAC

---

## Static SPA Server (`internal/api/static.go`)

```go
//go:embed ../../webroot
var webroot embed.FS

func RegisterStatic(r *mux.Router) {
    // serve /api/* handled by routes above
    // all other paths: serve from embed, fallback to index.html (SPA routing)
}
```

`webroot/` populated during Docker build (copy of `frontend/dist`). In local dev, Vite dev server proxies `/api` → backend; static serving not needed.

---

## Integration Tests (`backend/tests/integration/`)

Test against real Postgres + real HTTP (using `httptest.NewServer`).

**`api_test.go`** covers (each test truncates tables first):
- `POST /workflow-instances` → 201 + instance in start state
- `GET /workflow-instances/:id` → includes tasks + history
- `POST /task-instances/:id/complete` (linear, all done → transition)
- `POST /task-instances/:id/complete` (rejected → instance cancelled)
- `POST /task-instances/:id/complete` on cancelled instance → 409
- `POST /workflow-instances/:id/cancel` → instance cancelled
- `GET /workflow-defs` → list
- `POST /admin/reload-defs` → 200 + result
- `GET /healthz` → 200
- `GET /readyz` → 200
- 404 on unknown IDs
- Validation errors → 422

---

## `main.go` Wire-Up

```go
func main() {
    cfg := config.Load()
    log := config.NewLogger(cfg)

    pool := mustConnectDB(cfg.DatabaseURL)
    defer pool.Close()

    // repos
    defRepo  := repository.NewWorkflowDefRepo(pool)
    instRepo := repository.NewWorkflowInstanceRepo(pool)
    taskRepo := repository.NewTaskInstanceRepo(pool)
    transRepo:= repository.NewStateTransitionRepo(pool)
    outboxRepo:= repository.NewWebhookOutboxRepo(pool)

    // webhook
    subscribers, _ := webhook.LoadRegistry(cfg.WebhookConfig)
    worker := webhook.NewWorker(outboxRepo, subscribers, cfg, log)

    // loader
    loader := workflowloader.New(cfg.WorkflowsDir, defRepo, log)
    if _, err := loader.Load(ctx); err != nil { os.Exit(1) }  // hard errors

    // engine
    eng := engine.New(defRepo, instRepo, taskRepo, transRepo, pool, log)

    // services
    defSvc  := service.NewWorkflowDefService(defRepo, loader, log)
    instSvc := service.NewInstanceService(eng, instRepo, taskRepo, transRepo, outboxRepo, subscribers, log)
    taskSvc := service.NewTaskService(eng, taskRepo, outboxRepo, subscribers, log)

    // server
    r := api.NewRouter(defSvc, instSvc, taskSvc, loader, log)
    srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r}

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go worker.Run(ctx)
    go srv.ListenAndServe()

    <-ctx.Done()
    shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(shutCtx)
    worker.Stop(shutCtx)
}
```

---

## Exit Criteria
- [ ] `make test-integration` green (all API tests)
- [ ] `curl http://localhost:8080/healthz` → 200
- [ ] Webhook worker delivers to a test receiver (manual verify or integration test with httptest server)
- [ ] No internal errors leak in error responses (checked in tests)
- [ ] No `fmt.Println`
