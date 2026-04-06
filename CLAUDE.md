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

## Development Approach
- **Spec-driven development**: Document first, then implement.
- **All code must have tests**: Unit tests + integration tests. No exceptions.
- **Interface-based design**: Repository layer uses interfaces for testability.

## Tech Stack
- **Backend**: Go 1.22+, net/http + gorilla/mux, PostgreSQL 16, golang-migrate, testify, log/slog
- **Frontend**: React 18, TypeScript, Vite
- **Dev infra**: Docker Compose (PostgreSQL), Makefile

## Build & Run Commands
```bash
# Backend
cd backend && go build ./cmd/server        # Build
cd backend && go test ./...                 # Run all tests
cd backend && go test -tags=integration ./tests/integration/...  # Integration tests

# Frontend
cd frontend && npm install                  # Install deps
cd frontend && npm run dev                  # Dev server
cd frontend && npm test                     # Run tests

# Infrastructure
docker-compose up -d                        # Start PostgreSQL
make migrate-up                             # Run migrations
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
