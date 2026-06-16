# Test Details

Complete inventory of every test in the repository. No exceptions.

## Summary

| Layer | Files | Tests | Tool | Run Command |
|-------|-------|-------|------|-------------|
| Backend Unit | 5 files | 33 tests | `go test` | `make test` |
| Backend Integration | 6 files | 53 tests | `go test` + real Postgres | `make test-integration` |
| Frontend Unit | 3 files | 14 tests | Vitest + RTL | `make frontend-test` |
| Smoke | 1 script | ~10 assertions | `bash` + `curl` | `./tests/smoke/smoke.sh` |
| E2E | 5 specs | 5 tests | Playwright + Chromium | `cd tests/e2e && npx playwright test` |
| **Total** | | **~115** | | |

---

## Backend Unit Tests

Run: `go test github.com/workflow-engine/backend/internal/...`  
No database required. Uses hand-written mocks from `internal/testutil/mocks.go`.

---

### `internal/model/workflow_def_test.go`

Tests `WorkflowDefinition.Validate()` — hard validation rules enforced at load and reload time.

| Test | What it verifies |
|------|-----------------|
| `TestValidate_LinearDef` | Valid linear 2-state def passes validation |
| `TestValidate_InvalidName` | Name not matching `^[a-z][a-z0-9_-]*$` → error |
| `TestValidate_VersionZero` | Version 0 → error |
| `TestValidate_MissingStartState` | `start_state` not in states list → error |
| `TestValidate_MissingEndState` | `end_state` not in states list → error |
| `TestValidate_EndStateHasTasks` | Terminal `end_state` has tasks → error |
| `TestValidate_UnknownTaskType` | Task type not in valid set → error |
| `TestValidate_NonEndStateNoTasks` | Non-terminal state with no tasks and no outgoing → ok (valid secondary terminal) |
| `TestValidate_UnreachableState` | State unreachable from `start_state` → error |
| `TestValidate_Cycle` | Transition cycle detected → error |
| `TestValidate_TaskBoundMultiEdge` | Valid task-bound multi-edge def passes |
| `TestValidate_UserPicksMultiEdge` | Valid user-picks multi-edge def passes |
| `TestValidate_UserPicksMustHaveSingleTask` | User-picks state with >1 task → error |
| `TestFindState` | `FindState` returns correct state or false for missing |
| `TestTransitionMode` | `TransitionMode` returns correct mode enum per config shape |

Also covers helpers:
- `TestTransitionMode_Linear` — 1 outgoing, no trigger → `ModeLinear`
- `TestTransitionMode_TaskBound` — all edges have trigger → `ModeTaskBound`
- `TestTransitionMode_UserPicks` — multi-edge, no triggers → `ModeUserPicks`

---

### `internal/model/task_instance_test.go`

Tests `TaskInstance` status-checking helpers.

| Test | What it verifies |
|------|-----------------|
| `TestTaskInstance_IsCompletable` | Returns true only for `new` or `in_progress` |
| `TestTaskInstance_IsTerminal` | Returns true for `done`, `rejected`, `cancelled` |

---

### `internal/webhook/signer_test.go`

Tests HMAC-SHA256 signing used for `X-Webhook-Signature` header.

| Test | What it verifies |
|------|-----------------|
| `TestSign_ProducesHMACSHA256` | Output starts with `sha256=`, has non-trivial length |
| `TestSign_DifferentSecrets_DifferentSig` | Different secrets produce different signatures for same body |
| `TestSign_SameInputDeterministic` | Same secret + body always produces same signature |

---

### `internal/workflowloader/loader_test.go`

Tests `Loader.Load()` — file scanning, JSON decoding, validation, upsert logic.

