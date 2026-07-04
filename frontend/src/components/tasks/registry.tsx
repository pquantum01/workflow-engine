import React from 'react'
import type { WorkflowInstance, TaskInstance, TransitionDef } from '../../types'

export interface TaskRendererProps {
  task: TaskInstance
  instance: WorkflowInstance
  outgoingTransitions: TransitionDef[]
  onComplete: (status: 'done' | 'rejected', nextState?: string) => void
  isLoading: boolean
  onDetails?: () => void
}

export type TaskRenderer = (props: TaskRendererProps) => React.ReactElement

function ApprovalTaskRenderer({ task, outgoingTransitions, onComplete, isLoading, onDetails }: TaskRendererProps) {
  const isUserPicks = outgoingTransitions.length > 1 && outgoingTransitions.every((t) => !t.trigger)
  const completable = task.status === 'new' || task.status === 'in_progress'

  return (
    <TaskCard task={task} onDetails={onDetails}>
      {completable && (
        <div className="space-y-2">
          {isUserPicks && (
            <div>
              <label className="text-xs text-zinc-500 mb-1 block">Next step</label>
              <select
                id={`next-state-${task.id}`}
                className="w-full text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
              >
                {outgoingTransitions.map((t) => (
                  <option key={t.to_state} value={t.to_state}>{t.to_state}</option>
                ))}
              </select>
            </div>
          )}
          <div className="flex gap-2">
            <button
              disabled={isLoading}
              onClick={() => {
                if (isUserPicks) {
                  const sel = document.getElementById(`next-state-${task.id}`) as HTMLSelectElement
                  onComplete('done', sel?.value)
                } else {
                  onComplete('done')
                }
              }}
              className="flex-1 px-3 py-1.5 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            >
              Approve
            </button>
            <button
              disabled={isLoading}
              onClick={() => onComplete('rejected')}
              className="flex-1 px-3 py-1.5 text-sm font-medium rounded-lg border border-red-300 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors"
            >
              Reject
            </button>
          </div>
        </div>
      )}
    </TaskCard>
  )
}

function ManualTaskRenderer({ task, outgoingTransitions, onComplete, isLoading, onDetails }: TaskRendererProps) {
  const isUserPicks = outgoingTransitions.length > 1 && outgoingTransitions.every((t) => !t.trigger)
  const completable = task.status === 'new' || task.status === 'in_progress'

  return (
    <TaskCard task={task} onDetails={onDetails}>
      {completable && (
        <div className="space-y-2">
          {isUserPicks && (
            <div>
              <label className="text-xs text-zinc-500 mb-1 block">Next step</label>
              <select
                id={`next-state-${task.id}`}
                className="w-full text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
              >
                {outgoingTransitions.map((t) => (
                  <option key={t.to_state} value={t.to_state}>{t.to_state}</option>
                ))}
              </select>
            </div>
          )}
          <div className="flex gap-2">
            <button
              disabled={isLoading}
              onClick={() => {
                if (isUserPicks) {
                  const sel = document.getElementById(`next-state-${task.id}`) as HTMLSelectElement
                  onComplete('done', sel?.value)
                } else {
                  onComplete('done')
                }
              }}
              className="flex-1 px-3 py-1.5 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            >
              Mark Done
            </button>
            <button
              disabled={isLoading}
              onClick={() => onComplete('rejected')}
              className="flex-1 px-3 py-1.5 text-sm font-medium rounded-lg border border-zinc-300 dark:border-zinc-600 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors"
            >
              Reject
            </button>
          </div>
        </div>
      )}
    </TaskCard>
  )
}

function FormTaskRenderer(props: TaskRendererProps) {
  // v1 stub: same as manual
  return <ManualTaskRenderer {...props} />
}

function ChecklistTaskRenderer(props: TaskRendererProps) {
  // v1 stub: same as manual
  return <ManualTaskRenderer {...props} />
}

const registry: Record<string, TaskRenderer> = {
  approval: ApprovalTaskRenderer,
  form: FormTaskRenderer,
  checklist: ChecklistTaskRenderer,
  manual: ManualTaskRenderer,
}

export function rendererFor(type: string): TaskRenderer {
  return registry[type] ?? ManualTaskRenderer
}

// Shared task card wrapper
import { Badge } from '../ui/Badge'
import { PriorityBadge } from './PriorityBadge'
import { ChevronRight } from 'lucide-react'

function TaskCard({
  task,
  children,
  onDetails,
}: {
  task: TaskInstance
  children?: React.ReactNode
  onDetails?: () => void
}) {
  return (
    <div className="border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800/50 overflow-hidden">
      {/* Header row — clickable to open details */}
      <div
        className={`flex items-start gap-2 px-3 py-2.5 ${onDetails ? 'cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors' : ''}`}
        onClick={onDetails}
      >
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium text-zinc-800 dark:text-zinc-100">{task.task_label}</span>
            {task.priority && <PriorityBadge priority={task.priority as 'low' | 'medium' | 'high'} />}
            {task.is_adhoc && (
              <span className="rounded bg-purple-100 text-purple-600 dark:bg-purple-900/60 dark:text-purple-300 px-1.5 py-0.5 text-xs font-medium">
                Ad-hoc
              </span>
            )}
          </div>
          <div className="text-xs text-zinc-500 mt-0.5">{task.task_key} · {task.task_type}</div>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <Badge variant={task.status}>{task.status.replace('_', ' ')}</Badge>
          {onDetails && <ChevronRight size={14} className="text-zinc-400" />}
        </div>
      </div>
      {/* Action body — separate from header to avoid click conflict */}
      {children && (
        <div className="px-3 pb-3 border-t border-zinc-100 dark:border-zinc-700/60 pt-2.5 space-y-2">
          {children}
        </div>
      )}
    </div>
  )
}
