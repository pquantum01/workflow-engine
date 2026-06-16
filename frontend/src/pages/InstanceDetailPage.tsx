import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, AlertTriangle } from 'lucide-react'
import { useInstance, useCancelInstance, useCompleteTask } from '../hooks/useInstances'
import { useWorkflowDef } from '../hooks/useWorkflowDefs'
import { Badge } from '../components/ui/Badge'
import { Skeleton } from '../components/ui/Skeleton'
import { rendererFor } from '../components/tasks/registry'
import { getCompletedBy } from '../lib/session'
import type { TaskInstance, TransitionDef } from '../types'

export function InstanceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: inst, isLoading } = useInstance(id!)
  const { data: def } = useWorkflowDef(inst?.workflow_def_id ?? '')
  const cancelMutation = useCancelInstance()
  const completeMutation = useCompleteTask()
  const [showCancelConfirm, setShowCancelConfirm] = useState(false)

  if (isLoading || !inst) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    )
  }

  const tasks = inst.tasks ?? []
  const history = inst.history ?? []

  const openTasks = tasks.filter((t) => t.status === 'new' || t.status === 'in_progress')
  const doneTasks = tasks.filter((t) => t.status === 'done')
  const cancelledTasks = tasks.filter((t) => t.status === 'cancelled' || t.status === 'rejected')

  const handleComplete = (task: TaskInstance, status: 'done' | 'rejected', nextState?: string) => {
    completeMutation.mutate({
      taskId: task.id,
      status,
      nextState,
      completedBy: getCompletedBy() || 'user',
    })
  }

  const handleCancel = async () => {
    await cancelMutation.mutateAsync(inst.id)
    setShowCancelConfirm(false)
  }

  const hasContext = inst.context && Object.keys(inst.context).length > 0

  return (
    <div className="p-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate('/workflow-instances')} className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500">
            <ArrowLeft size={16} />
          </button>
          <div>
            <div className="flex items-center gap-2">
              <code className="text-sm font-mono text-zinc-600 dark:text-zinc-400">{inst.id.slice(0, 8)}…</code>
              <Badge variant={inst.status}>{inst.status}</Badge>
            </div>
            <div className="text-xs text-zinc-500 mt-0.5 space-x-2">
              <span>
                {def?.name ? `${def.name} v${def.version}` : inst.workflow_def_id.slice(0, 8)}
              </span>
              <span>·</span>
              <span>Started by <strong className="font-medium text-zinc-700 dark:text-zinc-300">{inst.created_by}</strong></span>
              <span>·</span>
              <span>{new Date(inst.started_at).toLocaleString()}</span>
              {inst.completed_at && (
                <>
                  <span>·</span>
                  <span>Completed {new Date(inst.completed_at).toLocaleString()}</span>
                </>
              )}
            </div>
          </div>
        </div>
        {inst.status === 'active' && (
          <button
            onClick={() => setShowCancelConfirm(true)}
            className="px-3 py-1.5 text-sm font-medium rounded-lg border border-red-300 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
          >
            Cancel Instance
          </button>
        )}
      </div>

      {/* Cancel confirm */}
      {showCancelConfirm && (
        <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-3">
          <AlertTriangle size={18} className="text-red-500 mt-0.5 shrink-0" />
          <div className="flex-1">
            <p className="text-sm font-medium text-red-800 dark:text-red-300">Cancel this instance?</p>
            <p className="text-xs text-red-600 dark:text-red-400 mt-0.5">All open tasks will be cancelled. This cannot be undone.</p>
          </div>
          <div className="flex gap-2">
            <button onClick={handleCancel} disabled={cancelMutation.isPending} className="px-3 py-1 text-sm font-medium rounded-lg bg-red-600 text-white hover:bg-red-700 disabled:opacity-50">
              {cancelMutation.isPending ? 'Cancelling…' : 'Confirm'}
            </button>
            <button onClick={() => setShowCancelConfirm(false)} className="px-3 py-1 text-sm rounded-lg border border-zinc-300 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Instance Context */}
      {hasContext && (
        <div className="mb-4 border border-zinc-200 dark:border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">Instance Context</h3>
          <pre className="text-xs font-mono text-zinc-600 dark:text-zinc-400 bg-zinc-50 dark:bg-zinc-800/50 rounded p-3 overflow-auto">
            {JSON.stringify(inst.context, null, 2)}
          </pre>
        </div>
      )}

      {/* Tasks */}
      <div className="space-y-3">
        {openTasks.length > 0 && (
          <TaskSection title="Open Tasks">
            {openTasks.map((task) => {
              const Renderer = rendererFor(task.task_type)
              const outgoing: TransitionDef[] = def?.definition?.transitions?.filter(
                (t) => t.from_state === task.state_id
              ) ?? []
              return (
                <Renderer
                  key={task.id}
                  task={task}
                  instance={inst}
                  outgoingTransitions={outgoing}
                  onComplete={(status, nextState) => handleComplete(task, status, nextState)}
                  isLoading={completeMutation.isPending}
                />
              )
            })}
          </TaskSection>
        )}

        {doneTasks.length > 0 && (
          <CollapsibleSection title={`Completed (${doneTasks.length})`}>
            {doneTasks.map((task) => (
              <SimpleTaskRow key={task.id} task={task} />
            ))}
          </CollapsibleSection>
        )}

        {cancelledTasks.length > 0 && (
          <CollapsibleSection title={`Cancelled/Rejected (${cancelledTasks.length})`}>
            {cancelledTasks.map((task) => (
              <SimpleTaskRow key={task.id} task={task} />
            ))}
          </CollapsibleSection>
        )}

        {tasks.length === 0 && (
          <p className="text-sm text-zinc-400 py-4">No tasks for this instance.</p>
        )}
      </div>

      {/* History timeline */}
      {history.length > 0 && (
        <div className="mt-6 border border-zinc-200 dark:border-zinc-700 rounded-lg p-4">
          <h3 className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-3">History</h3>
          <div className="space-y-2">
            {history.map((tr) => (
              <div key={tr.id} className="flex items-center gap-2 text-sm">
                <div className="w-2 h-2 rounded-full bg-indigo-400 shrink-0" />
                <code className="text-xs font-mono text-zinc-600 dark:text-zinc-400">{tr.from_state}</code>
                <span className="text-zinc-400">→</span>
                <code className="text-xs font-mono text-zinc-600 dark:text-zinc-400">{tr.to_state}</code>
                <span className="text-xs text-zinc-400 ml-auto">{new Date(tr.transitioned_at).toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function TaskSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="text-xs font-medium text-zinc-500 uppercase tracking-wide mb-2">{title}</h3>
      <div className="space-y-2">{children}</div>
    </div>
  )
}

function CollapsibleSection({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  return (
    <div>
      <button
        onClick={() => setOpen((o) => !o)}
        className="text-xs font-medium text-zinc-500 uppercase tracking-wide mb-2 flex items-center gap-1 hover:text-zinc-700 dark:hover:text-zinc-300"
      >
        {open ? '▾' : '▸'} {title}
      </button>
      {open && <div className="space-y-2">{children}</div>}
    </div>
  )
}

function SimpleTaskRow({ task }: { task: TaskInstance }) {
  return (
    <div className="flex items-center justify-between py-1.5 px-2 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-sm">
      <span className="text-zinc-700 dark:text-zinc-300">{task.task_label}</span>
      <Badge variant={task.status}>{task.status}</Badge>
    </div>
  )
}