| Test | What it verifies |
|------|-----------------|
| `TestLoader_ValidFile_Loads` | Valid JSON file → loaded into mock repo, `Loaded` non-empty |
| `TestLoader_InvalidJSON_Error` | Malformed JSON → `Errors` non-empty, no crash |
| `TestLoader_UnknownTaskType_Error` | Unknown task type in def → validation error in `Errors` |
| `TestLoader_SameContentNoOp` | Re-loading identical content → no duplicate insert, returns loaded |
| `TestLoader_DifferentContent_VersionConflict` | Same name+version, different content → version conflict error |
| `TestLoader_NonexistentDir_ReturnsEmpty` | Non-existent `dir` → empty result, no error |
| `TestLoader_SkipsNonJSONFiles` | `.txt` files in dir → ignored |
| `TestLoader_MultipleFiles_BadOneSkipped` | One bad file + one good file → good one loads, bad one in `Errors` |

---

### `internal/engine/engine_test.go`

Unit tests for transition mode derivation. Uses mock repos (no DB).

| Test | What it verifies |
|------|-----------------|
| `TestTransitionMode_Linear` | Single outgoing, no trigger → `ModeLinear` |
| `TestTransitionMode_TaskBound` | Multiple outgoing with trigger keys → `ModeTaskBound` |
| `TestTransitionMode_UserPicks` | Multiple outgoing, no triggers → `ModeUserPicks` |
| `TestEvent_Constants` | Event type string constants have expected values |
| `TestMockRepos_Compile` | Hand-written mocks satisfy all repo interfaces (compile check) |

---

## Backend Integration Tests

Run: `DATABASE_URL="postgres://workflow:workflow@localhost:5432/workflow_test?sslmode=disable" go test github.com/workflow-engine/backend/tests/integration -count=1`

Requires PostgreSQL. Tables truncated between tests via `truncateTables(t)` in `setup_test.go`.

---

### `tests/integration/workflow_def_test.go`

Tests `WorkflowDefRepo` against real Postgres.

| Test | What it verifies |
|------|-----------------|
| `TestWorkflowDefRepo_CreateAndGetByID` | Create row, fetch by UUID |
| `TestWorkflowDefRepo_GetByID_NotFound` | Missing UUID → `ErrNotFound` |
| `TestWorkflowDefRepo_GetByNameVersion` | Fetch by (name, version) compound key |
| `TestWorkflowDefRepo_List` | List with `include_deleted=false` excludes soft-deleted |
| `TestWorkflowDefRepo_SoftDelete` | Sets `deleted_at`, `GetByID` still returns it, `List` omits it |
| `TestWorkflowDefRepo_SoftDeleteAllowsReuseOfNameVersion` | After soft-delete, same (name, version) can be inserted again (partial unique index) |
| `TestWorkflowDefRepo_UniqueNameVersion` | Inserting duplicate (name, version) without delete → constraint error |

---

### `tests/integration/instance_repo_test.go`

Tests `WorkflowInstanceRepo`.

| Test | What it verifies |
|------|-----------------|
| `TestWorkflowInstanceRepo_CreateAndGetByID` | Create instance, retrieve by UUID |
| `TestWorkflowInstanceRepo_GetByID_NotFound` | Missing UUID → `ErrNotFound` |
| `TestWorkflowInstanceRepo_UpdateState` | `UpdateState` changes `current_state_id` |
| `TestWorkflowInstanceRepo_UpdateStatus` | `UpdateStatus` changes `status` + sets `completed_at` |
| `TestWorkflowInstanceRepo_List_WithFilters` | Filter by `def_id`, `status`; pagination via `limit`/`offset` |
| `TestWorkflowInstanceRepo_SoftDelete` | Sets `deleted_at` |
| `TestWorkflowInstanceRepo_GetByIDForUpdate` | `SELECT FOR UPDATE` returns correct row inside transaction |

---

### `tests/integration/task_repo_test.go`

Tests `TaskInstanceRepo`.

