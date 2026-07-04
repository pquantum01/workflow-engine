import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useInstance, useCompleteTask } from '../hooks/useInstances'
import { useWorkflowDef } from '../hooks/useWorkflowDefs'
import { tasksApi } from '../api/instances'
import { rendererFor } from '../components/tasks/registry'
import { PriorityBadge } from '../components/tasks/PriorityBadge'
import { CommentsPanel } from '../components/tasks/CommentsPanel'
import { AttachmentSection } from '../components/tasks/AttachmentSection'
import { Badge } from '../components/ui/Badge'
import { Skeleton } from '../components/ui/Skeleton'
import { getCompletedBy } from '../lib/session'
import type { TransitionDef } from '../types'

export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data: task, isLoading: taskLoading } = useQuery({
    queryKey: ['tasks', id],
    queryFn: () => tasksApi.getById(id!),
    enabled: !!id,
    refetchInterval: (q) => {
      const t = q.state.data
      if (!t) return 5_000
      return t.status === 'new' || t.status === 'in_progress' ? 5_000 : false
    },
  })
  const { data: inst } = useInstance(task?.workflow_instance_id ?? '')
  const { data: def } = useWorkflowDef(inst?.workflow_def_id ?? '')
  const completeMutation = useCompleteTask()

  if (taskLoading || !task) {
    return (
      <div className="p-6 space-y-4 max-w-lg">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  const outgoing: TransitionDef[] = def?.definition?.transitions?.filter(
    (t) => t.from_state === task.state_id
  ) ?? []

  const Renderer = rendererFor(task.task_type)

  return (
    <div className="p-6 max-w-lg">
      <div className="mb-4">
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-semibold">{task.task_label}</h1>
          {task.priority && <PriorityBadge priority={task.priority} />}
          {task.is_adhoc && (
            <span className="rounded bg-purple-100 text-purple-600 dark:bg-purple-900 dark:text-purple-300 px-2 py-0.5 text-xs font-medium">
              Ad-hoc
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 mt-1 text-sm text-zinc-500">
          <code className="font-mono text-xs">{task.task_key}</code>
          <span>·</span>
          <span>{task.task_type}</span>
          <Badge variant={task.status}>{task.status}</Badge>
        </div>
        {task.expected_completion_date && (
          <p className="text-xs text-zinc-400 mt-0.5">
            Due: {task.expected_completion_date}
            {task.expected_completion_time ? ` at ${task.expected_completion_time}` : ''}
          </p>
        )}
        {inst && (
          <p className="text-xs text-zinc-400 mt-1">
            Instance: <code className="font-mono">{inst.id.slice(0, 8)}…</code>
            {' '}· State: <code className="font-mono">{task.state_id}</code>
          </p>
        )}
      </div>

      {inst && (
        <Renderer
          task={task}
          instance={inst}
          outgoingTransitions={outgoing}
          onComplete={(status, nextState) => {
            completeMutation.mutate({
              taskId: task.id,
              status,
              nextState,
              completedBy: getCompletedBy() || 'user',
            })
          }}
          isLoading={completeMutation.isPending}
        />
      )}

      {completeMutation.isSuccess && (
        <div className="mt-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg text-sm text-green-700 dark:text-green-400">
          Task completed successfully.
        </div>
      )}
      {completeMutation.isError && (
        <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700">
          {(completeMutation.error as Error)?.message}
        </div>
      )}

      <div className="mt-6 space-y-6 border-t pt-6 dark:border-zinc-700">
        <CommentsPanel taskId={task.id} />
        <AttachmentSection
          taskId={task.id}
          hasMeta={task.has_attachment ?? false}
        />
      </div>
    </div>
  )
}
