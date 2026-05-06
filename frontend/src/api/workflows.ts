import { request } from './client';
import type { Workflow, CreateWorkflowInput, UpdateWorkflowInput } from '../types';

export function listWorkflows(processId: string): Promise<Workflow[]> {
  return request<Workflow[]>(`/processes/${processId}/workflows`);
}

export function getWorkflow(id: string): Promise<Workflow> {
  return request<Workflow>(`/workflows/${id}`);
}

export function createWorkflow(processId: string, input: CreateWorkflowInput): Promise<Workflow> {
  return request<Workflow>(`/processes/${processId}/workflows`, { method: 'POST', body: JSON.stringify(input) });
}

export function updateWorkflow(id: string, input: UpdateWorkflowInput): Promise<Workflow> {
  return request<Workflow>(`/workflows/${id}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteWorkflow(id: string): Promise<void> {
  return request<void>(`/workflows/${id}`, { method: 'DELETE' });
}
