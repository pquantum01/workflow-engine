# CLAUDE.md - Workflow Engine

## Project Overview
Standalone workflow engine with configurable state machines, tasks, and transitions. Backend Go, frontend React, single binary serves both. Target: plug in real-life scenarios from different services via webhooks and API.

## Repository Structure
```
workflow-engine/
├── backend/              Go backend (API server + embedded frontend)
│   ├── cmd/server/       Entry point
│   ├── internal/         All Go packages
│   │   ├── api/          handlers, middleware, dto, router, static
│   │   ├── config/       env vars, logger
│   │   ├── engine/       state machine logic
│   │   ├── model/        domain types
│   │   ├── repository/   interfaces + postgres implementations (flat, no subfolder)
│   │   ├── service/      business logic layer
│   │   ├── webhook/      dispatcher, worker, registry, signer
│   │   ├── workflowloader/ file scan → validate → upsert
│   │   └── testutil/     hand-written repo mocks
│   ├── migrations/       SQL migrations (up + down)
│   ├── webroot/          embed target (frontend/dist copied at build)
│   └── tests/integration/
├── frontend/             React 18 + TypeScript + Vite
├── workflows/            Sample workflow JSON defs (mounted in container)
├── tests/
│   ├── smoke/            smoke.sh (curl-based)
│   └── e2e/              Playwright + Chromium specs
├── spec/                 All documentation
├── docker-compose.yml    Full stack (postgres + app)
├── docker-compose.e2e.yml E2e test stack (isolated ports)
├── Makefile
└── go.work               Go workspace
```

## Key Documentation
- `spec/Requirement.md` — Functional and non-functional requirements
- `spec/Architecture.md` — Data model, API spec, engine semantics, tech stack
- `spec/Currentstatus.md` — Phase progress tracker + decisions log
- `spec/glossary.md` — Domain terms
- `spec/phase{N}-*.md` — Per-phase detailed specs
- `spec/plans/` — Implementation plans (per feature)

## File Organization Rules
- **All documentation goes in `spec/`** — design docs, plans, specs. Never create a `docs/` or `docs/superpowers/` or any other top-level folder for documentation.
- Read this file at the start of every session before engaging any skills or writing any files.

## Development Phases
| # | Phase | Status |
|---|-------|--------|
| 0 | Spec docs | ✅ Done |
| 1 | Foundation (scaffold, DB, models, repos) | ⬜ |
| 2 | Engine + Services | ⬜ |
| 3 | REST API + Webhooks | ⬜ |
| 4 | Frontend Admin (def list + graph) | ⬜ |
| 5 | Frontend Ops (instances + tasks) | ⬜ |
| 6 | Polish + Deploy (Dockerfile, e2e) | ⬜ |

## Development Approach
- **Spec-driven**: read phase spec before implementing. Update `spec/Currentstatus.md` when phase completes.
- **Tests mandatory**: every phase has unit + integration tests as exit criteria. No phase is done without green tests.
- **Backend-first**: complete and test all backend APIs before starting frontend.
- **Interface-based design**: repository layer uses interfaces (defined in `internal/repository/interfaces.go`). Postgres implementations in same package (flat, no subfolder).
- **Simple**: no unnecessary abstraction. Business logic in services. Handlers are thin. No global state beyond DI root.
- **No over-engineering**: if a feature isn't in the current phase spec, don't add it.

## Tech Stack

### Backend
- Go 1.22+, `gorilla/mux`, PostgreSQL 16
- `database/sql` + `lib/pq`, raw SQL (no ORM)
- `golang-migrate` for migrations
- `log/slog` JSON handler (mandatory — no `fmt.Println`)
- `testify` for assertions

### Frontend
- React 18, TypeScript, Vite
- Tailwind CSS + shadcn/ui
- TanStack Query v5, React Router v6
- React Flow + dagre (graph viz)
- Framer Motion (animation)
- Vitest + RTL (unit), Playwright + Chromium (e2e)

