# Task Extensions — Plan 4: Frontend

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add frontend API modules for task updates, comments, and attachments; extend types; add priority badge, comments panel, attachment section, and ad-hoc task modal to the instance detail page.

**Architecture:** Three new API modules in `src/api/`. `src/types/index.ts` gains new types. Task card in instance detail gains priority badge and expected-completion display. New `CommentsPanel`, `AttachmentSection`, and `AdhocTaskModal` components. TanStack Query v5 for all server state.

**Tech Stack:** React 18, TypeScript, Vite, TanStack Query v5, Tailwind CSS v4, shadcn/ui.

## Global Constraints

- TanStack Query v5 API: `useQuery({ queryKey, queryFn })` — no legacy options object pattern.
- `apiFetch` from `src/api/client.ts` for all HTTP calls.
- `ApiError` from `src/types/index.ts` has `{ code, message, details? }` — no constructor params.
- Tailwind v4 class names; use existing `dark:` variants.
- Run `npm test` green before each commit.

---

### Task 12: New Types

**Files:**
- Modify: `frontend/src/types/index.ts`

**Interfaces:**
- Produces:
  - `TaskInstance` extended with `priority`, `expected_completion_date`, `expected_completion_time`, `is_adhoc`, `comment_count`, `has_attachment`
  - `TaskComment` interface
  - `TaskAttachmentMeta` interface

- [ ] **Step 1: Write failing type test**

Create `frontend/src/types/types.test.ts`:

```typescript
import { describe, it, expectTypeOf } from 'vitest'
import type { TaskInstance, TaskComment, TaskAttachmentMeta } from './index'

describe('extended types', () => {
  it('TaskInstance has priority', () => {
    expectTypeOf<TaskInstance['priority']>().toEqualTypeOf<'low' | 'medium' | 'high'>()
  })
  it('TaskInstance has is_adhoc', () => {
    expectTypeOf<TaskInstance['is_adhoc']>().toEqualTypeOf<boolean>()
  })
  it('TaskComment exists', () => {
    expectTypeOf<TaskComment['author']>().toEqualTypeOf<string>()
  })
  it('TaskAttachmentMeta exists', () => {
    expectTypeOf<TaskAttachmentMeta['size_bytes']>().toEqualTypeOf<number>()
  })
})
```

- [ ] **Step 2: Run — expect fail**

```bash
cd frontend && npm test -- --run src/types/types.test.ts 2>&1 | tail -10
```

Expected: `Property 'priority' does not exist` or similar.

- [ ] **Step 3: Update `src/types/index.ts`**

Replace `TaskInstance` interface and add new types:

```typescript
export interface TaskInstance {
  id: string
  workflow_instance_id: string
  state_id: string
  task_key: string
  task_label: string
  task_type: string
  status: 'new' | 'in_progress' | 'done' | 'rejected' | 'cancelled'
  outcome?: string
  created_by: string
  created_at: string
  completed_at?: string
  // Phase 7
  priority: 'low' | 'medium' | 'high'
  expected_completion_date: string | null
  expected_completion_time: string | null
  is_adhoc: boolean
  comment_count: number
  has_attachment: boolean
}

export interface TaskComment {
  id: string
  task_instance_id: string
  author: string
  content: string
  created_at: string
}

export interface TaskAttachmentMeta {
  id: string
  filename: string
  content_type: string
  size_bytes: number
  uploaded_at: string
  updated_at: string
}
```

- [ ] **Step 4: Run — expect pass**

```bash
cd frontend && npm test -- --run src/types/types.test.ts
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/types/types.test.ts
git commit -m "feat: extend TaskInstance type, add TaskComment and TaskAttachmentMeta"
```

---

### Task 13: API Modules

**Files:**
- Create: `frontend/src/api/tasks.ts`
- Create: `frontend/src/api/comments.ts`
- Create: `frontend/src/api/attachment.ts`
- Modify: `frontend/src/api/index.ts`

**Interfaces:**
- Consumes: `apiFetch` from `src/api/client.ts`, `TaskInstance`, `TaskComment`, `TaskAttachmentMeta` types
- Produces:
  - `tasksApi.update(id, patch)` — PATCH `/task-instances/:id`
  - `tasksApi.createAdhoc(instanceId, body)` — POST `/workflow-instances/:id/tasks`
  - `commentsApi.add(taskId, body)` — POST `/task-instances/:id/comments`
  - `commentsApi.list(taskId)` — GET `/task-instances/:id/comments`
  - `attachmentApi.upload(taskId, body)` — POST `/task-instances/:id/attachment`
  - `attachmentApi.download(taskId)` — GET `/task-instances/:id/attachment` → `Blob`
  - `attachmentApi.delete(taskId)` — DELETE `/task-instances/:id/attachment`

