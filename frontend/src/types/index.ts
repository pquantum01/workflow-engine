export interface Process {
  id: string;
  name: string;
  description: string;
  version: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Workflow {
  id: string;
  process_id: string;
  name: string;
  description: string;
  state_machine: StateMachine;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  workflow_id: string;
  name: string;
  description: string;
  task_type: string;
  created_at: string;
  updated_at: string;
}

export interface StateMachine {
  states: State[];
  transitions: Transition[];
  initial_state: string;
  terminal_states: string[];
}

export interface State {
  name: string;
  description: string;
  metadata?: Record<string, string>;
}

export interface Transition {
  name: string;
  from_state: string;
  to_state: string;
  completion_rule?: CompletionRule;
}

export interface CompletionRule {
  type: string;
  task_refs: string[];
  n?: number;
}

export interface CreateProcessInput {
  name: string;
  description: string;
}

export interface UpdateProcessInput {
  name?: string;
  description?: string;
  status?: string;
}

export interface CreateWorkflowInput {
  name: string;
  description: string;
  state_machine: StateMachine;
}

export interface UpdateWorkflowInput {
  name?: string;
  description?: string;
  state_machine?: StateMachine;
}

export interface CreateTaskInput {
  name: string;
  description: string;
  task_type: string;
}

export interface UpdateTaskInput {
  name?: string;
  description?: string;
  task_type?: string;
}

// --- Instance types ---

export interface ProcessInstance {
  id: string;
  process_id: string;
  parent_id: string | null;
  status: string;
  started_at: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface WorkflowInstance {
  id: string;
  workflow_id: string;
  process_instance_id: string;
  current_state: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface TaskInstance {
  id: string;
  task_id: string;
  workflow_instance_id: string;
  status: string;
  assigned_to: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface StateTransition {
  id: string;
  workflow_instance_id: string;
  from_state: string;
  to_state: string;
  transition_name: string;
  triggered_by: string;
  trigger_type: string;
  created_at: string;
}

export interface StartProcessInput {
  process_id: string;
  parent_id?: string;
}

export interface UpdateTaskStatusInput {
  status: string;
  triggered_by?: string;
  trigger_type?: string;
}

export interface AssignTaskInput {
  assigned_to: string;
}

export interface TriggerTransitionInput {
  transition_name: string;
  triggered_by: string;
  trigger_type: string;
}

export interface ApiError {
  error: string;
}
