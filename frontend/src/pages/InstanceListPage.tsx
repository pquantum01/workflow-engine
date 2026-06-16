import { useState } from 'react'
import { Link } from 'react-router-dom'
import { List, Plus } from 'lucide-react'
import { useInstances } from '../hooks/useInstances'
import { useWorkflowDefs } from '../hooks/useWorkflowDefs'
import { Badge } from '../components/ui/Badge'
import { Skeleton } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'

export function InstanceListPage() {
  const [status, setStatus] = useState('')
  const [defId, setDefId] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  const { data, isLoading, error } = useInstances({ status: status || undefined, def_id: defId || undefined, limit, offset: page * limit })
  const { data: defsData } = useWorkflowDefs()

  const defMap = Object.fromEntries((defsData?.items ?? []).map((d) => [d.id, d]))

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Instances</h1>
        <Link
          to="/workflow-instances/new"
          className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 transition-colors"
        >
          <Plus size={14} />
          Start Instance
        </Link>
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-4">
        <select
          value={status}
          onChange={(e) => { setStatus(e.target.value); setPage(0) }}
          className="text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
        >
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select
          value={defId}
          onChange={(e) => { setDefId(e.target.value); setPage(0) }}
          className="text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-1.5 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
        >
          <option value="">All workflows</option>
          {(defsData?.items ?? []).map((d) => (
            <option key={d.id} value={d.id}>{d.name} v{d.version}</option>
          ))}
        </select>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}
        </div>
      )}

      {!isLoading && !error && (data?.items ?? []).length === 0 && (
        <EmptyState
          icon={List}
          title="No instances found"
          description="Start a workflow instance to see it here."
          action={
            <Link to="/workflow-instances/new" className="px-3 py-1.5 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 transition-colors">
              Start Instance
            </Link>
          }
        />
      )}

      {!isLoading && (data?.items ?? []).length > 0 && (
        <>
          <div className="overflow-hidden border border-zinc-200 dark:border-zinc-700 rounded-lg">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-zinc-50 dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-700">
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">ID</th>
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Workflow</th>
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Status</th>
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Current State</th>
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Created By</th>
                  <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Started</th>
                </tr>
              </thead>
              <tbody>
                {(data?.items ?? []).map((inst) => (
                  <tr
                    key={inst.id}
                    className="border-b border-zinc-100 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 cursor-pointer"
                  >
                    <td className="px-4 py-3">
                      <Link to={`/workflow-instances/${inst.id}`} className="font-mono text-xs text-indigo-600 hover:text-indigo-700">
                        {inst.id.slice(0, 8)}…
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-zinc-600 dark:text-zinc-400">
                      {defMap[inst.workflow_def_id]
                        ? `${defMap[inst.workflow_def_id].name} v${defMap[inst.workflow_def_id].version}`
                        : inst.workflow_def_id.slice(0, 8)}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={inst.status}>{inst.status}</Badge>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-zinc-600 dark:text-zinc-400">
                      {inst.current_state_id}
                    </td>
                    <td className="px-4 py-3 text-xs text-zinc-600 dark:text-zinc-400">
                      {inst.created_by}
                    </td>
                    <td className="px-4 py-3 text-xs text-zinc-500">
                      {new Date(inst.started_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {(data?.total ?? 0) > limit && (
            <div className="flex items-center justify-between mt-4 text-sm text-zinc-500">
              <span>{data?.total} total</span>
              <div className="flex gap-2">
                <button
                  disabled={page === 0}
                  onClick={() => setPage((p) => p - 1)}
                  className="px-3 py-1 rounded border border-zinc-300 disabled:opacity-40 hover:bg-zinc-50"
                >
                  Previous
                </button>
                <button
                  disabled={(page + 1) * limit >= (data?.total ?? 0)}
                  onClick={() => setPage((p) => p + 1)}
                  className="px-3 py-1 rounded border border-zinc-300 disabled:opacity-40 hover:bg-zinc-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