| Test | What it verifies |
|------|-----------------|
| `TestTaskInstanceRepo_BulkCreateAndList` | `BulkCreate` inserts multiple tasks, `ListByInstance` returns all |
| `TestTaskInstanceRepo_GetByID` | Fetch task by UUID |
| `TestTaskInstanceRepo_GetByID_NotFound` | Missing UUID → `ErrNotFound` |
| `TestTaskInstanceRepo_UpdateStatus` | Status + `completed_at` update |
| `TestTaskInstanceRepo_ListByInstance_StatusFilter` | Filter by `status` slice; `nil` filter returns all |
| `TestTaskInstanceRepo_ListByInstanceAndState` | Filter by `(instance_id, state_id)` |
| `TestTaskInstanceRepo_SoftDelete` | Sets `deleted_at` |

---

### `tests/integration/state_transition_test.go`

Tests `StateTransitionRepo` (append-only audit log).

| Test | What it verifies |
|------|-----------------|
| `TestStateTransitionRepo_AppendAndList` | Append transition, `ListByInstance` returns it |
| `TestStateTransitionRepo_ListByInstance_Empty` | No transitions → empty slice, no error |
| `TestStateTransitionRepo_MultipleTransitions_OrderedByTime` | Multiple appends → ordered by `transitioned_at ASC` |

---

### `tests/integration/engine_test.go`

Full engine flow tests against real Postgres. Each test creates fresh def + starts from scratch.

| Test | What it verifies |
|------|-----------------|
| `TestEngine_Start_CreatesInstanceAndTasks` | `Start()` creates instance with `active` status and tasks in `start_state` |
| `TestEngine_Linear_SingleTask_Done_Transitions` | Linear: single task done → transitions to next state, spawns new tasks |
| `TestEngine_Linear_MultiTask_WaitsForAll` | Linear: first task done doesn't transition; second done → fires |
| `TestEngine_TaskBound_CorrectEdgeFires` | Task-bound: completing `approve` task fires the `→ approved` edge |
| `TestEngine_TaskBound_RejectEdgeFires` | Task-bound: completing `reject` task fires the `→ rejected` edge, siblings cancelled |
| `TestEngine_UserPicks_ValidNextState` | User-picks: valid `next_state` provided → transitions |
| `TestEngine_UserPicks_RequiresNextState` | User-picks: `next_state=""` → `ErrNextStateRequired` |
| `TestEngine_UserPicks_InvalidNextState` | User-picks: unknown `next_state` → `ErrInvalidNextState` |
| `TestEngine_Rejection_CancelsInstance` | Rejected task → all siblings cancelled, instance becomes `cancelled` |
| `TestEngine_Cancel_CancelsAllOpenTasks` | `Cancel()` → all `new`/`in_progress` tasks become `cancelled`, instance `cancelled` |
| `TestEngine_CompleteOnTerminalInstance_Errors` | Completing task on already-terminal instance → `ErrInstanceTerminal` |
| `TestEngine_StateTransition_Recorded` | Transition appended to `state_transition` table with correct from/to |

---

### `tests/integration/api_test.go`

HTTP-level integration tests via `httptest.Server`. Each test uses a live router wired to real Postgres.

