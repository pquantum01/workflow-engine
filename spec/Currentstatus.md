# Current Status - Workflow Engine

## Phase: Phase 5 - Frontend Operations + Polish (COMPLETE)

### Completed
- [x] Phase 1: Backend Foundation
- [x] Phase 2: Core Engine + Services
- [x] Phase 3: API Layer + Integration Tests
- [x] Phase 4: Frontend Setup + Admin Portal
- [x] Phase 5: Frontend Operations + Polish
  - [x] Instance TypeScript types (ProcessInstance, WorkflowInstance, TaskInstance, StateTransition, inputs)
  - [x] API client layer (processInstances, workflowInstances, taskInstances, stateTransitions)
  - [x] InstanceListPage — table with status chips, Start Process dialog, row click navigation
  - [x] InstanceDetailPage — process detail with workflow accordion sections
  - [x] Task actions — Start (pending→in_progress), Complete, Fail buttons per status
  - [x] Task assignment — Assign popover with username input
  - [x] Manual transitions — buttons for available transitions from current state, trigger dialog
  - [x] Audit log — state transition history table per workflow instance (newest first)
  - [x] Sidebar nav updated — Instances link with PlayArrowIcon
  - [x] Backend: GET /api/v1/workflow-instances/{id}/transitions endpoint
  - [x] Backend integration test: TestAPI_StateTransitionsList
  - [x] Frontend tests: InstanceListPage (4 tests), api-instances (5 tests)
  - [x] Production build clean (tsc + vite build)

### Test Summary
```
Backend:
  workflow-engine/backend/internal/model    — PASS (48 tests)
  workflow-engine/backend/internal/engine   — PASS (15 tests)
  workflow-engine/backend/internal/service  — PASS (22 tests)
  workflow-engine/backend/tests/integration — 17 API integration tests

Frontend:
  src/__tests__/ProcessListPage.test.tsx    — PASS (4 tests)
  src/__tests__/api-client.test.ts          — PASS (4 tests)
  src/__tests__/InstanceListPage.test.tsx   — PASS (4 tests)
  src/__tests__/api-instances.test.ts       — PASS (5 tests)
```

### Frontend Commands
```bash
make fe-install    # Install dependencies
make fe-dev        # Dev server on :3000 (proxies /api to :8080)
make fe-build      # Production build
make fe-test       # Run tests
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
| 2026-04-06 | 1-level parent-child constraint | Enforced at application level in ProcessInstanceService |
| 2026-04-06 | Spec-driven development | Write specs/docs first, then implement with full test coverage |
| 2026-04-06 | Module path `workflow-engine/backend` | Short internal module path, no github.com prefix for local packages |
| 2026-04-06 | Shared testutil mocks | Mock repos in internal/testutil for reuse across engine and service tests |
| 2026-04-06 | Auto-evaluate engine on task completion | TaskInstanceService triggers engine.Evaluate after marking a task completed |
| 2026-04-08 | Response helpers in handler package | Avoids import cycle between api and handler packages |
| 2026-04-08 | Thin handlers, no handler unit tests | Handlers only parse HTTP and call services — business logic tested in service layer |
| 2026-04-08 | httptest.NewServer for integration tests | Full HTTP request/response cycle against real PostgreSQL |
| 2026-04-08 | `workflow_engine` schema | All tables in dedicated schema, search_path set on app connection |
| 2026-04-09 | MUI (Material UI) for frontend | Admin portal UI library — tables, forms, dialogs, Material Design aesthetic |
| 2026-04-09 | Vite + React 19 + TypeScript | Fast dev server, HMR, modern React with type safety |
| 2026-04-09 | Vitest + React Testing Library | Fast test runner compatible with Vite, standard React testing |
| 2026-04-13 | State transitions on WorkflowInstanceService | Added stateTransitionRepo to WorkflowInstanceService for ListTransitions, keeps handler thin |
