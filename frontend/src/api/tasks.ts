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
