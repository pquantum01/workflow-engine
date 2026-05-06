import { request } from './client';
import type { Task, CreateTaskInput, UpdateTaskInput } from '../types';

export function listTasks(workflowId: string): Promise<Task[]> {
  return request<Task[]>(`/workflows/${workflowId}/tasks`);
}

export function getTask(id: string): Promise<Task> {
  return request<Task>(`/tasks/${id}`);
}

export function createTask(workflowId: string, input: CreateTaskInput): Promise<Task> {
  return request<Task>(`/workflows/${workflowId}/tasks`, { method: 'POST', body: JSON.stringify(input) });
}

export function updateTask(id: string, input: UpdateTaskInput): Promise<Task> {
  return request<Task>(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteTask(id: string): Promise<void> {
  return request<void>(`/tasks/${id}`, { method: 'DELETE' });
}
