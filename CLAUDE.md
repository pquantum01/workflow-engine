# CLAUDE.md - Workflow Engine

## Project Overview
A standalone workflow engine (like Jira) with configurable processes, workflows, tasks, and state machines. Backend in Go, frontend in React, monorepo structure.

## Repository Structure
```
workflow-engine/
├── backend/          # Go backend (API server)
│   ├── cmd/server/   # Entry point
│   ├── internal/     # All Go packages (api, model, service, repository, engine, config)
│   ├── migrations/   # SQL migrations
│   └── tests/integration/
├── frontend/         # React + TypeScript frontend
├── docker-compose.yml
└── Makefile
```

## Key Documentation
- `spec/Requirement.md` — Functional and non-functional requirements
- `spec/Architecture.md` — System design, data model, API spec, tech stack
- `spec/Currentstatus.md` — Progress tracking, next steps, decisions log
- `spec/testDetails.md` — Test details and coverage

## Development Approach
- **Spec-driven development**: Document first, then implement.
- **All code must have tests**: Unit tests + integration tests. No exceptions.
- **Interface-based design**: Repository layer uses interfaces for testability.
- **Complexity management**: Keep business logic in services, avoid adding too much abstraction & nesting.
- **Code quality**: Follow Go best practices, use meaningful variable names, and write clear comments.
- **Keep it simple**: Avoid over-engineering and unnecessary complexity.
- **Documentation**: Keep documentation up-to-date with code changes.

## Tech Stack
- **Backend**: Go 1.22+, net/http + gorilla/mux, PostgreSQL 16, golang-migrate, testify, log/slog
- **Frontend**: React 18, TypeScript, Vite
- **Dev infra**: Docker Compose (PostgreSQL), Makefile

## Build & Run Commands
```bash
# All commands run from repo root (go.work workspace)
make build              # Build binary to ./bin/server
make run                # Build + run server
make test               # Run unit tests
make test-integration   # Run integration tests (requires PostgreSQL)
make docker-up          # Start PostgreSQL containers
make migrate-up         # Apply migrations
make migrate-up-test    # Apply migrations to test DB

# Frontend (later)
cd frontend && npm install    # Install deps
cd frontend && npm run dev    # Dev server
cd frontend && npm test       # Run tests
```

## Conventions
- Go package layout follows `internal/` convention (not importable externally).
- HTTP handlers are thin — business logic lives in the service layer.
- Repository interfaces live in `internal/repository/`, implementations in `internal/repository/postgres/`.
- State machine definitions are stored as JSONB in PostgreSQL.
- All IDs are UUIDs.
- API versioned under `/api/v1/`.
- Database constraints enforce the 1-level parent-child hierarchy for process instances.
- Completion rules (how tasks drive state transitions) are configured as part of the workflow definition's state machine JSONB.
- **No `fmt.Println` for logging** — use `log/slog` structured logger everywhere. This allows plugging in any backend (JSON, Loggly, Datadog, etc.) via slog handlers.
- **Backend-first development** — complete and test all backend APIs before starting frontend.
- DB table names: `process`, `workflow`, `task` (definitions); `process_instances`, `workflow_instances`, `task_instances` (instances).
