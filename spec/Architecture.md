# Workflow Engine — Architecture

## Tech Stack

| Layer | Choice |
|-------|--------|
| Backend language | Go 1.22+ |
| HTTP router | `gorilla/mux` |
| DB driver | `database/sql` + `github.com/lib/pq` |
| Query style | Raw SQL (no ORM) |
| Migrations | `golang-migrate/migrate` (`.up.sql` / `.down.sql`) |
| UUID | `github.com/google/uuid` |
| Validation (HTTP DTOs) | `github.com/go-playground/validator/v10` |
| Validation (workflow def) | Hand-written Go (richer rules) |
| Logger | `log/slog` JSON handler (no `fmt.Println`) |
| Test framework | `testing` + `github.com/stretchr/testify` |
| Database | PostgreSQL 16 |
| Frontend | React 18 + TypeScript + Vite |
| Styling | Tailwind CSS + shadcn/ui |
| Server state | TanStack Query v5 |
| Routing | React Router v6 |
| Graph viz | React Flow + dagre |
| Forms | react-hook-form + zod |
| Icons | lucide-react |
| Animation | Framer Motion |
| Frontend tests | Vitest + React Testing Library |
| E2e tests | Playwright + Chromium |
| Deployment | Single Alpine container (Go binary + embedded frontend dist) |

---

## Repository Layout

```
workflow-engine/
├── CLAUDE.md
├── README.md
├── Makefile
├── go.work                             Go workspace (backend module)
├── docker-compose.yml                  Postgres + app (full stack)
├── docker-compose.e2e.yml              E2e test stack (dedicated ports)
├── .gitignore
├── workflows/                          Sample JSON workflow defs (mounted in container)
│   ├── leave-approval.json
│   └── bug-triage.json
├── webhooks.json.example               Template (webhooks.json git-ignored)
├── scripts/                            (deprecated; smoke tests moved to tests/)
├── spec/                               All documentation
│   ├── Requirement.md
│   ├── Architecture.md
│   ├── Currentstatus.md
│   ├── glossary.md
│   ├── phase0-spec-docs.md
│   ├── phase1-foundation.md
│   ├── phase2-engine-services.md
│   ├── phase3-api-layer.md
│   ├── phase4-frontend-admin.md
│   ├── phase4-design.md
│   ├── phase5-frontend-operations.md
│   └── phase6-polish-deploy.md
├── backend/
│   ├── Dockerfile                      Multi-stage: node build + go build + alpine runtime
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/server/main.go              Entry point; wires DI, starts server
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/                HTTP handlers (thin — delegate to service)
│   │   │   ├── middleware/             CORS, request logging, panic recovery, request-id
│   │   │   ├── dto/                    Request / response structs
│   │   │   ├── static.go              Serves embedded frontend SPA with fallback
│   │   │   └── router.go              gorilla/mux route registration
│   │   ├── config/
│   │   │   ├── config.go              Env var loading
│   │   │   └── logger.go              slog setup
│   │   ├── engine/
│   │   │   ├── engine.go              State machine logic
│   │   │   └── engine_test.go
│   │   ├── model/
│   │   │   ├── workflow_def.go        WorkflowDef + nested types + Validate()
│   │   │   ├── workflow_instance.go
│   │   │   ├── task_instance.go
│   │   │   ├── state_transition.go
│   │   │   └── *_test.go
│   │   ├── repository/                 Interfaces + Postgres implementations (flat)
│   │   │   ├── interfaces.go
│   │   │   ├── errors.go              Sentinel errors
│   │   │   ├── workflow_def.go
│   │   │   ├── workflow_instance.go
│   │   │   ├── task_instance.go
│   │   │   ├── task_comment.go
│   │   │   ├── task_attachment.go
│   │   │   ├── state_transition.go
│   │   │   ├── webhook_outbox.go
│   │   │   └── helpers.go
│   │   ├── service/
│   │   │   ├── workflow_def.go        Load, reload, list, get, delete
│   │   │   ├── instance.go            Start, get, list, cancel, delete
│   │   │   ├── task.go                List, get, start, complete, update, add-adhoc
│   │   │   ├── comment.go             Add, list comments
│   │   │   ├── attachment.go          Upload, download, delete (S3-gated)
│   │   │   └── *_test.go
│   │   ├── storage/
│   │   │   ├── interface.go           FileStore interface (Upload/Download/Delete)
│   │   │   └── s3.go                  AWS S3 implementation (wired only when S3_ENABLED=true)
│   │   ├── webhook/
│   │   │   ├── dispatcher.go          Event → outbox rows (in-tx helper)
│   │   │   ├── worker.go              Background goroutine, retry logic
│   │   │   ├── registry.go            webhooks.json loader
│   │   │   └── signer.go              HMAC-SHA256
│   │   ├── workflowloader/
│   │   │   ├── loader.go              File scan → validate → upsert
│   │   │   └── loader_test.go
│   │   └── testutil/
│   │       └── mocks.go               Hand-written repo mocks
│   ├── migrations/
│   │   ├── 001_create_base_tables.up.sql
│   │   ├── 001_create_base_tables.down.sql
│   │   ├── 002_webhook_outbox.up.sql
│   │   └── 002_webhook_outbox.down.sql
│   ├── webroot/                        Embed target (frontend/dist copied here at build)
│   └── tests/integration/
│       ├── setup_test.go
│       ├── workflow_def_test.go
│       ├── instance_test.go
│       └── task_test.go
├── frontend/
│   ├── package.json
│   ├── vite.config.ts                  Dev proxy /api → localhost:8080
│   ├── tsconfig*.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── index.html
│   ├── public/
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api/                        client.ts + per-resource modules
│       ├── components/
│       │   ├── ui/                     shadcn components
│       │   ├── tasks/                  Task renderer registry + per-type renderers
│       │   ├── graph/                  ReactFlow wrappers, dagre layout
│       │   └── layout/                 Shell, sidebar, topbar
│       ├── hooks/
│       ├── pages/
│       ├── types/
│       └── lib/
└── tests/
    ├── smoke/
    │   └── smoke.sh                    curl-based end-to-end smoke
    └── e2e/
        ├── playwright.config.ts
        ├── package.json
        ├── fixtures/
        └── specs/
```

