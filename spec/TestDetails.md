# Test Details

All tests documented here, organized by layer. Use this as a quality review reference.

---

## How to run

```bash
# Unit tests (no DB required)
go test -cover ./backend/internal/...

# Integration tests (requires PostgreSQL on port 5432/5433)
go test -tags=integration -cover -coverpkg=./backend/internal/... ./backend/tests/integration/...

# Frontend tests
cd frontend && npm test
```

Go does not guarantee test order within a package — tests run in the order they appear in the file but packages run in dependency order. Integration tests each call `cleanupTables` or `setupTestServer` (which calls `cleanupTables`) at the start, so they are fully isolated regardless of order.

---

## 1. Unit Tests — `backend/internal/model`

Tests validate model structs using their `Validate()` method. No DB, no mocks.

### `model/process_test.go` — 6 tests

| Test | What it checks |
|------|---------------|
| `TestProcess_Validate_Valid` | A fully populated process with valid fields passes validation |
| `TestProcess_Validate_AllStatuses` | Each of `draft`, `active`, `archived` is accepted |
| `TestProcess_Validate_EmptyName` | Empty name returns error containing "name is required" |
| `TestProcess_Validate_LongName` | Name > 255 chars returns error mentioning "255" |
| `TestProcess_Validate_InvalidStatus` | Unknown status (e.g. `"invalid"`) returns error |
| `TestProcess_Validate_ZeroVersion` | Version = 0 returns error |
| `TestProcess_Validate_NegativeVersion` | Version = -1 returns error |

### `model/workflow_test.go` — 4 tests

| Test | What it checks |
|------|---------------|
| `TestWorkflow_Validate_Valid` | Valid workflow with a valid state machine passes |
| `TestWorkflow_Validate_EmptyName` | Empty name returns error containing "name is required" |
| `TestWorkflow_Validate_LongName` | Name > 255 chars returns error |
| `TestWorkflow_Validate_NilProcessID` | UUID nil process_id returns error |
| `TestWorkflow_Validate_InvalidStateMachine` | State machine with no states causes workflow validation to fail |

### `model/task_test.go` — 6 tests

| Test | What it checks |
|------|---------------|
| `TestTask_Validate_Valid` | Valid task passes |
| `TestTask_Validate_AllTypes` | Each of `manual`, `automated`, `approval` is accepted |
| `TestTask_Validate_EmptyName` | Empty name returns error containing "name is required" |
| `TestTask_Validate_LongName` | Name > 255 chars returns error |
| `TestTask_Validate_NilWorkflowID` | UUID nil workflow_id returns error |
| `TestTask_Validate_InvalidType` | Unknown task_type returns error |

### `model/state_machine_test.go` — 19 tests

StateMachine validation tests:

| Test | What it checks |
|------|---------------|
| `TestStateMachine_Validate_Valid` | A well-formed state machine passes |
| `TestStateMachine_Validate_NoStates` | No states returns error containing "at least one state" |
| `TestStateMachine_Validate_EmptyStateName` | State with empty name returns error |
| `TestStateMachine_Validate_DuplicateStateName` | Two states with the same name returns error containing "duplicate state name" |
| `TestStateMachine_Validate_EmptyInitialState` | Missing initial_state returns error |
| `TestStateMachine_Validate_InvalidInitialState` | initial_state referencing non-existent state returns error |
| `TestStateMachine_Validate_InvalidTerminalState` | terminal_states referencing non-existent state returns error |
| `TestStateMachine_Validate_EmptyTransitionName` | Transition with empty name returns error |
| `TestStateMachine_Validate_DuplicateTransitionName` | Two transitions with same name returns error |
| `TestStateMachine_Validate_InvalidFromState` | Transition from_state referencing non-existent state returns error |
| `TestStateMachine_Validate_InvalidToState` | Transition to_state referencing non-existent state returns error |
| `TestStateMachine_Validate_WithCompletionRule` | Valid completion rule on a transition passes |
| `TestStateMachine_Validate_InvalidCompletionRule` | Completion rule with invalid type returns error |

CompletionRule validation tests:

