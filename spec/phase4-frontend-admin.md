# Phase 4 — Frontend Admin

## Goal
React scaffold + workflow definition browsing + read-only state graph visualization. No instance operations (Phase 5).

---

## Stack
- React 18 + TypeScript + Vite
- Tailwind CSS + shadcn/ui
- TanStack Query v5
- React Router v6
- React Flow + dagre (graph)
- Framer Motion (animations)
- lucide-react (icons)
- Vitest + React Testing Library

Vite dev proxy: `server.proxy: { '/api': 'http://localhost:8080' }`

---

## Routes (Phase 4)

| Route | Component |
|-------|-----------|
| `/` | Redirect → `/workflow-defs` |
| `/workflow-defs` | `WorkflowDefListPage` |
| `/workflow-defs/:id` | `WorkflowDefDetailPage` |
| `/*` | `NotFoundPage` |

---

## Layout Shell (`src/components/layout/`)

```
┌─ Topbar ────────────────────────────────────────────────────────┐
│  [≡] Workflow Engine    ●health   [☀/☾ theme toggle]            │
├─ Sidebar ──┬─ Main Content ──────────────────────────────────────┤
│ Workflows  │                                                      │
│ Instances  │  (page content)                                      │
│ Task Types │                                                      │
│ Admin      │                                                      │
└────────────┴──────────────────────────────────────────────────────┘
```

Sidebar collapsible (icon-only mode). Active route highlighted. Sidebar items Phase 4 navigable; Instances + Admin grayed out until Phase 5 wires them.

Health dot: green (`/readyz` 200) / red (non-200) — polled every 30s.

Theme: system default on first load. Toggle stores in `localStorage`. Dark + light both fully designed.

---

## Workflow Def List Page (`/workflow-defs`)

**Features:**
- Table view (default) / Card grid toggle (persisted in `localStorage`)
- Search by name (client-side filter on loaded data — list is small)
- `include_deleted` toggle (shows soft-deleted rows with muted styling + "Deleted" badge)
- Sort: name (asc/desc), version (desc), created_at (desc)

**Table columns:** Name, Version, Description (truncated), States (#), Transitions (#), Created At, Status (Active / Deleted badge)

**Card fields:** Name, version badge, description, state count, created date, status badge

Click row/card → navigate to detail.

Empty state: illustration + "No workflow definitions found. Add JSON files to the workflows/ directory and reload."

Loading: skeleton rows (5).

---

## Workflow Def Detail Page (`/workflow-defs/:id`)

### Layout
```
┌─ Header ─────────────────────────────────────────────────────────────┐
│  "leave-approval"  v1   Active                                        │
│  Description text                   [↓ Download JSON]                │
├──────────────────────────────────────┬───────────────────────────────┤
│  State Graph (React Flow)            │  State Inspector Panel        │
│  (resizable / full-screen toggle)    │  (selected state details)     │
│                                      │                               │
├──────────────────────────────────────┴───────────────────────────────┤
│  Warnings banner (if any)                                            │
└──────────────────────────────────────────────────────────────────────┘
```

### State Graph (React Flow + dagre)

**Auto-layout:** dagre `rankdir: 'LR'` (left-to-right). Nodes sized to content.

**Node types:**
- `start` node: green outline + "START" chip
- `end` node: gold outline + "END" chip
- `regular` node: neutral bg, state label, task count badge

**Edge types:**
- Linear / user-picks: simple arrow, no label
- Task-bound: labeled with `trigger.task_key`
- All edges: animated dot on hover

**Interactivity (read-only):**
- Pan + zoom (React Flow default controls)
- Click node → highlights + populates inspector panel
- Fit-to-view button

### State Inspector Panel
Shown on right when node selected:
```
State ID:    "hr_review"
Label:       HR Review
Type:        Regular

Tasks (2):
  ┌─ hr_check · HR Verification · approval ─┐
  │                                          │
  └──────────────────────────────────────────┘
  ┌─ policy_check · Policy Compliance · checklist ─┐
  │                                                 │
  └─────────────────────────────────────────────────┘

Outgoing Transitions:
  → approved   (linear)
```

### Warnings Banner
If def has warnings: amber banner at top with list of warning messages. Dismissible per session.

### Download JSON
Button triggers `window.open('/api/v1/workflow-defs/:id')` — returns full def with JSONB.

---

## API Client (`src/api/`)

```ts
// client.ts
async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T>
// throws ApiError with code + message on non-2xx

// processes.ts (workflow defs)
export const workflowDefsApi = {
  list: (includeDeleted?: boolean) => apiFetch<WorkflowDefListResponse>('/api/v1/workflow-defs'),
  getById: (id: string) => apiFetch<WorkflowDefResponse>(`/api/v1/workflow-defs/${id}`),
}
```

---

## Types (`src/types/index.ts`)

```ts
export interface WorkflowDef {
  id: string
  name: string
  version: number
  description: string
  definition: WorkflowDefinition
  created_at: string
  deleted_at?: string
}

export interface WorkflowDefinition {
  name: string
  version: number
  start_state: string
  end_state: string
  states: StateDef[]
  transitions: TransitionDef[]
}

export interface StateDef {
  id: string
  label: string
  tasks: TaskDef[]
}

export interface TaskDef {
  key: string
  label: string
  type: string
}

export interface TransitionDef {
  from_state: string
  to_state: string
  trigger?: { task_key: string }
}
```

---

## TanStack Query Setup

```ts
// App.tsx
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
})

// hooks/useWorkflowDefs.ts
export function useWorkflowDefs(includeDeleted = false) {
  return useQuery({
    queryKey: ['workflow-defs', { includeDeleted }],
    queryFn: () => workflowDefsApi.list(includeDeleted),
  })
}

export function useWorkflowDef(id: string) {
  return useQuery({
    queryKey: ['workflow-defs', id],
    queryFn: () => workflowDefsApi.getById(id),
    enabled: !!id,
  })
}
```

---

## Graph Layout Helper (`src/components/graph/layout.ts`)

```ts
import dagre from '@dagrejs/dagre'

export function layoutGraph(def: WorkflowDefinition): { nodes: Node[], edges: Edge[] } {
  const g = new dagre.graphlib.Graph()
  g.setGraph({ rankdir: 'LR', nodesep: 60, ranksep: 80 })
  g.setDefaultEdgeLabel(() => ({}))
  // add nodes + edges
  dagre.layout(g)
  // return React Flow nodes + edges with positions
}
```

---

## Tests (Vitest + RTL)

- `WorkflowDefListPage`: renders list, shows skeleton on load, shows empty state, filters by search
- `WorkflowDefDetailPage`: renders graph nodes/edges, inspector updates on node click, warning banner shown
- `api/client.ts`: throws `ApiError` on non-2xx, passes correct headers
- `graph/layout.ts`: correct node positions for linear chain, multi-edge state

---

## Exit Criteria
- [ ] `npm run dev` works with Vite proxy to local backend
- [ ] `/workflow-defs` lists defs from real backend
- [ ] `/workflow-defs/:id` shows React Flow graph with correct nodes/edges
- [ ] Dark + light themes both look correct
- [ ] `npm test` green
- [ ] No TypeScript errors (`npm run build` succeeds)
