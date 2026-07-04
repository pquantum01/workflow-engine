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
