# Phase 5 — Frontend Operations

## Goal
Instance lifecycle management: start, monitor, complete tasks, cancel. Adds live graph (current state highlighted), task completion UI per type, admin reload page.

---

## New Routes

| Route | Component | Purpose |
|-------|-----------|---------|
| `/workflow-instances` | `InstanceListPage` | List + filter instances |
| `/workflow-instances/new` | `StartInstancePage` | Start new instance |
| `/workflow-instances/:id` | `InstanceDetailPage` | Live detail + tasks + history |
| `/task-instances/:id` | `TaskDetailPage` | Standalone task deep-link |
| `/admin/reload` | `AdminReloadPage` | Trigger reload-defs |

Sidebar items "Instances" and "Admin" become active links.

---

## Instance List Page (`/workflow-instances`)

**Filters (URL query params, persisted):**
- `def_id` — workflow def picker (searchable dropdown)
- `status` — `active | completed | cancelled | all` (default: all)
- `limit` / `offset` — pagination

**Table columns:** ID (monospace, truncated + copy button), Workflow Def, Status badge, Current State, Started At, Completed At

**Actions:**
- Row click → detail page
- "Start Instance" button (top-right) → `/workflow-instances/new`

**Polling:** every 15s when window focused.

---

## Start Instance Page (`/workflow-instances/new`)

Form fields:
1. **Workflow def** — searchable autocomplete. Shows `name @ v{version}`. Required.
2. **Created by** — free text. Required.
3. **Initial context** — KV editor (add row: key, value textarea) with "Raw JSON" toggle (textarea). Optional. Validated as valid JSON on submit.

On submit: `POST /api/v1/workflow-instances` → on success, redirect to `/workflow-instances/:id`.

Validation: react-hook-form + zod schema. Inline field errors.

"Start Instance" button also available as quick-action on `/workflow-defs/:id` detail page (opens modal with def pre-filled).

---

## Instance Detail Page (`/workflow-instances/:id`)

### Header
```
leave-approval v1   ●active                            [Cancel Instance]
ID: 3f8a...  Started: 2026-06-12 10:00  By: pankaj.gupta
```
Cancel button: confirmation dialog (shadcn `AlertDialog`). On confirm: `POST /instance/:id/cancel`.

### Main Layout (two-panel)
```
┌─ State Graph ─────────────────────────────────┬─ Task Panel ──────────────┐
│  (React Flow — live current state highlighted) │  Open Tasks               │
│  Visited states: filled                        │  [task cards]             │
│  Current state: pulsing indigo ring            │                           │
│  Future states: outlined                       │  Completed Tasks          │
│  Taken edges: solid + colored                  │  [collapsible]            │
│  Untaken edges: dashed                         │                           │
│                                                │  Cancelled Tasks          │
│                                                │  [collapsible]            │
├────────────────────────────────────────────────┴───────────────────────────┤
│  History Timeline (state_transition audit)                                  │
│  ● submitted → hr_review   2026-06-12 10:05                                │
│  ● hr_review → approved    2026-06-12 10:30                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Polling:** 5s while instance is `active`; paused when window blurred; stopped when terminal.

**Optimistic update on task complete:** immediately update local task status + instance state; rollback on API error with toast.

**Stale badge:** shown if last successful poll was >10s ago.

---

## Task Cards (in Instance Detail)

Each open task renders its **type-specific card** from the task renderer registry.

### Task Renderer Registry (`src/components/tasks/registry.ts`)
```ts
export const rendererFor = (type: string): TaskRenderer =>
  registry[type] ?? ManualTaskRenderer

const registry: Record<string, TaskRenderer> = {
  approval:  ApprovalTaskRenderer,
  form:      FormTaskRenderer,
  checklist: ChecklistTaskRenderer,
  manual:    ManualTaskRenderer,
}
```

### Per-Type Renderers

**ApprovalTaskRenderer:**
- Task label + type chip + status badge
- Outcome buttons: [✓ Approve] [✗ Reject]
- Optional comment textarea (sent as `data.comment`)
- If state has >1 outgoing + no triggers (user-picks): shows `next_state` dropdown above outcome buttons

**ManualTaskRenderer:**
- Task label + type chip
- [Mark Done] [Reject] buttons
- Optional note textarea

**FormTaskRenderer (v1 stub):**
- Task label + "Form" chip
- Free-text textarea (sent as `data.note`)
- [Submit] [Reject] buttons

**ChecklistTaskRenderer (v1 stub):**
- Same as ManualTaskRenderer (full checklist in v2)

### Decision State UI (User-Picks)
When state has >1 outgoing + no triggers:
```
[Next step]  ▾  [hr_review | fast_track | rejected]
[Mark Done]  [Reject]
```
`next_state` required before submit; validated client-side.

### Task-Bound State UI
Renderer hides `next_state` selector. Completing the bound task auto-selects edge. UI shows which edge this task fires ("→ hr_review" label, dimmed).

### Task Complete API Call
```ts
POST /api/v1/task-instances/:id/complete
{
  status: "done" | "rejected",
  next_state?: string,         // user-picks only
  completed_by: string,        // from a global session store (free text v1)
  outcome?: string,            // optional, v2 use
  data?: object                // optional, v2 use
}
```

---

## Task Detail Page (`/task-instances/:id`)

Standalone page for external deep-link (e.g., from webhook email notification):
```
Header: task label · type · status badge
Instance: [link to instance] workflow name
[TaskRenderer — same component, full page]
```
Polls every 10s. If task already terminal: shows read-only status + result.

---

## Admin Reload Page (`/admin/reload`)

```
[Reload Workflow Definitions]          (button)