---

## Data Model

### `workflow_def`
```sql
CREATE TABLE workflow_def (
  id          UUID PRIMARY KEY,
  name        TEXT NOT NULL,
  version     INT  NOT NULL,
  description TEXT,
  definition  JSONB NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ,
  UNIQUE (name, version) WHERE deleted_at IS NULL
);
```

### `workflow_instance`
```sql
CREATE TABLE workflow_instance (
  id                UUID PRIMARY KEY,
  workflow_def_id   UUID NOT NULL REFERENCES workflow_def(id),
  current_state_id  TEXT NOT NULL,
  status            TEXT NOT NULL CHECK (status IN ('active','completed','cancelled')),
  context           JSONB NOT NULL DEFAULT '{}',
  started_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at      TIMESTAMPTZ,
  deleted_at        TIMESTAMPTZ
);
CREATE INDEX ON workflow_instance (workflow_def_id) WHERE deleted_at IS NULL;
CREATE INDEX ON workflow_instance (status)          WHERE deleted_at IS NULL;
```

### `task_instance`
```sql
CREATE TABLE task_instance (
  id                      UUID PRIMARY KEY,
  workflow_instance_id    UUID NOT NULL REFERENCES workflow_instance(id),
  state_id                TEXT NOT NULL,
  task_key                TEXT NOT NULL,
  task_label              TEXT NOT NULL,
  task_type               TEXT NOT NULL,
  status                  TEXT NOT NULL CHECK (status IN ('new','in_progress','done','rejected','cancelled')),
  outcome                 TEXT,
  data                    JSONB,
  created_by              TEXT,
  priority                TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low','medium','high')),
  expected_completion_date DATE,
  expected_completion_time TIME,
  is_adhoc                BOOLEAN NOT NULL DEFAULT FALSE,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at            TIMESTAMPTZ,
  deleted_at              TIMESTAMPTZ,
  UNIQUE (workflow_instance_id, state_id, task_key)
);
CREATE INDEX ON task_instance (workflow_instance_id, status) WHERE deleted_at IS NULL;
```

