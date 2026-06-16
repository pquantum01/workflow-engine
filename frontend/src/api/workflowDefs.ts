import { apiFetch } from './client'
import type { WorkflowDef } from '../types'

const BASE = '/api/v1/workflow-defs'

export interface WorkflowDefListResponse {
  items: WorkflowDef[]
  total: number
}

export const workflowDefsApi = {
  list: (includeDeleted = false) =>
    apiFetch<WorkflowDefListResponse>(
      `${BASE}${includeDeleted ? '?include_deleted=true' : ''}`
    ),
  getById: (id: string) => apiFetch<WorkflowDef>(`${BASE}/${id}`),
}

export const adminApi = {
  reloadDefs: () =>
    apiFetch<{ loaded: Array<{ name: string; version: number }>; warnings: unknown[]; errors: unknown[] }>(
      '/api/v1/admin/reload-defs',
      { method: 'POST' }
    ),
}
