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