| Test | What it checks |
|------|---------------|
| `TestCompletionRule_Validate_Single` | `single` with exactly 1 task_ref passes |
| `TestCompletionRule_Validate_SingleMultipleRefs` | `single` with 2 task_refs returns error |
| `TestCompletionRule_Validate_AllOf` | `all_of` with multiple task_refs passes |
| `TestCompletionRule_Validate_AnyOf` | `any_of` with multiple task_refs passes |
| `TestCompletionRule_Validate_NOfM_Valid` | `n_of_m` with N < len(task_refs) passes |
| `TestCompletionRule_Validate_NOfM_ZeroN` | `n_of_m` with N = 0 returns error |
| `TestCompletionRule_Validate_NOfM_NGreaterThanM` | `n_of_m` with N > len(task_refs) returns error |
| `TestCompletionRule_Validate_EmptyTaskRefs` | Empty task_refs returns error |
| `TestCompletionRule_Validate_InvalidType` | Unknown type returns error |

### `model/instances_test.go` — 13 tests

| Test | What it checks |
|------|---------------|
| `TestProcessInstance_Validate_Valid` | Valid process instance passes |
| `TestProcessInstance_Validate_AllStatuses` | `running`, `completed`, `cancelled`, `failed` all pass |
| `TestProcessInstance_Validate_NilProcessID` | Nil process_id returns error |
| `TestProcessInstance_Validate_InvalidStatus` | Unknown status returns error |
| `TestWorkflowInstance_Validate_Valid` | Valid workflow instance passes |
| `TestWorkflowInstance_Validate_AllStatuses` | `active`, `completed`, `cancelled` all pass |
| `TestWorkflowInstance_Validate_MissingFields` | Table-driven: nil workflow_id, nil process_instance_id, empty current_state, invalid status each return errors |
| `TestTaskInstance_Validate_Valid` | Valid task instance passes |
| `TestTaskInstance_Validate_AllStatuses` | `pending`, `in_progress`, `completed`, `skipped`, `failed` all pass |
| `TestTaskInstance_Validate_MissingFields` | Table-driven: nil task_id, nil workflow_instance_id, invalid status each return errors |
| `TestStateTransition_Validate_Valid` | Valid state transition passes |
| `TestStateTransition_Validate_AllTriggerTypes` | `human`, `ai_agent`, `rule`, `process` all pass |
| `TestStateTransition_Validate_MissingFields` | Table-driven: nil workflow_instance_id, empty from_state, empty to_state, invalid trigger_type each return errors |

---

## 2. Unit Tests — `backend/internal/engine`

Tests use in-memory mock repositories. No DB. The `setupEngine` helper builds a full engine with mocked repos populated with the given state machine and task names.

### `engine/engine_test.go` — 15 tests

**Evaluate (auto-transition based on completion rules):**

| Test | What it checks |
|------|---------------|
| `TestEvaluate_NoTransitionsAvailable` | Transition with no completion rule → returns nil (no auto-transition) |
| `TestEvaluate_SingleRule_Satisfied` | Task completed → transitions to terminal state, IsTerminal = true |
| `TestEvaluate_SingleRule_NotSatisfied` | Task still pending → returns nil |
| `TestEvaluate_AllOf_Satisfied` | Both tasks completed → transitions |
| `TestEvaluate_AllOf_Partial` | Only one of two tasks completed → returns nil |
| `TestEvaluate_AnyOf_OneDone` | One of two tasks done → transitions |
| `TestEvaluate_NOfM_Satisfied` | 2 of 3 tasks done, N=2 → transitions |
| `TestEvaluate_NOfM_NotEnough` | 1 of 3 tasks done, N=2 → returns nil |
| `TestEvaluate_TerminalState_Error` | Workflow already in terminal state → returns error |
| `TestEvaluate_NonActiveWorkflow_Error` | Workflow instance status is `cancelled` → returns error |
| `TestEvaluate_TransitionSetsWorkflowCompleted` | When transition lands on terminal state, workflow instance status → `completed` |

**TriggerTransition (manual transition):**

| Test | What it checks |
|------|---------------|
| `TestTriggerTransition_Unguarded_Success` | Manual trigger succeeds, state updated, transition logged with triggered_by |
| `TestTriggerTransition_NotFound` | Non-existent transition name → returns error |
| `TestTriggerTransition_WrongState` | Transition valid for a different state → returns error |
| `TestTriggerTransition_TerminalState_Error` | Workflow in terminal state → returns error |
| `TestTriggerTransition_GuardedNotSatisfied` | Guarded transition but task not done → returns error |
| `TestTriggerTransition_GuardedSatisfied` | Guarded transition and task completed → succeeds, IsTerminal = true |

