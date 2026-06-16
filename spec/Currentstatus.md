# Current Status

## Phase Progress

| Phase | Name | Status | Notes |
|-------|------|--------|-------|
| 0 | Spec docs | ✅ Complete | All spec docs written |
| 1 | Foundation | ✅ Complete | Repo scaffold, DB schema, models, repos, integration tests |
| 2 | Engine + Services | ✅ Complete | Workflow loader, service layer, state machine engine |
| 3 | REST API | ✅ Complete | HTTP router, handlers, middleware, webhook dispatcher |
| 4 | Frontend Admin | ✅ Complete | React scaffold, workflow def list/detail, state graph viz |
| 5 | Frontend Ops | ✅ Complete | Start instance, task completion, live graph, admin reload |
| 6 | Polish & Deploy | ✅ Complete | Dockerfile, docker-compose, graceful shutdown, smoke + e2e |

## Exit Criteria Per Phase

### Phase 1
- [ ] `go.work` workspace + `backend/go.mod` with all deps
- [ ] `docker-compose.yml` (postgres only for dev)
- [ ] All migrations applied cleanly (`001_create_base_tables`, `002_webhook_outbox`)
- [ ] All domain models in `internal/model/` with unit tests
- [ ] Repository interfaces in `internal/repository/interfaces.go`
- [ ] All Postgres repo implementations with integration tests green
- [ ] `make test` passes (unit)
- [ ] `make test-integration` passes (requires postgres)

### Phase 2
- [ ] Workflow loader scans `./workflows/*.json`, validates, upserts
- [ ] Loader unit tests (valid + each error case)
- [ ] Service layer: `WorkflowDefService`, `InstanceService`, `TaskService`
- [ ] Engine: all three transition modes (linear, task-bound, user-picks)
- [ ] Engine: rejection → instance cancel
- [ ] Engine: `SELECT FOR UPDATE` concurrency guard
- [ ] Engine unit tests (mock repos, all branching paths)
- [ ] Service unit tests
- [ ] `make test` green

### Phase 3
- [ ] All endpoints from Architecture.md implemented
- [ ] Middleware: CORS, request-id, slog request log, panic recovery
- [ ] Webhook outbox dispatcher + worker + HMAC signer
- [ ] Integration tests: all endpoints, all error paths
- [ ] `POST /admin/reload-defs` returns warnings
- [ ] `GET /healthz` + `GET /readyz` work
- [ ] `make test-integration` green

### Phase 4
- [ ] React + Vite + TS + Tailwind + shadcn/ui scaffold
- [ ] API client (`src/api/client.ts` + resource modules)
- [ ] Layout shell (topbar, sidebar, routing)
- [ ] `/workflow-defs` list page (table default, card toggle)
- [ ] `/workflow-defs/:id` detail page with React Flow graph
- [ ] Dagre LR auto-layout, start/end node styling
- [ ] Warning banner for def warnings
- [ ] Vitest + RTL tests
- [ ] Dark/light theme toggle
- [ ] `npm test` green

### Phase 5
- [ ] `/workflow-instances` list + filters
- [ ] `/workflow-instances/new` start form (def picker, created_by, initial_context KV editor)
- [ ] `/workflow-instances/:id` detail (live graph, task panel, history timeline)
- [ ] `/task-instances/:id` standalone deep-link page
- [ ] Task renderer registry (approval, form stub, checklist stub, manual)
- [ ] Decision-state next-state dropdown on task card
- [ ] Polling: 5s active instance, 15s list
- [ ] Optimistic task complete with rollback
- [ ] `/admin/reload` page
- [ ] `npm test` green

### Phase 6
- [x] Backend Dockerfile (3 stages: node, go, alpine)
- [x] `docker-compose.full.yml` full stack; `docker-compose.yml` dev (postgres only)
- [x] Graceful shutdown (SIGTERM → drain HTTP, stop webhook worker)
- [x] `tests/smoke/smoke.sh` passes (verified locally)
- [x] Playwright e2e specs all green: golden-path, cancel, rejection, admin-reload, list-filter (5/5)
- [x] `README.md` quickstart complete

## Key Decisions Log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Def storage | JSONB blob | Defs read whole; easy field additions without migrations |
| Instance storage | Normalized tables | Runtime state needs indexed queries |
| Soft delete | `deleted_at` | Preserve history; hard deletes disallowed |
| Transaction handling | R2 (`WithTx`) | Ergonomic; tx explicit per call site |
| Concurrency | `SELECT FOR UPDATE` on instance | Serializes concurrent completions; Postgres-native |
| Router | gorilla/mux | Established, path vars, method routing |
| DB driver | `database/sql` + `lib/pq` | Standard library interface; pgx had dependency issues; both *sql.DB and *sql.Tx satisfy DB interface natively |
| Query style | Raw SQL | Simple queries; no ORM abstraction cost |
| Test DB isolation | Truncate between tests | Engine uses own txns; test-level wrapping awkward |
| Mock style | Hand-written | Small surface; no codegen tool dep |
| Task type registry | Go const (C1) | Validated at load; no migration per new type |
| State/task key scope | Workflow-local | Different domains need different vocab |
| Multi-edge | Unified (linear/task-bound/user-picks auto-derived) | Config shape determines semantics |
| Webhook registry | Static `webhooks.json` | Simple; v2 → DB-backed |
| Webhook delivery | Outbox + worker + retry | Durable; survives crash |
| Frontend serving | Embedded in Go binary | Single container, single port |
| Container base | alpine:3.20 | Debugger-friendly, small |
| CI | Deferred to v2 | Not needed for v1 |
| E2e testing | Playwright + Chromium | Full browser coverage |
| Terminal states | Multiple allowed (any state with 0 tasks + 0 outgoing) | Real workflows have multiple terminals (approved, denied, wont_fix); end_state = declared success terminal only |
| Validator warnings | Non-blocking (logged + returned in reload response) | Permissive; ops sees warnings in admin UI/API; hard errors only for structural breaks |
| Engine tx ownership | Engine holds `*sql.DB`, opens tx internally | Service layer calls engine; webhook outbox separate tx (acceptable for v1) |
| Frontend dev serving | Vite dev server proxied to :8080; prod = embedded in Go binary | Single port in prod; hot reload in dev |
