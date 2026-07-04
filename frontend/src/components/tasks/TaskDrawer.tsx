import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { X, Save } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useInstance } from '../../hooks/useInstances'
import { useWorkflowDef } from '../../hooks/useWorkflowDefs'
import { tasksApi } from '../../api/instances'
import { tasksExtApi } from '../../api/tasks'
import { CommentsPanel } from './CommentsPanel'
import { AttachmentSection } from './AttachmentSection'
import { PriorityBadge } from './PriorityBadge'
import { Badge } from '../ui/Badge'
import { getCompletedBy } from '../../lib/session'
import type { TaskInstance, TransitionDef } from '../../types'

interface Props {
  task: TaskInstance
  onClose: () => void
}

const PRIORITIES = ['low', 'medium', 'high'] as const

export function TaskDrawer({ task, onClose }: Props) {
  const qc = useQueryClient()
  const { data: inst } = useInstance(task.workflow_instance_id)
  const { data: def } = useWorkflowDef(inst?.workflow_def_id ?? '')

  // edit fields local state
  const [priority, setPriority] = useState<'low' | 'medium' | 'high' | ''>(task.priority ?? '')
  const [dueDate, setDueDate] = useState(task.expected_completion_date ?? '')
  const [dueTime, setDueTime] = useState(task.expected_completion_time?.slice(0, 5) ?? '')

  // sync if task prop changes (after mutation refreshes parent)
  useEffect(() => {
    setPriority(task.priority ?? '')
    setDueDate(task.expected_completion_date ?? '')
    setDueTime(task.expected_completion_time?.slice(0, 5) ?? '')
  }, [task.id, task.priority, task.expected_completion_date, task.expected_completion_time])

  // next-state for user-picks mode
  const outgoing: TransitionDef[] = def?.definition?.transitions?.filter(
    (t) => t.from_state === task.state_id
  ) ?? []
  const isUserPicks = outgoing.length > 1 && outgoing.every((t) => !t.trigger)
  const [nextState, setNextState] = useState(() => outgoing[0]?.to_state ?? '')
  useEffect(() => {
    if (outgoing.length > 0 && !nextState) setNextState(outgoing[0].to_state)
  }, [outgoing.length]) // eslint-disable-line react-hooks/exhaustive-deps

  const completeMutation = useMutation({
    mutationFn: ({ status, ns }: { status: 'done' | 'rejected'; ns?: string }) =>
      tasksApi.complete(task.id, {
        status,
        next_state: ns,
        completed_by: getCompletedBy() || 'user',
      }),
    onSuccess: (data) => {
      qc.setQueryData(['instances', data.id], data)
      qc.invalidateQueries({ queryKey: ['instances'] })
    },
  })

  const fieldsMutation = useMutation({
    mutationFn: () =>
      tasksExtApi.update(task.id, {
        priority: priority || undefined,
        expected_completion_date: dueDate || null,
        expected_completion_time: dueTime || null,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['instances', task.workflow_instance_id] })
    },
  })

  const completable = task.status === 'new' || task.status === 'in_progress'
  const active = inst?.status === 'active'

  const fieldsDirty =
    (priority || '') !== (task.priority ?? '') ||
    (dueDate || '') !== (task.expected_completion_date ?? '') ||
    (dueTime || '') !== (task.expected_completion_time?.slice(0, 5) ?? '')

  // close on Escape
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [onClose])

  return (
    <>
      <div
        className="fixed inset-0 bg-black/30 dark:bg-black/50 z-40"
        onClick={onClose}
      />
      <motion.div
        className="fixed inset-y-0 right-0 z-50 flex flex-col w-[520px] max-w-[95vw] bg-white dark:bg-zinc-900 border-l border-zinc-200 dark:border-zinc-800 shadow-2xl"
        initial={{ x: '100%' }}
        animate={{ x: 0 }}
        exit={{ x: '100%' }}
        transition={{ type: 'spring', damping: 28, stiffness: 280 }}
      >
        {/* ── Header ── */}
        <div className="flex items-start gap-3 px-5 py-4 border-b border-zinc-200 dark:border-zinc-800 shrink-0">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <h2 className="text-base font-semibold text-zinc-900 dark:text-zinc-50 leading-tight">
                {task.task_label}
              </h2>
              {task.priority && <PriorityBadge priority={task.priority as 'low' | 'medium' | 'high'} />}
              {task.is_adhoc && (
                <span className="rounded bg-purple-100 text-purple-600 dark:bg-purple-900/60 dark:text-purple-300 px-2 py-0.5 text-xs font-medium">
                  Ad-hoc
                </span>
              )}
              <Badge variant={task.status}>{task.status.replace('_', ' ')}</Badge>
            </div>
            <div className="flex items-center gap-1.5 mt-1 text-xs text-zinc-400 flex-wrap">
              <code className="font-mono">{task.task_key}</code>
              <span>·</span>
              <span>{task.task_type}</span>
              <span>·</span>
              <code className="font-mono">{task.state_id}</code>
              {task.expected_completion_date && (
                <>
                  <span>·</span>
                  <span className="text-zinc-500">
                    Due {task.expected_completion_date}
                    {task.expected_completion_time ? ` at ${task.expected_completion_time.slice(0, 5)}` : ''}
                  </span>
                </>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200 shrink-0 transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        {/* ── Scrollable body ── */}
        <div className="flex-1 overflow-y-auto divide-y divide-zinc-100 dark:divide-zinc-800">

          {/* Complete actions */}
          {completable && (
            <section className="p-5 space-y-3">
              <p className="text-xs font-medium text-zinc-500 uppercase tracking-wide">Actions</p>
              {isUserPicks && (
                <div>
                  <label className="text-xs text-zinc-500 mb-1 block">Next step</label>
                  <select
                    value={nextState}
                    onChange={(e) => setNextState(e.target.value)}
                    className="w-full text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:ring-2 focus:ring-indigo-500 outline-none"
                  >
                    {outgoing.map((t) => (
                      <option key={t.to_state} value={t.to_state}>{t.to_state}</option>
                    ))}
                  </select>
                </div>
              )}
              <div className="flex gap-2">
                {task.task_type === 'approval' ? (
                  <>
                    <button
                      disabled={completeMutation.isPending}
                      onClick={() => completeMutation.mutate({ status: 'done', ns: isUserPicks ? nextState : undefined })}
                      className="flex-1 px-3 py-2 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                    >
                      {completeMutation.isPending ? 'Saving…' : 'Approve'}
                    </button>
                    <button
                      disabled={completeMutation.isPending}
                      onClick={() => completeMutation.mutate({ status: 'rejected' })}
                      className="flex-1 px-3 py-2 text-sm font-medium rounded-lg border border-red-300 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors"
                    >
                      Reject
                    </button>
                  </>
                ) : (
                  <>
                    <button
                      disabled={completeMutation.isPending}
                      onClick={() => completeMutation.mutate({ status: 'done', ns: isUserPicks ? nextState : undefined })}
                      className="flex-1 px-3 py-2 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                    >
                      {completeMutation.isPending ? 'Saving…' : 'Mark Done'}
                    </button>
                    <button
                      disabled={completeMutation.isPending}
                      onClick={() => completeMutation.mutate({ status: 'rejected' })}
                      className="flex-1 px-3 py-2 text-sm font-medium rounded-lg border border-zinc-300 dark:border-zinc-600 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors"
                    >
                      Reject
                    </button>
                  </>
                )}
              </div>
              {completeMutation.isSuccess && (
                <div className="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg text-sm text-green-700 dark:text-green-400">
                  Task completed.
                </div>
              )}
              {completeMutation.isError && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-400">
                  {(completeMutation.error as Error)?.message}
                </div>
              )}
            </section>
          )}

          {/* Edit fields */}
          {active && (
            <section className="p-5 space-y-4">
              <p className="text-xs font-medium text-zinc-500 uppercase tracking-wide">Fields</p>

              {/* Priority */}
              <div>
                <label className="text-xs text-zinc-500 block mb-1.5">Priority</label>
                <div className="flex gap-1.5">
                  {PRIORITIES.map((p) => (
                    <button
                      key={p}
                      onClick={() => setPriority(p)}
                      className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors capitalize ${
                        priority === p
                          ? p === 'high'
                            ? 'bg-red-100 border-red-300 text-red-600 dark:bg-red-900/40 dark:border-red-700 dark:text-red-400'
                            : p === 'medium'
                            ? 'bg-yellow-100 border-yellow-300 text-yellow-700 dark:bg-yellow-900/40 dark:border-yellow-700 dark:text-yellow-400'
                            : 'bg-zinc-100 border-zinc-300 text-zinc-600 dark:bg-zinc-800 dark:border-zinc-600 dark:text-zinc-300'
                          : 'border-zinc-200 dark:border-zinc-700 text-zinc-400 hover:border-zinc-300 dark:hover:border-zinc-600'
                      }`}
                    >
                      {p.charAt(0).toUpperCase() + p.slice(1)}
                    </button>
                  ))}
                  {priority && (
                    <button
                      onClick={() => setPriority('')}
                      className="px-3 py-1 rounded-full text-xs border border-zinc-200 dark:border-zinc-700 text-zinc-400 hover:text-zinc-600 transition-colors"
                    >
                      Clear
                    </button>
                  )}
                </div>
              </div>

              {/* Due date / time */}
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="text-xs text-zinc-500 block mb-1.5">Due Date</label>
                  <input
                    type="date"
                    value={dueDate}
                    onChange={(e) => setDueDate(e.target.value)}
                    className="w-full text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:ring-2 focus:ring-indigo-500 outline-none"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-xs text-zinc-500 block mb-1.5">Due Time</label>
                  <input
                    type="time"
                    value={dueTime}
                    onChange={(e) => setDueTime(e.target.value)}
                    className="w-full text-sm border border-zinc-300 dark:border-zinc-600 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:ring-2 focus:ring-indigo-500 outline-none"
                  />
                </div>
              </div>

              {fieldsDirty && (
                <div className="flex items-center gap-2">
                  <button
                    disabled={fieldsMutation.isPending}
                    onClick={() => fieldsMutation.mutate()}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                  >
                    <Save size={12} />
                    {fieldsMutation.isPending ? 'Saving…' : 'Save Changes'}
                  </button>
                  {fieldsMutation.isError && (
                    <span className="text-xs text-red-500">Failed to save.</span>
                  )}
                  {fieldsMutation.isSuccess && (
                    <span className="text-xs text-green-600 dark:text-green-400">Saved.</span>
                  )}
                </div>
              )}
            </section>
          )}

          {/* Comments */}
          <section className="p-5">
            <CommentsPanel taskId={task.id} />
          </section>

          {/* Attachment */}
          <section className="p-5">
            <AttachmentSection taskId={task.id} hasMeta={task.has_attachment ?? false} />
          </section>
        </div>
      </motion.div>
    </>
  )
}