---

## 3. Unit Tests — `backend/internal/service`

Tests use mock repositories from `internal/testutil`. No DB.

### `service/process_test.go` — 6 tests

| Test | What it checks |
|------|---------------|
| `TestProcessService_Create` | Creates process with correct defaults: status=draft, version=1, UUID assigned |
| `TestProcessService_Create_EmptyName` | Empty name returns error |
| `TestProcessService_Update` | Updates name field correctly |
| `TestProcessService_Update_ArchivedProcess` | Archived process cannot be updated |
| `TestProcessService_Delete_Draft` | Draft process can be deleted; GetByID returns ErrNotFound after |
| `TestProcessService_Delete_ActiveProcess` | Active process cannot be deleted |
| `TestProcessService_List_DefaultLimit` | Creates 5 processes, List(0,0) returns all 5 |

### `service/workflow_test.go` — 5 tests

| Test | What it checks |
|------|---------------|
| `TestWorkflowService_Create` | Creates workflow linked to process, state machine preserved |
| `TestWorkflowService_Create_ArchivedProcess` | Cannot create workflow on archived process |
| `TestWorkflowService_Create_InvalidStateMachine` | Empty state machine fails validation |
| `TestWorkflowService_Delete_DraftProcess` | Workflow on draft process can be deleted |
| `TestWorkflowService_Delete_ActiveProcess` | Workflow on active process cannot be deleted |
| `TestWorkflowService_Update` | Updates workflow name correctly |

### `service/process_instance_test.go` — 6 tests

| Test | What it checks |
|------|---------------|
| `TestProcessInstanceService_Start` | Starting active process creates instance (status=running), 1 workflow instance at initial state, 1 task instance per task |
| `TestProcessInstanceService_Start_DraftProcess` | Draft process cannot be started |
| `TestProcessInstanceService_Start_ParentChild` | Child can be created with parent_id; grandchild (depth > 1) is rejected |
| `TestProcessInstanceService_Start_ParentNotRunning` | Cannot create child under a completed parent |
| `TestProcessInstanceService_Cancel` | Cancels instance (status=cancelled, CompletedAt set), all workflow instances also cancelled |
| `TestProcessInstanceService_Cancel_NotRunning` | Cancelling an already-cancelled instance returns error |

### `service/task_instance_test.go` — 5 tests

| Test | What it checks |
|------|---------------|
| `TestTaskInstanceService_UpdateStatus_Complete` | `in_progress → completed` sets CompletedAt, calls engine.Evaluate once |
| `TestTaskInstanceService_UpdateStatus_InvalidTransition` | Table-driven: `pending→completed`, `pending→failed`, `completed→in_progress`, `in_progress→pending` all rejected |
| `TestTaskInstanceService_UpdateStatus_InactiveWorkflow` | Cannot update task on a completed workflow instance |
| `TestTaskInstanceService_Assign` | Sets AssignedTo field |
| `TestTaskInstanceService_UpdateStatus_PendingToInProgress` | `pending → in_progress` succeeds; engine.Evaluate NOT called (only called on completion) |

---

## 4. Integration Tests — Repository Layer (`backend/tests/integration`)

Require PostgreSQL. Each test calls `cleanupTables` at the start for isolation. These test real SQL against the `workflow_engine` schema.

### `process_repo_test.go` — 7 tests

| Test | What it checks |
|------|---------------|
| `TestProcessRepo_Create_And_GetByID` | Round-trip: create then retrieve, verify name/status/version |
| `TestProcessRepo_GetByID_NotFound` | Random UUID → ErrNotFound |
| `TestProcessRepo_List` | Create 3, list → 3 results |
| `TestProcessRepo_List_Pagination` | Create 5; page1(0,2)=2, page2(2,2)=2 |
| `TestProcessRepo_Update` | Update name and status, verify persisted |
| `TestProcessRepo_Update_NotFound` | Update non-existent ID → ErrNotFound |
| `TestProcessRepo_Delete` | Delete then GetByID → ErrNotFound |
| `TestProcessRepo_Delete_NotFound` | Delete non-existent ID → ErrNotFound |

### `workflow_repo_test.go` — 7 tests