### `task_comment`
```sql
CREATE TABLE task_comment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL REFERENCES task_instance(id),
  author           TEXT NOT NULL,
  content          VARCHAR(2048) NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX task_comment_task_idx ON task_comment (task_instance_id, created_at);
```

### `task_attachment`
```sql
CREATE TABLE task_attachment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL UNIQUE,
  filename         TEXT NOT NULL,
  content_type     TEXT NOT NULL,
  size_bytes       INT  NOT NULL,
  s3_url           TEXT NOT NULL,
  uploaded_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
```
One attachment per task (UNIQUE on `task_instance_id`). Upload replaces existing. `s3_url` is the full internal S3 URL — never exposed to clients.

### `state_transition` (append-only audit)
```sql
CREATE TABLE state_transition (
  id                    UUID PRIMARY KEY,
  workflow_instance_id  UUID NOT NULL REFERENCES workflow_instance(id),
  from_state            TEXT NOT NULL,
  to_state              TEXT NOT NULL,
  transitioned_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON state_transition (workflow_instance_id, transitioned_at);
```

### `webhook_outbox`
```sql
CREATE TABLE webhook_outbox (
  id              UUID PRIMARY KEY,
  event_type      TEXT NOT NULL,
  payload         JSONB NOT NULL,
  target_url      TEXT NOT NULL,
  attempts        INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  delivered_at    TIMESTAMPTZ,
  failed_at       TIMESTAMPTZ,
  last_error      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON webhook_outbox (next_attempt_at) WHERE delivered_at IS NULL AND failed_at IS NULL;
```

---

## Workflow Definition JSON Schema

```json
{
  "name": "leave-approval",
  "version": 1,
  "description": "...",
  "start_state": "submitted",
  "end_state": "approved",
  "states": [
    {
      "id": "submitted",
      "label": "Submitted",
      "tasks": [
        {
          "key": "manager_review",
          "label": "Manager Review",
          "type": "approval"
        }
      ]
    },
    {
      "id": "hr_review",
      "label": "HR Review",
      "tasks": [
        { "key": "hr_check",      "label": "HR Verification",  "type": "approval"  },
        { "key": "policy_check",  "label": "Policy Compliance", "type": "checklist" }
      ]
    },
    {
      "id": "approved",
      "label": "Approved",
      "tasks": []
    }
  ],
  "transitions": [
    { "from_state": "submitted", "to_state": "hr_review" },
    { "from_state": "hr_review", "to_state": "approved" }
  ]
}
```

Multi-edge (task-bound) example:
```json
"transitions": [
  { "from_state": "manager_review", "to_state": "hr_review",
    "trigger": { "task_key": "approve" } },
  { "from_state": "manager_review", "to_state": "denied",
    "trigger": { "task_key": "reject" } }
]
```

Multi-edge (user-picks) example (state must have exactly 1 task):
```json
"transitions": [
  { "from_state": "triage", "to_state": "fast_track" },
  { "from_state": "triage", "to_state": "standard_queue" },
  { "from_state": "triage", "to_state": "rejected" }
]
```

---

## Validation Rules

### Hard Errors (reject load)
1. `name` non-empty, matches `^[a-z][a-z0-9-]*$`
2. `version` integer ≥ 1
3. `start_state` and `end_state` exist in `states[]`
4. `states[].id` unique within def, matches `^[a-z][a-z0-9_]*$`
5. `states[].label` non-empty
6. `states[].tasks[].key` unique within state, matches `^[a-z][a-z0-9_]*$`
7. `states[].tasks[].label` non-empty
8. `states[].tasks[].type` ∈ `{approval, form, checklist, manual}`
9. End state has zero tasks
10. Non-end state has ≥1 task
11. Transition `from_state` and `to_state` resolve to existing state IDs
12. Non-end state has ≥1 outgoing transition
13. End state has zero outgoing transitions
14. All states reachable from start (BFS)
15. Every state has path to end_state (reverse BFS from end)
16. No cycles (DFS)
17. Mixed triggers on same state's edges → reject (ambiguous)
18. Task-bound mode: each task in state referenced by exactly one trigger; no unreferenced tasks; trigger references existing task key
19. User-picks mode: exactly 1 task in state
20. v1: `trigger.task_outcome` field absent; `task.outcomes` absent or empty
21. Unknown JSON fields → reject

