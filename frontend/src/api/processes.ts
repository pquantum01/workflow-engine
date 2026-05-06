import { request } from './client';
import type { Process, CreateProcessInput, UpdateProcessInput } from '../types';

export function listProcesses(offset = 0, limit = 20): Promise<Process[]> {
  return request<Process[]>(`/processes?offset=${offset}&limit=${limit}`);
}

export function getProcess(id: string): Promise<Process> {
  return request<Process>(`/processes/${id}`);
}

export function createProcess(input: CreateProcessInput): Promise<Process> {
  return request<Process>('/processes', { method: 'POST', body: JSON.stringify(input) });
}

export function updateProcess(id: string, input: UpdateProcessInput): Promise<Process> {
  return request<Process>(`/processes/${id}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteProcess(id: string): Promise<void> {
  return request<void>(`/processes/${id}`, { method: 'DELETE' });
}