Last result:
  Loaded (3):  leave-approval v1, bug-triage v1, expense-review v2
  Warnings (1): leave-approval v1 — W1: State "triage" has multiple outgoing...
  Errors (0):
```

On click: `POST /api/v1/admin/reload-defs`. Disables button while in-flight. Shows spinner. Result displayed below.

---

## Global State (Minimal)

```ts
// src/lib/session.ts — simple localStorage key
export const getCompletedBy = () => localStorage.getItem('completed_by') ?? ''
export const setCompletedBy = (v: string) => localStorage.setItem('completed_by', v)
```

First task completion: if `completed_by` empty, prompt user to enter name (persisted). Not authentication — just identifier.

---

## New API Modules

```ts
// src/api/instances.ts
export const instancesApi = {
  list:    (filter: InstanceFilter) => apiFetch<InstanceListResponse>('/api/v1/workflow-instances?' + qs),
  getById: (id: string) => apiFetch<WorkflowInstanceDetailResponse>(`/api/v1/workflow-instances/${id}`),
  start:   (req: StartInstanceRequest) => apiFetch<WorkflowInstanceResponse>('/api/v1/workflow-instances', { method: 'POST', body: JSON.stringify(req) }),
  cancel:  (id: string) => apiFetch<WorkflowInstanceResponse>(`/api/v1/workflow-instances/${id}/cancel`, { method: 'POST' }),
}

// src/api/tasks.ts
export const tasksApi = {
  getById:  (id: string) => apiFetch<TaskInstanceResponse>(`/api/v1/task-instances/${id}`),
  start:    (id: string) => apiFetch<TaskInstanceResponse>(`/api/v1/task-instances/${id}/start`, { method: 'POST' }),
  complete: (id: string, req: CompleteTaskRequest) => apiFetch<WorkflowInstanceResponse>(`/api/v1/task-instances/${id}/complete`, { method: 'POST', body: JSON.stringify(req) }),
  listByInstance: (instanceId: string, status?: string) => apiFetch<TaskInstanceResponse[]>(`/api/v1/workflow-instances/${instanceId}/tasks`),
}

// src/api/admin.ts
export const adminApi = {
  reloadDefs: () => apiFetch<ReloadResult>('/api/v1/admin/reload-defs', { method: 'POST' }),
}
```

---

## New Hooks

```ts
useInstances(filter)               // with polling 15s
useInstance(id)                    // with polling 5s, stops when terminal
useTasksByInstance(instanceId)
useCompleteTask()                  // mutation, optimistic
useCancelInstance()                // mutation
useReloadDefs()                    // mutation
```

---

## Tests (Vitest + RTL)

- `InstanceListPage`: renders list, shows filters, pagination
- `StartInstancePage`: form validation, submit fires correct API, redirects on success
- `InstanceDetailPage`: shows live graph, task cards, history timeline
- `ApprovalTaskRenderer`: approve fires complete with status=done, reject fires with status=rejected, next_state required when user-picks
- `TaskDetailPage`: renders task, shows read-only when terminal
- `AdminReloadPage`: button calls reload API, shows result

---

## Exit Criteria
- [ ] Start instance flow works end-to-end (form → instance detail)
- [ ] Task completion progresses instance state (graph updates)
- [ ] Rejection cancels instance + shows cancelled status
- [ ] Decision state: next_state dropdown present + validated
- [ ] Task-bound: sibling task cards show `cancelled` status after bound task completed
- [ ] Instance cancel: all open tasks show cancelled
- [ ] Deep-link `/task-instances/:id` renders and allows completion
- [ ] Admin reload page shows correct warnings
- [ ] Polling stops for terminal instances
- [ ] `npm test` green