| Test | What it checks |
|------|---------------|
| `TestWorkflowRepo_Create_And_GetByID` | Round-trip, verify state machine JSONB (states, initial_state) |
| `TestWorkflowRepo_Create_ForeignKeyViolation` | Non-existent process_id → ErrForeignKey |
| `TestWorkflowRepo_ListByProcessID` | Create 3 workflows for a process, list → 3 results |
| `TestWorkflowRepo_ListByProcessID_Empty` | No workflows for a random ID → 0 results |
| `TestWorkflowRepo_Update` | Update name, verify persisted |
| `TestWorkflowRepo_Delete` | Delete then GetByID → ErrNotFound |
| `TestWorkflowRepo_GetByID_NotFound` | Random UUID → ErrNotFound |

### `task_repo_test.go` — 5 tests

| Test | What it checks |
|------|---------------|
| `TestTaskRepo_Create_And_GetByID` | Round-trip, verify name and task_type |
| `TestTaskRepo_Create_ForeignKeyViolation` | Non-existent workflow_id → ErrForeignKey |
| `TestTaskRepo_ListByWorkflowID` | Create 3 tasks, list → 3 results |
| `TestTaskRepo_Update` | Update name and task_type, verify persisted |
| `TestTaskRepo_Delete` | Delete then GetByID → ErrNotFound |

### `instance_repo_test.go` — 8 tests

| Test | What it checks |
|------|---------------|
| `TestProcessInstanceRepo_Create_And_GetByID` | Round-trip, verify status and nil parent_id |
| `TestProcessInstanceRepo_ParentChild` | Create parent + child, ListByParentID → 1 child |
| `TestProcessInstanceRepo_ForeignKeyViolation` | Non-existent process_id → ErrForeignKey |
| `TestProcessInstanceRepo_Update` | Set status=completed and CompletedAt, verify persisted |
| `TestProcessInstanceRepo_List` | Create 3 instances, list → 3 results |
| `TestWorkflowInstanceRepo_Create_And_GetByID` | Round-trip, verify current_state |
| `TestWorkflowInstanceRepo_ListByProcessInstanceID` | Create 2 workflow instances, list → 2 results |
| `TestWorkflowInstanceRepo_Update` | Update current_state and status to completed, verify persisted |
| `TestTaskInstanceRepo_CRUD` | Full create/read/update/list cycle including assigned_to and completed_at |
| `TestStateTransitionRepo_Create_And_List` | Create transition, list by workflow_instance_id → 1 result with correct from_state and trigger_type |

---

## 5. Integration Tests — API Layer (`backend/tests/integration/api_test.go`)

Full HTTP tests using `httptest.NewServer` with a real PostgreSQL DB. Each test calls `setupTestServer` which calls `cleanupTables` — fully isolated.

### Order tests appear in file (Go runs them in this order within the package):

| # | Test | What it checks |
|---|------|----------------|
| 1 | `TestAPI_ProcessCRUD` | Full create → get → list → update → delete cycle; verifies list length = 1; confirms 404 after delete |
| 2 | `TestAPI_Process_CreateEmptyName` | `POST /processes` with empty name → 400 |
| 3 | `TestAPI_Process_DeleteActiveBlocked` | Create process, activate it, attempt delete → 409 |
| 4 | `TestAPI_WorkflowCRUD` | Create process, create workflow (with state machine), list, get, update, delete |
| 5 | `TestAPI_TaskCRUD` | Create process + workflow, create task, list, get, update, delete |
| 6 | `TestAPI_EmptyListReturnsArray` | `GET /processes`, `GET /process-instances`, `GET /processes/:id/workflows` on empty DB all return `[]` not `null` |
| 7 | `TestAPI_ProcessInstanceLifecycle` | Start active process → running, get, list (1 result), list workflow instances (1), cancel → cancelled |
| 8 | `TestAPI_ProcessInstance_StartDraftBlocked` | Start instance on draft process → 409 |
| 9 | `TestAPI_ProcessInstance_ParentChild` | Start parent, start child with parent_id, list children (1), start grandchild → 409 |
| 10 | `TestAPI_TaskInstanceStatusAndTransition` | Full: start instance → get workflow instance (state=open) → get task instances → assign → in_progress → complete → verify workflow state=done, status=completed |
| 11 | `TestAPI_TaskInstance_InvalidTransition` | `pending → completed` directly → 409 |
| 12 | `TestAPI_WorkflowInstance_ManualTransition` | `POST /workflow-instances/:id/transition` with `close` transition → 200 with from_state=open, to_state=closed |
| 13 | `TestAPI_NotFoundError` | GET non-existent process UUID → 404 |
| 14 | `TestAPI_InvalidUUID` | GET `/processes/not-a-uuid` → 400 |
| 15 | `TestAPI_HealthCheck` | `GET /health` → 200 |
| 16 | `TestAPI_CORSHeaders` | `OPTIONS /api/v1/processes` → 204 with `Access-Control-Allow-Origin: *` |
| 17 | `TestAPI_StateTransitionsList` | `GET /workflow-instances/:id/transitions` returns empty initially, 1 after a manual transition, 2 after a second transition; verifies from/to states and triggered_by |

