# Phase 5: Frontend Operations + Polish

## Overview

Add the runtime/operations side of the frontend: starting and managing process instances, viewing workflow and task states, performing task actions (assign, update status, trigger transitions), and viewing state transition audit history. Extends the existing MUI-based admin portal.

## Scope

| Feature | Description |
|---|---|
| Instance API client | API functions for process instances, workflow instances, task instances, transitions |
| Instance types | TypeScript interfaces for all instance models |
| Instances list page | Table of all process instances with status, process name, timestamps |
| Instance detail page | Process instance detail: workflow instances, task instances, current states |
| Task actions | Assign, move to in_progress, complete, fail, skip tasks |
| Manual transitions | Trigger unguarded workflow transitions |
| Audit log | State transition history per workflow instance |
| Sidebar nav update | Add "Instances" link to sidebar |

## New Routes

| Route | Page | Description |
|---|---|---|
| `/instances` | InstanceListPage | Table of all process instances |
| `/instances/:id` | InstanceDetailPage | Detail view: workflows, tasks, transitions |

## Types to Add (`types/index.ts`)

```typescript
interface ProcessInstance {
  id: string;
  process_id: string;
  parent_id: string | null;
  status: string;          // running, completed, cancelled, failed
  started_at: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

interface WorkflowInstance {
  id: string;
  workflow_id: string;
  process_instance_id: string;
  current_state: string;
  status: string;          // active, completed, cancelled
  created_at: string;
  updated_at: string;
}

interface TaskInstance {
  id: string;
  task_id: string;
  workflow_instance_id: string;
  status: string;          // pending, in_progress, completed, skipped, failed
  assigned_to: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

interface StateTransition {
  id: string;
  workflow_instance_id: string;
  from_state: string;
  to_state: string;
  transition_name: string;
  triggered_by: string;
  trigger_type: string;    // human, ai_agent, rule, process
  created_at: string;
}

interface StartProcessInput {
  process_id: string;
  parent_id?: string;
}

interface UpdateTaskStatusInput {
  status: string;
  triggered_by?: string;
  trigger_type?: string;
}

interface AssignTaskInput {
  assigned_to: string;
}

interface TriggerTransitionInput {
  transition_name: string;
  triggered_by: string;
  trigger_type: string;
}
```

## API Client Functions

### `api/processInstances.ts`
```typescript
listProcessInstances(offset?, limit?): Promise<ProcessInstance[]>
getProcessInstance(id): Promise<ProcessInstance>
startProcessInstance(input: StartProcessInput): Promise<ProcessInstance>
cancelProcessInstance(id): Promise<void>
listChildren(id): Promise<ProcessInstance[]>
```

### `api/workflowInstances.ts`
```typescript
listWorkflowInstances(processInstanceId): Promise<WorkflowInstance[]>
getWorkflowInstance(id): Promise<WorkflowInstance>
triggerTransition(id, input: TriggerTransitionInput): Promise<StateTransition>
```

### `api/taskInstances.ts`
```typescript
listTaskInstances(workflowInstanceId): Promise<TaskInstance[]>
getTaskInstance(id): Promise<TaskInstance>
updateTaskStatus(id, input: UpdateTaskStatusInput): Promise<TaskInstance>
assignTask(id, input: AssignTaskInput): Promise<TaskInstance>
```

### `api/stateTransitions.ts`
```typescript
listStateTransitions(workflowInstanceId): Promise<StateTransition[]>
```

**Note:** State transitions endpoint does not exist yet in backend. We will add a `GET /api/v1/workflow-instances/{id}/transitions` route.

## Pages

### InstanceListPage (`/instances`)

- MUI `Table` listing all process instances
- Columns: process name (looked up from process_id), status, started_at, completed_at
- Status chips: `running`=info, `completed`=success, `cancelled`=warning, `failed`=error
- "Start Process" button → dialog with process selector (dropdown of active processes)
- Row click → navigate to `/instances/:id`
- Data loaded from `GET /api/v1/process-instances`
- Process names resolved by fetching processes list once and building a lookup map