### Warnings (logged + returned in reload response)
- W1: State has >1 outgoing, multiple tasks, no triggers → callers must supply `next_state` on first completion
- W2: Mixed binding within a state (some edges triggered, some not) — not a hard error if validator rule 17 above is relaxed; currently this is a hard error
- W3: Single outgoing edge with explicit trigger → trigger is metadata only

---

## Engine Semantics

### Transition Mode (auto-derived per state)

| Mode | Config | Semantics |
|------|--------|-----------|
| **Linear** | k=1 outgoing, no trigger (or trigger present = metadata) | All tasks done → fire the single edge |
| **Task-Bound** | k>1 outgoing, all edges have `trigger.task_key` | Completing bound task fires its edge; siblings → cancelled |
| **User-Picks** | k>1 outgoing, no triggers, exactly 1 task | Caller supplies `next_state` on completion; siblings (none, since 1 task) N/A |

### Complete Task Flow
```
Complete(taskID, status, nextState?, createdBy)
  SELECT workflow_instance FOR UPDATE  -- serialization lock
  validate task.status ∈ {new, in_progress}
  validate instance.status == active

  if status == rejected:
    task.status = rejected
    instance.status = cancelled
    cancel all open tasks in instance
    append no state_transition (cancelled, not transitioned)
    emit instance.cancelled event
    return

  task.status = done
  mode = derive(state)

  switch mode:
    Linear:
      if not all non-cancelled tasks in state are done: return (wait)
      nextStateID = state.outgoing[0].to_state
    Task-Bound:
      edge = outgoing[trigger.task_key == task.key]
      nextStateID = edge.to_state
      cancel sibling tasks in state
    User-Picks:
      require nextState arg
      validate nextState ∈ state.outgoing[*].to_state
      nextStateID = nextState
      (no siblings to cancel — single task state)

  append state_transition(from=state.id, to=nextStateID)
  instance.current_state_id = nextStateID

  if nextStateID == def.end_state:
    instance.status = completed
    emit instance.completed event
  else:
    spawn task_instances for nextStateID tasks
    emit task.created events

  emit state.transitioned event
  emit task.completed event (for completing task)
```

---

## API Endpoints

All endpoints under `/api/v1/`. JSON request/response. All IDs are UUIDs. Timestamps UTC ISO-8601.

### Workflow Definitions
```
GET    /workflow-defs                    list (active; ?include_deleted=true for all)
GET    /workflow-defs/{id}               get def + full JSONB
POST   /admin/reload-defs               rescan ./workflows/*.json → { loaded, warnings, errors }
```

### Workflow Instances
```
POST   /workflow-instances              start { workflow_def_id, created_by, initial_context? }
GET    /workflow-instances              list ?def_id=&status=&limit=&offset=
GET    /workflow-instances/{id}         full detail (instance + tasks + history)
POST   /workflow-instances/{id}/cancel  cancel instance
DELETE /workflow-instances/{id}         soft-delete
```

### Task Instances
```
GET    /workflow-instances/{id}/tasks           list tasks ?status=
POST   /workflow-instances/{id}/tasks           create ad-hoc task { state_id, label, task_type, created_by, priority?, expected_completion_date?, expected_completion_time? }
GET    /task-instances/{id}                     get task (includes priority, dates, is_adhoc, comment_count, has_attachment)
PATCH  /task-instances/{id}                     partial update { priority?, expected_completion_date?, expected_completion_time? } — locked post-completion
POST   /task-instances/{id}/start               mark in_progress
POST   /task-instances/{id}/complete            { status, next_state?, completed_by, outcome?, data? }
POST   /task-instances/{id}/comments            add comment { author, content } — allowed anytime incl. post-completion
GET    /task-instances/{id}/comments            list comments ordered by created_at ASC
POST   /task-instances/{id}/attachment          upload/replace file { filename, content_type, data: base64 } — S3-gated
GET    /task-instances/{id}/attachment          download file (server-proxied from S3) — S3-gated
DELETE /task-instances/{id}/attachment          remove file — S3-gated
```

### Misc
```
GET    /workflow-instances/{id}/history  state transitions (audit)
GET    /task-types                       list task type constants
GET    /healthz                          liveness
GET    /readyz                           readiness (DB ping)
```

