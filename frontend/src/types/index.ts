export interface WorkflowDef {
  id: string
  name: string
  version: number
  description: string
  definition: WorkflowDefinition
  created_at: string
  deleted_at?: string
}

export interface WorkflowDefinition {
  name: string
  version: number
  description?: string
  start_state: string
  end_state: string
  states: StateDef[]
  transitions: TransitionDef[]
}

export interface StateDef {
  id: string
  label: string
  tasks: TaskDef[]
}

export interface TaskDef {
  key: string
  label: string
  type: string
}

export interface TransitionDef {
  from_state: string
  to_state: string
  trigger?: { task_key: string }
}

export interface WorkflowInstance {
  id: string
  workflow_def_id: string
  current_state_id: string
  status: 'active' | 'completed' | 'cancelled'
  created_by: string
  context: Record<string, unknown>
  started_at: string
  completed_at?: string
  tasks?: TaskInstance[]
  history?: StateTransition[]
}

export interface TaskInstance {
  id: string
  workflow_instance_id: string
  state_id: string
  task_key: string
  task_label: string
  task_type: string
  status: 'new' | 'in_progress' | 'done' | 'rejected' | 'cancelled'
  outcome?: string
  created_by: string
  created_at: string
  completed_at?: string
  priority: 'low' | 'medium' | 'high'
  expected_completion_date: string | null
  expected_completion_time: string | null
  is_adhoc: boolean
  comment_count: number
  has_attachment: boolean
}

export interface TaskComment {
  id: string
  task_instance_id: string
  author: string
  content: string
  created_at: string
}

export interface TaskAttachmentMeta {
  id: string
  filename: string
  content_type: string
  size_bytes: number
  uploaded_at: string
  updated_at: string
}

export interface StateTransition {
  id: string
  from_state: string
  to_state: string
  transitioned_at: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

export interface ApiError {
  code: string
  message: string
  details?: unknown
}