### InstanceDetailPage (`/instances/:id`)

Top-level layout:
- Header: process name, instance status chip, started_at, cancel button (if running)
- Section per workflow instance (accordion panels, like ProcessDetailPage)

Each workflow instance section:
- **Header**: workflow name, current state chip, status chip
- **Task instances table**:
  - Columns: task name, status, assigned_to, actions
  - Task name resolved via task definition lookup
  - Actions column (per-row buttons based on current status):
    - `pending` → "Start" button (moves to in_progress)
    - `in_progress` → "Complete" / "Fail" buttons
    - Any non-terminal → "Assign" button → small popover/dialog for username
    - `completed`/`failed`/`skipped` → no actions (read-only)
  - Status chips: `pending`=default, `in_progress`=info, `completed`=success, `failed`=error, `skipped`=warning
- **Manual transitions section** (below tasks):
  - Show available transitions from the workflow's state machine for the current state
  - Each rendered as a button: "Trigger: {transition_name}"
  - Only show transitions without a completion_rule (unguarded) or where completion rule is already satisfied (simpler: show all valid-from-state transitions, backend will reject if guard fails)
  - Clicking opens a small dialog: triggered_by (text field), trigger_type (dropdown: human/ai_agent)
  - On success, reload the instance detail
- **Audit log section** (below transitions):
  - MUI `Table` showing state transition history for this workflow instance
  - Columns: from_state, to_state, transition_name, triggered_by, trigger_type, timestamp
  - Ordered by created_at descending (newest first)
  - Data from `GET /api/v1/workflow-instances/{id}/transitions`

## Backend Addition

### New endpoint: list state transitions

**Route:** `GET /api/v1/workflow-instances/{id}/transitions`

The `StateTransitionRepo.ListByWorkflowInstanceID` method already exists. We need:
1. A handler method on `WorkflowInstanceHandler` or a new `StateTransitionHandler`
2. Register the route in `router.go`

This is a simple read-only endpoint — add it to the existing `WorkflowInstanceHandler` to keep things simple.

## Sidebar Update

Add "Instances" nav item to `Layout.tsx` sidebar:
```
Processes    (AccountTreeIcon)    → /
Instances    (PlayArrowIcon)      → /instances
```

Highlight based on `location.pathname`.

## Test Plan

### Frontend unit tests (Vitest + RTL)

| Test | What it verifies |
|---|---|
| InstanceListPage renders | Shows table, "Start Process" button |
| InstanceListPage empty state | Shows "No instances yet" when list is empty |
| InstanceDetailPage renders | Shows process name, workflow sections, task tables |
| Task action buttons | Correct buttons shown per task status (Start for pending, Complete/Fail for in_progress) |
| API client functions | Correct URL construction for all instance endpoints |

### Backend integration test

| Test | What it verifies |
|---|---|
| TestAPI_StateTransitionsList | `GET /workflow-instances/{id}/transitions` returns transition history after a manual transition |

## File Summary

New files:
```
frontend/src/api/processInstances.ts
frontend/src/api/workflowInstances.ts
frontend/src/api/taskInstances.ts
frontend/src/api/stateTransitions.ts
frontend/src/pages/InstanceListPage.tsx
frontend/src/pages/InstanceDetailPage.tsx
frontend/src/__tests__/InstanceListPage.test.tsx
frontend/src/__tests__/api-instances.test.ts
```

Modified files:
```
frontend/src/types/index.ts          — add instance types
frontend/src/components/Layout.tsx   — add Instances nav item
frontend/src/App.tsx                 — add instance routes
backend/internal/api/handler/workflow_instance.go — add ListTransitions handler
backend/internal/api/router.go       — add transitions route
backend/tests/integration/api_test.go — add transitions test
spec/testDetails.md                  — update with new tests
spec/Currentstatus.md               — update status
```