### Error Envelope
```json
{
  "error": {
    "code": "INSTANCE_TERMINAL",
    "message": "Cannot complete task on cancelled instance",
    "details": { "instance_id": "..." }
  }
}
```

Error codes: `VALIDATION_FAILED` `NOT_FOUND` `CONFLICT` `INSTANCE_TERMINAL` `TASK_NOT_COMPLETABLE` `INVALID_NEXT_STATE` `TASK_COMPLETED` `INVALID_STATE` `FEATURE_DISABLED` `FILE_TOO_LARGE` `INTERNAL` `BAD_REQUEST`

HTTP status: `200` `201` `204` `400` `404` `409` `422` `500`

Pagination: offset+limit. Default `limit=50`, max `200`. Response: `{ items, total, limit, offset }`.

---

## Webhook Design

### `webhooks.json`
```json
[
  {
    "url": "https://example.com/hooks/workflow",
    "secret": "your-hmac-secret",
    "event_types": ["task.created", "state.transitioned"],
    "workflow_def_names": ["leave-approval"]
  }
]
```
`workflow_def_names` omit = all workflows.

### Payload
```json
{
  "id": "<outbox-id>",
  "event_type": "state.transitioned",
  "occurred_at": "2026-06-12T10:00:00Z",
  "workflow_instance_id": "...",
  "data": { "from_state": "submitted", "to_state": "hr_review" }
}
```
Headers: `X-Webhook-Event`, `X-Webhook-Delivery`, `X-Webhook-Signature: sha256=<hex>`.

### Retry Schedule
Attempts 1–6 at: immediate, +1m, +5m, +30m, +2h, +12h. After 6 failures → `failed_at` set (dead).

---

## Repository Interface Pattern (R2)

```go
type WorkflowDefRepo struct { db DB }

func NewWorkflowDefRepo(db DB) *WorkflowDefRepo { ... }

func (r *WorkflowDefRepo) WithTx(tx DB) *WorkflowDefRepo {
    return &WorkflowDefRepo{db: tx}
}

// Usage in service:
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)
defRepo.WithTx(tx).Create(ctx, def)
tx.Commit(ctx)
```

---

## Environment Variables

```
# Backend
DATABASE_URL              postgres://user:pass@host:5432/workflow
LOG_LEVEL                 info | debug | warn | error  (default: info)
LOG_FORMAT                json | text                   (default: json)
HTTP_ADDR                 :8080                         (default: :8080)
WORKFLOWS_DIR             /workflows                    (default: ./workflows)
WEBHOOK_CONFIG            /etc/webhook/webhooks.json    (default: ./webhooks.json)
WEBHOOK_WORKER_INTERVAL   5s                            (default: 5s)
WEBHOOK_MAX_ATTEMPTS      6                             (default: 6)
CORS_ALLOWED_ORIGINS      http://localhost:3000,...

# S3 File Upload (feature-flagged)
S3_ENABLED                false                         (default: false — disables all attachment endpoints)
S3_BUCKET                                               (required when S3_ENABLED=true)
S3_REGION                 us-east-1                     (default: us-east-1)
S3_ACCESS_KEY                                           (required when S3_ENABLED=true)
S3_SECRET_KEY                                           (required when S3_ENABLED=true)
S3_ENDPOINT                                             (optional — for MinIO/LocalStack)

# Frontend (build-time)
VITE_API_BASE_URL         /api   (default; nginx or embedded static serves /api/* to backend)
```

---

## Concurrency

Instance operations (start, complete task, cancel) acquire `SELECT * FROM workflow_instance WHERE id=$1 FOR UPDATE` at the start of each transaction. Guarantees serializability per instance. No distributed locking needed (single process, single Postgres).

---

## Deployment

Single Docker image (`backend/Dockerfile`):
1. `node:20-alpine` stage: `npm ci && npm run build` → `/web/dist`
2. `golang:1.22-alpine` stage: `go build` (embeds `frontend/dist` via `//go:embed webroot/*`)
3. `alpine:3.20` runtime: runs Go binary

`/api/*` handled by Go router. `/*` served from embedded `webroot/` with SPA fallback to `index.html`.