| Test | Method + Path | What it verifies |
|------|--------------|-----------------|
| `TestAPI_Healthz` | `GET /healthz` | 200 OK, `{"status":"ok"}` |
| `TestAPI_Readyz` | `GET /readyz` | 200 OK when DB is reachable |
| `TestAPI_ListWorkflowDefs_Empty` | `GET /api/v1/workflow-defs` | 200 with `total=0` on empty DB |
| `TestAPI_ListWorkflowDefs_WithDefs` | `GET /api/v1/workflow-defs` | 200 with `total=1` after seeding |
| `TestAPI_GetWorkflowDef_NotFound` | `GET /api/v1/workflow-defs/:id` | 404 for non-existent UUID |
| `TestAPI_GetWorkflowDef_Found` | `GET /api/v1/workflow-defs/:id` | 200, returns `definition` blob |
| `TestAPI_StartInstance` | `POST /api/v1/workflow-instances` | 201 with `status=active`, `current_state_id=s1` |
| `TestAPI_StartInstance_MissingFields` | `POST /api/v1/workflow-instances` | 422 when `workflow_def_id` missing |
| `TestAPI_GetInstance_WithTasksAndHistory` | `GET /api/v1/workflow-instances/:id` | 200 with `tasks[]` and `history[]` always present |
| `TestAPI_CompleteTask_Linear_Completes` | `POST /api/v1/task-instances/:id/complete` | 200, returned instance `status=completed` after last task |
| `TestAPI_CompleteTask_Reject_CancelsInstance` | `POST /api/v1/task-instances/:id/complete` | 200 with `status=rejected`, returned instance `status=cancelled` |
| `TestAPI_CompleteTask_OnTerminalInstance_Returns409` | `POST /api/v1/task-instances/:id/complete` | 409 when instance already cancelled |
| `TestAPI_CancelInstance` | `POST /api/v1/workflow-instances/:id/cancel` | 200 with `status=cancelled` |
| `TestAPI_ListInstances_FilterByStatus` | `GET /api/v1/workflow-instances?status=active` | 200 with `total=1` |
| `TestAPI_TaskTypes` | `GET /api/v1/task-types` | 200 with exactly 4 task types |
| `TestAPI_ReloadDefs` | `POST /api/v1/admin/reload-defs` | 200, response contains `loaded`, `errors` arrays |
| `TestAPI_GetTask_NotFound` | `GET /api/v1/task-instances/:id` | 404 for non-existent UUID |

---

## Frontend Unit Tests

Run: `cd frontend && npm test`  
Tool: Vitest 4 + Testing Library / JSDOM. Config in `frontend/vitest.config.ts`.

---

### `src/__tests__/api-client.test.ts`

Tests `apiFetch` — the central HTTP wrapper in `src/api/client.ts`.

| Test | What it verifies |
|------|-----------------|
| `returns parsed JSON on success` | 200 response → decoded body returned |
| `throws ApiError on non-2xx with error body` | 404 with `{error:{code,message}}` → `ApiError` with matching fields |
| `throws ApiError on non-2xx without body` | 500 with unparseable body → `ApiError` still thrown |
| `returns undefined for 204` | 204 No Content → `undefined` returned, no parse attempt |

---

### `src/__tests__/graph.test.ts`

Tests `layoutGraph()` in `src/lib/graph.ts` — dagre layout computation for React Flow.

| Test | What it verifies |
|------|-----------------|
| `produces correct node count` | 2-state def → 2 nodes |
| `produces correct edge count` | 1-transition def → 1 edge |
| `marks start node` | `isStart=true` on `start_state` node |
| `marks end node as terminal` | Terminal state (0 tasks, 0 outgoing) → `isTerminal=true` |
| `nodes have positive positions after layout` | dagre assigns x≥0, y≥0 to all nodes |
| `edge has trigger label when trigger present` | Edge with `trigger.task_key` → `label=task_key`, `isTriggered=true` |

---

### `src/__tests__/WorkflowDefListPage.test.tsx`

Tests `WorkflowDefListPage` React component via RTL. `useWorkflowDefs` mocked with `vi.mock`.

| Test | What it verifies |
|------|-----------------|
| `shows loading skeletons` | `isLoading=true` → skeleton elements rendered |
| `renders workflow def rows` | `data.items=[mockDef]` → name `leave-approval` and `v1` visible in table |
| `shows empty state when no defs` | `data.items=[]` → empty state message rendered |
| `shows error message on failure` | `error` set → error message rendered |

---

## Smoke Tests

Script: `tests/smoke/smoke.sh <BASE_URL>`  
Run against live server. No framework — pure `bash` + `curl` + `python3 -c`.

Requires: running backend, `leave-approval` workflow def loaded.

