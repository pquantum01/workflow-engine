# Current Status - Workflow Engine

## Phase: Phase 1 - Backend Foundation (COMPLETE)

### Completed
- [x] Requirements documented (Requirement.md)
- [x] Architecture designed (Architecture.md)
- [x] Phase 1 spec written (phase1-foundation.md)
- [x] Go module initialized (`workflow-engine/backend`)
- [x] Go environment setup (GOPATH, GOBIN, GOROOT in ~/.zshrc)
- [x] Docker Compose for PostgreSQL (dev on :5432, test on :5433)
- [x] Makefile for common commands
- [x] Database migrations (7 tables: process, workflow, task, process_instances, workflow_instances, task_instances, state_transitions)
- [x] Domain model structs with validation (Process, Workflow, Task, StateMachine, CompletionRule, instances)
- [x] Repository interfaces (7 interfaces in internal/repository/)
- [x] PostgreSQL repository implementations (7 repos in internal/repository/postgres/)
- [x] Structured logging with log/slog (JSON handler)
- [x] Unit tests for all models (process, workflow, task, state_machine, instances)
- [x] Integration tests for all repositories (process, workflow, task, instances, state_transitions)
- [x] Server entry point with health endpoint

### Next Up: Phase 2 - Core Engine + Services
- [ ] State machine engine (transition evaluation, guard checks)
- [ ] Completion rule evaluation (single, all_of, n_of_m, any_of)
- [ ] Service layer for definitions (Process, Workflow, Task CRUD)
- [ ] Service layer for instances (start process, complete tasks, trigger transitions)
- [ ] Parent-child process hierarchy with 1-level enforcement
- [ ] Unit tests for engine and all services

### Blocked
- Nothing currently blocked.

### How to Run Integration Tests
```bash
make docker-up                # Start PostgreSQL containers
make migrate-up               # Apply migrations to dev DB
make migrate-up-test          # Apply migrations to test DB
make test-integration         # Run integration tests
```

### Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-06 | PostgreSQL for storage | JSONB for flexible state machine config, strong consistency for state transitions |
| 2026-04-06 | net/http + gorilla/mux for routing | Standard library server with mux for path params and route matching |
| 2026-04-06 | log/slog for structured logging | Pluggable handlers allow integration with Loggly, Datadog, etc. No fmt logging |
| 2026-04-06 | Backend-first development | Complete and test all backend APIs before starting frontend |
| 2026-04-06 | Simplified table names | `process`, `workflow`, `task` for definitions; `_instances` suffix for runtime tables |
| 2026-04-06 | Interface-based repository pattern | Enables unit testing services without a database |
| 2026-04-06 | JSONB for state machine definitions | Avoids complex relational modeling for flexible graph structures |
| 2026-04-06 | 1-level parent-child constraint | Enforced at DB level with CHECK constraint + application-level validation |
| 2026-04-06 | Spec-driven development | Write specs/docs first, then implement with full test coverage |
| 2026-04-06 | Module path `workflow-engine/backend` | Short internal module path, no github.com prefix for local packages |