---

## 6. Frontend Tests (`frontend/src/__tests__`)

Run with Vitest + jsdom + React Testing Library. Each test file gets a fresh module scope.

### `api-client.test.ts` — 4 tests

| Test | What it checks |
|------|----------------|
| `makes a GET request and returns JSON` | `request('/processes')` calls `fetch` with correct URL and Content-Type header, returns parsed JSON |
| `returns undefined for 204 responses` | DELETE returning 204 → result is undefined |
| `throws ApiError on non-ok response` | 404 response with JSON body → throws `ApiError` with the error message from body |
| `handles non-JSON error bodies` | 500 response where `json()` throws → throws ApiError with generic "Request failed with status 500" |

### `ProcessListPage.test.tsx` — 4 tests

`processApi` is mocked. Default mock returns 2 processes: `Bug Triage` (draft) and `Onboarding` (active).

| Test | What it checks |
|------|----------------|
| `renders the process table` | After load, both process names appear in the DOM |
| `shows the New Process button` | "New Process" button is present on render |
| `opens create dialog on button click` | Clicking "New Process" opens a dialog with `<h2>New Process</h2>` |
| `shows empty state when no processes` | Mock returns `[]` → "No processes yet" text appears |

### `InstanceListPage.test.tsx` — 4 tests

`processApi` and `processInstanceApi` are mocked. Default mock returns 1 active process and 2 instances (running + completed).

| Test | What it checks |
|------|----------------|
| `renders the instance table` | After load, both instances show process name, running and completed status chips |
| `shows the Start Process button` | "Start Process" button is present on render |
| `opens start dialog on button click` | Clicking "Start Process" opens a dialog with `<h2>Start Process</h2>` |
| `shows empty state when no instances` | Mock returns `[]` → "No instances yet" text appears |

### `api-instances.test.ts` — 5 tests

| Test | What it checks |
|------|----------------|
| `listProcessInstances calls correct URL` | Calls `/api/v1/process-instances?offset=0&limit=20` |
| `startProcessInstance sends POST with body` | POST to `/api/v1/process-instances` with `{ process_id }` |
| `listWorkflowInstances calls correct URL` | Calls `/api/v1/process-instances/{id}/workflows` |
| `listTaskInstances calls correct URL` | Calls `/api/v1/workflow-instances/{id}/tasks` |
| `listStateTransitions calls correct URL` | Calls `/api/v1/workflow-instances/{id}/transitions` |

---

## Coverage summary

| Layer | Type | Tests |
|-------|------|-------|
| model | unit | 48 |
| engine | unit | 15 |
| service | unit | 22 |
| repository (postgres) | integration | 37 |
| API handlers | integration | 17 |
| Frontend API client | unit | 9 |
| Frontend components | unit | 8 |
| **Total** | | **156** |

### Known gaps

- **Empty list at repo layer**: `TestWorkflowRepo_ListByProcessID_Empty` covers the workflow repo. Process, task, process_instance, workflow_instance, task_instance repos do not have a dedicated empty-list repo test (covered at the API layer by `TestAPI_EmptyListReturnsArray`).
- **Task service**: No unit tests for `TaskService` (create/update/delete task). Covered indirectly by API integration tests.
- **Frontend pages**: `ProcessDetailPage`, `InstanceDetailPage`, `StateMachineEditor`, form components (`ProcessForm`, `WorkflowForm`, `TaskForm`) have no unit tests.
- **Error paths**: Service-level error logging paths (e.g. engine.Evaluate failure after task completion) are not tested via unit tests — covered by integration test `TestAPI_TaskInstanceStatusAndTransition`.
- **Frontend instance detail**: `InstanceDetailPage` task actions (Start, Complete, Fail, Assign) and transition triggering are not covered by frontend unit tests — tested manually or via integration tests.
