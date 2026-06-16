import { useReloadDefs } from '../hooks/useWorkflowDefs'
import { RefreshCw, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'

export function AdminReloadPage() {
  const { mutate, isPending, data, error } = useReloadDefs()

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-2xl font-semibold tracking-tight mb-6">Admin</h1>

      <div className="border border-zinc-200 dark:border-zinc-700 rounded-lg p-5">
        <h2 className="text-base font-medium mb-2">Reload Workflow Definitions</h2>
        <p className="text-sm text-zinc-500 mb-4">
          Rescans the <code className="font-mono bg-zinc-100 dark:bg-zinc-800 px-1 py-0.5 rounded text-xs">workflows/</code> directory and upserts any new or unchanged definitions.
        </p>
        <button
          onClick={() => mutate()}
          disabled={isPending}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
        >
          <RefreshCw size={14} className={isPending ? 'animate-spin' : ''} />
          {isPending ? 'Reloading…' : 'Reload Definitions'}
        </button>

        {error && (
          <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700">
            Reload failed: {(error as Error).message}
          </div>
        )}

        {data && (
          <div className="mt-4 space-y-3">
            <ResultSection
              icon={CheckCircle}
              color="text-green-600"
              title={`Loaded (${data.loaded.length})`}
              items={data.loaded.map((l: any) => `${l.name} v${l.version}`)}
              empty="Nothing new to load"
            />
            {data.warnings.length > 0 && (
              <ResultSection
                icon={AlertTriangle}
                color="text-amber-500"
                title={`Warnings (${data.warnings.length})`}
                items={(data.warnings as any[]).map((w) => `${w.file}: ${w.message}`)}
              />
            )}
            {data.errors.length > 0 && (
              <ResultSection
                icon={XCircle}
                color="text-red-500"
                title={`Errors (${data.errors.length})`}
                items={(data.errors as any[]).map((e) => `${e.file}: ${e.message}`)}
              />
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function ResultSection({ icon: Icon, color, title, items, empty }: {
  icon: any; color: string; title: string; items: string[]; empty?: string
}) {
  return (
    <div>
      <div className={`flex items-center gap-1.5 text-sm font-medium ${color} mb-1`}>
        <Icon size={14} />
        {title}
      </div>
      {items.length === 0 ? (
        <p className="text-sm text-zinc-400 pl-5">{empty}</p>
      ) : (
        <ul className="pl-5 space-y-0.5">
          {items.map((item, i) => (
            <li key={i} className="text-sm text-zinc-600 dark:text-zinc-400">{item}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
