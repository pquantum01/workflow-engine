import { request } from './client';
import type { ProcessInstance, StartProcessInput } from '../types';

export function listProcessInstances(offset = 0, limit = 20): Promise<ProcessInstance[]> {
  return request<ProcessInstance[]>(`/process-instances?offset=${offset}&limit=${limit}`);
}

export function getProcessInstance(id: string): Promise<ProcessInstance> {
  return request<ProcessInstance>(`/process-instances/${id}`);
}

export function startProcessInstance(input: StartProcessInput): Promise<ProcessInstance> {
  return request<ProcessInstance>('/process-instances', { method: 'POST', body: JSON.stringify(input) });
}

export function cancelProcessInstance(id: string): Promise<void> {
  return request<void>(`/process-instances/${id}/cancel`, { method: 'POST' });
}

export function listChildren(id: string): Promise<ProcessInstance[]> {
  return request<ProcessInstance[]>(`/process-instances/${id}/children`);
}
