import { request } from './client';
import type { StateTransition } from '../types';

export function listStateTransitions(workflowInstanceId: string): Promise<StateTransition[]> {
  return request<StateTransition[]>(`/workflow-instances/${workflowInstanceId}/transitions`);
}