| Step | Assertion |
|------|-----------|
| `GET /healthz` | `status=ok` |
| `GET /readyz` | `status=ok` |
| `POST /api/v1/admin/reload-defs` | Response contains `loaded` key |
| `GET /api/v1/workflow-defs` | `total > 0` |
| `POST /api/v1/workflow-instances` (leave-approval) | `status=active` |
| `GET /api/v1/workflow-instances/:id` | Returned `id` matches |
| `GET /api/v1/workflow-instances/:id/tasks` | At least 1 task |
| `GET /api/v1/task-instances/:id` | Returned `id` matches |
| `POST /api/v1/task-instances/:id/complete` (done) | Valid JSON response |
| `GET /api/v1/workflow-instances/:id/history` | ≥1 transition recorded |

Skips gracefully if no workflow defs loaded or start state has no tasks.

---

## E2E Tests (Playwright + Chromium)

Location: `tests/e2e/specs/`  
Config: `tests/e2e/playwright.config.ts`  
Browser: Chromium (headless)  
Run: `cd tests/e2e && BASE_URL=http://localhost:5173 npx playwright test`

Requires: backend on `:8080`, Vite dev server on `:5173` (or full built app on `:8080`).  
No DB reset between tests — each spec creates its own instances via API helpers in `helpers.ts`.

---

### `specs/golden-path.spec.ts`

| Test | Steps | Assertion |
|------|-------|-----------|
| `start instance → complete all tasks → instance completed` | Start instance via API → navigate to detail page → complete task 1 (submitted) via API → reload → complete task 2 (hr_review) via API → reload | Page shows `completed` status badge |

---

### `specs/cancel.spec.ts`

| Test | Steps | Assertion |
|------|-------|-----------|
| `cancel instance → status cancelled` | Start instance via API → navigate to detail page → click "Cancel Instance" button → click "Confirm" | Page shows `cancelled` status badge |

---

### `specs/rejection.spec.ts`

| Test | Steps | Assertion |
|------|-------|-----------|
| `reject task → instance cancelled` | Start instance via API → reject first task via API → navigate to detail page | Page shows `cancelled` badge (scoped to `.rounded-full` first match) |

---

### `specs/list-filter.spec.ts`

| Test | Steps | Assertion |
|------|-------|-----------|
| `filter by status shows correct subset` | Start 2 instances via API → navigate to `/workflow-instances` → select `active` filter from status dropdown | Table has ≥1 row |

---

### `specs/admin-reload.spec.ts`

| Test | Steps | Assertion |
|------|-------|-----------|
| `reload-defs shows result` | Navigate to `/admin/reload` → click "Reload Definitions" button | `Loaded` section appears in result panel |

---

## Test Infrastructure

### DB Isolation (`setup_test.go`)

Integration tests use a shared `testDB *sql.DB` opened once via `TestMain`. Between tests, `truncateTables(t)` issues:
```sql
TRUNCATE workflow_def, workflow_instance, task_instance, state_transition, webhook_outbox RESTART IDENTITY CASCADE
```

### Hand-written Mocks (`internal/testutil/mocks.go`)

All repository interfaces have mock implementations:
- `MockWorkflowDefRepo` — in-memory map, returns `ErrNotFound` on miss
- `MockWorkflowInstanceRepo` — in-memory, `WithTx` returns same instance
- `MockTaskInstanceRepo` — in-memory, supports `BulkCreate`, `ListByInstanceAndState`
- `MockStateTransitionRepo` — append-only slice
- `MockWebhookOutboxRepo` — in-memory pending queue

`TestMockRepos_Compile` ensures all mocks satisfy their interfaces at compile time.

### E2E Helpers (`tests/e2e/specs/helpers.ts`)

```
reloadDefs(request)          — POST /api/v1/admin/reload-defs
startInstance(request, name) — finds def by name, POST /api/v1/workflow-instances
getFirstTask(request, id)    — GET /api/v1/workflow-instances/:id/tasks → first open task
completeTask(request, id, status?, nextState?) — POST /api/v1/task-instances/:id/complete
```
