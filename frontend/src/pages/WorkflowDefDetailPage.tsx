import { useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Download } from 'lucide-react'
import { useWorkflowDef } from '../hooks/useWorkflowDefs'
import { WorkflowGraph } from '../components/graph/WorkflowGraph'
import { Badge } from '../components/ui/Badge'
import { Skeleton } from '../components/ui/Skeleton'
import type { StateDef } from '../types'

export function WorkflowDefDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: def, isLoading, error } = useWorkflowDef(id!)
  const [selectedState, setSelectedState] = useState<string | null>(null)

  const selected = def?.definition?.states?.find((s) => s.id === selectedState)

  if (isLoading) {
    return (
      <div className="p-6 space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    )
  }

  if (error || !def) {
    return (
      <div className="p-6">
        <div className="text-red-600">Workflow definition not found.</div>
      </div>
    )
  }

  const handleDownload = () => {
    const blob = new Blob([JSON.stringify(def.definition, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${def.name}-v${def.version}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="p-6">
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => navigate('/workflow-defs')}
          className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500"
        >
          <ArrowLeft size={16} />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight font-mono">{def.name}</h1>
            <span className="text-sm text-zinc-500 font-normal">v{def.version}</span>
            <Badge variant={def.deleted_at ? 'deleted' : 'active'}>
              {def.deleted_at ? 'Deleted' : 'Active'}
            </Badge>
          </div>
          {def.description && (
            <p className="text-sm text-zinc-500 mt-0.5">{def.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Link
            to={`/workflow-instances/new?def_id=${def.id}`}
            className="px-3 py-1.5 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 transition-colors"
          >
            Start Instance
          </Link>
          <button
            onClick={handleDownload}
            className="p-1.5 rounded-lg border border-zinc-300 dark:border-zinc-600 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors"
          >
            <Download size={16} />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4" style={{ height: 'calc(100vh - 220px)' }}>
        {/* Graph */}
        <div className="col-span-2 border border-zinc-200 dark:border-zinc-700 rounded-lg overflow-hidden">
          {def.definition ? (
            <WorkflowGraph
              definition={def.definition}
              onNodeClick={setSelectedState}
            />
          ) : (
            <div className="flex items-center justify-center h-full text-zinc-400">
              No graph data
            </div>
          )}
        </div>

        {/* Inspector */}
        <div className="border border-zinc-200 dark:border-zinc-700 rounded-lg overflow-auto">
          {selected ? (
            <StateInspector state={selected} def={def.definition} />
          ) : (
            <div className="flex items-center justify-center h-full text-sm text-zinc-400">
              Click a state to inspect
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function StateInspector({ state, def }: { state: StateDef; def: any }) {
  const outgoing = (def?.transitions ?? []).filter((t: any) => t.from_state === state.id)
  const isStart = def?.start_state === state.id
  const isEnd = def?.end_state === state.id

  return (
    <div className="p-4 space-y-4">
      <div>
        <div className="flex items-center gap-2 mb-1">
          <span className="text-base font-medium">{state.label}</span>
          {isStart && <Badge variant="active">START</Badge>}
          {isEnd && <Badge variant="completed">END</Badge>}
        </div>
        <code className="text-xs text-zinc-500 font-mono">{state.id}</code>
      </div>

      {state.tasks.length > 0 && (
        <div>
          <h3 className="text-xs font-medium text-zinc-500 uppercase tracking-wide mb-2">
            Tasks ({state.tasks.length})
          </h3>
          <div className="space-y-1.5">
            {state.tasks.map((t) => (
              <div
                key={t.key}
                className="border border-zinc-200 dark:border-zinc-700 rounded-lg p-2.5"
              >
                <div className="text-sm font-medium">{t.label}</div>
                <div className="text-xs text-zinc-500 mt-0.5">
                  <code className="font-mono">{t.key}</code> · {t.type}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {outgoing.length > 0 && (
        <div>
          <h3 className="text-xs font-medium text-zinc-500 uppercase tracking-wide mb-2">
            Transitions
          </h3>
          <div className="space-y-1">
            {outgoing.map((t: any, i: number) => (
              <div key={i} className="flex items-center gap-1.5 text-sm">
                <span className="text-zinc-400">→</span>
                <code className="font-mono text-xs text-indigo-600">{t.to_state}</code>
                {t.trigger && (
                  <span className="text-xs text-zinc-400">via {t.trigger.task_key}</span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
