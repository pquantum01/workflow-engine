import { request } from './client';
import type { TaskInstance, UpdateTaskStatusInput, AssignTaskInput } from '../types';

export function listTaskInstances(workflowInstanceId: string): Promise<TaskInstance[]> {
  return request<TaskInstance[]>(`/workflow-instances/${workflowInstanceId}/tasks`);
}

export function getTaskInstance(id: string): Promise<TaskInstance> {
  return request<TaskInstance>(`/task-instances/${id}`);
}

export function updateTaskStatus(id: string, input: UpdateTaskStatusInput): Promise<TaskInstance> {
  return request<TaskInstance>(`/task-instances/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export function assignTask(id: string, input: AssignTaskInput): Promise<TaskInstance> {
  return request<TaskInstance>(`/task-instances/${id}/assign`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}