## Build & Run Commands
```bash
# Backend
make build              # Build binary to ./bin/server
make run                # Build + run server (needs postgres)
make test               # Unit tests
make test-integration   # Integration tests (requires postgres)

# Docker
make docker-up          # Start postgres only (local dev)
make docker-down        # Stop postgres
make migrate-up         # Apply migrations (dev DB)
make migrate-up-test    # Apply migrations (test DB)

# Frontend
make frontend-dev       # cd frontend && npm run dev (Vite proxy → localhost:8080)
make frontend-build     # cd frontend && npm run build
make frontend-test      # cd frontend && npm test

# Full stack
make stack-up           # docker compose up --build -d (full app)
make stack-down         # docker compose down -v

# Tests
make smoke              # curl-based smoke against full stack
make e2e                # Playwright e2e (spins up e2e stack, runs tests, tears down)
```

## Conventions

### Go
- HTTP handlers are thin — all business logic in service layer.
- Repository interfaces in `internal/repository/interfaces.go`. Postgres implementations in `internal/repository/` (same package, flat).
- Repository pattern R2: repos bound to `DB` at construction, `WithTx(tx DB)` returns new instance for transactional use.
- Concurrency: `SELECT … FOR UPDATE` on `workflow_instance` row at start of Complete/Cancel transactions.
- Engine returns `[]Event` alongside updated instance. Service passes events to webhook dispatcher.
- **No `fmt.Println`** — use `log/slog` structured logger everywhere.
- Sentinel errors in `repository/errors.go` and `service/` files; handlers translate to HTTP error codes.
- All IDs are UUIDs (`github.com/google/uuid`).

### Database
- Table naming: `workflow_def`, `workflow_instance`, `task_instance`, `state_transition`, `webhook_outbox`.
- Soft delete: `deleted_at TIMESTAMPTZ`. Default queries filter `WHERE deleted_at IS NULL`.
- Workflow def unique by `(name, version) WHERE deleted_at IS NULL`.
- Definitions immutable once loaded — new content requires new version.

### API
- All endpoints under `/api/v1/`.
- JSON in, JSON out. UTC ISO-8601 timestamps.
- Error envelope: `{ "error": { "code": "...", "message": "...", "details": ... } }`.
- Pagination: offset+limit, default 50, max 200.

### Workflow Definitions (JSON files in `./workflows/`)
- Slug format for `name`, `states[].id`, `tasks[].key`: `^[a-z][a-z0-9_-]*$`
- Hard validation errors → reject file (fail boot on startup; return error on reload).
- Validator warnings → logged + returned in reload response.
- Unknown JSON fields → rejected.
- Engine transition modes auto-derived from config shape (linear / task-bound / user-picks). See `spec/Architecture.md`.

### Frontend
- API calls via `src/api/client.ts` wrapper. Per-resource modules in `src/api/`.
- Hand-written types in `src/types/index.ts` mirroring backend DTOs.
- Task renderer registry in `src/components/tasks/registry.ts` — extensible per task type.
- TanStack Query for server state. No Redux/Zustand unless required.
- Dark + light theme (Tailwind `dark:` class, system default, manual toggle).
- See `spec/phase4-design.md` for full design language (colors, typography, motion).

### Testing
- Unit tests: mock repos via `internal/testutil/mocks.go` (hand-written, not generated).
- Integration tests: real Postgres, truncate tables between tests.
- E2e: Playwright + Chromium in `tests/e2e/specs/`. 7 spec files covering full user journeys.
- `make smoke`: curl-based golden path against running stack.

## Webhook Integration
- Subscribers defined in `webhooks.json` (git-ignored; example at `webhooks.json.example`).
- Events: `task.created`, `task.completed`, `task.cancelled`, `state.transitioned`, `instance.started`, `instance.completed`, `instance.cancelled`.
- Outbox pattern: durable delivery, retry schedule (1m/5m/30m/2h/12h/dead).
- HMAC-SHA256 signature on every delivery (`X-Webhook-Signature: sha256=<hex>`).

## Environment Variables
```
DATABASE_URL              postgres://user:pass@host:5432/workflow
LOG_LEVEL                 info | debug | warn | error
LOG_FORMAT                json | text
HTTP_ADDR                 :8080
WORKFLOWS_DIR             /workflows
WEBHOOK_CONFIG            /etc/webhook/webhooks.json
WEBHOOK_WORKER_INTERVAL   5s
WEBHOOK_MAX_ATTEMPTS      6
CORS_ALLOWED_ORIGINS      http://localhost:3000,...
```
