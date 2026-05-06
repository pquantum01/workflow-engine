import { request } from './client';
import type { WorkflowInstance, StateTransition, TriggerTransitionInput } from '../types';

export function listWorkflowInstances(processInstanceId: string): Promise<WorkflowInstance[]> {
  return request<WorkflowInstance[]>(`/process-instances/${processInstanceId}/workflows`);
}

export function getWorkflowInstance(id: string): Promise<WorkflowInstance> {
  return request<WorkflowInstance>(`/workflow-instances/${id}`);
}

export function triggerTransition(id: string, input: TriggerTransitionInput): Promise<StateTransition> {
  return request<StateTransition>(`/workflow-instances/${id}/transition`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
}
