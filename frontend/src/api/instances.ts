import { apiFetch } from './client'
import type { WorkflowInstance, TaskInstance, StateTransition, PaginatedResponse } from '../types'

const BASE = '/api/v1'

export interface InstanceFilter {
  def_id?: string
  status?: string
  limit?: number
  offset?: number
}

export const instancesApi = {
  list: (filter: InstanceFilter = {}) => {
    const params = new URLSearchParams()
    if (filter.def_id) params.set('def_id', filter.def_id)
    if (filter.status) params.set('status', filter.status)
    if (filter.limit) params.set('limit', String(filter.limit))
    if (filter.offset) params.set('offset', String(filter.offset))
    const qs = params.toString()
    return apiFetch<PaginatedResponse<WorkflowInstance>>(
      `${BASE}/workflow-instances${qs ? '?' + qs : ''}`
    )
  },
  getById: (id: string) =>
    apiFetch<WorkflowInstance>(`${BASE}/workflow-instances/${id}`),
  start: (body: { workflow_def_id: string; created_by: string; initial_context?: Record<string, unknown> }) =>
    apiFetch<WorkflowInstance>(`${BASE}/workflow-instances`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  cancel: (id: string) =>
    apiFetch<WorkflowInstance>(`${BASE}/workflow-instances/${id}/cancel`, { method: 'POST' }),
  softDelete: (id: string) =>
    apiFetch<void>(`${BASE}/workflow-instances/${id}`, { method: 'DELETE' }),
  listTasks: (id: string, status?: string) => {
    const qs = status ? `?status=${status}` : ''
    return apiFetch<TaskInstance[]>(`${BASE}/workflow-instances/${id}/tasks${qs}`)
  },
  getHistory: (id: string) =>
    apiFetch<StateTransition[]>(`${BASE}/workflow-instances/${id}/history`),
}

export const tasksApi = {
  getById: (id: string) => apiFetch<TaskInstance>(`${BASE}/task-instances/${id}`),
  start: (id: string) =>
    apiFetch<TaskInstance>(`${BASE}/task-instances/${id}/start`, { method: 'POST' }),
  complete: (id: string, body: { status: 'done' | 'rejected'; next_state?: string; completed_by?: string }) =>
    apiFetch<WorkflowInstance>(`${BASE}/task-instances/${id}/complete`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
}
