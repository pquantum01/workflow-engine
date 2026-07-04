import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksExtApi } from '../../api/tasks'
import type { StateDef } from '../../types'

interface Props {
  instanceId: string
  states: StateDef[]
  onClose: () => void
}

const TASK_TYPES = ['manual', 'approval', 'form', 'checklist']
const PRIORITIES = ['low', 'medium', 'high'] as const

export function AdhocTaskModal({ instanceId, states, onClose }: Props) {
  const qc = useQueryClient()
  const [stateId, setStateId] = useState(states[0]?.id ?? '')
  const [label, setLabel] = useState('')
  const [taskType, setTaskType] = useState('manual')
  const [priority, setPriority] = useState<'low' | 'medium' | 'high'>('medium')
  const [createdBy, setCreatedBy] = useState('')
  const [expDate, setExpDate] = useState('')
  const [expTime, setExpTime] = useState('')

  const mutation = useMutation({
    mutationFn: () =>
      tasksExtApi.createAdhoc(instanceId, {
        state_id: stateId,
        label,
        task_type: taskType,
        created_by: createdBy,
        priority,
        expected_completion_date: expDate || undefined,
        expected_completion_time: expTime || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['instance-tasks', instanceId] })
      onClose()
    },
  })

  const canSubmit = label.trim().length > 0 && stateId.length > 0

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md p-6 space-y-4">
        <h2 className="text-base font-semibold">Add Ad-hoc Task</h2>

        <div className="space-y-3">
          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            State
            <select
              value={stateId}
              onChange={(e) => setStateId(e.target.value)}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {states.map((s) => (
                <option key={s.id} value={s.id}>{s.label}</option>
              ))}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Label *
            <input
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="Task label"
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            />
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Task Type
            <select
              value={taskType}
              onChange={(e) => setTaskType(e.target.value)}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {TASK_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Priority
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as 'low' | 'medium' | 'high')}
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            >
              {PRIORITIES.map((p) => <option key={p} value={p}>{p}</option>)}
            </select>
          </label>

          <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
            Created By
            <input
              value={createdBy}
              onChange={(e) => setCreatedBy(e.target.value)}
              placeholder="Your name"
              className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
            />
          </label>

          <div className="grid grid-cols-2 gap-2">
            <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
              Due Date
              <input
                type="date"
                value={expDate}
                onChange={(e) => setExpDate(e.target.value)}
                className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
              />
            </label>
            <label className="block text-xs font-medium text-gray-600 dark:text-gray-400">
              Due Time
              <input
                type="time"
                value={expTime}
                onChange={(e) => setExpTime(e.target.value)}
                className="mt-1 block w-full rounded border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-800"
              />
            </label>
          </div>
        </div>

        {mutation.isError && (
          <p className="text-xs text-red-500">Failed to create task.</p>
        )}

        <div className="flex justify-end gap-2 pt-2">
          <button
            onClick={onClose}
            className="rounded border border-gray-300 px-3 py-1 text-xs dark:border-gray-600"
          >
            Cancel
          </button>
          <button
            disabled={!canSubmit || mutation.isPending}
            onClick={() => mutation.mutate()}
            className="rounded bg-blue-600 px-3 py-1 text-xs text-white disabled:opacity-50"
          >
            {mutation.isPending ? 'Adding…' : 'Add Task'}
          </button>
        </div>
      </div>
    </div>
  )
}
