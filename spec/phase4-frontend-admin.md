# Phase 4: Frontend Setup + Admin Portal

## Overview

React + TypeScript app (Vite) with MUI (Material UI) for the admin portal. Covers process/workflow/task definition management and state machine configuration.

## Tech Stack

| Tool | Purpose |
|---|---|
| React 18 | UI framework |
| TypeScript | Type safety |
| Vite | Build tool + dev server |
| MUI (Material UI) | UI component library (tables, forms, dialogs, layout) |
| React Router | Client-side routing |
| Vitest + React Testing Library | Unit tests |

## Directory Structure

```
frontend/
├── src/
│   ├── api/                    # API client functions (one file per resource)
│   │   ├── client.ts           # Base fetch wrapper
│   │   ├── processes.ts
│   │   ├── workflows.ts
│   │   └── tasks.ts
│   ├── types/                  # TypeScript type definitions matching backend models
│   │   └── index.ts
│   ├── pages/                  # Page-level components (one per route)
│   │   ├── ProcessListPage.tsx
│   │   ├── ProcessDetailPage.tsx
│   │   └── NotFoundPage.tsx
│   ├── components/             # Reusable UI components
│   │   ├── Layout.tsx          # App shell with sidebar nav
│   │   ├── ProcessForm.tsx
│   │   ├── WorkflowForm.tsx
│   │   ├── TaskForm.tsx
│   │   └── StateMachineEditor.tsx
│   ├── App.tsx                 # Routes + layout
│   ├── main.tsx                # Entry point
│   └── index.css               # Global styles
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── vitest.config.ts
```

## Pages & Routes

| Route | Page | Description |
|---|---|---|
| `/` | ProcessListPage | Table of all process definitions with create/edit/delete |
| `/processes/:id` | ProcessDetailPage | Process detail with workflow list, task list per workflow, state machine view |
| `*` | NotFoundPage | 404 fallback |

## Page Details

### ProcessListPage (`/`)
- MUI (Material UI) `Table` listing all processes (name, status, version, created_at)
- Status column with color tags (draft=blue, active=green, archived=gray)
- "New Process" button → modal with `ProcessForm`
- Row actions: Edit (modal), Delete (confirm popover), Activate/Archive (status change)

### ProcessDetailPage (`/processes/:id`)
- Header: process name, status tag, description
- **Workflows section**: collapsible panels, one per workflow
  - Each panel shows: workflow name, state machine summary
  - "Add Workflow" button → modal with `WorkflowForm`
  - Per-workflow actions: Edit, Delete
  - **Tasks sub-section** inside each workflow panel
    - Table of tasks (name, type)
    - "Add Task" button → modal with `TaskForm`
    - Per-task actions: Edit, Delete
  - **State Machine section** inside each workflow panel
    - Visual display of states and transitions
    - `StateMachineEditor` component for editing

## API Client Layer

### Base client (`api/client.ts`)
```typescript
const API_BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T>
// Handles: JSON encoding, error extraction, base URL
```

### Resource clients
Each file exports functions that call the backend API:
```typescript
// processes.ts
listProcesses(offset?, limit?): Promise<Process[]>
getProcess(id): Promise<Process>
createProcess(input): Promise<Process>
updateProcess(id, input): Promise<Process>
deleteProcess(id): Promise<void>
```

Same pattern for workflows and tasks.

## TypeScript Types (`types/index.ts`)

Mirror backend models:
```typescript
interface Process { id: string; name: string; description: string; version: number; status: string; created_at: string; updated_at: string }
interface Workflow { id: string; process_id: string; name: string; description: string; state_machine: StateMachine; sort_order: number; ... }
interface Task { id: string; workflow_id: string; name: string; description: string; task_type: string; ... }
interface StateMachine { states: State[]; transitions: Transition[]; initial_state: string; terminal_states: string[] }
interface State { name: string; description: string; metadata?: Record<string, string> }
interface Transition { name: string; from_state: string; to_state: string; completion_rule?: CompletionRule }
interface CompletionRule { type: string; task_refs: string[]; n?: number }
```

## Components

### Layout
- MUI (Material UI) `Layout` with `Sider` (sidebar nav) + `Content`
- Sidebar: "Processes" link (expandable later with Instances)

### ProcessForm
- Fields: name (required), description (optional)
- Used for both create and edit (edit pre-fills values)

### WorkflowForm
- Fields: name (required), description (optional)
- State machine configured separately via StateMachineEditor

### TaskForm
- Fields: name (required), description (optional), task_type (select: manual/automated/approval)

### StateMachineEditor
- Add/remove states (name + description)
- Add/remove transitions (name, from_state dropdown, to_state dropdown)
- Set initial state (dropdown)
- Set terminal states (multi-select)
- Completion rule per transition (optional): type dropdown + task_refs multi-select + N input for n_of_m
- Task refs populated from the workflow's task list

## Vite Dev Proxy

`vite.config.ts` proxies `/api` to `http://localhost:8080` so frontend dev server talks to the Go backend.

## Test Plan

Unit tests with Vitest + React Testing Library:

| Test | What it verifies |
|---|---|
| ProcessListPage renders | Shows table, create button |
| ProcessListPage create flow | Modal opens, form submits, table refreshes |
| ProcessDetailPage renders | Shows process info, workflows, tasks |
| ProcessForm validation | Name required |
| WorkflowForm submit | Calls API with correct payload |
| TaskForm submit | Calls API with correct payload |
| StateMachineEditor | Add/remove states and transitions |
| API client functions | Correct URL construction, error handling |
