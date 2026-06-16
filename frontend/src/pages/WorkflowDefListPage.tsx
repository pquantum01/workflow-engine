import { useState } from 'react'
import { Link } from 'react-router-dom'
import { GitBranch, Search } from 'lucide-react'
import { useWorkflowDefs } from '../hooks/useWorkflowDefs'
import { Badge } from '../components/ui/Badge'
import { Skeleton } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'

export function WorkflowDefListPage() {
  const [search, setSearch] = useState('')
  const [includeDeleted, setIncludeDeleted] = useState(false)
  const { data, isLoading, error } = useWorkflowDefs(includeDeleted)

  const defs = (data?.items ?? []).filter(
    (d) =>
      !search ||
      d.name.toLowerCase().includes(search.toLowerCase()) ||
      d.description?.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Workflow Definitions</h1>
        <label className="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-400">
          <input
            type="checkbox"
            checked={includeDeleted}
            onChange={(e) => setIncludeDeleted(e.target.checked)}
            className="rounded border-zinc-300"
          />
          Show deleted
        </label>
      </div>

      <div className="relative mb-4">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by name..."
          className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
        />
      </div>

      {isLoading && (
        <div className="space-y-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      )}

      {error && (
        <div className="text-red-600 text-sm">Failed to load workflow definitions.</div>
      )}

      {!isLoading && !error && defs.length === 0 && (
        <EmptyState
          icon={GitBranch}
          title="No workflow definitions"
          description="Add JSON files to the workflows/ directory and reload."
        />
      )}

      {!isLoading && defs.length > 0 && (
        <div className="overflow-hidden border border-zinc-200 dark:border-zinc-700 rounded-lg">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-zinc-50 dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-700">
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Name</th>
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Version</th>
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Description</th>
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">States</th>
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Status</th>
                <th className="text-left px-4 py-3 font-medium text-xs uppercase text-zinc-500 tracking-wide">Created</th>
              </tr>
            </thead>
            <tbody>
              {defs.map((def) => (
                <tr
                  key={def.id}
                  className="border-b border-zinc-100 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 cursor-pointer transition-colors"
                >
                  <td className="px-4 py-3">
                    <Link
                      to={`/workflow-defs/${def.id}`}
                      className="font-medium text-indigo-600 hover:text-indigo-700 font-mono text-sm"
                    >
                      {def.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-zinc-500">v{def.version}</td>
                  <td className="px-4 py-3 text-zinc-600 dark:text-zinc-400 max-w-xs truncate">
                    {def.description || '—'}
                  </td>
                  <td className="px-4 py-3 text-zinc-500">
                    {def.definition?.states?.length ?? '—'}
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={def.deleted_at ? 'deleted' : 'active'}>
                      {def.deleted_at ? 'Deleted' : 'Active'}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-zinc-500 text-xs">
                    {new Date(def.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