- [ ] **Step 1: Create `src/api/tasks.ts`**

```typescript
import { apiFetch } from './client'
import type { TaskInstance } from '../types'

const BASE = '/api/v1'

export interface TaskPatch {
  priority?: 'low' | 'medium' | 'high'
  expected_completion_date?: string | null
  expected_completion_time?: string | null
}

export interface AdhocTaskBody {
  state_id: string
  label: string
  task_type: string
  created_by: string
  priority?: 'low' | 'medium' | 'high'
  expected_completion_date?: string
  expected_completion_time?: string
}

export const tasksExtApi = {
  update: (id: string, patch: TaskPatch) =>
    apiFetch<void>(`${BASE}/task-instances/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(patch),
    }),

  createAdhoc: (instanceId: string, body: AdhocTaskBody) =>
    apiFetch<TaskInstance>(`${BASE}/workflow-instances/${instanceId}/tasks`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
}
```

- [ ] **Step 2: Create `src/api/comments.ts`**

```typescript
import { apiFetch } from './client'
import type { TaskComment } from '../types'

const BASE = '/api/v1'

export interface AddCommentBody {
  author: string
  content: string
}

export const commentsApi = {
  add: (taskId: string, body: AddCommentBody) =>
    apiFetch<TaskComment>(`${BASE}/task-instances/${taskId}/comments`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  list: (taskId: string) =>
    apiFetch<TaskComment[]>(`${BASE}/task-instances/${taskId}/comments`),
}
```

- [ ] **Step 3: Create `src/api/attachment.ts`**

```typescript
import { apiFetch } from './client'
import type { TaskAttachmentMeta } from '../types'

const BASE = '/api/v1'

export interface UploadAttachmentBody {
  filename: string
  content_type: string
  data: string // base64
}

export const attachmentApi = {
  upload: (taskId: string, body: UploadAttachmentBody) =>
    apiFetch<TaskAttachmentMeta>(`${BASE}/task-instances/${taskId}/attachment`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  download: async (taskId: string): Promise<Blob> => {
    const res = await fetch(`${BASE}/task-instances/${taskId}/attachment`)
    if (!res.ok) throw new Error(`download failed: ${res.status}`)
    return res.blob()
  },

  delete: (taskId: string) =>
    apiFetch<void>(`${BASE}/task-instances/${taskId}/attachment`, {
      method: 'DELETE',
    }),
}
```

- [ ] **Step 4: Export from index**

In `frontend/src/api/index.ts`, add:

```typescript
export * from './tasks'
export * from './comments'
export * from './attachment'
```

- [ ] **Step 5: Build check**

```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: no TypeScript errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/tasks.ts frontend/src/api/comments.ts frontend/src/api/attachment.ts frontend/src/api/index.ts
git commit -m "feat: frontend API modules for task updates, comments, attachments"
```

---

### Task 14: PriorityBadge Component

**Files:**
- Create: `frontend/src/components/tasks/PriorityBadge.tsx`
- Create: `frontend/src/components/tasks/PriorityBadge.test.tsx`

**Interfaces:**
- Produces: `<PriorityBadge priority="high" />` renders colored chip

- [ ] **Step 1: Write test**

```typescript
// frontend/src/components/tasks/PriorityBadge.test.tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { PriorityBadge } from './PriorityBadge'

describe('PriorityBadge', () => {
  it('renders high priority in red', () => {
    render(<PriorityBadge priority="high" />)
    const el = screen.getByText('High')
    expect(el).toBeTruthy()
    expect(el.className).toContain('red')
  })

  it('renders medium priority in yellow', () => {
    render(<PriorityBadge priority="medium" />)
    expect(screen.getByText('Medium').className).toContain('yellow')
  })

  it('renders low priority in gray', () => {
    render(<PriorityBadge priority="low" />)
    expect(screen.getByText('Low').className).toContain('gray')
  })
})
```

- [ ] **Step 2: Run — expect fail**

```bash
cd frontend && npm test -- --run src/components/tasks/PriorityBadge.test.tsx 2>&1 | tail -10
```

Expected: module not found.

- [ ] **Step 3: Create component**

```typescript
// frontend/src/components/tasks/PriorityBadge.tsx
interface Props {
  priority: 'low' | 'medium' | 'high'
}

const config = {
  low:    { label: 'Low',    cls: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300' },
  medium: { label: 'Medium', cls: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300' },
  high:   { label: 'High',   cls: 'bg-red-100 text-red-600 dark:bg-red-900 dark:text-red-300' },
}

export function PriorityBadge({ priority }: Props) {
  const { label, cls } = config[priority] ?? config.medium
  return (
    <span className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-medium ${cls}`}>
      {label}
    </span>
  )
}
```

- [ ] **Step 4: Run — expect pass**

```bash
cd frontend && npm test -- --run src/components/tasks/PriorityBadge.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/tasks/PriorityBadge.tsx frontend/src/components/tasks/PriorityBadge.test.tsx
git commit -m "feat: PriorityBadge component (low/medium/high)"
```

---

### Task 15: CommentsPanel Component

**Files:**
- Create: `frontend/src/components/tasks/CommentsPanel.tsx`
- Create: `frontend/src/components/tasks/CommentsPanel.test.tsx`

**Interfaces:**
- Consumes: `commentsApi.list`, `commentsApi.add` (Task 13), `TaskComment` type
- Produces: `<CommentsPanel taskId={string} />` — renders comment list + add form

- [ ] **Step 1: Write test**

```typescript
// frontend/src/components/tasks/CommentsPanel.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { CommentsPanel } from './CommentsPanel'

vi.mock('../../api/comments', () => ({
  commentsApi: {
    list: vi.fn().mockResolvedValue([
      { id: '1', task_instance_id: 'tid', author: 'alice', content: 'first comment', created_at: '2026-07-03T00:00:00Z' },
    ]),
    add: vi.fn().mockResolvedValue(
      { id: '2', task_instance_id: 'tid', author: 'bob', content: 'new comment', created_at: '2026-07-03T01:00:00Z' }
    ),
  },
}))

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('CommentsPanel', () => {
  it('shows existing comment', async () => {
    wrap(<CommentsPanel taskId="tid" />)
    await waitFor(() => screen.getByText('first comment'))
    expect(screen.getByText('alice')).toBeTruthy()
  })

  it('submit button disabled when fields empty', async () => {
    wrap(<CommentsPanel taskId="tid" />)
    await waitFor(() => screen.getByText('first comment'))
    const btn = screen.getByRole('button', { name: /add comment/i })
    expect(btn).toBeDisabled()
  })
})
```

- [ ] **Step 2: Run — expect fail**

```bash
cd frontend && npm test -- --run src/components/tasks/CommentsPanel.test.tsx 2>&1 | tail -10
```

Expected: module not found.

- [ ] **Step 3: Create component**

```typescript
// frontend/src/components/tasks/CommentsPanel.tsx
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { commentsApi } from '../../api/comments'
import type { TaskComment } from '../../types'

interface Props {
  taskId: string
}

export function CommentsPanel({ taskId }: Props) {
  const qc = useQueryClient()
  const [author, setAuthor] = useState('')
  const [content, setContent] = useState('')

  const { data: comments = [] } = useQuery<TaskComment[]>({
    queryKey: ['task-comments', taskId],
    queryFn: () => commentsApi.list(taskId),
    staleTime: 30_000,
  })

  const addMutation = useMutation({
    mutationFn: () => commentsApi.add(taskId, { author, content }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task-comments', taskId] })
      setAuthor('')
      setContent('')
    },
  })

  const canSubmit = author.trim().length > 0 && content.trim().length > 0

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Comments</h3>
      <div className="space-y-3">
        {comments.map((c) => (
          <div key={c.id} className="rounded border border-gray-200 dark:border-gray-700 p-3">
            <div className="flex items-center justify-between mb-1">
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{c.author}</span>
              <span className="text-xs text-gray-400">{new Date(c.created_at).toLocaleString()}</span>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{c.content}</p>
          </div>
        ))}
        {comments.length === 0 && (
          <p className="text-xs text-gray-400">No comments yet.</p>
        )}
      </div>
      <div className="space-y-2 border-t pt-3 dark:border-gray-700">
        <input
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Your name"
          className="w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
        />
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Add a comment…"
          rows={3}
          maxLength={2048}
          className="w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
        />
        <button
          disabled={!canSubmit || addMutation.isPending}
          onClick={() => addMutation.mutate()}
          className="rounded bg-blue-600 px-3 py-1 text-xs text-white disabled:opacity-50"
        >
          {addMutation.isPending ? 'Adding…' : 'Add Comment'}
        </button>
        {addMutation.isError && (
          <p className="text-xs text-red-500">Failed to add comment.</p>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run — expect pass**

```bash
cd frontend && npm test -- --run src/components/tasks/CommentsPanel.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/tasks/CommentsPanel.tsx frontend/src/components/tasks/CommentsPanel.test.tsx
git commit -m "feat: CommentsPanel component with list + add"
```

---

### Task 16: AttachmentSection Component

**Files:**
- Create: `frontend/src/components/tasks/AttachmentSection.tsx`
- Create: `frontend/src/components/tasks/AttachmentSection.test.tsx`

**Interfaces:**
- Consumes: `attachmentApi.upload/download/delete` (Task 13), `TaskAttachmentMeta`
- Produces: `<AttachmentSection taskId={string} hasMeta={boolean} />` — upload/download/delete file

- [ ] **Step 1: Write test**

```typescript
// frontend/src/components/tasks/AttachmentSection.test.tsx
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { AttachmentSection } from './AttachmentSection'

vi.mock('../../api/attachment', () => ({
  attachmentApi: {
    upload: vi.fn(),
    download: vi.fn(),
    delete: vi.fn(),
  },
}))

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AttachmentSection', () => {
  it('shows upload button when no attachment', () => {
    wrap(<AttachmentSection taskId="tid" hasMeta={false} />)
    expect(screen.getByRole('button', { name: /upload/i })).toBeTruthy()
  })

  it('shows download button when attachment exists', () => {
    wrap(<AttachmentSection taskId="tid" hasMeta={true} filename="doc.pdf" />)
    expect(screen.getByRole('button', { name: /download/i })).toBeTruthy()
  })
})
```

- [ ] **Step 2: Run — expect fail**

```bash
cd frontend && npm test -- --run src/components/tasks/AttachmentSection.test.tsx 2>&1 | tail -10
```

Expected: module not found.

- [ ] **Step 3: Create component**

```typescript
// frontend/src/components/tasks/AttachmentSection.tsx
import { useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { attachmentApi } from '../../api/attachment'

const MAX_BYTES = 2 * 1024 * 1024

interface Props {
  taskId: string
  hasMeta: boolean
  filename?: string
}

export function AttachmentSection({ taskId, hasMeta, filename }: Props) {
  const qc = useQueryClient()
  const inputRef = useRef<HTMLInputElement>(null)

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      if (file.size > MAX_BYTES) throw new Error('FILE_TOO_LARGE')
      const buf = await file.arrayBuffer()
      const b64 = btoa(String.fromCharCode(...new Uint8Array(buf)))
      return attachmentApi.upload(taskId, {
        filename: file.name,
        content_type: file.type || 'application/octet-stream',
        data: b64,
      })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['task', taskId] }),
  })

  const deleteMutation = useMutation({
    mutationFn: () => attachmentApi.delete(taskId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['task', taskId] }),
  })

  const handleDownload = async () => {
    const blob = await attachmentApi.download(taskId)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename ?? 'attachment'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-2">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Attachment</h3>
      {hasMeta && filename ? (
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600 dark:text-gray-400">{filename}</span>
          <button
            onClick={handleDownload}
            className="rounded bg-blue-600 px-2 py-0.5 text-xs text-white"
          >
            Download
          </button>
          <button
            onClick={() => deleteMutation.mutate()}
            disabled={deleteMutation.isPending}
            className="rounded border border-red-400 px-2 py-0.5 text-xs text-red-500 disabled:opacity-50"
          >
            {deleteMutation.isPending ? 'Removing…' : 'Remove'}
          </button>
        </div>
      ) : null}
      <div>
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) uploadMutation.mutate(file)
            if (inputRef.current) inputRef.current.value = ''
          }}
        />
        <button
          onClick={() => inputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 disabled:opacity-50"
        >
          {uploadMutation.isPending ? 'Uploading…' : hasMeta ? 'Replace File' : 'Upload File'}
        </button>
        {uploadMutation.isError && (
          <p className="text-xs text-red-500 mt-1">
            {(uploadMutation.error as Error).message === 'FILE_TOO_LARGE'
              ? 'File must be under 2MB.'
              : 'Upload failed.'}
          </p>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run — expect pass**

```bash
cd frontend && npm test -- --run src/components/tasks/AttachmentSection.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/tasks/AttachmentSection.tsx frontend/src/components/tasks/AttachmentSection.test.tsx
git commit -m "feat: AttachmentSection component with upload/download/delete"
```

---

### Task 17: AdhocTaskModal Component

**Files:**
- Create: `frontend/src/components/tasks/AdhocTaskModal.tsx`
- Create: `frontend/src/components/tasks/AdhocTaskModal.test.tsx`

**Interfaces:**
- Consumes: `tasksExtApi.createAdhoc`, `StateDef[]` from `WorkflowDefinition`
- Produces: `<AdhocTaskModal instanceId={string} states={StateDef[]} onClose={() => void} />` — modal form

- [ ] **Step 1: Write test**

```typescript
// frontend/src/components/tasks/AdhocTaskModal.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { AdhocTaskModal } from './AdhocTaskModal'

vi.mock('../../api/tasks', () => ({
  tasksExtApi: { createAdhoc: vi.fn().mockResolvedValue({}) },
}))

const states = [{ id: 'review', label: 'Review', tasks: [] }]

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AdhocTaskModal', () => {
  it('renders state dropdown', () => {
    wrap(<AdhocTaskModal instanceId="iid" states={states} onClose={() => {}} />)
    expect(screen.getByRole('combobox')).toBeTruthy()
    expect(screen.getByText('Review')).toBeTruthy()
  })

  it('submit button disabled when label empty', () => {
    wrap(<AdhocTaskModal instanceId="iid" states={states} onClose={() => {}} />)
    const btn = screen.getByRole('button', { name: /add task/i })
    expect(btn).toBeDisabled()
  })
})
```

- [ ] **Step 2: Run — expect fail**

```bash
cd frontend && npm test -- --run src/components/tasks/AdhocTaskModal.test.tsx 2>&1 | tail -10
```

Expected: module not found.

- [ ] **Step 3: Create component**

```typescript
// frontend/src/components/tasks/AdhocTaskModal.tsx
import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksExtApi } from '../../api/tasks'
import type { StateDef } from '../../types'

interface Props {
  instanceId: string
  states: StateDef[]
  onClose: () => void
}

const TASK_TYPES = ['manual', 'approval', 'form', 'checklist']
const PRIORITIES = ['low', 'medium', 'high'] as const

export function AdhocTaskModal({ instanceId, states, onClose }: Props) {
  const qc = useQueryClient()
  const [stateId, setStateId] = useState(states[0]?.id ?? '')
  const [label, setLabel] = useState('')
  const [taskType, setTaskType] = useState('manual')
  const [priority, setPriority] = useState<'low' | 'medium' | 'high'>('medium')
  const [createdBy, setCreatedBy] = useState('')
  const [expDate, setExpDate] = useState('')
  const [expTime, setExpTime] = useState('')

  const mutation = useMutation({
    mutationFn: () =>
      tasksExtApi.createAdhoc(instanceId, {
        state_id: stateId,
        label,
        task_type: taskType,
        created_by: createdBy,
        priority,
        expected_completion_date: expDate || undefined,
        expected_completion_time: expTime || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['instance-tasks', instanceId] })
      onClose()
    },
  })

  const canSubmit = label.trim().length > 0 && stateId.length > 0

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6 space-y-4">
        <h2 className="text-base font-semibold">Add Ad-hoc Task</h2>

        <div className="space-y-3">
          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            State
            <select
              value={stateId}
              onChange={(e) => setStateId(e.target.value)}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {states.map((s) => (
                <option key={s.id} value={s.id}>{s.label}</option>
              ))}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Label *
            <input
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="Task label"
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            />
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Task Type
            <select
              value={taskType}
              onChange={(e) => setTaskType(e.target.value)}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {TASK_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Priority
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as 'low' | 'medium' | 'high')}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {PRIORITIES.map((p) => <option key={p} value={p}>{p}</option>)}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Created By
            <input
              value={createdBy}
              onChange={(e) => setCreatedBy(e.target.value)}
              placeholder="Your name"
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            />
          </label>

          <div className="grid grid-cols-2 gap-2">
            <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
              Due Date
              <input
                type="date"
                value={expDate}
                onChange={(e) => setExpDate(e.target.value)}
                className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
              />
            </label>
            <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
              Due Time
              <input
                type="time"
                value={expTime}
                onChange={(e) => setExpTime(e.target.value)}
                className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
              />
            </label>
          </div>
        </div>

        {mutation.isError && (
          <p className="text-xs text-red-500">Failed to create task.</p>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="rounded border border-gray-300 px-3 py-1 text-xs dark:border-gray-600"
          >
            Cancel
          </button>
          <button
            disabled={!canSubmit || mutation.isPending}
            onClick={() => mutation.mutate()}
            className="rounded bg-blue-600 px-3 py-1 text-xs text-white disabled:opacity-50"
          >
            {mutation.isPending ? 'Adding…' : 'Add Task'}
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run — expect pass**

```bash
cd frontend && npm test -- --run src/components/tasks/AdhocTaskModal.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/tasks/AdhocTaskModal.tsx frontend/src/components/tasks/AdhocTaskModal.test.tsx
git commit -m "feat: AdhocTaskModal component for ad-hoc task creation"
```

---

### Task 18: Wire Components into Instance Detail Page

**Files:**
- Modify: existing task card / instance detail page component (locate with `grep -r "TaskInstance\|task_label" frontend/src/pages/ --include="*.tsx" -l`)
- Use `PriorityBadge`, `CommentsPanel`, `AttachmentSection` in the task detail section
- Show "Add Task" button (visible only when `instance.status === 'active'`) that opens `AdhocTaskModal`

**Interfaces:**
- Consumes: all components from Tasks 14-17

- [ ] **Step 1: Locate the task card rendering**

```bash
grep -r "task_label\|TaskInstance" frontend/src/pages/ --include="*.tsx" -l
```

Note the file path for the next steps.

- [ ] **Step 2: Import and add PriorityBadge to task card**

In the task card JSX (where `task.task_label` is rendered), add alongside the label:

```tsx
import { PriorityBadge } from '../components/tasks/PriorityBadge'

// In JSX:
<div className="flex items-center gap-2">
  <span className="font-medium text-sm">{task.task_label}</span>
  <PriorityBadge priority={task.priority ?? 'medium'} />
  {task.is_adhoc && (
    <span className="rounded bg-purple-100 text-purple-600 dark:bg-purple-900 dark:text-purple-300 px-2 py-0.5 text-xs font-medium">
      Ad-hoc
    </span>
  )}
</div>
{task.expected_completion_date && (
  <p className="text-xs text-gray-400 mt-0.5">
    Due: {task.expected_completion_date}
    {task.expected_completion_time ? ` at ${task.expected_completion_time}` : ''}
  </p>
)}
```

- [ ] **Step 3: Add CommentsPanel and AttachmentSection to task detail (expanded or on click)**

Where task detail expands (or in a separate task detail route), add:

```tsx
import { CommentsPanel } from '../components/tasks/CommentsPanel'
import { AttachmentSection } from '../components/tasks/AttachmentSection'

// Below existing task metadata:
<div className="mt-4 space-y-6">
  <CommentsPanel taskId={task.id} />
  <AttachmentSection
    taskId={task.id}
    hasMeta={task.has_attachment}
  />
</div>
```

- [ ] **Step 4: Add "Add Task" button and AdhocTaskModal**

In the instance detail page, where tasks are listed and `instance.status === 'active'`:

```tsx
import { useState } from 'react'
import { AdhocTaskModal } from '../components/tasks/AdhocTaskModal'

const [showAdhoc, setShowAdhoc] = useState(false)

// In JSX (near task list header):
{instance.status === 'active' && (
  <button
    onClick={() => setShowAdhoc(true)}
    className="rounded bg-blue-600 px-3 py-1 text-xs text-white"
  >
    + Add Task
  </button>
)}
{showAdhoc && (
  <AdhocTaskModal
    instanceId={instance.id}
    states={instance.definition?.states ?? []}
    onClose={() => setShowAdhoc(false)}
  />
)}
```

- [ ] **Step 5: Run full frontend test suite**

```bash
cd frontend && npm test
```

Expected: all tests pass.

- [ ] **Step 6: Build**

```bash
cd frontend && npm run build
```

Expected: no TypeScript errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/
git commit -m "feat: wire PriorityBadge, CommentsPanel, AttachmentSection, AdhocTaskModal into instance detail"
```
